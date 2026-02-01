package backup

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	backupDirName    = ".arc-to-zen/backups"
	sessionFileName  = "zen-sessions.jsonlz4"
	backupTimeFormat = "2006-01-02_15-04-05"
)

// BackupInfo represents metadata about a backup
type BackupInfo struct {
	Path      string
	Timestamp time.Time
	Name      string
}

// getBackupDir returns the path to the backup directory
func getBackupDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, backupDirName), nil
}

// ensureBackupDir creates the backup directory if it doesn't exist
func ensureBackupDir() (string, error) {
	backupDir, err := getBackupDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	return backupDir, nil
}

// CreateBackup creates a timestamped backup of the zen-sessions.jsonlz4 file
func CreateBackup(profilePath string) error {
	backupDir, err := ensureBackupDir()
	if err != nil {
		return err
	}

	sessionPath := filepath.Join(profilePath, sessionFileName)
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return fmt.Errorf("zen-sessions.jsonlz4 not found at: %s", sessionPath)
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format(backupTimeFormat)
	backupFilename := fmt.Sprintf("zen-sessions_%s.jsonlz4", timestamp)
	backupPath := filepath.Join(backupDir, backupFilename)

	// Copy the file
	if err := copyFile(sessionPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	fmt.Printf("✓ Backup created: %s\n", backupFilename)
	return nil
}

// ListBackups returns a list of all backups sorted chronologically (newest first)
func ListBackups() ([]BackupInfo, error) {
	backupDir, err := getBackupDir()
	if err != nil {
		return nil, err
	}

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonlz4") {
			continue
		}

		// Parse timestamp from filename
		timestamp, err := parseBackupTimestamp(entry.Name())
		if err != nil {
			continue // Skip files with invalid format
		}

		backups = append(backups, BackupInfo{
			Path:      filepath.Join(backupDir, entry.Name()),
			Timestamp: timestamp,
			Name:      entry.Name(),
		})
	}

	// Sort by timestamp, newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// RestoreBackup presents a menu to select and restore a backup
func RestoreBackup(profilePath string) error {
	backups, err := ListBackups()
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		return fmt.Errorf("no backups found in ~/.arc-to-zen/backups/")
	}

	// Display backups
	fmt.Println("\nAvailable backups:")
	fmt.Println(strings.Repeat("-", 60))
	for i, backup := range backups {
		fmt.Printf("[%d] %s (%s)\n", i+1, backup.Name, backup.Timestamp.Format("Mon Jan 2, 2006 at 3:04 PM"))
	}
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("[0] Cancel\n\n")

	// Prompt for selection
	fmt.Print("Select backup to restore: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	var selection int
	if _, err := fmt.Sscanf(input, "%d", &selection); err != nil {
		return fmt.Errorf("invalid selection")
	}

	if selection == 0 {
		fmt.Println("Restore cancelled.")
		return nil
	}

	if selection < 1 || selection > len(backups) {
		return fmt.Errorf("invalid selection: must be between 0 and %d", len(backups))
	}

	// Restore the selected backup
	selectedBackup := backups[selection-1]
	sessionPath := filepath.Join(profilePath, sessionFileName)

	// Create a backup of the current state before restoring
	fmt.Println("\nCreating backup of current state before restore...")
	if err := CreateBackup(profilePath); err != nil {
		fmt.Printf("Warning: failed to backup current state: %v\n", err)
		fmt.Print("Continue anyway? (y/N): ")
		confirm, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
			fmt.Println("Restore cancelled.")
			return nil
		}
	}

	// Copy backup to session path
	if err := copyFile(selectedBackup.Path, sessionPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	fmt.Printf("\n✓ Successfully restored backup: %s\n", selectedBackup.Name)
	fmt.Printf("  Restored to: %s\n", sessionPath)
	return nil
}

// parseBackupTimestamp extracts the timestamp from a backup filename
func parseBackupTimestamp(filename string) (time.Time, error) {
	// Format: zen-sessions_2006-01-02_15-04-05.jsonlz4
	parts := strings.Split(filename, "_")
	if len(parts) < 3 {
		return time.Time{}, fmt.Errorf("invalid backup filename format")
	}

	// Extract date and time parts
	datePart := parts[1]
	timePart := strings.TrimSuffix(parts[2], ".jsonlz4")
	timestampStr := fmt.Sprintf("%s_%s", datePart, timePart)

	return time.Parse(backupTimeFormat, timestampStr)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(dst, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}
