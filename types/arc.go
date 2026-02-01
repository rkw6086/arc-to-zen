package types

// ArcData represents the top-level Arc browser data structure
type ArcData struct {
	Sidebar *ArcSidebar `json:"sidebar"`
}

// ArcSidebar contains Arc's sidebar configuration
type ArcSidebar struct {
	Containers []*ArcContainer `json:"containers"`
}

// ArcContainer represents a container in Arc (spaces + items)
type ArcContainer struct {
	Spaces []interface{} `json:"spaces"` // Can be objects or strings
	Items  []interface{} `json:"items"`  // Can be objects or strings
}

// ArcSpace represents an Arc workspace/space
type ArcSpace struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	Icon         string        `json:"icon"`
	Color        string        `json:"color"`
	ContainerIDs []interface{} `json:"containerIDs"` // Can be strings or objects
	CustomInfo   *ArcCustomInfo `json:"customInfo"`
}

// ArcCustomInfo contains custom space configuration
type ArcCustomInfo struct {
	IconType *ArcIconType `json:"iconType"`
}

// ArcIconType contains icon information
type ArcIconType struct {
	Icon string `json:"icon"`
}

// ArcItem represents a tab or folder in Arc
type ArcItem struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	ParentID    string      `json:"parentID"`
	ChildrenIds []string    `json:"childrenIds"`
	Data        *ArcItemData `json:"data"`
}

// ArcItemData contains tab or container data
type ArcItemData struct {
	Tab           *ArcTab           `json:"tab"`
	ItemContainer *ArcItemContainer `json:"itemContainer"`
}

// ArcTab represents a browser tab
type ArcTab struct {
	SavedTitle string `json:"savedTitle"`
	SavedURL   string `json:"savedURL"`
}

// ArcItemContainer represents Arc internal containers (to be skipped)
type ArcItemContainer struct {
	ContainerType interface{} `json:"containerType"` // Can be string or object
}
