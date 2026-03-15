package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var messagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Manage messages",
}

var messagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List messages in a room",
	Run:   runMessagesList,
}

var messagesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Send a message to a room",
	Run:   runMessagesCreate,
}

var messagesNearCmd = &cobra.Command{
	Use:   "near <message-id>",
	Short: "Show messages around a specific message",
	Args:  cobra.ExactArgs(1),
	Run:   runMessagesNear,
}

var messagesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a message",
	Args:  cobra.ExactArgs(1),
	Run:   runMessagesDelete,
}

func init() {
	rootCmd.AddCommand(messagesCmd)
	messagesCmd.AddCommand(messagesListCmd, messagesCreateCmd, messagesNearCmd, messagesDeleteCmd)

	messagesListCmd.Flags().String("room-id", "", "Room ID")
	messagesListCmd.Flags().String("after", "", "Only messages after this ID")
	messagesListCmd.Flags().String("before", "", "Only messages before this ID")
	messagesListCmd.Flags().String("limit", "", "Max messages to return")
	_ = messagesListCmd.MarkFlagRequired("room-id")

	messagesCreateCmd.Flags().String("room-id", "", "Room ID")
	messagesCreateCmd.Flags().String("body", "", "Message body")
	messagesCreateCmd.Flags().String("body-file", "", "Read message body from file")
	_ = messagesCreateCmd.MarkFlagRequired("room-id")

	messagesNearCmd.Flags().String("room-id", "", "Room ID")
	messagesNearCmd.Flags().Int("limit", 5, "Number of messages on each side")
	_ = messagesNearCmd.MarkFlagRequired("room-id")

	messagesDeleteCmd.Flags().String("room-id", "", "Room ID")
	messagesDeleteCmd.Flags().Bool("force", false, "Skip confirmation")
	_ = messagesDeleteCmd.MarkFlagRequired("room-id")
}

func runMessagesList(cmd *cobra.Command, args []string) {
	roomID, _ := cmd.Flags().GetString("room-id")
	after, _ := cmd.Flags().GetString("after")
	before, _ := cmd.Flags().GetString("before")
	limit, _ := cmd.Flags().GetString("limit")

	c := newClient()
	body, err := c.ListMessages(roomID, map[string]string{
		"after":  after,
		"before": before,
		"limit":  limit,
	})
	if err != nil {
		exitWithError("listing messages", err)
	}

	count := countItems(body)
	summary := fmt.Sprintf("%d messages", count)

	switch {
	case jsonOutput:
		outputList(body, summary, func(item map[string]interface{}) []Breadcrumb {
			id := itemStr(item, "id")
			return []Breadcrumb{
				{Action: "reply", Cmd: fmt.Sprintf("campfire messages create --room-id %s --body \"{your_reply}\"", roomID), Description: "Reply in this room"},
				{Action: "boost", Cmd: fmt.Sprintf("campfire boosts create --message-id %s --content \"{emoji}\"", id), Description: "React with emoji"},
				{Action: "view_context", Cmd: fmt.Sprintf("campfire messages near %s --room-id %s", id, roomID), Description: "Show surrounding messages"},
			}
		}, func(items []map[string]interface{}) []Breadcrumb {
			if len(items) == 0 {
				return nil
			}
			lastID := itemStr(items[len(items)-1], "id")
			return []Breadcrumb{{Action: "next_page", Cmd: fmt.Sprintf("campfire messages list --room-id %s --after %s", roomID, lastID), Description: "Load more messages"}}
		})
		return
	case markdownOutput:
		markdownList(body, summary, Columns{
			{"ID", "id"},
			{"FROM", "creator_name"},
			{"BODY", "body"},
			{"TIME", "created_at"},
		})
		return
	}

	var messages []struct {
		ID          int    `json:"id"`
		CreatorName string `json:"creator_name"`
		Body        string `json:"body"`
		CreatedAt   string `json:"created_at"`
	}
	if err := json.Unmarshal(body, &messages); err != nil {
		exitWithError("parsing response", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tFROM\tBODY\tTIME")
	for _, m := range messages {
		// Truncate body for table display
		bodyPreview := m.Body
		if len(bodyPreview) > 60 {
			bodyPreview = bodyPreview[:57] + "..."
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", m.ID, m.CreatorName, bodyPreview, m.CreatedAt)
	}
	w.Flush()
}

func runMessagesCreate(cmd *cobra.Command, args []string) {
	roomID, _ := cmd.Flags().GetString("room-id")
	msgBody, _ := cmd.Flags().GetString("body")
	bodyFile, _ := cmd.Flags().GetString("body-file")

	if bodyFile != "" {
		data, err := os.ReadFile(bodyFile)
		if err != nil {
			exitWithError("reading body file", err)
		}
		msgBody = string(data)
	}
	if msgBody == "" {
		exitWithError("message body is required (use --body or --body-file)", nil)
	}

	c := newClient()
	body, err := c.CreateMessage(roomID, msgBody)
	if err != nil {
		exitWithError("creating message", err)
	}

	item, _ := parseSingleItem(body)
	id := itemStr(item, "id")
	summary := fmt.Sprintf("Message sent (ID: %s)", id)

	switch {
	case jsonOutput:
		outputSingle(body, summary, func(item map[string]interface{}) []Breadcrumb {
			mid := itemStr(item, "id")
			return []Breadcrumb{
				{Action: "view_context", Cmd: fmt.Sprintf("campfire messages near %s --room-id %s", mid, roomID), Description: "Show this message in context"},
			}
		})
		return
	case markdownOutput:
		markdownMutation(summary)
		return
	}

	var msg struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(body, &msg); err != nil {
		exitWithError("parsing response", err)
	}

	fmt.Printf("Message sent (ID: %d)\n", msg.ID)
}

func runMessagesNear(cmd *cobra.Command, args []string) {
	messageID := args[0]
	roomID, _ := cmd.Flags().GetString("room-id")
	limit, _ := cmd.Flags().GetInt("limit")
	limitStr := strconv.Itoa(limit)

	c := newClient()
	body, err := c.ListMessages(roomID, map[string]string{
		"around": messageID,
		"limit":  limitStr,
	})
	if err != nil {
		exitWithError("fetching context", err)
	}

	count := countItems(body)
	summary := fmt.Sprintf("%d messages around #%s", count, messageID)

	switch {
	case jsonOutput:
		outputList(body, summary, func(item map[string]interface{}) []Breadcrumb {
			id := itemStr(item, "id")
			return []Breadcrumb{
				{Action: "reply", Cmd: fmt.Sprintf("campfire messages create --room-id %s --body \"{your_reply}\"", roomID), Description: "Reply in this room"},
				{Action: "boost", Cmd: fmt.Sprintf("campfire boosts create --message-id %s --content \"{emoji}\"", id), Description: "React with emoji"},
			}
		}, func(items []map[string]interface{}) []Breadcrumb {
			return []Breadcrumb{
				{Action: "expand", Cmd: fmt.Sprintf("campfire messages near %s --room-id %s --limit %d", messageID, roomID, limit*2), Description: "Show more surrounding messages"},
			}
		})
		return
	case markdownOutput:
		markdownList(body, summary, Columns{
			{"ID", "id"},
			{"FROM", "creator_name"},
			{"BODY", "body"},
			{"TIME", "created_at"},
		})
		return
	}

	var messages []struct {
		ID          int    `json:"id"`
		CreatorName string `json:"creator_name"`
		Body        string `json:"body"`
		CreatedAt   string `json:"created_at"`
	}
	if err := json.Unmarshal(body, &messages); err != nil {
		exitWithError("parsing response", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  \tID\tFROM\tBODY\tTIME")
	for _, m := range messages {
		bodyPreview := m.Body
		if len(bodyPreview) > 60 {
			bodyPreview = bodyPreview[:57] + "..."
		}
		marker := " "
		if strconv.Itoa(m.ID) == messageID {
			marker = ">"
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n", marker, m.ID, m.CreatorName, bodyPreview, m.CreatedAt)
	}
	w.Flush()
}

func runMessagesDelete(cmd *cobra.Command, args []string) {
	messageID := args[0]
	roomID, _ := cmd.Flags().GetString("room-id")
	force, _ := cmd.Flags().GetBool("force")

	if !force {
		fmt.Printf("Delete message %s? [y/N] ", messageID)
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			fmt.Println("Cancelled.")
			return
		}
	}

	c := newClient()
	if err := c.DeleteMessage(roomID, messageID); err != nil {
		exitWithError("deleting message", err)
	}

	fmt.Printf("Message %s deleted.\n", messageID)
}
