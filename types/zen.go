package types

// ZenSession represents the Zen browser session structure
type ZenSession struct {
	Spaces        []ZenSpace      `json:"spaces"`
	Tabs          []ZenTab        `json:"tabs"`
	Folders       []ZenFolder     `json:"folders"`
	Groups        []ZenGroup      `json:"groups"`
	SplitViewData []interface{}   `json:"splitViewData"`
	LastCollected int64           `json:"lastCollected"`
}

// ZenSpace represents a Zen workspace
type ZenSpace struct {
	UUID                   string    `json:"uuid"`
	Name                   string    `json:"name"`
	Icon                   string    `json:"icon"`
	ContainerTabID         int       `json:"containerTabId"`
	Position               int       `json:"position"`
	Theme                  ZenTheme  `json:"theme"`
	HasCollapsedPinnedTabs bool      `json:"hasCollapsedPinnedTabs"`
}

// ZenTheme represents workspace theme configuration
type ZenTheme struct {
	Type           string        `json:"type"`
	GradientColors []interface{} `json:"gradientColors"`
	Opacity        float64       `json:"opacity"`
	Rotation       interface{}   `json:"rotation"`
	Texture        interface{}   `json:"texture"`
}

// ZenTab represents a browser tab
type ZenTab struct {
	Entries                 []ZenTabEntry `json:"entries"`
	LastAccessed            int64         `json:"lastAccessed"`
	Pinned                  bool          `json:"pinned"`
	Hidden                  bool          `json:"hidden"`
	ZenWorkspace            string        `json:"zenWorkspace"`
	ZenSyncID               string        `json:"zenSyncId"`
	ZenEssential            bool          `json:"zenEssential"`
	ZenDefaultUserContextID interface{}   `json:"zenDefaultUserContextId"` // Can be int or string
	ZenPinnedIcon           interface{}   `json:"zenPinnedIcon"`
	ZenIsEmpty              bool          `json:"zenIsEmpty"`
	ZenHasStaticIcon        bool          `json:"zenHasStaticIcon"`
	ZenGlanceID             interface{}   `json:"zenGlanceId"`
	ZenIsGlance             bool          `json:"zenIsGlance"`
	ZenStaticLabel          string        `json:"zenStaticLabel,omitempty"`
	ZenPinnedInitialState   interface{}   `json:"_zenPinnedInitialState"`
	SearchMode              interface{}   `json:"searchMode"`
	UserContextID           int           `json:"userContextId"`
	Attributes              interface{}   `json:"attributes"`
	Index                   int           `json:"index"`
	UserTypedValue          string        `json:"userTypedValue"`
	UserTypedClear          int           `json:"userTypedClear"`
	Image                   interface{}   `json:"image"`
	GroupID                 string        `json:"groupId,omitempty"`
}

// ZenTabEntry represents a tab's history entry
type ZenTabEntry struct {
	URL                      string `json:"url"`
	Title                    string `json:"title"`
	TriggeringPrincipalBase64 string `json:"triggeringPrincipal_base64"`
}

// ZenFolder represents a pinned folder
type ZenFolder struct {
	Pinned             bool        `json:"pinned"`
	SplitViewGroup     bool        `json:"splitViewGroup"`
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	Collapsed          bool        `json:"collapsed"`
	SaveOnWindowClose  bool        `json:"saveOnWindowClose"`
	ParentID           string      `json:"parentId,omitempty"`
	PrevSiblingInfo    interface{} `json:"prevSiblingInfo"`
	EmptyTabIDs        []string    `json:"emptyTabIds"`
	UserIcon           string      `json:"userIcon"`
	WorkspaceID        string      `json:"workspaceId"`
}

// ZenGroup represents a tab group (Firefox requirement)
type ZenGroup struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Color     interface{} `json:"color"`
	Collapsed bool        `json:"collapsed"`
	Pinned    bool        `json:"pinned"`
	Essential bool        `json:"essential"`
	SplitView bool        `json:"splitView"`
}

// ContainersData represents the containers.json structure
type ContainersData struct {
	Version           int                 `json:"version"`
	LastUserContextID *int                `json:"lastUserContextId,omitempty"` // Tracks highest used ID
	Identities        []ContainerIdentity `json:"identities"`
}

// ContainerIdentity represents a single container
// Note: User-created containers have "name", built-in containers have "l10nId" for localization
type ContainerIdentity struct {
	UserContextID *int   `json:"userContextId,omitempty"` // Pointer to detect null vs 0
	Name          string `json:"name,omitempty"`          // User-defined name (not for built-in containers)
	Icon          string `json:"icon"`
	Color         string `json:"color"`
	Public        bool   `json:"public"`
	L10nID        string `json:"l10nId,omitempty"` // Localization ID for built-in containers (lowercase 'd'!)
	AccessKey     string `json:"accessKey,omitempty"`
}

// GetUserContextID returns the userContextId or 0 if nil
func (c *ContainerIdentity) GetUserContextID() int {
	if c.UserContextID == nil {
		return 0
	}
	return *c.UserContextID
}

// HasValidUserContextID returns true if the container has a valid (non-nil, non-zero) userContextId
func (c *ContainerIdentity) HasValidUserContextID() bool {
	return c.UserContextID != nil && *c.UserContextID > 0
}
