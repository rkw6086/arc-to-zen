# Arc to Zen Browser Import Tool

A Go implementation of the Arc browser data import tool for Zen browser. This tool imports Arc browser spaces, folders, and tabs into your Zen browser profile.

## Features

- ✅ **Auto-discovery** of Zen profiles (no manual path needed)
- ✅ Import Arc spaces as Zen workspaces
- ✅ Import Arc folders and nested folder hierarchies
- ✅ Import Arc tabs with full metadata
- ✅ Automatic container management
- ✅ Icon and color mapping from Arc to Zen
- ✅ Automatic session backup before import
- ✅ Merge mode: updates existing spaces or creates new ones
- ✅ **Reset function** to restore profile to default state
- ✅ **List profiles** to see all available Zen profiles
- ✅ **Dry-run mode**: Preview changes without modifying anything
- ✅ Verbose output for debugging

## Installation

### Via Homebrew (Recommended)

```bash
# Add the tap
brew tap rkw6086/arc-to-zen

# Install
brew install arc-to-zen
```

To update:
```bash
brew update
brew upgrade arc-to-zen
```

### Build from source

```bash
# Clone the repository
git clone https://github.com/rkw6086/arc-to-zen.git
cd arc-to-zen

# Install dependencies
go mod download

# Build the binary
go build -o arc-to-zen cmd/arc-to-zen/main.go

# Or use make
make build
```

### Install to system

```bash
make install
```

This installs the `arc-to-zen` binary to `~/bin/`.

## Usage

### Basic Usage (Auto-discovery)

```bash
# Import using auto-discovered default profile
arc-to-zen

# List all available profiles
arc-to-zen -list

# Reset profile to default state
arc-to-zen -reset

# Dry-run to see what would happen
arc-to-zen -dry-run
arc-to-zen -reset -dry-run
```

### Advanced Usage (Explicit Profile Path)

```bash
# Import with explicit profile path
arc-to-zen "/Users/username/Library/Application Support/zen/Profiles/xxx.default"

# Or using tilde expansion
arc-to-zen "~/Library/Application Support/zen/Profiles/xxx.default"

# Verbose output
arc-to-zen -verbose "~/Library/Application Support/zen/Profiles/xxx.default"
```

### Profile Auto-Discovery

The tool automatically discovers your Zen profiles at:
`~/Library/Application Support/zen/Profiles/`

If you have multiple profiles, it will use the default one. Use `-list` to see all available profiles and their paths.

### Finding your Zen profile path manually

1. Open Zen Browser
2. Go to `about:profiles`
3. Find your profile's "Root Directory" path
4. Copy that path and use it with the tool

### Commands

#### Import (Default)

Imports Arc browser data into Zen:

```bash
arc-to-zen [options] [profile-path]
```

Options:
- `-dry-run` - Show what would be imported without making changes
- `-verbose` - Show detailed output during import

#### List Profiles

List all available Zen profiles:

```bash
arc-to-zen -list
```

#### Reset Profile

Reset a Zen profile to default state by removing session files:

```bash
arc-to-zen -reset [profile-path]
```

This removes:
- `zen-sessions.jsonlz4` - Current session file
- `zen-sessions-backup/` - Session backup directory
- `sessionstore.jsonlz4` - Firefox-compatible session store
- `sessionstore-backups/` - Session store backups

Use with `-dry-run` to preview what would be removed:
```bash
arc-to-zen -reset -dry-run
```

## How it works

1. **Reads Arc data** from `~/Library/Application Support/Arc/StorableSidebar.json`
2. **Reads Zen session** from your profile's `zen-sessions.jsonlz4` file
3. **Creates backups** of your Zen session before making changes
4. **Imports spaces** - each Arc space becomes a Zen workspace
5. **Imports items** - Arc folders and tabs are imported with their hierarchy preserved
6. **Updates containers** - creates or updates container identities for each space
7. **Writes back** the updated session and container data

## Technical Details

### Mozilla LZ4 Compression

Zen browser stores its session data in Mozilla's custom LZ4 format:
- 8-byte header: `"mozLz40\0"`
- 4-byte uncompressed size (little-endian uint32)
- LZ4 block-compressed data

This tool handles compression/decompression automatically.

### Data Structures

The tool properly handles:
- Arc spaces → Zen workspaces with themes and icons
- Arc folders → Zen folders with group metadata
- Arc tabs → Zen pinned tabs with full history entries
- Arc containers → Zen container identities
- Nested folder hierarchies
- Tab ordering and positioning

### Icon Mapping

Arc icon names are mapped to Zen's SVG icon paths. Over 100 icon mappings are included, covering:
- Work & Business (briefcase, office, etc.)
- Communication (mail, chat, phone)
- Development (code, terminal, bug)
- Media (music, video, game)
- And many more...

### Color Mapping

Arc color themes are mapped to Zen's container colors:
- blue, red, green, yellow, orange, purple, pink
- Shades automatically map to base colors
- Default fallback to gray

## Safety

- ✅ **Automatic backups** - session is backed up before any changes
- ✅ **Non-destructive** - existing Zen data is preserved
- ✅ **Merge mode** - updates existing spaces by name instead of duplicating
- ✅ **Validation** - validates all paths and data before importing
- ✅ **Dry-run mode** - preview changes before applying them

## Requirements

- macOS (Arc browser is macOS-only)
- Arc Browser installed with some data
- Zen Browser installed with a valid profile
- Go 1.21+ (for building from source)

## Development

### Project Structure

```
arc-to-zen/
├── cmd/arc-to-zen/     # CLI application
├── backup/             # Backup and restore functionality
├── favicon/            # Favicon fetching and caching
├── importer/           # Core import logic
├── mappings/           # Icon/color mappings
├── mozlz4/             # Mozilla LZ4 compression
├── profiles/           # Profile discovery and reset
├── types/              # Data structure definitions
├── go.mod              # Go module definition
├── Makefile            # Build automation
└── README.md           # This file
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms (macOS, Linux, Windows)
make build-all

# Run tests
make test

# Clean build artifacts
make clean
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Comparison with JavaScript Version

This Go implementation offers several advantages over the original Node.js version:

- **Single binary** - no Node.js or npm dependencies
- **Faster** - significantly faster execution
- **Type safety** - compile-time type checking
- **Better error handling** - explicit error handling throughout
- **Easier distribution** - just copy the binary
- **Cross-compilation** - easy to build for different platforms

## License

MIT

