package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Stream real-time events from your rooms",
	Run:   runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().String("room-id", "", "Only show events from this room")
}

func runWatch(cmd *cobra.Command, args []string) {
	roomFilter, _ := cmd.Flags().GetString("room-id")

	c := newClient()

	// Handle Ctrl-C gracefully
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		fmt.Fprintln(os.Stderr, "\nDisconnected.")
		os.Exit(0)
	}()

	if !jsonOutput && !markdownOutput {
		fmt.Fprintln(os.Stderr, "Watching for messages... (Ctrl-C to stop)")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	seen := make(map[string]bool)

	err := c.Watch(func(raw []byte) {
		var event struct {
			Event     string `json:"event"`
			Mentioned bool   `json:"mentioned"`
			Message   struct {
				ID          int    `json:"id"`
				RoomID      int    `json:"room_id"`
				CreatorID   int    `json:"creator_id"`
				CreatorName string `json:"creator_name"`
				Body        string `json:"body"`
				CreatedAt   string `json:"created_at"`
			} `json:"message"`
			Room struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"room"`
		}
		if err := json.Unmarshal(raw, &event); err != nil {
			return
		}

		// Deduplicate (server may deliver via multiple streams)
		dedupKey := fmt.Sprintf("%s:%d", event.Event, event.Message.ID)
		if seen[dedupKey] {
			return
		}
		seen[dedupKey] = true

		// Apply room filter
		if roomFilter != "" && fmt.Sprintf("%d", event.Room.ID) != roomFilter {
			return
		}

		mid := fmt.Sprintf("%d", event.Message.ID)
		rid := fmt.Sprintf("%d", event.Room.ID)

		switch {
		case jsonOutput:
			watchOutputJSON(raw, event.Event, mid, rid)
		case markdownOutput:
			watchOutputMarkdown(event.Event, event.Mentioned, mid, rid, event.Room.Name, event.Message.CreatorName, event.Message.Body, event.Message.CreatedAt)
		default:
			switch event.Event {
			case "message_created":
				mention := ""
				if event.Mentioned {
					mention = " [MENTIONED]"
				}
				fmt.Fprintf(w, "[%s]\t%s\t%s:\t%s%s\n", event.Room.Name, event.Message.CreatedAt, event.Message.CreatorName, event.Message.Body, mention)
			case "message_removed":
				fmt.Fprintf(w, "[%s]\t%s\t\tMessage %d removed\n", event.Room.Name, event.Message.CreatedAt, event.Message.ID)
			}
			w.Flush()
		}
	})
	if err != nil {
		exitWithError("watch", err)
	}
}

func watchBreadcrumbs(event, messageID, roomID string) []Breadcrumb {
	switch event {
	case "message_created":
		return []Breadcrumb{
			{Action: "reply", Cmd: fmt.Sprintf("campfire messages create --room-id %s --body \"{your_reply}\"", roomID), Description: "Reply in this room"},
			{Action: "boost", Cmd: fmt.Sprintf("campfire boosts create --message-id %s --content \"{emoji}\"", messageID), Description: "React with emoji"},
			{Action: "view_context", Cmd: fmt.Sprintf("campfire messages near %s --room-id %s", messageID, roomID), Description: "Show surrounding messages"},
			{Action: "search_room", Cmd: fmt.Sprintf("campfire search --query \"{query}\" --room-id %s", roomID), Description: "Search this room"},
		}
	case "message_removed":
		return []Breadcrumb{
			{Action: "view_context", Cmd: fmt.Sprintf("campfire messages list --room-id %s --limit 10", roomID), Description: "See recent messages"},
		}
	}
	return nil
}

func watchOutputJSON(raw []byte, event, messageID, roomID string) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		fmt.Println(string(raw))
		return
	}
	parsed["breadcrumbs"] = watchBreadcrumbs(event, messageID, roomID)
	out, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		fmt.Println(string(raw))
		return
	}
	fmt.Println(string(out))
}

func watchOutputMarkdown(event string, mentioned bool, messageID, roomID, roomName, creator, body, createdAt string) {
	switch event {
	case "message_created":
		fmt.Printf("---\n")
		mentionTag := ""
		if mentioned {
			mentionTag = " (**mentioned you**)"
		}
		fmt.Printf("**%s** in **%s**%s (%s):\n\n", creator, roomName, mentionTag, createdAt)
		fmt.Printf("%s\n\n", body)
		for _, b := range watchBreadcrumbs(event, messageID, roomID) {
			fmt.Printf("- %s: `%s`\n", b.Description, b.Cmd)
		}
		fmt.Println()
	case "message_removed":
		fmt.Printf("---\n")
		fmt.Printf("~~Message %s removed from **%s**~~ (%s)\n\n", messageID, roomName, createdAt)
		for _, b := range watchBreadcrumbs(event, messageID, roomID) {
			fmt.Printf("- %s: `%s`\n", b.Description, b.Cmd)
		}
		fmt.Println()
	}
}
