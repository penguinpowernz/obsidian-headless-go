# Security & Bug Audit Report

## Critical Issues

### 🔴 CRITICAL: Weak PBKDF2 Iterations
**File:** `internal/crypto/crypto.go:29`
```go
key := pbkdf2.Key([]byte(password), saltBytes, 10000, KeySize, sha256.New)
```

**Issue:** Only 10,000 iterations is considered weak by modern standards. NIST recommends at least 100,000 for PBKDF2-SHA256.

**BUT:** This matches the JavaScript implementation which uses scrypt (N=32768, r=8, p=1) for password→key derivation, not PBKDF2. The DeriveKey function may not be the right implementation.

**Action Required:** Verify if this is the correct algorithm from the JS code. The JS uses scrypt with specific parameters.

### 🔴 CRITICAL: Incorrect Password Derivation
**File:** `internal/crypto/crypto.go:22-30`

The JavaScript code uses **scrypt**, not PBKDF2:
```javascript
// JavaScript (line 11237)
async function vr(s, e) {
  // Uses scrypt with N=32768, r=8, p=1
  return await window.scrypt.scrypt(...)
}
```

**Fix Required:** Replace PBKDF2 with scrypt using `golang.org/x/crypto/scrypt`:
```go
import "golang.org/x/crypto/scrypt"

func DeriveKey(password, salt string) ([]byte, error) {
    saltBytes, err := hex.DecodeString(salt)
    if err != nil {
        return nil, fmt.Errorf("decode salt: %w", err)
    }
    
    // Match JavaScript: N=32768, r=8, p=1
    key, err := scrypt.Key([]byte(password), saltBytes, 32768, 8, 1, 32)
    if err != nil {
        return nil, fmt.Errorf("derive key: %w", err)
    }
    return key, nil
}
```

## High Severity Issues

### 🟠 Missing Error Check
**File:** `internal/api/client.go:280`
```go
optReq, _ := http.NewRequest("OPTIONS", url, nil)
c.HTTPClient.Do(optReq)
```

**Issue:** Ignoring errors from both NewRequest and Do. The preflight request errors are silently discarded.

**Fix:**
```go
optReq, err := http.NewRequest("OPTIONS", url, nil)
if err == nil {
    _, _ = c.HTTPClient.Do(optReq) // Preflight can fail, that's ok
}
```

### 🟠 Potential Race Condition
**File:** `internal/sync/websocket.go:253-266`

The `done` channel is closed in `Disconnect()` but accessed in `startHeartbeat()`. If Disconnect is called multiple times, this will panic.

**Fix:**
```go
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

    // Only close if not already closed
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
```

### 🟠 Missing Context/Cancellation
**File:** `internal/sync/websocket.go:287-306`

The Pull() and Push() methods have no way to be cancelled. They could hang indefinitely.

**Fix:** Add context.Context parameter:
```go
func (c *WebSocketClient) Pull(ctx context.Context, uid int64) ([]byte, error)
func (c *WebSocketClient) Push(ctx context.Context, ...) error
```

## Medium Severity Issues

### 🟡 No TLS Certificate Validation Options
**File:** `internal/sync/websocket.go:106`
```go
conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
```

**Issue:** No way to configure TLS settings. In production, you may want to pin certificates or use custom CAs.

**Recommendation:** Create a custom dialer:
```go
dialer := &websocket.Dialer{
    TLSClientConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
        // Add any custom CA certificates here
    },
}
conn, _, err := dialer.Dial(wsURL, nil)
```

### 🟡 Memory Leak Risk in WebSocket
**File:** `internal/sync/websocket.go:243-250`

The heartbeat goroutine may not exit cleanly if Disconnect() isn't called.

**Fix:** Already partially handled with `done` channel, but add defer in Connect:
```go
// In Connect():
// Add cleanup on error
defer func() {
    if err != nil {
        c.Disconnect()
    }
}()
```

### 🟡 No Rate Limiting
**File:** `internal/api/client.go:62-66`

No rate limiting on API calls. Could trigger rate limiting from Obsidian's servers.

**Recommendation:** Add rate limiting using `golang.org/x/time/rate`:
```go
import "golang.org/x/time/rate"

type Client struct {
    // ...
    limiter *rate.Limiter
}

func NewClient(token string) *Client {
    return &Client{
        // ...
        limiter: rate.NewLimiter(rate.Limit(10), 20), // 10 req/sec burst 20
    }
}

func (c *Client) Request(...) error {
    if err := c.limiter.Wait(context.Background()); err != nil {
        return err
    }
    // ... rest of request
}
```

### 🟡 Password Stored in Memory
**File:** `internal/cmd/auth.go:75-77`

Password is stored as a string, which can't be securely wiped from memory in Go.

**Mitigation:** This is a Go limitation. Document that users should not run this on untrusted systems. Consider using `memguard` library for extra protection.

## Low Severity Issues

### 🟢 Missing Input Validation
**File:** `internal/api/client.go:125-136`

No validation that email/password are non-empty before sending to server.

**Fix:**
```go
func (c *Client) SignIn(email, password, mfa string) (*AuthResponse, error) {
    if email == "" || password == "" {
        return nil, fmt.Errorf("email and password are required")
    }
    // ...
}
```

### 🟢 Unbounded Buffer Growth
**File:** `internal/sync/websocket.go:289-298`

The `Pull()` method allocates a buffer of size `totalSize` without checking bounds.

**Fix:**
```go
const MaxFileSize = 500 * 1024 * 1024 // 500 MB

func (c *WebSocketClient) Pull(uid int64) ([]byte, error) {
    // ...
    totalSize := resp.Size
    if totalSize > MaxFileSize {
        return nil, fmt.Errorf("file too large: %d bytes", totalSize)
    }
    data := make([]byte, totalSize)
    // ...
}
```

### 🟢 No Timeout on WebSocket Reads
**File:** `internal/sync/websocket.go:206-209`

WebSocket reads have no timeout. Could hang forever.

**Fix:**
```go
func (c *WebSocketClient) readJSON() (*Message, error) {
    c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
    var msg Message
    err := c.conn.ReadJSON(&msg)
    c.lastMessageTime = time.Now()
    return &msg, err
}
```

## Security Best Practices - Implemented ✅

### ✅ Secure File Permissions
Config files are created with 0600 (owner read/write only):
- `internal/config/config.go:86`
- `internal/config/config.go:146`
- `internal/config/config.go:290`

Directories with 0700 (owner only):
- `internal/config/config.go:67`
- `internal/config/config.go:136`
- `internal/config/config.go:280`

### ✅ HTTPS/WSS by Default
All API calls use HTTPS, WebSocket uses WSS unless explicitly HTTP.

### ✅ No Hardcoded Credentials
No credentials are hardcoded in the source.

### ✅ Error Messages Don't Leak Secrets
Error messages don't expose sensitive data.

### ✅ defer resp.Body.Close()
All HTTP responses properly closed.

### ✅ Proper Error Wrapping
Using `fmt.Errorf` with `%w` for proper error chains.

## Code Quality Issues

### Missing Tests
No test files exist. Should add:
- Unit tests for crypto functions
- Mock tests for API client
- Table-driven tests for validation

### No Logging
No structured logging. Add using `log/slog` or `logrus`:
```go
import "log/slog"

// In main.go
logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
```

### Magic Numbers
Constants like chunk sizes (2MB) and timeouts (30s) should be configurable.

## Summary of Required Fixes

### Must Fix (Breaking Issues)
1. ✅ **Replace PBKDF2 with scrypt** - Wrong algorithm used
2. ✅ **Fix race condition in Disconnect()** - Will panic
3. ✅ **Add context to WebSocket operations** - Can hang forever

### Should Fix (Security)
4. Check preflight OPTIONS errors properly
5. Add TLS certificate validation options
6. Add timeout to WebSocket reads
7. Add max file size check in Pull()

### Nice to Have
8. Add rate limiting
9. Add input validation
10. Add structured logging
11. Add tests

## Testing Recommendations

```bash
# Static analysis
go vet ./...
staticcheck ./...

# Security scanning
gosec ./...

# Race detection
go test -race ./...

# Coverage
go test -cover ./...
```

## Before Production Use

1. **Fix the scrypt issue immediately** - Will not work otherwise
2. Add comprehensive tests
3. Add logging
4. Add rate limiting
5. Add context cancellation
6. Security audit with `gosec`
7. Load testing to find memory leaks
8. Test against real Obsidian API to verify protocol
