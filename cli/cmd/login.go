package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with a Campfire instance",
	Run:   runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("url", "", "Campfire instance URL")
	loginCmd.Flags().String("email", "", "Email address")
	loginCmd.Flags().String("password", "", "Password")
}

func runLogin(cmd *cobra.Command, args []string) {
	url, _ := cmd.Flags().GetString("url")
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")

	reader := bufio.NewReader(os.Stdin)

	if url == "" {
		url = viper.GetString("url")
	}
	if url == "" {
		fmt.Print("Campfire URL: ")
		url, _ = reader.ReadString('\n')
		url = strings.TrimSpace(url)
	}
	if url == "" {
		exitWithError("URL is required", nil)
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

	c := newUnauthClient(url)
	body, err := c.Login(email, password)
	if err != nil {
		exitWithError("login failed", err)
	}

	var result struct {
		User struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email_address"`
			Role  string `json:"role"`
		} `json:"user"`
		APIToken string `json:"api_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		exitWithError("parsing response", err)
	}

	if err := saveConfig(url, result.APIToken, result.User.ID); err != nil {
		exitWithError("saving config", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
	} else {
		fmt.Printf("Logged in as %s (%s)\n", result.User.Name, result.User.Email)
		fmt.Println("Config saved to ~/.config/campfire/config.toml")
	}
}
