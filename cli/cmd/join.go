package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/spf13/cobra"
)

var joinCmd = &cobra.Command{
	Use:   "join <invite-url>",
	Short: "Join a Campfire instance via invite URL",
	Args:  cobra.ExactArgs(1),
	Run:   runJoin,
}

func init() {
	rootCmd.AddCommand(joinCmd)
	joinCmd.Flags().String("name", "", "Your display name")
	joinCmd.Flags().String("email", "", "Email address")
	joinCmd.Flags().String("password", "", "Password")
}

func runJoin(cmd *cobra.Command, args []string) {
	inviteURL := args[0]
	name, _ := cmd.Flags().GetString("name")
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")

	// Parse invite URL: http://host:port/join/CODE
	parsed, err := url.Parse(inviteURL)
	if err != nil {
		exitWithError("parsing invite URL", err)
	}

	baseURL := parsed.Scheme + "://" + parsed.Host
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "join" {
		exitWithError("invalid invite URL format (expected .../join/<code>)", nil)
	}
	joinCode := parts[1]

	reader := bufio.NewReader(os.Stdin)

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

	c := newUnauthClient(baseURL)
	body, err := c.Join(joinCode, name, email, password)
	if err != nil {
		exitWithError("join failed", err)
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

	if err := saveConfig(baseURL, result.APIToken, result.User.ID); err != nil {
		exitWithError("saving config", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
	} else {
		fmt.Printf("Joined as %s (%s)\n", result.User.Name, result.User.Email)
		fmt.Printf("Config saved to ~/.config/campfire/config.toml [profile: %s]\n", activeProfile())
	}
}
