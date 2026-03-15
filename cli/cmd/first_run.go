package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/spf13/cobra"
)

var firstRunCmd = &cobra.Command{
	Use:   "first-run",
	Short: "Set up a new Campfire instance",
	Run:   runFirstRun,
}

func init() {
	rootCmd.AddCommand(firstRunCmd)
	firstRunCmd.Flags().String("url", "", "Campfire instance URL")
	firstRunCmd.Flags().String("name", "", "Admin display name")
	firstRunCmd.Flags().String("email", "", "Admin email address")
	firstRunCmd.Flags().String("password", "", "Admin password")
}

func runFirstRun(cmd *cobra.Command, args []string) {
	serverURL, _ := cmd.Flags().GetString("url")
	name, _ := cmd.Flags().GetString("name")
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")

	reader := bufio.NewReader(os.Stdin)

	if serverURL == "" {
		fmt.Print("Campfire URL: ")
		serverURL, _ = reader.ReadString('\n')
		serverURL = strings.TrimSpace(serverURL)
	}
	if serverURL == "" {
		exitWithError("URL is required", nil)
	}

	if name == "" {
		fmt.Print("Name: ")
		name, _ = reader.ReadString('\n')
		name = strings.TrimSpace(name)
	}
	if email == "" {
		fmt.Print("Email: ")
		email, _ = reader.ReadString('\n')
		email = strings.TrimSpace(email)
	}
	if password == "" {
		fmt.Print("Password: ")
		pw, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			exitWithError("reading password", err)
		}
		fmt.Println()
		password = string(pw)
	}

	c := newUnauthClient(serverURL)
	body, err := c.FirstRun(name, email, password)
	if err != nil {
		exitWithError("first run failed", err)
	}

	var result struct {
		User struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email_address"`
		} `json:"user"`
		APIToken string `json:"api_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		exitWithError("parsing response", err)
	}

	if err := saveConfig(serverURL, result.APIToken); err != nil {
		exitWithError("saving config", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
	} else {
		fmt.Printf("Instance created! Logged in as %s (%s)\n", result.User.Name, result.User.Email)
		fmt.Println("Config saved to ~/.config/campfire/config.toml")
	}
}
