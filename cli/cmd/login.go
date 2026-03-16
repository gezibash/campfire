package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
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
	url = strings.TrimRight(url, "/")

	// If no email/password provided and stdin is a terminal, try browser auth
	if email == "" && password == "" && term.IsTerminal(int(os.Stdin.Fd())) {
		if browserLogin(url) {
			return
		}
		// Browser auth failed or was skipped — fall through to interactive prompts
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
		fmt.Printf("Config saved to ~/.config/campfire/config.toml [profile: %s]\n", activeProfile())
	}
}

// browserLogin starts a local callback server, opens the browser to the Campfire
// authorize endpoint, and waits for the redirect with the token.
func browserLogin(baseURL string) bool {
	// Start local server on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return false
	}
	port := listener.Addr().(*net.TCPAddr).Port

	type callbackResult struct {
		token  string
		userID string
		name   string
	}
	resultCh := make(chan callbackResult, 1)

	srv := &http.Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		userID := r.URL.Query().Get("user_id")
		name := r.URL.Query().Get("name")

		if token == "" {
			http.Error(w, "Missing token", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><h2>Logged in as %s</h2><p>You can close this tab and return to the terminal.</p></body></html>`, name)

		resultCh <- callbackResult{token: token, userID: userID, name: name}
	})
	srv.Handler = mux

	go srv.Serve(listener)
	defer srv.Shutdown(context.Background())

	authorizeURL := fmt.Sprintf("%s/cli/authorize?port=%d", baseURL, port)
	fmt.Printf("Opening browser to log in...\n")
	fmt.Printf("If the browser doesn't open, visit: %s\n", authorizeURL)

	openBrowser(authorizeURL)

	// Wait for callback
	result := <-resultCh

	userID := 0
	fmt.Sscanf(result.userID, "%d", &userID)

	if err := saveConfig(baseURL, result.token, userID); err != nil {
		exitWithError("saving config", err)
	}

	fmt.Printf("Logged in as %s\n", result.name)
	fmt.Printf("Config saved to ~/.config/campfire/config.toml [profile: %s]\n", activeProfile())
	return true
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}
