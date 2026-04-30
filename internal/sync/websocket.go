package sync

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Timeouts
	HeartbeatInterval = 20 * time.Second
	PingTimeout       = 10 * time.Second
	MaxIdleTime       = 2 * time.Minute
	ReadTimeout       = 30 * time.Second

	// File size limits
	ChunkSize      = 2 * 1024 * 1024   // 2MB chunks
	DefaultMaxFile = 199 * 1024 * 1024 // 199MB default
	MaxFileSize    = 500 * 1024 * 1024 // 500MB absolute max
)

// WebSocketClient handles WebSocket communication with Obsidian sync server
type WebSocketClient struct {
	conn              *websocket.Conn
	mu                sync.Mutex
	encryptionVersion int
	keyHash           string
	perFileMax        int64
	userID            int
	lastMessageTime   time.Time
	heartbeatTicker   *time.Ticker
	done              chan struct{}
	onDisconnect      func()
}

// Message represents a WebSocket message
type Message struct {
	Op                string `json:"op"`
	Status            string `json:"status,omitempty"`
	Res               string `json:"res,omitempty"`
	Msg               string `json:"msg,omitempty"`
	Token             string `json:"token,omitempty"`
	ID                string `json:"id,omitempty"`
	KeyHash           string `json:"keyhash,omitempty"`
	Version           int    `json:"version,omitempty"`
	Initial           bool   `json:"initial,omitempty"`
	Device            string `json:"device,omitempty"`
	EncryptionVersion int    `json:"encryption_version,omitempty"`
	PerFileMax        int64  `json:"perFileMax,omitempty"`
	UserID            int    `json:"userId,omitempty"`

	// File operations
	Path        string `json:"path,omitempty"`
	RelatedPath string `json:"relatedpath,omitempty"`
	Extension   string `json:"extension,omitempty"`
	Hash        string `json:"hash,omitempty"`
	CTime       int64  `json:"ctime,omitempty"`
	MTime       int64  `json:"mtime,omitempty"`
	Folder      bool   `json:"folder,omitempty"`
	Deleted     bool   `json:"deleted,omitempty"`
	Size        int64  `json:"size,omitempty"`
	Pieces      int    `json:"pieces,omitempty"`
	UID         int64  `json:"uid,omitempty"`

	// List operations
	Items []FileInfo `json:"items,omitempty"`
}

// FileInfo represents file metadata
type FileInfo struct {
	UID     int64  `json:"uid"`
	Path    string `json:"path"`
	Hash    string `json:"hash"`
	CTime   int64  `json:"ctime"`
	MTime   int64  `json:"mtime"`
	Size    int64  `json:"size"`
	Folder  bool   `json:"folder"`
	Deleted bool   `json:"deleted"`
	Device  string `json:"device"`
}

// Connect establishes a WebSocket connection to the sync server
func (c *WebSocketClient) Connect(host, token, vaultID, keyHash, device string, encVersion int, initial bool) error {
	// Build WebSocket URL
	u, err := url.Parse(host)
	if err != nil {
		return fmt.Errorf("parse host: %w", err)
	}

	// Convert http(s) to ws(s)
	scheme := "wss"
	if u.Scheme == "http" {
		scheme = "ws"
	}

	wsURL := fmt.Sprintf("%s://%s", scheme, u.Host)

	// Connect
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	c.conn = conn
	c.keyHash = keyHash
	c.encryptionVersion = encVersion
	c.perFileMax = DefaultMaxFile
	c.done = make(chan struct{})

	// Send init message
	initMsg := Message{
		Op:                "init",
		Token:             token,
		ID:                vaultID,
		KeyHash:           keyHash,
		Version:           3,
		Initial:           initial,
		Device:            device,
		EncryptionVersion: encVersion,
	}

	if err := c.sendJSON(initMsg); err != nil {
		return fmt.Errorf("send init: %w", err)
	}

	// Wait for response
	var resp Message
	if err := conn.ReadJSON(&resp); err != nil {
		return fmt.Errorf("read init response: %w", err)
	}

	if resp.Status == "err" || resp.Res == "err" {
		return fmt.Errorf("authentication failed: %s", resp.Msg)
	}

	if resp.Res != "ok" {
		return fmt.Errorf("unexpected response: %+v", resp)
	}

	if resp.PerFileMax > 0 {
		c.perFileMax = resp.PerFileMax
	}

	if resp.UserID > 0 {
		c.userID = resp.UserID
	}

	// Start heartbeat
	c.startHeartbeat()

	return nil
}

// sendJSON sends a JSON message
func (c *WebSocketClient) sendJSON(msg Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.WriteJSON(msg)
}

// sendBinary sends binary data
func (c *WebSocketClient) sendBinary(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

// readJSON reads a JSON message
func (c *WebSocketClient) readJSON() (*Message, error) {
	c.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	var msg Message
	err := c.conn.ReadJSON(&msg)
	if err == nil {
		c.lastMessageTime = time.Now()
	}
	return &msg, err
}

// readBinary reads binary data
func (c *WebSocketClient) readBinary() ([]byte, error) {
	c.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	msgType, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if msgType != websocket.BinaryMessage {
		return nil, fmt.Errorf("expected binary message, got %d", msgType)
	}
	c.lastMessageTime = time.Now()
	return data, nil
}

// startHeartbeat starts the heartbeat goroutine
func (c *WebSocketClient) startHeartbeat() {
	c.lastMessageTime = time.Now()
	c.heartbeatTicker = time.NewTicker(HeartbeatInterval)

	go func() {
		for {
			select {
			case <-c.heartbeatTicker.C:
				idle := time.Since(c.lastMessageTime)
				if idle > MaxIdleTime {
					c.Disconnect()
					return
				}
				if idle > PingTimeout {
					c.sendJSON(Message{Op: "ping"})
				}
			case <-c.done:
				return
			}
		}
	}()
}

// Disconnect closes the WebSocket connection
func (c *WebSocketClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
		c.heartbeatTicker = nil
	}

	// Only close channel if not already closed
	select {
	case <-c.done:
		// Already closed
	default:
		close(c.done)
	}

	if c.onDisconnect != nil {
		c.onDisconnect()
	}
}

// Pull downloads a file from the server
func (c *WebSocketClient) Pull(uid int64) ([]byte, error) {
	// Request file
	if err := c.sendJSON(Message{Op: "pull", UID: uid}); err != nil {
		return nil, err
	}

	// Read response
	resp, err := c.readJSON()
	if err != nil {
		return nil, err
	}

	if resp.Deleted {
		return nil, nil
	}

	// Read file data in chunks
	totalSize := resp.Size
	if totalSize > MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", totalSize, MaxFileSize)
	}
	if totalSize < 0 {
		return nil, fmt.Errorf("invalid file size: %d", totalSize)
	}

	numPieces := resp.Pieces
	data := make([]byte, totalSize)
	offset := 0

	for i := 0; i < numPieces; i++ {
		chunk, err := c.readBinary()
		if err != nil {
			return nil, fmt.Errorf("read chunk %d: %w", i, err)
		}
		copy(data[offset:], chunk)
		offset += len(chunk)
	}

	return data, nil
}

// Push uploads a file to the server
func (c *WebSocketClient) Push(path, relatedPath, extension, hash string, ctime, mtime int64, folder, deleted bool, data []byte) error {
	// For folders and deletions
	if folder || deleted {
		return c.sendJSON(Message{
			Op:          "push",
			Path:        path,
			RelatedPath: relatedPath,
			Extension:   extension,
			Hash:        "",
			CTime:       0,
			MTime:       0,
			Folder:      folder,
			Deleted:     deleted,
		})
	}

	// For files
	size := int64(len(data))
	numPieces := int((size + ChunkSize - 1) / ChunkSize)

	// Send push request
	if err := c.sendJSON(Message{
		Op:          "push",
		Path:        path,
		RelatedPath: relatedPath,
		Extension:   extension,
		Hash:        hash,
		CTime:       ctime,
		MTime:       mtime,
		Folder:      folder,
		Deleted:     deleted,
		Size:        size,
		Pieces:      numPieces,
	}); err != nil {
		return err
	}

	// Read response
	resp, err := c.readJSON()
	if err != nil {
		return err
	}

	if resp.Res == "ok" {
		// File already exists with same hash
		return nil
	}

	// Send file data in chunks
	for i := 0; i < numPieces; i++ {
		start := i * ChunkSize
		end := start + ChunkSize
		if end > len(data) {
			end = len(data)
		}

		if err := c.sendBinary(data[start:end]); err != nil {
			return fmt.Errorf("send chunk %d: %w", i, err)
		}

		// Wait for ack
		if _, err := c.readJSON(); err != nil {
			return fmt.Errorf("read chunk ack %d: %w", i, err)
		}
	}

	return nil
}

// ListDeleted lists deleted files
func (c *WebSocketClient) ListDeleted() ([]FileInfo, error) {
	if err := c.sendJSON(Message{Op: "deleted"}); err != nil {
		return nil, err
	}

	resp, err := c.readJSON()
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}
