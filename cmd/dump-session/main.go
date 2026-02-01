package main

import (
	"encoding/json"
	"fmt"
	"os"

	"arc-to-zen/mozlz4"
)

func main() {
	data, err := os.ReadFile("/Users/rkw6086/Library/Application Support/zen/Profiles/5dyj5mgm.Default (release)/zen-sessions.jsonlz4")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	decompressed, err := mozlz4.Decompress(data)
	if err != nil {
		fmt.Println("Error decompressing:", err)
		return
	}

	var session map[string]interface{}
	json.Unmarshal(decompressed, &session)

	// Find tabs that match the prevSiblingInfo IDs
	if tabs, ok := session["tabs"].([]interface{}); ok {
		fmt.Println("=== TABS WITH TIMESTAMP-FORMAT IDs ===")
		for _, t := range tabs {
			tab := t.(map[string]interface{})
			syncId, _ := tab["zenSyncId"].(string)
			// Look for timestamp-format IDs (not UUIDs)
			if len(syncId) > 0 && syncId[0] != '{' {
				fmt.Printf("  zenSyncId: %v, groupId: %v, zenIsEmpty: %v\n", syncId, tab["groupId"], tab["zenIsEmpty"])
			}
		}
		fmt.Println()
	}

	// Pretty print folders with parentId
	if folders, ok := session["folders"].([]interface{}); ok {
		fmt.Println("=== NESTED FOLDERS (with parentId) ===")
		for _, f := range folders {
			folder := f.(map[string]interface{})
			parentId := folder["parentId"]
			if parentId != nil {
				out, _ := json.MarshalIndent(folder, "", "  ")
				fmt.Println(string(out))
			}
		}
	}
}
