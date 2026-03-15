package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Run:   runUsersList,
}

var usersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user (admin only)",
	Run:   runUsersCreate,
}

func init() {
	rootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(usersListCmd, usersCreateCmd)

	usersCreateCmd.Flags().String("name", "", "User's display name")
	usersCreateCmd.Flags().String("email", "", "User's email address")
	usersCreateCmd.Flags().String("password", "", "User's password (auto-generated if omitted)")
	usersCreateCmd.Flags().String("role", "member", "User's role (member or administrator)")
	_ = usersCreateCmd.MarkFlagRequired("name")
	_ = usersCreateCmd.MarkFlagRequired("email")
}

func runUsersList(cmd *cobra.Command, args []string) {
	c := newClient()
	body, err := c.ListUsers()
	if err != nil {
		exitWithError("listing users", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
		return
	}

	var users []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email_address"`
		Role  string `json:"role"`
		Admin bool   `json:"admin"`
	}
	if err := json.Unmarshal(body, &users); err != nil {
		exitWithError("parsing response", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tROLE")
	for _, u := range users {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", u.ID, u.Name, u.Email, u.Role)
	}
	w.Flush()
}

func runUsersCreate(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")
	role, _ := cmd.Flags().GetString("role")

	params := map[string]interface{}{
		"name":          name,
		"email_address": email,
		"role":          role,
	}
	if password != "" {
		params["password"] = password
	}

	c := newClient()
	body, err := c.CreateUser(params)
	if err != nil {
		exitWithError("creating user", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
		return
	}

	var user struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email_address"`
		Password string `json:"password"`
		APIToken string `json:"api_token"`
	}
	if err := json.Unmarshal(body, &user); err != nil {
		exitWithError("parsing response", err)
	}

	fmt.Printf("Created user %s (ID: %d)\n", user.Name, user.ID)
	if user.Password != "" {
		fmt.Printf("Password: %s\n", user.Password)
	}
	fmt.Printf("API Token: %s\n", user.APIToken)
}
