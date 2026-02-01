package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"arc-to-zen/backup"
	"arc-to-zen/favicon"
	"arc-to-zen/importer"
	"arc-to-zen/mozlz4"
	"arc-to-zen/profiles"
)

func main() {
	// Define flags
	dryRun := flag.Bool("dry-run", false, "Show what would be imported without making changes")
	verbose := flag.Bool("verbose", false, "Show detailed output")
	reset := flag.Bool("reset", false, "Reset the profile to default state (removes session files)")
	listProfiles := flag.Bool("list", false, "List available Zen profiles")
	decompress := flag.String("decompress", "", "Decompress a Mozilla LZ4 (.jsonlz4) file and print JSON to stdout")
	backupSession := flag.Bool("backup", false, "Create a backup of the zen-sessions.jsonlz4 file")
	restoreSession := flag.Bool("restore", false, "Restore a backup of the zen-sessions.jsonlz4 file")
	faviconStats := flag.Bool("favicon-stats", false, "Show favicon cache statistics")
	faviconRetryFailed := flag.Bool("favicon-retry-failed", false, "Clear failed favicon cache entries so they will be retried on next import")
	faviconClearCache := flag.Bool("favicon-clear-cache", false, "Clear entire favicon cache for a fresh re-fetch on next import")
	flag.Usage = printUsage
	flag.Parse()

	// Handle favicon cache commands
	if *faviconStats || *faviconRetryFailed || *faviconClearCache {
		f := favicon.New()

		if *faviconStats {
			total, successful, failed, err := f.GetCacheStats()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Favicon Cache Statistics:")
			fmt.Printf("  Total entries:  %d\n", total)
			fmt.Printf("  Successful:     %d\n", successful)
			fmt.Printf("  Failed:         %d\n", failed)
			os.Exit(0)
		}

		if *faviconRetryFailed {
			// First show stats
			_, _, failed, err := f.GetCacheStats()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if failed == 0 {
				fmt.Println("No failed favicon entries to clear.")
				os.Exit(0)
			}

			removed, err := f.ClearFailedCache()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Cleared %d failed favicon entries.\n", removed)
			fmt.Println("Run 'arc-to-zen' again to retry fetching these favicons.")
			os.Exit(0)
		}

		if *faviconClearCache {
			// First show stats
			total, _, _, err := f.GetCacheStats()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if total == 0 {
				fmt.Println("Favicon cache is already empty.")
				os.Exit(0)
			}

			removed, err := f.ClearCache()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Cleared %d favicon cache entries.\n", removed)
			fmt.Println("Run 'arc-to-zen' again to fetch all favicons fresh.")
			os.Exit(0)
		}
	}

	// Handle decompress command
	if *decompress != "" {
		filePath := *decompress

		// If "default" is specified, use the default profile's zen-sessions.jsonlz4
		if filePath == "default" {
			defaultProfile, err := profiles.GetDefaultProfile()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not find default profile: %v\n", err)
				os.Exit(1)
			}
			filePath = filepath.Join(defaultProfile.Path, "zen-sessions.jsonlz4")
		}
		
		if err := decompressFile(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Handle list profiles command
	if *listProfiles {
		profileList, err := profiles.DiscoverProfiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(profiles.ListProfiles(profileList))
		os.Exit(0)
	}

	// Handle backup and restore commands (need profile path)
	if *backupSession || *restoreSession {
		args := flag.Args()
		var zenProfilePath string

		if len(args) > 0 {
			zenProfilePath = args[0]
		} else {
			defaultProfile, err := profiles.GetDefaultProfile()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: No profile path provided and auto-discovery failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "\nPlease provide a profile path or use --list to see available profiles.\n")
				os.Exit(1)
			}
			zenProfilePath = defaultProfile.Path
			fmt.Printf("Using auto-discovered profile: %s\n", defaultProfile.Name)
			fmt.Printf("Profile path: %s\n\n", zenProfilePath)
		}

		if *backupSession {
			if err := backup.CreateBackup(zenProfilePath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: backup failed: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}

		if *restoreSession {
			if err := backup.RestoreBackup(zenProfilePath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: restore failed: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	// Get profile path (optional if auto-discovery works)
	args := flag.Args()
	var zenProfilePath string
	
	if len(args) > 0 {
		// Profile path provided as argument
		zenProfilePath = args[0]
	} else {
		// Try auto-discovery
		defaultProfile, err := profiles.GetDefaultProfile()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: No profile path provided and auto-discovery failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "\nPlease provide a profile path or use --list to see available profiles.\n")
			printUsage()
			os.Exit(1)
		}
		zenProfilePath = defaultProfile.Path
		fmt.Printf("Using auto-discovered profile: %s\n", defaultProfile.Name)
		fmt.Printf("Profile path: %s\n\n", zenProfilePath)
	}

	// Handle reset command
	if *reset {
		if err := profiles.ResetProfile(zenProfilePath, *dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Error: reset failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Default Arc data path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	arcDataPath := filepath.Join(homeDir, "Library", "Application Support", "Arc", "StorableSidebar.json")

	// Check if Arc data exists
	if _, err := os.Stat(arcDataPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Arc browser data not found at: %s\n", arcDataPath)
		fmt.Fprintf(os.Stderr, "Make sure Arc is installed and has been used.\n")
		os.Exit(1)
	}

	// Create importer with options
	opts := importer.ImportOptions{
		DryRun:  *dryRun,
		Verbose: *verbose,
	}
	imp := importer.NewWithOptions(zenProfilePath, nil, opts)

	// Perform import
	result, err := imp.Import(arcDataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: import failed: %v\n", err)
		os.Exit(1)
	}

	if result.Success {
		if *dryRun {
			fmt.Println("\n✓ Dry-run completed successfully (no changes made)")
		} else {
			fmt.Println("\n✓ Import completed successfully")
		}
		os.Exit(0)
	} else {
		fmt.Fprintf(os.Stderr, "\nError: import failed\n")
		os.Exit(1)
	}
}

func decompressFile(path string) error {
	// Read compressed file
	compressedData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Decompress
	decompressedData, err := mozlz4.Decompress(compressedData)
	if err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}

	// Parse JSON to pretty print
	var jsonData interface{}
	if err := json.Unmarshal(decompressedData, &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Pretty print to stdout
	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	fmt.Println(string(prettyJSON))
	return nil
}

func printUsage() {
	fmt.Println("Arc to Zen Browser Import Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  arc-to-zen [options] [zen-profile-path]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -dry-run              Show what would be imported/reset without making changes")
	fmt.Println("  -verbose              Show detailed output during import")
	fmt.Println("  -reset                Reset the profile to default state (removes session files)")
	fmt.Println("  -list                 List all available Zen profiles")
	fmt.Println("  -decompress <file>    Decompress a Mozilla LZ4 file and print JSON to stdout")
	fmt.Println("  -backup               Create a timestamped backup of zen-sessions.jsonlz4")
	fmt.Println("  -restore              Restore a backup of zen-sessions.jsonlz4")
	fmt.Println("")
	fmt.Println("Favicon Cache:")
	fmt.Println("  -favicon-stats        Show favicon cache statistics")
	fmt.Println("  -favicon-retry-failed Clear failed entries so they retry on next import")
	fmt.Println("  -favicon-clear-cache  Clear entire cache for fresh fetch on next import")
	fmt.Println("")
	fmt.Println("Profile Path:")
	fmt.Println("  If no profile path is provided, the tool will auto-discover your default")
	fmt.Println("  Zen profile. Use -list to see all available profiles.")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Auto-discover and import")
	fmt.Println("  arc-to-zen")
	fmt.Println("")
	fmt.Println("  # List available profiles")
	fmt.Println("  arc-to-zen -list")
	fmt.Println("")
	fmt.Println("  # Import with explicit profile path")
	fmt.Println("  arc-to-zen \"~/Library/Application Support/zen/Profiles/xxx.default\"")
	fmt.Println("")
	fmt.Println("  # Reset profile to default state")
	fmt.Println("  arc-to-zen -reset")
	fmt.Println("")
	fmt.Println("  # Dry-run to see what would happen")
	fmt.Println("  arc-to-zen -dry-run")
	fmt.Println("  arc-to-zen -reset -dry-run")
	fmt.Println("")
	fmt.Println("  # Decompress a .jsonlz4 file")
	fmt.Println("  arc-to-zen -decompress default")
	fmt.Println("  arc-to-zen -decompress default > output.json")
	fmt.Println("  arc-to-zen -decompress /path/to/zen-sessions.jsonlz4")
	fmt.Println("")
	fmt.Println("  # Create a backup of zen-sessions.jsonlz4")
	fmt.Println("  arc-to-zen -backup")
	fmt.Println("")
	fmt.Println("  # Restore a backup (interactive menu)")
	fmt.Println("  arc-to-zen -restore")
	fmt.Println("")
	fmt.Println("  # View favicon cache stats")
	fmt.Println("  arc-to-zen -favicon-stats")
	fmt.Println("")
	fmt.Println("  # Retry failed favicon fetches")
	fmt.Println("  arc-to-zen -favicon-retry-failed")
	fmt.Println("")
	fmt.Println("  # Clear favicon cache for fresh fetch")
	fmt.Println("  arc-to-zen -favicon-clear-cache")
	fmt.Println("")
	fmt.Println("This tool imports Arc browser spaces, folders, and tabs into Zen browser.")
}
