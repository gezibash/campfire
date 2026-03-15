package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var involvementCmd = &cobra.Command{
	Use:   "involvement",
	Short: "Update notification level for a room",
	Run:   runInvolvement,
}

func init() {
	rootCmd.AddCommand(involvementCmd)
	involvementCmd.Flags().String("room-id", "", "Room ID")
	involvementCmd.Flags().String("level", "", "Notification level (invisible, nothing, mentions, everything)")
	_ = involvementCmd.MarkFlagRequired("room-id")
	_ = involvementCmd.MarkFlagRequired("level")
}

func runInvolvement(cmd *cobra.Command, args []string) {
	roomID, _ := cmd.Flags().GetString("room-id")
	level, _ := cmd.Flags().GetString("level")

	c := newClient()
	body, err := c.UpdateInvolvement(roomID, level)
	if err != nil {
		exitWithError("updating involvement", err)
	}

	item, _ := parseSingleItem(body)
	involvement := itemStr(item, "involvement")
	summary := fmt.Sprintf("Notification level set to %s", involvement)

	switch {
	case jsonOutput:
		outputSingle(body, summary, func(item map[string]interface{}) []Breadcrumb {
			rid := itemStr(item, "room_id")
			return []Breadcrumb{
				{Action: "read", Cmd: fmt.Sprintf("campfire messages list --room-id %s", rid), Description: "Read messages in this room"},
			}
		})
		return
	case markdownOutput:
		markdownMutation(summary)
		return
	}

	var result struct {
		RoomID      int    `json:"room_id"`
		Involvement string `json:"involvement"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		exitWithError("parsing response", err)
	}

	fmt.Printf("Room %d notification level set to: %s\n", result.RoomID, result.Involvement)
}
