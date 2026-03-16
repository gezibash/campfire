package cmd

import (
	"fmt"
	"os"

	"campfire/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	jsonOutput     bool
	markdownOutput bool
	versionInfo    string
	rootCmd     = &cobra.Command{
		Use:   "campfire",
		Short: "CLI for Campfire chat",
		Long:  "Command-line interface for interacting with a Campfire instance.",
	}
)

func SetVersionInfo(version, commit, date string) {
	versionInfo = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
	rootCmd.Version = versionInfo
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&markdownOutput, "markdown", false, "Output in GitHub-Flavored Markdown")
}

func configDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error finding home directory:", err)
		os.Exit(1)
	}
	return home + "/.config/campfire"
}

func initConfig() {
	viper.AddConfigPath(configDir())
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.SetEnvPrefix("CAMPFIRE")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()
}

func getURL() string {
	url := viper.GetString("url")
	if url == "" {
		fmt.Fprintln(os.Stderr, "Error: Campfire URL not configured.")
		fmt.Fprintln(os.Stderr, "Run 'campfire login --url <URL>' or set CAMPFIRE_URL.")
		os.Exit(1)
	}
	return url
}

func getToken() string {
	token := viper.GetString("token")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: Not authenticated.")
		fmt.Fprintln(os.Stderr, "Run 'campfire login' or set CAMPFIRE_TOKEN.")
		os.Exit(1)
	}
	return token
}

func newClient() *client.Client {
	return client.New(getURL(), getToken())
}

// newUnauthClient creates a client without requiring a token (for login/join/first-run)
func newUnauthClient(url string) *client.Client {
	return client.New(url, "")
}

func exitWithError(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s: %v\n", msg, err)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	}
	os.Exit(1)
}

func getUserID() int {
	return viper.GetInt("user_id")
}

func saveConfig(url, token string, userID int) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	viper.Set("url", url)
	viper.Set("token", token)
	viper.Set("user_id", userID)
	return viper.WriteConfigAs(dir + "/config.toml")
}
