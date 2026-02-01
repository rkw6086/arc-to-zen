package profiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Profile represents a Zen browser profile
type Profile struct {
	Name    string
	Path    string
	Default bool
}

// ProfilesIni represents the profiles.ini structure
type ProfilesIni struct {
	Install    map[string]InstallSection
	Profile    []ProfileSection
	General    GeneralSection
}

type InstallSection struct {
	Default  string
	Locked   int
}

type ProfileSection struct {
	Name      string
	IsRelative int
	Path      string
	Default   int
}

type GeneralSection struct {
	StartWithLastProfile int
	Version              int
}

// DiscoverProfiles finds all Zen browser profiles on the system
func DiscoverProfiles() ([]Profile, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	// Zen profiles are stored at ~/Library/Application Support/zen/Profiles/
	zenDir := filepath.Join(homeDir, "Library", "Application Support", "zen")
	
	// Check if Zen directory exists
	if _, err := os.Stat(zenDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Zen browser directory not found at: %s (is Zen installed?)", zenDir)
	}

	profilesDir := filepath.Join(zenDir, "Profiles")
	
	// Check if Profiles directory exists
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Zen profiles directory not found at: %s", profilesDir)
	}

	// Read all profile directories
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, fmt.Errorf("could not read profiles directory: %w", err)
	}

	var profiles []Profile
	var defaultProfileName string

	// Try to read profiles.ini to determine default profile
	profilesIni := filepath.Join(zenDir, "profiles.ini")
	if data, err := os.ReadFile(profilesIni); err == nil {
		defaultProfileName = parseDefaultProfile(string(data))
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		profilePath := filepath.Join(profilesDir, entry.Name())
		
		// Verify it's a valid profile by checking for key files
		sessionFile := filepath.Join(profilePath, "zen-sessions.jsonlz4")
		if _, err := os.Stat(sessionFile); err == nil {
			// Extract profile name (remove hash prefix if present)
			name := entry.Name()
			parts := strings.SplitN(name, ".", 2)
			if len(parts) == 2 {
				name = parts[1]
			}

			isDefault := (entry.Name() == defaultProfileName) || strings.Contains(strings.ToLower(name), "default")
			
			profiles = append(profiles, Profile{
				Name:    name,
				Path:    profilePath,
				Default: isDefault,
			})
		}
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no valid Zen profiles found in: %s", profilesDir)
	}

	return profiles, nil
}

// GetDefaultProfile returns the default Zen profile
func GetDefaultProfile() (*Profile, error) {
	profiles, err := DiscoverProfiles()
	if err != nil {
		return nil, err
	}

	// Look for default profile
	for _, profile := range profiles {
		if profile.Default {
			return &profile, nil
		}
	}

	// If no default found, return the first profile
	return &profiles[0], nil
}

// parseDefaultProfile parses profiles.ini to find the default profile
func parseDefaultProfile(content string) string {
	lines := strings.Split(content, "\n")
	var currentSection string
	var defaultPath string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		// Look for Default=1 in Profile sections
		if strings.HasPrefix(currentSection, "Profile") {
			if strings.HasPrefix(line, "Default=1") || strings.HasPrefix(line, "Default=true") {
				// Found default profile, look for Path in this section
				for _, l := range lines {
					l = strings.TrimSpace(l)
					if strings.HasPrefix(l, "Path=") {
						defaultPath = strings.TrimPrefix(l, "Path=")
						return defaultPath
					}
				}
			}
			if strings.HasPrefix(line, "Path=") && defaultPath == "" {
				// Store path in case this is the default
				path := strings.TrimPrefix(line, "Path=")
				if currentSection == "Profile0" {
					defaultPath = path
				}
			}
		}

		// Look for Install section default
		if strings.HasPrefix(currentSection, "Install") {
			if strings.HasPrefix(line, "Default=") {
				defaultPath = strings.TrimPrefix(line, "Default=")
				if strings.Contains(defaultPath, "/") {
					// Extract just the profile directory name
					parts := strings.Split(defaultPath, "/")
					if len(parts) > 0 {
						defaultPath = parts[len(parts)-1]
					}
				}
			}
		}
	}

	return defaultPath
}

// ListProfiles returns a formatted list of profiles for display
func ListProfiles(profiles []Profile) string {
	var sb strings.Builder
	sb.WriteString("Available Zen profiles:\n\n")
	
	for i, profile := range profiles {
		defaultMarker := ""
		if profile.Default {
			defaultMarker = " (default)"
		}
		sb.WriteString(fmt.Sprintf("  %d. %s%s\n", i+1, profile.Name, defaultMarker))
		sb.WriteString(fmt.Sprintf("     Path: %s\n", profile.Path))
		sb.WriteString("\n")
	}

	return sb.String()
}
