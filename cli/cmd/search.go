package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search messages across rooms",
	Run:   runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().String("query", "", "Search query")
	searchCmd.Flags().String("room-id", "", "Limit search to a specific room")
	searchCmd.Flags().String("after", "", "Only results after this message ID")
	searchCmd.Flags().String("before", "", "Only results before this message ID")
	searchCmd.Flags().String("limit", "", "Max results per page (default 50, max 200)")
	_ = searchCmd.MarkFlagRequired("query")
}

func runSearch(cmd *cobra.Command, args []string) {
	query, _ := cmd.Flags().GetString("query")
	roomID, _ := cmd.Flags().GetString("room-id")
	after, _ := cmd.Flags().GetString("after")
	before, _ := cmd.Flags().GetString("before")
	limit, _ := cmd.Flags().GetString("limit")

	c := newClient()
	body, err := c.Search(query, map[string]string{
		"room_id": roomID,
		"after":   after,
		"before":  before,
		"limit":   limit,
	})
	if err != nil {
		exitWithError("searching", err)
	}

	count := countItems(body)
	summary := fmt.Sprintf("%d search results for %q", count, query)

	switch {
	case jsonOutput:
		outputList(body, summary, func(item map[string]interface{}) []Breadcrumb {
			id := itemStr(item, "id")
			rid := itemStr(item, "room_id")
			breadcrumbs := []Breadcrumb{
				{Action: "view_context", Cmd: fmt.Sprintf("campfire messages near %s --room-id %s", id, rid), Description: "Show surrounding messages"},
				{Action: "reply", Cmd: fmt.Sprintf("campfire messages create --room-id %s --body \"{your_reply}\"", rid), Description: "Reply in this room"},
				{Action: "boost", Cmd: fmt.Sprintf("campfire boosts create --message-id %s --content \"{emoji}\"", id), Description: "React with emoji"},
			}
			if roomID == "" {
				breadcrumbs = append(breadcrumbs, Breadcrumb{Action: "search_room", Cmd: fmt.Sprintf("campfire search --query %q --room-id %s", query, rid), Description: "Narrow search to this room"})
			}
			return breadcrumbs
		}, func(items []map[string]interface{}) []Breadcrumb {
			if len(items) == 0 {
				return nil
			}
			lastID := itemStr(items[len(items)-1], "id")
			cmd := fmt.Sprintf("campfire search --query %q --after %s", query, lastID)
			if roomID != "" {
				cmd += fmt.Sprintf(" --room-id %s", roomID)
			}
			return []Breadcrumb{{Action: "next_page", Cmd: cmd, Description: "Load more results"}}
		})
		return
	case markdownOutput:
		markdownList(body, summary, Columns{
			{"ID", "id"},
			{"ROOM", "room_name"},
			{"FROM", "creator_name"},
			{"BODY", "body"},
			{"TIME", "created_at"},
		})
		return
	}

	var results []struct {
		ID          int    `json:"id"`
		RoomName    string `json:"room_name"`
		CreatorName string `json:"creator_name"`
		Body        string `json:"body"`
		CreatedAt   string `json:"created_at"`
	}
	if err := json.Unmarshal(body, &results); err != nil {
		exitWithError("parsing response", err)
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tROOM\tFROM\tBODY\tTIME")
	for _, r := range results {
		bodyPreview := r.Body
		if len(bodyPreview) > 50 {
			bodyPreview = bodyPreview[:47] + "..."
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", r.ID, r.RoomName, r.CreatorName, bodyPreview, r.CreatedAt)
	}
	w.Flush()

	if len(results) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d results. For more: --after %d\n", len(results), results[len(results)-1].ID)
	}
}
