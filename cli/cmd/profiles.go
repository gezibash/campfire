package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage login profiles",
}

var profilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Run:   runProfilesList,
}

var profilesSetDefaultCmd = &cobra.Command{
	Use:   "default <name>",
	Short: "Set the default profile",
	Args:  cobra.ExactArgs(1),
	Run:   runProfilesSetDefault,
}

func init() {
	rootCmd.AddCommand(profilesCmd)
	profilesCmd.AddCommand(profilesListCmd, profilesSetDefaultCmd)
}

func runProfilesList(cmd *cobra.Command, args []string) {
	profiles := viper.GetStringMap("profiles")
	if len(profiles) == 0 {
		fmt.Println("No profiles configured. Run 'campfire login' to create one.")
		return
	}

	current := activeProfile()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  \tPROFILE\tURL\tUSER ID")
	for name := range profiles {
		marker := " "
		if name == current {
			marker = "*"
		}
		url := viper.GetString("profiles." + name + ".url")
		userID := viper.GetInt("profiles." + name + ".user_id")
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", marker, name, url, userID)
	}
	w.Flush()
}

func runProfilesSetDefault(cmd *cobra.Command, args []string) {
	name := args[0]

	profiles := viper.GetStringMap("profiles")
	if _, ok := profiles[name]; !ok {
		exitWithError(fmt.Sprintf("profile %q not found", name), nil)
	}

	viper.Set("default_profile", name)
	if err := viper.WriteConfigAs(configDir() + "/config.toml"); err != nil {
		exitWithError("saving config", err)
	}

	fmt.Printf("Default profile set to %q\n", name)
}
