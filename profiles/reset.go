package profiles

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResetProfile resets a Zen profile to default state by removing session files and backups
func ResetProfile(profilePath string, dryRun bool) error {
	// Files and directories to remove
	itemsToRemove := []string{
		"zen-sessions.jsonlz4",
		"zen-sessions-backup",
		"sessionstore.jsonlz4",
		"sessionstore-backups",
	}

	if dryRun {
		fmt.Println("DRY-RUN MODE - Showing what would be removed:")
		fmt.Println()
	} else {
		fmt.Println("Resetting Zen profile...")
		fmt.Println()
	}

	removedCount := 0
	notFoundCount := 0

	for _, item := range itemsToRemove {
		itemPath := filepath.Join(profilePath, item)
		
		// Check if item exists
		info, err := os.Stat(itemPath)
		if os.IsNotExist(err) {
			if dryRun {
				fmt.Printf("  ⊘ %s (not found, skipping)\n", item)
			} else {
				fmt.Printf("  ⊘ %s (not found)\n", item)
			}
			notFoundCount++
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking %s: %w", item, err)
		}

		// Display what will be or was removed
		if info.IsDir() {
			if dryRun {
				fmt.Printf("  🗑  Would remove directory: %s\n", item)
			} else {
				fmt.Printf("  🗑  Removing directory: %s\n", item)
				if err := os.RemoveAll(itemPath); err != nil {
					return fmt.Errorf("failed to remove %s: %w", item, err)
				}
			}
		} else {
			if dryRun {
				fmt.Printf("  🗑  Would remove file: %s\n", item)
			} else {
				fmt.Printf("  🗑  Removing file: %s\n", item)
				if err := os.Remove(itemPath); err != nil {
					return fmt.Errorf("failed to remove %s: %w", item, err)
				}
			}
		}
		removedCount++
	}

	fmt.Println()
	if dryRun {
		fmt.Printf("Summary: Would remove %d items (%d not found)\n", removedCount, notFoundCount)
		fmt.Println()
		fmt.Println("Run without --dry-run to perform the actual reset.")
	} else {
		fmt.Printf("✓ Reset complete: Removed %d items (%d not found)\n", removedCount, notFoundCount)
		fmt.Println()
		fmt.Println("The profile has been reset to default state.")
		fmt.Println("Next time you start Zen, it will create fresh session files.")
	}

	return nil
}
