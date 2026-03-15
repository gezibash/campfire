package cmd

import (
	"encoding/json"
	"fmt"
	"os"
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

var messagesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a message",
	Args:  cobra.ExactArgs(1),
	Run:   runMessagesDelete,
}

func init() {
	rootCmd.AddCommand(messagesCmd)
	messagesCmd.AddCommand(messagesListCmd, messagesCreateCmd, messagesDeleteCmd)

	messagesListCmd.Flags().String("room-id", "", "Room ID")
	messagesListCmd.Flags().String("after", "", "Only messages after this ID")
	messagesListCmd.Flags().String("before", "", "Only messages before this ID")
	messagesListCmd.Flags().String("limit", "", "Max messages to return")
	_ = messagesListCmd.MarkFlagRequired("room-id")

	messagesCreateCmd.Flags().String("room-id", "", "Room ID")
	messagesCreateCmd.Flags().String("body", "", "Message body")
	messagesCreateCmd.Flags().String("body-file", "", "Read message body from file")
	_ = messagesCreateCmd.MarkFlagRequired("room-id")

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

	if jsonOutput {
		fmt.Println(string(body))
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

	if jsonOutput {
		fmt.Println(string(body))
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
