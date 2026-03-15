package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var presenceCmd = &cobra.Command{
	Use:   "presence",
	Short: "View who is online",
	Run:   runPresence,
}

func init() {
	rootCmd.AddCommand(presenceCmd)
}

func runPresence(cmd *cobra.Command, args []string) {
	c := newClient()
	body, err := c.ListPresence()
	if err != nil {
		exitWithError("fetching presence", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
		return
	}

	var users []struct {
		ID         int     `json:"id"`
		Name       string  `json:"name"`
		LastSeenAt *string `json:"last_seen_at"`
	}
	if err := json.Unmarshal(body, &users); err != nil {
		exitWithError("parsing response", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSTATUS\tLAST SEEN")
	for _, u := range users {
		status := "offline"
		lastSeen := "never"
		if u.LastSeenAt != nil {
			lastSeen = *u.LastSeenAt
			t, err := time.Parse(time.RFC3339, *u.LastSeenAt)
			if err == nil && time.Since(t) < 5*time.Minute {
				status = "online"
			}
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", u.ID, u.Name, status, lastSeen)
	}
	w.Flush()
}
