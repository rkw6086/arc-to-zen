package importer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"arc-to-zen/mappings"
	"arc-to-zen/types"
)

// parseArcSpaces converts interface{} slice to ArcSpace slice
func parseArcSpaces(rawSpaces []interface{}) ([]*types.ArcSpace, error) {
	var spaces []*types.ArcSpace

	for _, raw := range rawSpaces {
		// Skip if not an object
		objMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if it has an ID (required for valid space)
		if _, hasID := objMap["id"]; !hasID {
			continue
		}

		// Marshal back to JSON and unmarshal to struct
		jsonData, err := json.Marshal(objMap)
		if err != nil {
			continue
		}

		var space types.ArcSpace
		if err := json.Unmarshal(jsonData, &space); err != nil {
			continue
		}

		spaces = append(spaces, &space)
	}

	return spaces, nil
}

// parseArcItems converts interface{} slice to ArcItem slice
func parseArcItems(rawItems []interface{}) ([]*types.ArcItem, error) {
	var items []*types.ArcItem

	for _, raw := range rawItems {
		// Skip if not an object
		objMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if it has an ID (required for valid item)
		if _, hasID := objMap["id"]; !hasID {
			continue
		}

		// Marshal back to JSON and unmarshal to struct
		jsonData, err := json.Marshal(objMap)
		if err != nil {
			continue
		}

		var item types.ArcItem
		if err := json.Unmarshal(jsonData, &item); err != nil {
			continue
		}

		items = append(items, &item)
	}

	return items, nil
}

// calculateNextContainerID finds the next available container ID
func calculateNextContainerID(containersData *types.ContainersData) int {
	// Start with lastUserContextId if available
	maxID := 0
	if containersData.LastUserContextID != nil {
		maxID = *containersData.LastUserContextID
	}
	
	// Also check all existing containers to be safe
	for _, container := range containersData.Identities {
		id := container.GetUserContextID()
		if id != maxUint32 && id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}

// findSpaceByName finds a space by its name
func findSpaceByName(spaces []types.ZenSpace, name string) *types.ZenSpace {
	for i := range spaces {
		if spaces[i].Name == name {
			return &spaces[i]
		}
	}
	return nil
}

// findContainerByName finds a container by its name (only returns containers with valid userContextId)
func findContainerByName(containers []types.ContainerIdentity, name string) *types.ContainerIdentity {
	for i := range containers {
		if containers[i].Name == name && containers[i].HasValidUserContextID() {
			return &containers[i]
		}
	}
	return nil
}

// updateContainer updates an existing container's properties
func updateContainer(containers []types.ContainerIdentity, containerID int, name, icon, color string) {
	for i := range containers {
		if containers[i].GetUserContextID() == containerID {
			containers[i].Name = name
			containers[i].Icon = mappings.MapArcIconToContainerIcon(icon)
			containers[i].Color = mappings.MapArcColorToZen(color)
			break
		}
	}
}

// filterTabs removes tabs for a specific workspace that are pinned
func filterTabs(tabs []types.ZenTab, workspaceUUID string) []types.ZenTab {
	var filtered []types.ZenTab
	for _, tab := range tabs {
		if tab.ZenWorkspace != workspaceUUID || !tab.Pinned {
			filtered = append(filtered, tab)
		}
	}
	return filtered
}

// filterFolders removes folders for a specific workspace
func filterFolders(folders []types.ZenFolder, workspaceID string) []types.ZenFolder {
	var filtered []types.ZenFolder
	for _, folder := range folders {
		if folder.WorkspaceID != workspaceID {
			filtered = append(filtered, folder)
		}
	}
	return filtered
}

// buildItemToSpaceMap builds a map of item ID to space ID
func buildItemToSpaceMap(spaces []*types.ArcSpace, itemsMap map[string]*types.ArcItem) map[string]string {
	itemToSpaceMap := make(map[string]string)

	for _, space := range spaces {
		// Get container IDs (can be mixed strings and objects)
		var containerIDs []string
		if space.ContainerIDs != nil {
			for _, raw := range space.ContainerIDs {
				if str, ok := raw.(string); ok {
					containerIDs = append(containerIDs, str)
				}
			}
		}

		// Mark direct children
		for _, itemID := range containerIDs {
			itemToSpaceMap[itemID] = space.ID

			// Mark all descendants
			markDescendants(itemID, space.ID, itemsMap, itemToSpaceMap)
		}
	}

	return itemToSpaceMap
}

// markDescendants recursively marks all descendants as belonging to a space
func markDescendants(parentID, spaceID string, itemsMap map[string]*types.ArcItem, itemToSpaceMap map[string]string) {
	parent := itemsMap[parentID]
	if parent == nil {
		return
	}

	for _, childID := range parent.ChildrenIds {
		itemToSpaceMap[childID] = spaceID
		markDescendants(childID, spaceID, itemsMap, itemToSpaceMap)
	}
}

// getProfileName extracts the profile name from an Arc space
// Returns "default" for the default profile, or the directoryBasename for custom profiles
func getProfileName(space *types.ArcSpace) string {
	if space.Profile == nil {
		return "default"
	}
	if space.Profile.Default != nil {
		return "default"
	}
	if space.Profile.Custom != nil && space.Profile.Custom.Data != nil {
		return space.Profile.Custom.Data.DirectoryBasename
	}
	return "default"
}

// ProfileInfo holds information about a unique profile
type ProfileInfo struct {
	Name        string // Profile directory name (e.g., "Profile 1")
	DisplayName string // First space that uses this profile (for container name)
	ContainerID int    // Zen container ID
	Icon        string // Icon from first space
	Color       string // Color from first space
}

// Firefox container colors
var containerColors = []string{"blue", "turquoise", "green", "yellow", "orange", "red", "pink", "purple"}

// collectUniqueProfiles finds all unique profiles and their first associated space
func collectUniqueProfiles(spaces []*types.ArcSpace) map[string]*ProfileInfo {
	profiles := make(map[string]*ProfileInfo)
	colorIndex := 0
	
	for _, space := range spaces {
		profileName := getProfileName(space)
		
		// Only record the first space that uses this profile
		if _, exists := profiles[profileName]; !exists {
			// Get icon
			icon := ""
			if space.CustomInfo != nil && space.CustomInfo.IconType != nil {
				icon = space.CustomInfo.IconType.Icon
			}
			if icon == "" {
				icon = space.Icon
			}
			
			displayName := space.Title
			if displayName == "" {
				displayName = profileName
			}
			
			// Arc doesn't have simple color names, so we rotate through colors
			color := containerColors[colorIndex%len(containerColors)]
			colorIndex++
			
			profiles[profileName] = &ProfileInfo{
				Name:        profileName,
				DisplayName: displayName,
				Icon:        icon,
				Color:       color,
			}
		}
	}
	
	return profiles
}

// filterArcContainers filters out Arc internal containers
func filterArcContainers(items []*types.ArcItem) []*types.ArcItem {
	var filtered []*types.ArcItem
	for _, item := range items {
		isArcContainer := item.Data != nil && item.Data.ItemContainer != nil && item.Data.ItemContainer.ContainerType != nil
		if !isArcContainer {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// getRootItemsForSpace gets the root-level items for a space
func getRootItemsForSpace(space *types.ArcSpace, itemsMap map[string]*types.ArcItem) []*types.ArcItem {
	var rootItems []*types.ArcItem

	if space.ContainerIDs == nil {
		return rootItems
	}

	// Arc's containerIDs contains item IDs directly (and some string markers like "pinned"/"unpinned")
	// Just try to look up each ID in the items map
	for _, raw := range space.ContainerIDs {
		// Can be string or object
		var itemID string
		switch v := raw.(type) {
		case string:
			itemID = v
		case map[string]interface{}:
			// Skip non-string entries
			continue
		default:
			continue
		}

		// Try to look up this ID in the items map
		if item := itemsMap[itemID]; item != nil {
			rootItems = append(rootItems, item)
		}
		// If not found (e.g. "pinned", "unpinned" markers), just skip it
	}
	return rootItems
}

// getIconOrDefault returns icon or default if empty
func getIconOrDefault(icon, defaultIcon string) string {
	if icon == "" {
		return defaultIcon
	}
	return icon
}

// getTitleOrDefault returns title or default if empty
func getTitleOrDefault(title, defaultTitle string) string {
	if title == "" {
		return defaultTitle
	}
	return title
}

// insertItemWithChildren recursively inserts an item and its children
func (imp *Importer) insertItemWithChildren(
	arcItem *types.ArcItem,
	parentFolderID string,
	spaceID string,
	spaceUUIDMap map[string]string,
	space *types.ArcSpace,
	itemsMap map[string]*types.ArcItem,
	arcToZenUUIDMap map[string]string,
	containersData *types.ContainersData,
	zenSession *types.ZenSession,
	now int64,
	level int,
	lastFolderByParent map[string]string, // Tracks the last folder created for each parent (for sibling references)
) int {
	indent := strings.Repeat("  ", level)
	itemsCreated := 0

	// Skip Arc containers but process their children
	isArcContainer := arcItem.Data != nil && arcItem.Data.ItemContainer != nil && arcItem.Data.ItemContainer.ContainerType != nil
	if isArcContainer {
		imp.logger.Info("%sSkipping Arc container \"%s\"", indent, getTitleOrDefault(arcItem.Title, arcItem.ID))
		// Arc containers at root level - process in normal order
		for _, childID := range arcItem.ChildrenIds {
			if child := itemsMap[childID]; child != nil {
			itemsCreated += imp.insertItemWithChildren(
					child, parentFolderID, spaceID, spaceUUIDMap, space,
					itemsMap, arcToZenUUIDMap, containersData, zenSession, now, level,
					lastFolderByParent,
				)
			}
		}
		return itemsCreated
	}

	// Get space info
	workspaceUUID := spaceUUIDMap[spaceID]
	spaceName := space.Title
	if spaceName == "" {
		spaceName = fmt.Sprintf("Workspace %s", spaceID)
	}

	// Find container ID
	containerID := 0
	for _, container := range containersData.Identities {
		if container.Name == spaceName && container.HasValidUserContextID() {
			containerID = container.GetUserContextID()
			break
		}
	}

	zenUUID := arcToZenUUIDMap[arcItem.ID]
	isFolder := len(arcItem.ChildrenIds) > 0

	// Get title and URL
	title := arcItem.Title
	if title == "" && arcItem.Data != nil && arcItem.Data.Tab != nil {
		title = arcItem.Data.Tab.SavedTitle
	}
	if title == "" {
		title = "Untitled"
	}

	url := ""
	if arcItem.Data != nil && arcItem.Data.Tab != nil {
		url = arcItem.Data.Tab.SavedURL
	}

	if isFolder {
		if !imp.options.DryRun {
			imp.logger.Info("%sCreating \"%s\" (FOLDER)", indent, title)
		} else {
			imp.logger.Info("%s[DRY-RUN] Would create folder: \"%s\"", indent, title)
		}

		// Create folder
		folderID := fmt.Sprintf("%d-%d", now, len(zenSession.Folders))

		// Create an anchor tab for this folder - required for Firefox to create the tab-group.
		// Without at least one tab with groupId=folderID, no tab-group element is created,
		// and Zen can't restore the folder structure.
		anchorTabID := fmt.Sprintf("{%s}", uuid.New().String())
		anchorTab := types.ZenTab{
			Entries: []types.ZenTabEntry{{
				URL:                       "about:blank",
				Title:                     "",
				TriggeringPrincipalBase64: "eyIzIjp7fX0=",
			}},
			LastAccessed:            now,
			Pinned:                  true,
			Hidden:                  false,
			ZenWorkspace:            workspaceUUID,
			ZenSyncID:               anchorTabID,
			ZenEssential:            false,
			ZenDefaultUserContextID: containerID,
			ZenPinnedIcon:           nil,
			ZenIsEmpty:              true, // Mark as empty so Zen treats it as folder placeholder
			ZenHasStaticIcon:        false,
			ZenGlanceID:             nil,
			ZenIsGlance:             false,
			ZenStaticLabel:          "",
			ZenPinnedInitialState:   nil,
			SearchMode:              nil,
			UserContextID:           containerID,
			Attributes:              map[string]interface{}{},
			Index:                   len(zenSession.Tabs),
			UserTypedValue:          "",
			UserTypedClear:          0,
			Image:                   nil,
			GroupID:                 folderID, // Critical: links tab to folder's tab-group
		}
		zenSession.Tabs = append(zenSession.Tabs, anchorTab)

		// Determine prevSiblingInfo for nested folders
		// For nested folders, reference the previous sibling FOLDER (not tabs) to ensure
		// proper ordering. Folder IDs persist through Zen save/restore cycles, unlike empty tabs.
		var prevSiblingInfo interface{}
		if parentFolderID != "" {
			// This is a nested folder
			if prevFolderID, exists := lastFolderByParent[parentFolderID]; exists {
				// Reference the previous sibling folder
				prevSiblingInfo = map[string]interface{}{
					"type": "group",
					"id":   prevFolderID,
				}
			}
			// else: first nested folder - nil means "insert at start" (default case in Zen's restore)
		}
		// Record this folder as the last one for this parent
		lastFolderByParent[parentFolderID] = folderID

		// Add folder metadata
		zenSession.Folders = append(zenSession.Folders, types.ZenFolder{
			Pinned:            true,
			SplitViewGroup:    false,
			ID:                folderID,
			Name:              title,
			Collapsed:         true, // Import collapsed by default
			SaveOnWindowClose: true,
			ParentID:          parentFolderID,
			PrevSiblingInfo:   prevSiblingInfo, // Reference previous sibling folder (or nil for first)
			EmptyTabIDs:       []string{anchorTabID}, // Track the anchor tab so Zen marks it properly
			UserIcon:          "",
			WorkspaceID:       workspaceUUID,
		})

		// Add tab group entry
		zenSession.Groups = append(zenSession.Groups, types.ZenGroup{
			ID:        folderID,
			Name:      title,
			Color:     nil,
			Collapsed: true, // Import collapsed by default
			Pinned:    true,
			Essential: false,
			SplitView: false,
		})

		itemsCreated++

		// Process children in FORWARD order - with folder-based prevSiblingInfo chaining,
		// each folder references its predecessor, maintaining natural order
		for _, childID := range arcItem.ChildrenIds {
			if child := itemsMap[childID]; child != nil {
				itemsCreated += imp.insertItemWithChildren(
					child, folderID, spaceID, spaceUUIDMap, space,
					itemsMap, arcToZenUUIDMap, containersData, zenSession, now, level+1,
					lastFolderByParent,
				)
			}
		}
	} else {
		if !imp.options.DryRun {
			imp.logger.Info("%sCreating \"%s\" (TAB)", indent, title)
		} else {
			imp.logger.Info("%s[DRY-RUN] Would create tab: \"%s\" → %s", indent, title, url)
		}

		// Fetch favicon
		var faviconDataURL string
		if url != "" {
			faviconDataURL = imp.faviconFetcher.FetchAsDataURL(url)
			if faviconDataURL != "" && imp.options.Verbose {
				imp.logger.Info("%s  ✓ Fetched favicon", indent)
			}
		}

		// Create tab
		tabEntry := types.ZenTabEntry{
			URL:                      url,
			Title:                    title,
			TriggeringPrincipalBase64: "eyIzIjp7fX0=",
		}

		// Prepare image field (nil if no favicon)
		var imageField interface{}
		if faviconDataURL != "" {
			imageField = faviconDataURL
		} else {
			imageField = nil
		}

		tab := types.ZenTab{
			Entries:                 []types.ZenTabEntry{tabEntry},
			LastAccessed:            now,
			Pinned:                  true,
			Hidden:                  false,
			ZenWorkspace:            workspaceUUID,
			ZenSyncID:               zenUUID,
			ZenEssential:            false,
			ZenDefaultUserContextID: containerID,
			ZenPinnedIcon:           nil,
			ZenIsEmpty:              false,
			ZenHasStaticIcon:        false,
			ZenGlanceID:             nil,
			ZenIsGlance:             false,
			ZenStaticLabel:          title,
			ZenPinnedInitialState: map[string]interface{}{
				"entry": map[string]interface{}{
					"url":                        url,
					"title":                      title,
					"triggeringPrincipal_base64": "eyIzIjp7fX0=",
				},
				"image": imageField,
			},
			SearchMode:     nil,
			UserContextID:  containerID,
			Attributes:     map[string]interface{}{},
			Index:          len(zenSession.Tabs),
			UserTypedValue: "",
			UserTypedClear: 0,
			Image:          imageField,
		}

		// Only set groupId if in a folder
		if parentFolderID != "" {
			tab.GroupID = parentFolderID
		}

		zenSession.Tabs = append(zenSession.Tabs, tab)
		itemsCreated++
	}

	return itemsCreated
}

// collectAllURLs recursively collects all URLs from Arc items
func collectAllURLs(items []*types.ArcItem, itemsMap map[string]*types.ArcItem) []string {
	var urls []string
	seen := make(map[string]bool)

	var collect func(item *types.ArcItem)
	collect = func(item *types.ArcItem) {
		if item == nil {
			return
		}

		// Skip Arc containers
		isArcContainer := item.Data != nil && item.Data.ItemContainer != nil && item.Data.ItemContainer.ContainerType != nil
		if isArcContainer {
			// Process children of containers
			for _, childID := range item.ChildrenIds {
				if child := itemsMap[childID]; child != nil {
					collect(child)
				}
			}
			return
		}

		// Extract URL from tab data
		if item.Data != nil && item.Data.Tab != nil && item.Data.Tab.SavedURL != "" {
			url := item.Data.Tab.SavedURL
			if !seen[url] {
				urls = append(urls, url)
				seen[url] = true
			}
		}

		// Process children recursively
		for _, childID := range item.ChildrenIds {
			if child := itemsMap[childID]; child != nil {
				collect(child)
			}
		}
	}

	// Collect URLs from all items
	for _, item := range items {
		collect(item)
	}

	return urls
}
