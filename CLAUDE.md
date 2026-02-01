# Arc Sync - Claude Code Guidelines

## What is this project?
A Go CLI tool (`arc-to-zen`) that migrates Arc browser data into Zen browser. It imports spaces, folders, and tabs while handling Mozilla's custom LZ4 compression format for Zen's session files.

## Quick Start
```bash
make build          # Build binary to ./build/arc-to-zen
make install        # Install to ~/bin/arc-to-zen
make test           # Run tests
```

## Project Layout
- `cmd/arc-to-zen/main.go` - CLI entrypoint, flag parsing
- `cmd/dump-session/main.go` - Debug tool to inspect session structure
- `backup/backup.go` - Backup and restore zen-sessions
- `importer/importer.go` - Main import orchestration
- `importer/helpers.go` - Parsing, filtering, item insertion
- `mappings/mappings.go` - Arc → Zen icon/color mappings
- `mozlz4/mozlz4.go` - Mozilla LZ4 compress/decompress
- `profiles/discovery.go` - Auto-discover Zen profiles
- `profiles/reset.go` - Reset profile to defaults
- `types/arc.go` - Arc data structures
- `types/zen.go` - Zen data structures

## Key Technical Details
1. **Mozilla LZ4 Format:** `mozLz40\0` header (8 bytes) + uncompressed size (4 bytes LE) + LZ4 block
2. **Arc Data Source:** `~/Library/Application Support/Arc/StorableSidebar.json`
3. **Zen Session File:** `{profile}/zen-sessions.jsonlz4` (compressed)
4. **Main Container:** Arc spaces/items are in `sidebar.containers[1]` (index 1, not 0)
5. **Item Ordering:** Children processed in forward order with folder-based sibling chaining
6. **Parallel Favicon Pre-caching:** Uses 10 concurrent workers to fetch favicons before import
7. **Fresh Session Support:** Can create new session from scratch if no existing session file exists

## CLI Flags
- `-dry-run` - Preview changes without writing
- `-verbose` - Detailed output
- `-reset` - Remove session files to reset profile
- `-list` - Show available Zen profiles
- `-backup` - Create timestamped backup of zen-sessions.jsonlz4
- `-restore` - Restore a backup (interactive menu)

## Common Tasks

### Adding a new icon mapping
Edit `mappings/mappings.go`, add to `arcIconToZenSvg` map:
```go
"iconname": "chrome://browser/skin/zen-icons/selectable/iconname.svg",
```

### Adding a new color mapping
Edit `mappings/mappings.go`, add to `arcColorToZen` map:
```go
"arc-color": "zen-color",
```

### Modifying import behavior
Edit `importer/importer.go`:
- `doImport()` - Main import logic
- `insertItemWithChildren()` in `helpers.go` - Tab/folder creation

## Dependencies
- `github.com/google/uuid` - UUIDs for Zen entities
- `github.com/pierrec/lz4/v4` - LZ4 compression

## Testing
```bash
go test ./...           # Run all tests
go test -cover ./...    # With coverage
go test ./mozlz4/...    # Test specific package
```

## Error Handling
- Use `fmt.Errorf("context: %w", err)` for wrapping
- Check `os.IsNotExist(err)` for file existence
- Validate paths before operations

## Backup & Restore
- Backups stored in `~/.arc-to-zen/backups/`
- Filename format: `zen-sessions_YYYY-MM-DD_HH-MM-SS.jsonlz4`
- Sorted chronologically (newest first)
- Restore creates backup of current state before restoring
- Interactive selection menu for restores

## Performance
- **Favicon pre-caching:** Collects all URLs upfront and fetches in parallel (10 workers)
- Significantly faster than sequential fetching
- Cache-aware: skips already cached favicons
- Reports stats: cached/fetched/failed counts

## Nested Folder Structure (CRITICAL)
This was a complex fix - Zen browser has specific requirements for nested folders to work:

1. **Anchor tabs required:** Each folder needs at least one tab with `zenIsEmpty: true` AND `groupId: folderID` for Firefox to create the tab-group element
2. **Folder-based sibling references:** `prevSiblingInfo` must use `{type: "group", id: folderID}` to reference sibling FOLDERS (not tabs). Tab IDs get discarded by Zen on save.
3. **Forward processing order:** Children are processed in forward order with each folder referencing its predecessor
4. **emptyTabIds tracking:** The anchor tab's UUID must be in `emptyTabIds` so Zen marks it as a folder placeholder

**Why this matters:**
- Zen's `#filterUnusedTabs()` removes empty tabs without groupId
- Tab IDs created by arc-to-zen don't persist through Zen's save/restore cycle
- Folder IDs DO persist, so sibling references must use folder IDs

**Common issues if broken:**
- Zen crashes on second open → `prevSiblingInfo` references tab IDs that were discarded
- Nested folders appear flat → No tab with folder's `groupId` exists
- Folder order reversed → Wrong processing order or sibling chaining

## Notes
- macOS only (Arc browser requirement)
- Always backs up session before writing
- Merge mode: matches existing spaces by name
- Dry-run mode is safe for testing
