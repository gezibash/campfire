# Presence & Involvement

## View who's online

```bash
campfire presence
campfire presence --json
```

Output columns: ID, NAME, STATUS, LAST SEEN

Status is determined by last activity:
- **online** — active within the last 5 minutes
- **offline** — no recent activity, or never seen

## Update notification level

```bash
campfire involvement --room-id 5 --level mentions
campfire involvement --room-id 5 --level everything --json
```

| Flag | Required | Notes |
|------|----------|-------|
| `--room-id` | yes | Room to configure |
| `--level` | yes | Notification level |

Notification levels:
- **everything** — notify on all messages
- **mentions** — only notify on @mentions
- **nothing** — no notifications, room still visible
- **invisible** — no notifications, room hidden from sidebar

## Scripting patterns

**Check if a specific user is online:**
```bash
campfire presence --json | jq -r '.data[] | select(.name == "Jane") | .last_seen_at'
```

**List only online users:**
```bash
campfire presence --json | jq -r '[.data[] | select(.last_seen_at != null)] | map(select((.last_seen_at | fromdateiso8601) > (now - 300))) | .[].name'
```

**Mute all rooms except one:**
```bash
KEEP=5
campfire rooms list --json | jq -r '.data[].id' | while read ROOM_ID; do
  if [ "$ROOM_ID" != "$KEEP" ]; then
    campfire involvement --room-id "$ROOM_ID" --level nothing
  fi
done
```
