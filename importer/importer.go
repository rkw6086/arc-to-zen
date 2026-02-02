package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"arc-to-zen/favicon"
	"arc-to-zen/mappings"
	"arc-to-zen/mozlz4"
	"arc-to-zen/types"
)

const maxUint32 = 4294967295

// ImportOptions configures the import behavior
type ImportOptions struct {
	DryRun  bool // If true, only show what would be imported
	Verbose bool // If true, show detailed output
}

// Importer handles Arc to Zen browser data import
type Importer struct {
	zenProfilePath  string
	logger          Logger
	options         ImportOptions
	faviconFetcher  *favicon.Fetcher
}

// Logger interface for custom logging
type Logger interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// defaultLogger implements Logger with standard output
type defaultLogger struct{}

func (l *defaultLogger) Info(format string, args ...interface{}) {
	fmt.Printf("[ARC-IMPORT] "+format+"\n", args...)
}

func (l *defaultLogger) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}

// New creates a new Importer
func New(zenProfilePath string, logger Logger) *Importer {
	return NewWithOptions(zenProfilePath, logger, ImportOptions{})
}

// NewWithOptions creates a new Importer with custom options
func NewWithOptions(zenProfilePath string, logger Logger, options ImportOptions) *Importer {
	if logger == nil {
		logger = &defaultLogger{}
	}
	return &Importer{
		zenProfilePath:  zenProfilePath,
		logger:          logger,
		options:         options,
		faviconFetcher:  favicon.New(),
	}
}

// ImportResult contains statistics about the import
type ImportResult struct {
	Success         bool
	SpacesCreated   int
	ItemsImported   int
	ContainersCount int
}

// Import performs the Arc to Zen import
func (imp *Importer) Import(arcDataPath string) (*ImportResult, error) {
	imp.logger.Info(strings.Repeat("=", 80))
	if imp.options.DryRun {
		imp.logger.Info("DRY-RUN MODE - NO CHANGES WILL BE MADE")
		imp.logger.Info(strings.Repeat("=", 80))
	}
	imp.logger.Info("STARTING ARC IMPORT")
	imp.logger.Info(strings.Repeat("=", 80))
	imp.logger.Info("Zen Profile: %s", imp.zenProfilePath)

	// Validate Zen profile
	if err := imp.validateZenProfile(); err != nil {
		return nil, err
	}

	// Read Arc data
	arcData, err := imp.readArcData(arcDataPath)
	if err != nil {
		return nil, err
	}

	// Read Zen session
	zenSession, err := imp.readZenSession()
	if err != nil {
		return nil, err
	}

	// Read containers
	containersData, err := imp.readContainers()
	if err != nil {
		return nil, err
	}

	// Perform import
	result, err := imp.doImport(arcData, zenSession, containersData)
	if err != nil {
		return nil, err
	}

	// Write back (skip in dry-run mode)
	if !imp.options.DryRun {
		if err := imp.writeContainers(containersData); err != nil {
			return nil, err
		}

		if err := imp.writeZenSession(zenSession); err != nil {
			return nil, err
		}
	} else {
		imp.logger.Info("")
		imp.logger.Info("[DRY-RUN] Skipping file writes")
	}

	imp.logger.Info("")
	imp.logger.Info(strings.Repeat("=", 80))
	if imp.options.DryRun {
		imp.logger.Info("🔍 DRY-RUN COMPLETE")
	} else {
		imp.logger.Info("🎉 IMPORT COMPLETE")
	}
	imp.logger.Info(strings.Repeat("=", 80))
	imp.logger.Info("")
	imp.logger.Info("Summary:")
	imp.logger.Info("  • Spaces to import: %d", result.SpacesCreated)
	imp.logger.Info("  • Items to import: %d", result.ItemsImported)
	imp.logger.Info("  • Containers to create/update: %d", result.ContainersCount)
	imp.logger.Info("")
	if imp.options.DryRun {
		imp.logger.Info("This was a dry-run. No changes were made.")
		imp.logger.Info("Run without -dry-run to perform the actual import.")
	} else {
		imp.logger.Info("Next steps:")
		imp.logger.Info("  1. Close Zen Browser if it's running")
		imp.logger.Info("  2. Restart Zen Browser")
		imp.logger.Info("  3. Your Arc spaces should now appear in Zen!")
	}
	imp.logger.Info("")

	return result, nil
}

func (imp *Importer) validateZenProfile() error {
	imp.logger.Info("Validating Zen profile path...")
	if _, err := os.Stat(imp.zenProfilePath); os.IsNotExist(err) {
		return fmt.Errorf("Zen profile directory not found: %s", imp.zenProfilePath)
	}
	imp.logger.Info("✓ Zen profile found")

	sessionPath := filepath.Join(imp.zenProfilePath, "zen-sessions.jsonlz4")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		imp.logger.Info("  Session file not found - will create fresh session")
	} else {
		imp.logger.Info("✓ Session file found")
	}

	return nil
}

func (imp *Importer) readArcData(arcDataPath string) (*types.ArcData, error) {
	imp.logger.Info("Reading Arc data from: %s", arcDataPath)

	data, err := os.ReadFile(arcDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Arc data: %w", err)
	}

	var arcData types.ArcData
	if err := json.Unmarshal(data, &arcData); err != nil {
		return nil, fmt.Errorf("failed to parse Arc data: %w", err)
	}

	return &arcData, nil
}

func (imp *Importer) readZenSession() (*types.ZenSession, error) {
	sessionPath := filepath.Join(imp.zenProfilePath, "zen-sessions.jsonlz4")
	imp.logger.Info("Reading Zen session from: %s", sessionPath)

	// Check if file exists
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		imp.logger.Info("✓ Creating fresh session (no existing session file)")
		return &types.ZenSession{
			Spaces:        []types.ZenSpace{},
			Tabs:          []types.ZenTab{},
			Folders:       []types.ZenFolder{},
			Groups:        []types.ZenGroup{},
			SplitViewData: []interface{}{},
			LastCollected: 0,
		}, nil
	}

	// Read and decompress
	compressedData, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	decompressedData, err := mozlz4.Decompress(compressedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress session: %w", err)
	}

	var session types.ZenSession
	if err := json.Unmarshal(decompressedData, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	imp.logger.Info("✓ Session loaded: %d spaces, %d tabs", len(session.Spaces), len(session.Tabs))

	return &session, nil
}

func (imp *Importer) readContainers() (*types.ContainersData, error) {
	containersPath := filepath.Join(imp.zenProfilePath, "containers.json")

	// Check if file exists
	if _, err := os.Stat(containersPath); os.IsNotExist(err) {
		// Create default containers data
		return &types.ContainersData{
			Version:    5,
			Identities: []types.ContainerIdentity{},
		}, nil
	}

	data, err := os.ReadFile(containersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read containers: %w", err)
	}

	var containersData types.ContainersData
	if err := json.Unmarshal(data, &containersData); err != nil {
		return nil, fmt.Errorf("failed to parse containers: %w", err)
	}

	return &containersData, nil
}

func (imp *Importer) writeContainers(data *types.ContainersData) error {
	containersPath := filepath.Join(imp.zenProfilePath, "containers.json")
	imp.logger.Info("Writing containers.json...")

	// Clean up invalid containers (those with null/missing userContextId)
	// Keep: internal containers (public=false), containers with valid userContextId
	var validIdentities []types.ContainerIdentity
	invalidCount := 0
	for _, container := range data.Identities {
		// Keep internal containers (public=false) regardless of userContextId
		if !container.Public {
			validIdentities = append(validIdentities, container)
			continue
		}
		// For public containers, require valid userContextId
		if container.HasValidUserContextID() {
			validIdentities = append(validIdentities, container)
		} else {
			invalidCount++
			imp.logger.Info("  Removing invalid container: %q (null/missing userContextId)", container.Name)
		}
	}
	if invalidCount > 0 {
		imp.logger.Info("  Cleaned up %d invalid containers", invalidCount)
	}
	data.Identities = validIdentities

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal containers: %w", err)
	}

	if err := os.WriteFile(containersPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write containers: %w", err)
	}

	imp.logger.Info("✓ Updated containers.json")
	return nil
}

func (imp *Importer) writeZenSession(session *types.ZenSession) error {
	sessionPath := filepath.Join(imp.zenProfilePath, "zen-sessions.jsonlz4")
	imp.logger.Info("Writing Zen session file...")

	// Create backup first
	if err := imp.backupSession(); err != nil {
		imp.logger.Error("Warning: failed to create backup: %v", err)
	}

	// Marshal to JSON (no indentation for compression)
	jsonData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Compress
	compressedData, err := mozlz4.Compress(jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress session: %w", err)
	}

	// Write
	if err := os.WriteFile(sessionPath, compressedData, 0644); err != nil {
		return fmt.Errorf("failed to write session: %w", err)
	}

	imp.logger.Info("✓ Session file written successfully")
	return nil
}

func (imp *Importer) backupSession() error {
	// Skip backup in dry-run mode
	if imp.options.DryRun {
		imp.logger.Info("[DRY-RUN] Would create backup")
		return nil
	}

	sessionPath := filepath.Join(imp.zenProfilePath, "zen-sessions.jsonlz4")
	backupDir := filepath.Join(imp.zenProfilePath, "zen-sessions-backup")

	// Check if session exists
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return nil // No session to backup
	}

	// Create backup directory
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	// Create timestamped backup
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("zen-sessions-%s.jsonlz4", timestamp))

	// Copy file
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return err
	}

	imp.logger.Info("✓ Session backed up to: %s", backupPath)
	return nil
}

func (imp *Importer) doImport(
	arcData *types.ArcData,
	zenSession *types.ZenSession,
	containersData *types.ContainersData,
) (*ImportResult, error) {
	// Extract main container (index 1)
	if arcData.Sidebar == nil || len(arcData.Sidebar.Containers) < 2 {
		return nil, fmt.Errorf("no main container found in Arc data")
	}

	mainContainer := arcData.Sidebar.Containers[1]

	// Parse spaces
	spaces, err := parseArcSpaces(mainContainer.Spaces)
	if err != nil {
		return nil, err
	}
	imp.logger.Info("Found %d Arc spaces", len(spaces))

	// Parse items
	items, err := parseArcItems(mainContainer.Items)
	if err != nil {
		return nil, err
	}
	imp.logger.Info("Found %d Arc items", len(items))

	// Collect unique profiles (Arc profiles map to Zen containers)
	// Multiple Arc spaces can share the same profile/container
	profiles := collectUniqueProfiles(spaces)
	imp.logger.Info("Found %d unique profiles", len(profiles))

	// Calculate next container ID
	nextContainerID := calculateNextContainerID(containersData)

	// Create containers for each unique profile
	for profileName, profile := range profiles {
		// Check if container already exists for this profile
		existingContainer := findContainerByName(containersData.Identities, profile.DisplayName)
		if existingContainer != nil {
			profile.ContainerID = existingContainer.GetUserContextID()
			if !imp.options.DryRun {
				imp.logger.Info("Reusing existing container \"%s\" for profile \"%s\" (ID: %d)", 
					profile.DisplayName, profileName, profile.ContainerID)
			} else {
				imp.logger.Info("[DRY-RUN] Would reuse container \"%s\" for profile \"%s\" (ID: %d)", 
					profile.DisplayName, profileName, profile.ContainerID)
			}
		} else {
			// Create new container for this profile
			profile.ContainerID = nextContainerID
			nextContainerID++

			containersData.Identities = append(containersData.Identities, types.ContainerIdentity{
				UserContextID: &profile.ContainerID,
				Name:          profile.DisplayName,
				Icon:          mappings.MapArcIconToContainerIcon(profile.Icon),
				Color:         mappings.MapArcColorToZen(profile.Color),
				Public:        true,
			})

			// Update lastUserContextId
			containersData.LastUserContextID = &profile.ContainerID

			if !imp.options.DryRun {
				imp.logger.Info("Created container \"%s\" for profile \"%s\" (ID: %d)", 
					profile.DisplayName, profileName, profile.ContainerID)
			} else {
				imp.logger.Info("[DRY-RUN] Would create container \"%s\" for profile \"%s\" (ID: %d)", 
					profile.DisplayName, profileName, profile.ContainerID)
			}
		}
	}

	// Map space IDs to UUIDs
	spaceUUIDMap := make(map[string]string)

	// Get max space position
	maxSpacePosition := 0
	for _, space := range zenSession.Spaces {
		if space.Position > maxSpacePosition {
			maxSpacePosition = space.Position
		}
	}
	nextSpacePosition := maxSpacePosition + 1000

	spacesCreated := 0

	// Process each space
	for _, space := range spaces {
		spaceName := space.Title
		if spaceName == "" {
			spaceName = fmt.Sprintf("Workspace %s", space.ID)
		}

		// Extract icon
		arcIcon := ""
		if space.CustomInfo != nil && space.CustomInfo.IconType != nil {
			arcIcon = space.CustomInfo.IconType.Icon
		}
		if arcIcon == "" {
			arcIcon = space.Icon
		}

		// Get the container ID from the space's profile
		profileName := getProfileName(space)
		profile := profiles[profileName]
		containerID := profile.ContainerID

		// Check if space exists
		var spaceUUID string
		existingSpace := findSpaceByName(zenSession.Spaces, spaceName)

		if existingSpace != nil {
			// Merge into existing
			spaceUUID = existingSpace.UUID

			if !imp.options.DryRun {
				imp.logger.Info("Merging into existing space \"%s\" (profile: %s, container: %d)", spaceName, profileName, containerID)
			} else {
				imp.logger.Info("[DRY-RUN] Would merge into existing space: \"%s\" (profile: %s)", spaceName, profileName)
			}

			// Delete old pins for this workspace
			zenSession.Tabs = filterTabs(zenSession.Tabs, spaceUUID)
			zenSession.Folders = filterFolders(zenSession.Folders, spaceUUID)

			// Update space icon and container
			for i := range zenSession.Spaces {
				if zenSession.Spaces[i].UUID == spaceUUID {
					zenSession.Spaces[i].Icon = mappings.MapArcIconToSvg(arcIcon)
					zenSession.Spaces[i].ContainerTabID = containerID
					break
				}
			}
		} else {
			// Create new space
			spaceUUID = fmt.Sprintf("{%s}", uuid.New().String())

			// Add new space using the profile's container
			zenSession.Spaces = append(zenSession.Spaces, types.ZenSpace{
				UUID:           spaceUUID,
				Name:           spaceName,
				Icon:           mappings.MapArcIconToSvg(arcIcon),
				ContainerTabID: containerID,
				Position:       nextSpacePosition,
				Theme: types.ZenTheme{
					Type:           "gradient",
					GradientColors: []interface{}{},
					Opacity:        0.5,
					Rotation:       nil,
					Texture:        nil,
				},
				HasCollapsedPinnedTabs: false,
			})

			if !imp.options.DryRun {
				imp.logger.Info("Created space \"%s\" (profile: %s, container: %d)", spaceName, profileName, containerID)
			} else {
				imp.logger.Info("[DRY-RUN] Would create space: \"%s\" (profile: %s)", spaceName, profileName)
			}
			nextSpacePosition += 1000
		}

		spaceUUIDMap[space.ID] = spaceUUID
		spacesCreated++
	}

	// Build item lookup map
	itemsMap := make(map[string]*types.ArcItem)
	for _, item := range items {
		itemsMap[item.ID] = item
	}

	// Build item-to-space mapping (for future use)
	// itemToSpaceMap := buildItemToSpaceMap(spaces, itemsMap)

	// Pre-cache favicons for all URLs
	allURLs := collectAllURLs(items, itemsMap)
	if len(allURLs) > 0 {
		imp.logger.Info("Pre-caching favicons for %d URLs...", len(allURLs))
		
		// Spinner characters for animation
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinIdx := 0
		
		// Progress callback with spinner
		progress := func(processed, total int) {
			fmt.Printf("\r%s Fetching favicons... %d/%d", spinner[spinIdx%len(spinner)], processed, total)
			spinIdx++
		}
		
		result := imp.faviconFetcher.PreCacheFaviconsWithProgress(allURLs, 10, progress)
		fmt.Print("\r") // Clear the spinner line
		imp.logger.Info("✓ Favicon pre-cache complete: %d cached, %d fetched, %d failed", 
			result.Cached, result.Fetched, result.Failed)
	} else {
		imp.logger.Info("No URLs to fetch favicons for")
	}

	// Filter out Arc internal containers
	itemsToProcess := filterArcContainers(items)

	// Pre-generate UUIDs for all items
	arcToZenUUIDMap := make(map[string]string)
	for _, item := range itemsToProcess {
		arcToZenUUIDMap[item.ID] = fmt.Sprintf("{%s}", uuid.New().String())
	}

	// Process items
	now := time.Now().UnixMilli()
	pinsCreated := 0

	// Track the last folder created for each parent (for sibling references in nested folders)
	lastFolderByParent := make(map[string]string)

	imp.logger.Info("Creating items...")
	for _, space := range spaces {
		spaceTitle := space.Title
		if spaceTitle == "" {
			spaceTitle = fmt.Sprintf("Workspace %s", space.ID)
		}
		imp.logger.Info("Processing space: \"%s\"", spaceTitle)

		// Get root items for this space
		rootItems := getRootItemsForSpace(space, itemsMap)
		imp.logger.Info("Found %d root items", len(rootItems))

		for _, rootItem := range rootItems {
			pinsCreated += imp.insertItemWithChildren(
				rootItem,
				"",
				space.ID,
				spaceUUIDMap,
				space,
				itemsMap,
				arcToZenUUIDMap,
				containersData,
				zenSession,
				now,
				0,
				lastFolderByParent,
			)
		}
	}

	return &ImportResult{
		Success:         true,
		SpacesCreated:   spacesCreated,
		ItemsImported:   pinsCreated,
		ContainersCount: len(containersData.Identities),
	}, nil
}

// Helper functions continue in next part...
