# Obsidian Headless (Go Port)

**A clean, readable Go implementation of the Obsidian headless CLI.**

Headless client for [Obsidian Sync](https://obsidian.md/sync) and [Obsidian Publish](https://obsidian.md/publish).
Sync and publish your vaults from the command line without the desktop app.

In 2026 using NPM is security nightmare. This repository will be deleted if Obsidian releases a client in safe language like Golang or Rust.

## Quick Start

```bash
# Build
go build -o ob ./cmd/ob

# Install
go install ./cmd/ob

# Login and list vaults
ob login --email your@email.com
ob sync-list-remote

# Setup a vault for syncing
ob sync-setup --vault "My Vault"
ob sync-config --path /path/to/vault

# List publish sites
ob publish-list-sites
```

**Current limitation:** Authentication and vault/site management work. File sync/publish protocols are implemented but not yet wired to file operations.

## Status

This is a **successful port** from the heavily minified JavaScript version. The original code appeared impossible to reverse-engineer, but after beautifying with Prettier, all API endpoints and protocols became clear.

### ✅ Fully Implemented

**Core Infrastructure:**
- ✅ Project structure and Go module setup
- ✅ Configuration management (vault and site configs with 0600 permissions)
- ✅ CLI framework using Cobra with all commands
- ✅ Helper packages for code reuse (auth, vault, site, UI)

**Cryptography:**
- ✅ Scrypt key derivation (N=32768, r=8, p=1) - matching JavaScript
- ✅ AES-GCM encryption/decryption
- ✅ HKDF key derivation
- ✅ Secure salt generation

**API Client:**
- ✅ Full HTTP client with proper headers and CORS preflight
- ✅ Authentication endpoints (signin, signout, user info)
- ✅ Vault management (list, create, access, regions)
- ✅ Publish endpoints (list sites, create, slug management)
- ✅ Error handling and response parsing

**Commands:**
- ✅ `login` / `logout` - Working authentication
- ✅ `sync-list-remote` / `sync-list-local` - List vaults
- ✅ `sync-create-remote` - Create new vaults
- ✅ `sync-setup` - Configure vault sync
- ✅ `sync-config` / `sync-status` / `sync-unlink` - Management
- ✅ `publish-list-sites` - List publish sites
- ✅ `publish-create-site` - Create sites with slugs
- ✅ `publish-setup` / `publish-config` / `publish-unlink` - Management

**WebSocket Sync Protocol:**
- ✅ WebSocket client with TLS support
- ✅ Connection with authentication and encryption
- ✅ Heartbeat/ping mechanism (20s interval, 2min max idle)
- ✅ File pull (download) with chunked transfer (2MB chunks)
- ✅ File push (upload) with chunked transfer
- ✅ List deleted files
- ✅ JSON message protocol (ops: init, ping, pong, pull, push, deleted, history)

### 🔧 Needs Integration

**Sync Engine:**
- File scanning and state tracking
- Diff calculation between local and remote
- Encryption/decryption integration
- Conflict resolution
- File watching for continuous mode

**Publish Engine:**
- YAML frontmatter parsing
- File filtering by `publish: true`
- Upload to publish sites
- Deletion handling

**Quality Improvements:**
- Comprehensive test coverage
- Structured logging
- Progress bars for uploads/downloads
- Rate limiting for API calls

## Why Port to Go?

The original JavaScript implementation is heavily minified (167 lines with 11,000+ character line lengths), making it difficult to:

- NPM presents an untenable security risk in 2026
- Understand the sync/publish protocols
- Modify or extend functionality
- Debug issues
- Audit security

This Go port provides:
- **Clean, readable code** - 2,462 lines of well-structured Go vs minified JavaScript
- **Better performance** - Native binary vs Node.js overhead (~10MB standalone)
- **Improved maintainability** - Proper separation of concerns, helper packages
- **Enhanced security** - Easy to audit, secure file permissions (0600/0700)
- **Linux-focused** - No unnecessary cross-platform complexity

## Building

```bash
go mod download
go build -o ob ./cmd/ob
```

## Installation

```bash
go install ./cmd/ob
```

## Usage

The command interface matches the original JavaScript version:

```bash
# Authentication
ob login                        # Login with email/password
ob logout                       # Logout and remove credentials

# Sync - List and Setup
ob sync-list-remote            # List remote vaults
ob sync-list-local             # List locally configured vaults
ob sync-create-remote          # Create a new remote vault
ob sync-setup --vault "My Vault"  # Setup sync for a vault

# Sync - Configuration
ob sync-config                 # View/update vault sync config
  --path /path/to/vault
  --sync-mode bidirectional    # or: pull-only, mirror-remote
  --conflict-strategy merge    # or: local, remote
  --exclude-folders .git
  --file-types md,pdf,png
  --device-name "My Device"

# Sync - Operations
ob sync                        # Perform sync (not yet wired up)
ob sync-status                 # View sync status
ob sync-unlink                 # Unlink vault from sync

# Publish - List and Setup
ob publish-list-sites          # List publish sites
ob publish-create-site         # Create a new publish site
ob publish-setup --site "my-site"  # Setup publish for a site

# Publish - Configuration
ob publish-config              # View/update publish config
  --path /path/to/vault
  --includes "*.md"
  --excludes "private/**"

# Publish - Operations
ob publish                     # Publish vault (not yet wired up)
ob publish-unlink              # Unlink vault from publish
```

## Testing the Current Implementation

Since this requires a real Obsidian account:

```bash
# 1. Build the binary
go build -o ob ./cmd/ob

# 2. Test authentication
./ob login --email your@email.com
# Enter your password when prompted
# Enter MFA code if enabled

# 3. List your vaults
./ob sync-list-remote

# 4. List your publish sites
./ob publish-list-sites

# 5. Configure a vault (replace with your vault ID)
./ob sync-setup --vault "Your Vault Name"

# 6. View configuration
./ob sync-status --path /path/to/your/vault
```

**Note:** The `sync` and `publish` commands are scaffolded but not yet wired to file operations. They won't sync files yet.

## Configuration

Configuration files are stored in `~/.obsidian-headless/`:
- `auth.json` - Authentication token
- `vaults/*.json` - Vault configurations
- `sites/*.json` - Publish site configurations

## Reverse Engineering Success

The original JavaScript appeared impossible to reverse-engineer due to heavy minification. However, using `npx prettier` to beautify the code made everything clear.

### Discovered API Endpoints

**Base URLs:**
- Main API: `https://api.obsidian.md`
- Publish API: `https://publish.obsidian.md`

**Authentication:**
- `POST /user/signin` - Login with email/password/MFA
- `POST /user/signout` - Logout
- `POST /user/info` - Get user info

**Vault Management:**
- `POST /vault/list` - List vaults
- `POST /vault/create` - Create vault
- `POST /vault/access` - Validate vault access
- `POST /vault/regions` - Get available regions

**Publish:**
- `POST /publish/list` - List sites
- `POST /publish/create` - Create site
- `POST /api/slug` - Set site slug
- `POST /api/slugs` - Get site slugs

### Sync Protocol Details

**WebSocket-based real-time sync:**
- Connects via WSS to vault host
- JSON messages with `op` field
- Operations: `init`, `ping`, `pong`, `pull`, `push`, `deleted`, `history`
- Files transferred in 2MB chunks
- Heartbeat every 20s, max idle 2 minutes

**Encryption:**
- Version 0: Simple AES-GCM
- Version 2/3: HKDF-based with AES-SIV for deterministic filename encryption
- Key derivation: Scrypt (N=32768, r=8, p=1) for password→key
- HKDF with info strings: "ObsidianKeyHash", "ObsidianAesGcm"
- File content: AES-GCM (random IV)
- Filenames: AES-SIV (deterministic for deduplication)

## Security Considerations

### ✅ Implemented Security Features
- **Secure file permissions**: Config files (0600), directories (0700)
- **No credential storage**: Passwords never stored, only derived keys
- **TLS/WSS by default**: All API calls use HTTPS, WebSocket uses WSS
- **Proper error handling**: No secrets leaked in error messages
- **Scrypt key derivation**: Industry-standard parameters (N=32768, r=8, p=1)
- **AES-GCM encryption**: Authenticated encryption for file content
- **HKDF for subkeys**: Proper key derivation for multiple purposes

### ⚠️ Known Issues (from Security Audit)
See `SECURITY_AUDIT.md` for detailed analysis. Key items:
- ✅ **Fixed**: Scrypt implementation (was incorrectly using PBKDF2)
- 🔧 **TODO**: Add context cancellation to WebSocket operations
- 🔧 **TODO**: Add max file size validation (currently 500MB limit)
- 🔧 **TODO**: Add rate limiting for API calls
- 🔧 **TODO**: Add read timeouts for WebSocket messages

### Best Practices
- Use a **test account** for development
- Always **backup your vaults** before testing
- Store credentials securely (config dir is `~/.obsidian-headless/`)
- Review `SECURITY_AUDIT.md` before production use

## Project Structure

```
cmd/ob/main.go                 # Entry point
internal/
  ├── cmd/
  │   ├── root.go              # CLI root
  │   ├── auth.go              # Login/logout (133 lines)
  │   ├── sync.go              # Sync commands (405 lines)
  │   ├── publish.go           # Publish commands (310 lines)
  │   └── util.go              # Utility functions
  ├── api/
  │   └── client.go            # HTTP client with all endpoints
  ├── config/
  │   └── config.go            # Config management with secure permissions
  ├── crypto/
  │   └── crypto.go            # Scrypt, AES-GCM, HKDF encryption
  ├── sync/
  │   └── websocket.go         # WebSocket client for real-time sync
  ├── helpers/
  │   ├── auth.go              # Authentication helpers
  │   ├── vault.go             # Vault config helpers (173 lines)
  │   └── site.go              # Publish site helpers (70 lines)
  └── ui/
      └── output.go            # User interface and output formatting (94 lines)
```

**Total:** 2,462 lines of clean, readable Go code

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/gorilla/websocket` - WebSocket client
- `golang.org/x/crypto` - Cryptographic primitives (scrypt, HKDF)
- `golang.org/x/term` - Terminal password input

## License

This is an unofficial port. The original obsidian-headless is UNLICENSED by Dynalist Inc.
This port is provided as-is for educational and personal use only.

## What's Left to Do

The core infrastructure and protocols are complete. What remains is integration:

### High Priority
1. **Sync Engine Integration** - Wire up the WebSocket client to file operations
   - File scanning and state tracking
   - Diff calculation (local vs remote)
   - Integrate encryption/decryption
   - Conflict resolution
   - File watching for continuous mode (using `fsnotify`)

2. **Publish Implementation** - Connect publish commands to actual operations
   - YAML frontmatter parsing
   - Filter files with `publish: true`
   - Upload to publish sites
   - Handle deletions

3. **Testing & Quality**
   - Unit tests (especially for crypto)
   - Integration tests with mock server
   - Comprehensive error handling
   - Structured logging (slog or logrus)
   - Progress bars for large operations

### Medium Priority
4. **Performance & Reliability**
   - Rate limiting for API calls
   - Context cancellation for long operations
   - Better WebSocket reconnection logic
   - Parallel file uploads/downloads

5. **User Experience**
   - Better error messages
   - Shell completions (bash, zsh, fish)
   - Configuration validation
   - Dry-run mode

## Development Notes

### How the Port Was Completed

The JavaScript code initially appeared impossible to reverse-engineer:
- 167 lines of heavily minified code
- Line lengths exceeding 11,000 characters
- Meaningless variable names (`s`, `e`, `t`, etc.)
- Claude Code CLI

**The breakthrough:** Using `npx prettier cli.js > cli-beautified.js` made the code completely readable:
- API endpoints became visible as string literals
- Protocol logic was clear and well-structured
- Encryption scheme was documented in code
- WebSocket operations were straightforward

**Result:** What seemed like weeks of reverse engineering took hours once beautified.

### Code Quality Improvements

The refactoring focused on:
- **Reducing complexity**: Functions reduced by 40-86% in size
- **DRY principle**: Extracted repeated patterns (path resolution, config updates, output)
- **Separation of concerns**: Helper packages for auth, vault, site, UI
- **Better maintainability**: Single source of truth for common operations


## Comparison: JavaScript vs Go

| Aspect | JavaScript (Original) | Go (Port) |
|--------|----------------------|-----------|
| Lines of code | 167 (minified) | 2,462 (readable) |
| Readability | Nearly impossible | Clear and documented |
| Dependencies | 2 (commander, better-sqlite3) | 3 (cobra, crypto, websocket) |
| Binary size | N/A (needs Node.js runtime) | ~10MB standalone binary |
| Performance | Node.js overhead | Native performance |
| Debugging | Nearly impossible | Standard Go tools |
| Security audit | Extremely difficult | Easy to review |
| Modifications | Requires JS expertise | Standard Go patterns |
| Malware | NPM | No NPM |

## Disclaimer

This is an unofficial tool and is not affiliated with or endorsed by Obsidian or Dynalist Inc.

**Legal:** Reverse engineering the protocol may violate Obsidian's Terms of Service. This project is for **educational purposes** and **personal use** only.

**Support Obsidian:** Consider subscribing to official [Obsidian Sync](https://obsidian.md/sync) and [Publish](https://obsidian.md/publish) services 
rather than using third-party clients. **Convince them to use languages that don't present security nightmares via their distribution channels**

**Use at your own risk.** Always backup your vaults before using sync tools. The authors are not responsible for data loss.
