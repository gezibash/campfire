# Rooms

## List rooms

```bash
campfire rooms list
campfire rooms list --json
```

Output columns: ID, NAME, TYPE, DIRECT

Lists all rooms you're a member of. Type values: `open`, `closed`, `direct`.

## Create a room

```bash
campfire rooms create --name "General"
campfire rooms create --name "Engineering" --type closed --user-ids 1,2,3
campfire rooms create --name "Announcements" --type open --json
```

| Flag | Required | Notes |
|------|----------|-------|
| `--name` | yes | Room name |
| `--type` | no | `open` (default) or `closed` |
| `--user-ids` | no | Comma-separated user IDs for closed rooms. Creator is always included |

**Open rooms** are visible to all users. **Closed rooms** require explicit membership — only listed users can see and post.

If the instance restricts room creation to administrators, non-admin users will get a 403 error.

## Update a room

```bash
campfire rooms update 5 --name "New Name"
campfire rooms update 5 --user-ids 1,2,3,4
campfire rooms update 5 --name "Renamed" --user-ids 1,2 --json
```

| Flag | Notes |
|------|-------|
| `--name` | New room name |
| `--user-ids` | Replace member list (closed rooms only). Users not in the new list are removed |

Only flags you pass are changed. Requires administrator access or room ownership.

For closed rooms, `--user-ids` does a full replacement — users in the current list but not in the new list are revoked.

## Delete a room

```bash
campfire rooms delete 5
campfire rooms delete 5 --force
```

Prompts for confirmation unless `--force` is passed. Requires administrator access or room ownership. This is a **hard delete** — the room and all its messages are permanently destroyed.

## Direct messages

```bash
campfire rooms direct --user-id 3
campfire rooms direct --user-id 3 --json
```

Finds an existing DM room with the specified user, or creates one if it doesn't exist. Returns the room details including its ID, which you can then use with `messages` commands.

## Scripting patterns

**Get room ID by name:**
```bash
ROOM_ID=$(campfire rooms list --json | jq -r '.data[] | select(.name == "General") | .id')
```

**Create a room and immediately post:**
```bash
ROOM_ID=$(campfire rooms create --name "Standup" --json | jq -r '.data.id')
campfire messages create --room-id "$ROOM_ID" --body "Room created! 👋"
```

**Find or create a DM and send a message:**
```bash
DM_ID=$(campfire rooms direct --user-id 3 --json | jq -r '.data.id')
campfire messages create --room-id "$DM_ID" --body "Hey, got a minute?"
```
