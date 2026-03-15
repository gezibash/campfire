package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var boostsCmd = &cobra.Command{
	Use:   "boosts",
	Short: "Manage boosts (reactions)",
}

var boostsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add a boost to a message",
	Run:   runBoostsCreate,
}

var boostsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Remove a boost",
	Args:  cobra.ExactArgs(1),
	Run:   runBoostsDelete,
}

func init() {
	rootCmd.AddCommand(boostsCmd)
	boostsCmd.AddCommand(boostsCreateCmd, boostsDeleteCmd)

	boostsCreateCmd.Flags().String("message-id", "", "Message ID to boost")
	boostsCreateCmd.Flags().String("content", "", "Boost content (emoji)")
	_ = boostsCreateCmd.MarkFlagRequired("message-id")
	_ = boostsCreateCmd.MarkFlagRequired("content")

	boostsDeleteCmd.Flags().String("message-id", "", "Message ID the boost belongs to")
	boostsDeleteCmd.Flags().Bool("force", false, "Skip confirmation")
	_ = boostsDeleteCmd.MarkFlagRequired("message-id")
}

func runBoostsCreate(cmd *cobra.Command, args []string) {
	messageID, _ := cmd.Flags().GetString("message-id")
	content, _ := cmd.Flags().GetString("content")

	c := newClient()
	body, err := c.CreateBoost(messageID, content)
	if err != nil {
		exitWithError("creating boost", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
		return
	}

	var boost struct {
		ID      int    `json:"id"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(body, &boost); err != nil {
		exitWithError("parsing response", err)
	}

	fmt.Printf("Boost added: %s (ID: %d)\n", boost.Content, boost.ID)
}

func runBoostsDelete(cmd *cobra.Command, args []string) {
	boostID := args[0]
	messageID, _ := cmd.Flags().GetString("message-id")
	force, _ := cmd.Flags().GetBool("force")

	if !force {
		fmt.Printf("Remove boost %s? [y/N] ", boostID)
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			fmt.Println("Cancelled.")
			return
		}
	}

	c := newClient()
	if err := c.DeleteBoost(messageID, boostID); err != nil {
		exitWithError("deleting boost", err)
	}

	fmt.Printf("Boost %s removed.\n", boostID)
}
