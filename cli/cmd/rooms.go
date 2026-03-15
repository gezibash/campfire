package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var roomsCmd = &cobra.Command{
	Use:   "rooms",
	Short: "Manage rooms",
}

var roomsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List rooms you belong to",
	Run:   runRoomsList,
}

var roomsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new room",
	Run:   runRoomsCreate,
}

var roomsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a room",
	Args:  cobra.ExactArgs(1),
	Run:   runRoomsUpdate,
}

var roomsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a room",
	Args:  cobra.ExactArgs(1),
	Run:   runRoomsDelete,
}

var roomsDirectCmd = &cobra.Command{
	Use:   "direct",
	Short: "Find or create a direct message room",
	Run:   runRoomsDirect,
}

func init() {
	rootCmd.AddCommand(roomsCmd)
	roomsCmd.AddCommand(roomsListCmd, roomsCreateCmd, roomsUpdateCmd, roomsDeleteCmd, roomsDirectCmd)

	roomsCreateCmd.Flags().String("name", "", "Room name")
	roomsCreateCmd.Flags().String("type", "open", "Room type (open or closed)")
	roomsCreateCmd.Flags().StringSlice("user-ids", nil, "User IDs for closed rooms")
	_ = roomsCreateCmd.MarkFlagRequired("name")

	roomsUpdateCmd.Flags().String("name", "", "New room name")
	roomsUpdateCmd.Flags().StringSlice("user-ids", nil, "User IDs (for closed rooms)")

	roomsDeleteCmd.Flags().Bool("force", false, "Skip confirmation")

	roomsDirectCmd.Flags().String("user-id", "", "User ID to start a DM with")
	_ = roomsDirectCmd.MarkFlagRequired("user-id")
}

type roomJSON struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Direct    bool   `json:"direct"`
	MemberIDs []int  `json:"member_ids"`
}

func runRoomsList(cmd *cobra.Command, args []string) {
	c := newClient()
	body, err := c.ListRooms()
	if err != nil {
		exitWithError("listing rooms", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
		return
	}

	var rooms []roomJSON
	if err := json.Unmarshal(body, &rooms); err != nil {
		exitWithError("parsing response", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTYPE\tDIRECT")
	for _, r := range rooms {
		typeName := friendlyType(r.Type)
		fmt.Fprintf(w, "%d\t%s\t%s\t%v\n", r.ID, r.Name, typeName, r.Direct)
	}
	w.Flush()
}

func runRoomsCreate(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	roomType, _ := cmd.Flags().GetString("type")
	userIDs, _ := cmd.Flags().GetStringSlice("user-ids")

	params := map[string]interface{}{
		"name": name,
		"type": roomType,
	}
	if len(userIDs) > 0 {
		params["user_ids"] = userIDs
	}

	c := newClient()
	body, err := c.CreateRoom(params)
	if err != nil {
		exitWithError("creating room", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
		return
	}

	var room roomJSON
	if err := json.Unmarshal(body, &room); err != nil {
		exitWithError("parsing response", err)
	}

	fmt.Printf("Created room \"%s\" (ID: %d, type: %s)\n", room.Name, room.ID, friendlyType(room.Type))
}

func runRoomsUpdate(cmd *cobra.Command, args []string) {
	id := args[0]
	params := map[string]interface{}{}

	if cmd.Flags().Changed("name") {
		name, _ := cmd.Flags().GetString("name")
		params["name"] = name
	}
	if cmd.Flags().Changed("user-ids") {
		userIDs, _ := cmd.Flags().GetStringSlice("user-ids")
		params["user_ids"] = userIDs
	}

	if len(params) == 0 {
		exitWithError("no fields to update (use --name or --user-ids)", nil)
	}

	c := newClient()
	body, err := c.UpdateRoom(id, params)
	if err != nil {
		exitWithError("updating room", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
		return
	}

	var room roomJSON
	if err := json.Unmarshal(body, &room); err != nil {
		exitWithError("parsing response", err)
	}

	fmt.Printf("Updated room \"%s\" (ID: %d)\n", room.Name, room.ID)
}

func runRoomsDelete(cmd *cobra.Command, args []string) {
	id := args[0]
	force, _ := cmd.Flags().GetBool("force")

	if !force {
		fmt.Printf("Delete room %s? [y/N] ", id)
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			fmt.Println("Cancelled.")
			return
		}
	}

	c := newClient()
	if err := c.DeleteRoom(id); err != nil {
		exitWithError("deleting room", err)
	}

	fmt.Printf("Room %s deleted.\n", id)
}

func runRoomsDirect(cmd *cobra.Command, args []string) {
	userID, _ := cmd.Flags().GetString("user-id")

	c := newClient()
	body, err := c.DirectRoom(userID)
	if err != nil {
		exitWithError("finding/creating DM", err)
	}

	if jsonOutput {
		fmt.Println(string(body))
		return
	}

	var room roomJSON
	if err := json.Unmarshal(body, &room); err != nil {
		exitWithError("parsing response", err)
	}

	fmt.Printf("Direct room \"%s\" (ID: %d)\n", room.Name, room.ID)
}

func friendlyType(t string) string {
	switch t {
	case "Rooms::Open":
		return "open"
	case "Rooms::Closed":
		return "closed"
	case "Rooms::Direct":
		return "direct"
	default:
		return t
	}
}
