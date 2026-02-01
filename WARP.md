# Arc Sync - Warp AI Guidelines

## Project Overview
A Go CLI tool that imports Arc browser data (spaces, folders, tabs) into Zen browser. It reads Arc's `StorableSidebar.json` and writes to Zen's compressed session files using Mozilla's custom LZ4 format.

## Tech Stack
- **Language:** Go 1.21+
- **Dependencies:** 
  - `github.com/google/uuid` - UUID generation for Zen entities
  - `github.com/pierrec/lz4/v4` - Mozilla LZ4 compression/decompression
- **Platform:** macOS (Arc browser is macOS-only)
- **Cache:** Favicons cached under `~/.arc-to-zen/favicons/` as data URLs for faster re-runs

## Project Structure
```
arc-to-zen/
├── cmd/arc-to-zen/     # CLI entrypoint
├── cmd/dump-session/   # Debug tool to inspect session structure
├── backup/             # Backup and restore functionality for zen-sessions
├── favicon/            # Favicon fetching and encoding
├── importer/           # Core import logic (importer.go, helpers.go)
├── mappings/           # Arc → Zen icon/color mappings
├── mozlz4/             # Mozilla LZ4 compression library
├── profiles/           # Profile discovery and reset functionality
├── types/              # Data structure definitions (arc.go, zen.go)
├── go.mod              # Go module definition
└── Makefile            # Build automation
```

## Build & Run Commands
```bash
# Build
make build              # Build for current platform
make build-all          # Build for all platforms (darwin, linux, windows)

# Install
make install            # Install to ~/bin/

# Test
make test               # Run all tests
go test ./...           # Run tests directly

# Development
make run                # Build and run
make clean              # Remove build artifacts
make deps               # Download dependencies
make tidy               # Tidy go.mod
```

## CLI Usage
```bash
# Basic usage (auto-discovers default Zen profile)
arc-to-zen

# List available profiles
arc-to-zen -list

# Import with dry-run (preview only)
arc-to-zen -dry-run

# Verbose output
arc-to-zen -verbose

# Reset profile to default state
arc-to-zen -reset
arc-to-zen -reset -dry-run

# Backup zen-sessions.jsonlz4
arc-to-zen -backup

# Restore a backup (interactive menu)
arc-to-zen -restore

# Explicit profile path
arc-to-zen "~/Library/Application Support/zen/Profiles/xxx.default"
```

## Key File Locations
- **Arc data:** `~/Library/Application Support/Arc/StorableSidebar.json`
- **Zen profiles:** `~/Library/Application Support/zen/Profiles/`
- **Zen session:** `{profile}/zen-sessions.jsonlz4`
- **Zen containers:** `{profile}/containers.json`
- **Backups:** `~/.arc-to-zen/backups/` (timestamped zen-sessions backups)

## Data Flow
1. Read Arc's `StorableSidebar.json` (plain JSON)
2. Read Zen's `zen-sessions.jsonlz4` (Mozilla LZ4 compressed)
3. Read/create `containers.json`
4. Transform Arc spaces → Zen workspaces
5. Transform Arc folders/tabs → Zen folders/tabs
6. **Pre-cache favicons in parallel** (10 concurrent workers)
   - Collects all unique URLs from Arc data
   - Fetches favicons concurrently with 10 parallel workers
   - Caches to disk at `~/.arc-to-zen/favicons/`
   - Skips already cached favicons
7. Process tabs and apply cached favicons
8. Encode favicons as base64 data URLs
9. Backup existing session
10. Write updated session and containers with favicon data

## Important Implementation Details
- **Mozilla LZ4 format:** 8-byte header (`mozLz40\0`) + 4-byte size (LE) + LZ4 block data
- **Arc containers at index 1:** Main container with spaces/items is at `sidebar.containers[1]`
- **Folder children forward:** Children processed in forward order with folder-based sibling chaining
- **Merge mode:** Existing spaces matched by name are updated, not duplicated
- **Fresh session support:** Can create new session from scratch if no existing session file exists
- **Favicon fetching:** Automatically fetches favicons during import and stores as base64 data URLs (cached on disk)
  - **Parallel pre-caching:** Uses 10 concurrent workers to fetch favicons in parallel before import
  - Collects all unique URLs from Arc data before fetching
  - Cache checked first - only fetches uncached favicons
  - Format: `data:image/x-icon;base64,{base64_data}`
  - 5 second timeout per request
  - 1MB size limit
  - Graceful failure: if fetch fails, tab still imports without favicon
  - Supports multiple formats (ico, png, jpeg, gif, svg, webp)
  - Significant performance improvement for large imports
- **Backup and restore:**
  - Backups stored in `~/.arc-to-zen/backups/` with timestamp format `zen-sessions_YYYY-MM-DD_HH-MM-SS.jsonlz4`
  - Backups sorted chronologically (newest first)
  - Restore creates a backup of current state before restoring
  - Interactive menu allows selection of backup to restore
- **Nested folder structure:**
  - Each folder gets an anchor tab (`zenIsEmpty: true` + `groupId`) to ensure Firefox creates the tab-group
  - `prevSiblingInfo` uses `{type: "group", id: folderID}` to reference sibling folders (not tabs)
  - Folder IDs persist through Zen save/restore cycles, unlike empty tab IDs which Zen discards
  - First nested folder: `prevSiblingInfo: null` (default positioning)
  - Subsequent nested folders: chain to previous sibling folder for correct ordering
  - `emptyTabIds` tracks anchor tab UUIDs so Zen marks them as folder placeholders
  - **Collapsed by default:** All folders are imported in collapsed state for cleaner initial view
- **Icon mapping:** Two separate mapping systems in `mappings/mappings.go`:
  - **Workspace icons:** `MapArcIconToSvg()` maps Arc icons to Zen's selectable SVG icons (chrome://browser/skin/zen-icons/selectable/*.svg)
    - `star` → `star-1.svg` (classic 5-pointed star, not the asterisk/sparkle shape in `star.svg`)
  - **Container icons:** `MapArcIconToContainerIcon()` maps Arc icons to Firefox's built-in container icons
    - Firefox containers use a fixed set: fingerprint, briefcase, dollar, cart, circle, gift, vacation, food, fruit, pet, tree, chill, fence
    - Default container icon: "briefcase"

## Code Conventions
- Use `fmt.Errorf` with `%w` for error wrapping
- Logger interface allows custom logging implementation
- Options pattern for configurable behavior (`ImportOptions`)
- Explicit error handling (no panics)

## Testing Notes
- Tests exist in `mozlz4/mozlz4_test.go` and `favicon/favicon_test.go`
- Dry-run mode allows safe testing without file modifications
- Always backup session before writes (automatic)
- Favicon tests use mock HTTP servers to avoid external dependencies
- Caching tests use temp directories and verify cache is used when network is unavailable
