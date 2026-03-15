# Search

Search messages across all rooms you have access to, including DMs. Results are paginated with cursor-based navigation.

## Search messages

```bash
campfire search --query "deploy"
campfire search --query "bug fix" --room-id 5
campfire search --query "meeting" --limit 20
campfire search --query "deadline" --json
```

| Flag | Required | Notes |
|------|----------|-------|
| `--query` | yes | Search text (FTS5 with Porter stemming) |
| `--room-id` | no | Limit search to a specific room (including DMs) |
| `--after` | no | Only results after this message ID (for pagination) |
| `--before` | no | Only results before this message ID |
| `--limit` | no | Max results per page (default 50, max 200) |

Output columns: ID, ROOM, FROM, BODY, TIME

Results span all rooms you're a member of — open, closed, and direct messages. Each result includes the room name. After each page, a hint is printed showing the `--after` value to fetch the next page.

## Pagination

Results are cursor-based using message IDs. The CLI prints a hint after each page:

```
50 results. For more: --after 1234
```

To paginate forward:
```bash
campfire search --query "deploy" --after 1234
```

## JSON output

With `--json`, search results are wrapped in an envelope with `ok`, `summary`, and `breadcrumbs`:

```bash
campfire search --query "deploy" --json
```

```json
{
  "ok": true,
  "data": [
    {
      "id": 456, "room_id": 7, "body": "Anyone know when we deploy next?",
      "breadcrumbs": [
        {"action": "view_context", "cmd": "campfire messages near 456 --room-id 7", "description": "Show surrounding messages"},
        {"action": "reply", "cmd": "campfire messages create --room-id 7 --body \"{your_reply}\"", "description": "Reply in this room"},
        {"action": "boost", "cmd": "campfire boosts create --message-id 456 --content \"{emoji}\"", "description": "React with emoji"},
        {"action": "search_room", "cmd": "campfire search --query \"deploy\" --room-id 7", "description": "Narrow search to this room"}
      ]
    }
  ],
  "summary": "1 search results for \"deploy\"",
  "breadcrumbs": [
    {"action": "next_page", "cmd": "campfire search --query \"deploy\" --after 456", "description": "Load more results"}
  ]
}
```

- `ok` indicates success, `summary` describes the result
- `data` contains the results array, each item with `breadcrumbs`
- Top-level `breadcrumbs` has pagination (`next_page`)
- Each breadcrumb has `action`, `cmd`, and `description`
- Use `jq '.data[]'` to iterate results, `jq '.breadcrumbs'` for pagination

## Scripting patterns

**Search within a DM:**
```bash
DM_ID=$(campfire rooms direct --user-id 3 --json | jq -r '.data.id')
campfire search --query "lunch" --room-id "$DM_ID"
```

**Search and extract room/message pairs:**
```bash
campfire search --query "action item" --json | jq -r '.data[] | {room: .room_name, from: .creator_name, body}'
```

**Paginate through all matches:**
```bash
LAST_ID=0
while true; do
  RESULTS=$(campfire search --query "error" --after "$LAST_ID" --limit 200 --json)
  COUNT=$(echo "$RESULTS" | jq '.data | length')
  [ "$COUNT" -eq 0 ] && break

  echo "$RESULTS" | jq -r '.data[] | "\(.room_name) | \(.creator_name): \(.body)"'
  LAST_ID=$(echo "$RESULTS" | jq -r '.data[-1].id')
done
```

**Search hit → context → reply (agentic workflow):**
```bash
# Get the view_context command from the first search result
campfire search --query "deploy" --json | jq -r '.data[0].breadcrumbs[] | select(.action == "view_context") | .cmd'
```

**Count mentions of a topic:**
```bash
campfire search --query "kubernetes" --limit 200 --json | jq '.data | length'
```

**Find messages from a specific person:**
```bash
campfire search --query "deadline" --json | jq '.data[] | select(.creator_name == "Jane")'
```
