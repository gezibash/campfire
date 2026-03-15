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
	searchCmd.Flags().String("limit", "", "Max results (default 100, max 200)")
	_ = searchCmd.MarkFlagRequired("query")
}

func runSearch(cmd *cobra.Command, args []string) {
	query, _ := cmd.Flags().GetString("query")
	limit, _ := cmd.Flags().GetString("limit")

	c := newClient()
	body, err := c.Search(query, limit)
	if err != nil {
		exitWithError("searching", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
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
}
