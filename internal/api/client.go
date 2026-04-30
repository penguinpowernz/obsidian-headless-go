package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultAPIURL    = "https://api.obsidian.md"
	PublishAPIURL    = "https://publish.obsidian.md"
	UserAgent        = "obsidian-headless-go/0.1.0"
	OriginHeader     = "https://obsidian.md"
)

// Client handles API requests to Obsidian services
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	PublishURL string
	Token      string
}

// NewClient creates a new API client
func NewClient(token string) *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL:    DefaultAPIURL,
		PublishURL: PublishAPIURL,
		Token:      token,
	}
}

// Request makes an HTTP request to the API
func (c *Client) Request(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// Vault represents a remote vault
type Vault struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Host              string `json:"host"`
	Region            string `json:"region"`
	Salt              string `json:"salt"`
	Password          string `json:"password,omitempty"`
	EncryptionVersion int    `json:"encryption_version"`
}

// VaultListResponse contains vault lists
type VaultListResponse struct {
	Vaults []Vault `json:"vaults"`
	Shared []Vault `json:"shared"`
}

// Site represents a publish site
type Site struct {
	ID   string `json:"id"`
	Host string `json:"host"`
}

// SiteListResponse contains site lists
type SiteListResponse struct {
	Sites  []Site `json:"sites"`
	Shared []Site `json:"shared"`
}

// Region represents a server region
type Region struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

// RegionsResponse contains available regions
type RegionsResponse struct {
	Regions []Region `json:"regions"`
}

// Auth functions

// SignIn authenticates with email and password
func (c *Client) SignIn(email, password, mfa string) (*AuthResponse, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	req := map[string]string{
		"email":    email,
		"password": password,
	}
	if mfa != "" {
		req["mfa"] = mfa
	}

	var resp AuthResponse
	err := c.requestWithOrigin("POST", "/user/signin", req, &resp)
	return &resp, err
}

// SignOut logs out the current session
func (c *Client) SignOut() error {
	req := map[string]string{"token": c.Token}
	return c.Request("POST", "/user/signout", req, nil)
}

// GetUserInfo retrieves user account information
func (c *Client) GetUserInfo() (*UserInfo, error) {
	req := map[string]string{"token": c.Token}
	var resp UserInfo
	err := c.Request("POST", "/user/info", req, &resp)
	return &resp, err
}

// Vault functions

// ListVaults returns all vaults accessible to the user
func (c *Client) ListVaults(encryptionVersion int) (*VaultListResponse, error) {
	req := map[string]interface{}{
		"token":                        c.Token,
		"supported_encryption_version": encryptionVersion,
	}
	var resp VaultListResponse
	err := c.Request("POST", "/vault/list", req, &resp)
	return &resp, err
}

// GetRegions returns available server regions
func (c *Client) GetRegions(host string) (*RegionsResponse, error) {
	req := map[string]string{
		"token": c.Token,
		"host":  host,
	}
	var resp RegionsResponse
	err := c.Request("POST", "/vault/regions", req, &resp)
	return &resp, err
}

// CreateVault creates a new remote vault
func (c *Client) CreateVault(name, keyhash, salt, region string, encVersion int) (*Vault, error) {
	req := map[string]interface{}{
		"token":              c.Token,
		"name":               name,
		"keyhash":            keyhash,
		"salt":               salt,
		"region":             region,
		"encryption_version": encVersion,
	}
	var resp Vault
	err := c.Request("POST", "/vault/create", req, &resp)
	return &resp, err
}

// AccessVault validates access to a vault
func (c *Client) AccessVault(vaultUID, keyhash, host string, encVersion int) error {
	req := map[string]interface{}{
		"token":              c.Token,
		"vault_uid":          vaultUID,
		"keyhash":            keyhash,
		"host":               host,
		"encryption_version": encVersion,
	}
	return c.Request("POST", "/vault/access", req, nil)
}

// Publish functions

// ListSites returns all publish sites
func (c *Client) ListSites() (*SiteListResponse, error) {
	req := map[string]string{"token": c.Token}
	var resp SiteListResponse
	err := c.requestPublish("/publish/list", req, &resp)
	return &resp, err
}

// CreateSite creates a new publish site
func (c *Client) CreateSite() (*Site, error) {
	req := map[string]string{"token": c.Token}
	var resp Site
	err := c.requestPublish("/publish/create", req, &resp)
	return &resp, err
}

// SetSlug sets the slug for a publish site
func (c *Client) SetSlug(siteID, host, slug string) error {
	req := map[string]string{
		"token": c.Token,
		"id":    siteID,
		"host":  host,
		"slug":  slug,
	}
	return c.requestPublish("/api/slug", req, nil)
}

// GetSlugs retrieves slugs for multiple sites
func (c *Client) GetSlugs(siteIDs []string) (map[string]string, error) {
	req := map[string]interface{}{
		"token": c.Token,
		"ids":   siteIDs,
	}
	var resp map[string]string
	err := c.requestPublish("/api/slugs", req, &resp)
	return resp, err
}

// Response types

type AuthResponse struct {
	Token string `json:"token"`
}

type UserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// requestWithOrigin makes a request with Origin header (required for signin)
func (c *Client) requestWithOrigin(method, path string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", OriginHeader)

	// Preflight OPTIONS request for signin
	if path == "/user/signin" {
		optReq, err := http.NewRequest("OPTIONS", url, nil)
		if err == nil {
			optReq.Header.Set("Origin", OriginHeader)
			_, _ = c.HTTPClient.Do(optReq) // Preflight failure is non-fatal
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// requestPublish makes a request to the publish API
func (c *Client) requestPublish(path string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.PublishURL + path
	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
