# Messages

## List messages

```bash
campfire messages list --room-id 5
campfire messages list --room-id 5 --limit 100
campfire messages list --room-id 5 --after 42
campfire messages list --room-id 5 --before 100 --limit 20
campfire messages list --room-id 5 --json
```

| Flag | Required | Notes |
|------|----------|-------|
| `--room-id` | yes | Room to list messages from |
| `--after` | no | Only messages after this message ID (for pagination) |
| `--before` | no | Only messages before this message ID |
| `--limit` | no | Max messages to return (default 50, max 200) |

Output columns: ID, FROM, BODY, TIME

Messages are returned in chronological order. Use `--after` with the last message ID to paginate forward.

## Send a message

```bash
campfire messages create --room-id 5 --body "Hello everyone!"
campfire messages create --room-id 5 --body-file ./announcement.md
campfire messages create --room-id 5 --body "Deploy complete ✅" --json
```

| Flag | Required | Notes |
|------|----------|-------|
| `--room-id` | yes | Room to post in |
| `--body` | no* | Message text |
| `--body-file` | no* | Read message body from a file |

*One of `--body` or `--body-file` is required. Use `--body-file` for anything longer than a sentence — it reads the file contents as the message body.

## View context around a message

```bash
campfire messages near 42 --room-id 5
campfire messages near 42 --room-id 5 --limit 10
campfire messages near 42 --room-id 5 --json
```

| Flag | Required | Notes |
|------|----------|-------|
| `--room-id` | yes | Room the message belongs to |
| `--limit` | no | Number of messages on each side (default 5, so up to 11 total) |

Shows symmetric context around a message — N messages before + the target + N messages after. The target message is marked with `> ` in table output. This is the critical bridge between a search hit and understanding the conversation.

## Delete a message

```bash
campfire messages delete 42 --room-id 5
campfire messages delete 42 --room-id 5 --force
```

| Flag | Required | Notes |
|------|----------|-------|
| `--room-id` | yes | Room the message belongs to |
| `--force` | no | Skip confirmation prompt |

Requires administrator access or message ownership (you authored it). This is a **hard delete** — the message is permanently destroyed and broadcast removal is sent to connected clients.

## Scripting patterns

**Send a file's contents as a message:**
```bash
campfire messages create --room-id 5 --body-file ./status-report.md
```

**Paginate through all messages:**
```bash
ROOM_ID=5
LAST_ID=0

while true; do
  MSGS=$(campfire messages list --room-id "$ROOM_ID" --after "$LAST_ID" --limit 200 --json)
  COUNT=$(echo "$MSGS" | jq '.data | length')
  [ "$COUNT" -eq 0 ] && break

  echo "$MSGS" | jq -r '.data[] | "\(.creator_name): \(.body)"'
  LAST_ID=$(echo "$MSGS" | jq -r '.data[-1].id')
done
```

**Post to multiple rooms:**
```bash
MESSAGE="Scheduled maintenance tonight at 10pm PST"
for ROOM_ID in 1 2 5 8; do
  campfire messages create --room-id "$ROOM_ID" --body "$MESSAGE"
done
```

**Get the latest message in a room:**
```bash
campfire messages list --room-id 5 --limit 1 --json | jq '.data[0]'
```

**Search hit → context → reply (agentic workflow):**
```bash
# Find a message via search, then view surrounding context
campfire messages near 42 --room-id 5 --json
# The breadcrumbs in each result suggest ready-made reply and boost commands
```
