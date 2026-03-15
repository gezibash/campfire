---
name: campfire
description: Manage Campfire via the CLI — send messages, create rooms, react with boosts, search chat history, check who's online. Use this skill whenever the user mentions Campfire, wants to send or read messages, create or manage rooms, react to messages, search conversations, check presence, or manage notification levels. Also triggers for "the chat app", "group chat", "DMs", "direct messages", messaging workflows, or automating chat operations.
---

# Campfire CLI

Campfire is a self-hosted group chat app. The `campfire` CLI manages rooms, messages, boosts, users, and presence from the terminal.

## Reference files

Read these when you need exact flags, examples, or details for a specific resource:

- `references/auth.md` — login, join, first-run, env vars, agentic usage
- `references/rooms.md` — list, create, update, delete rooms, direct messages
- `references/messages.md` — list, send, delete, **near** (context around a message), bulk messaging
- `references/boosts.md` — add and remove emoji reactions
- `references/users.md` — list and create users
- `references/search.md` — search messages across rooms
- `references/presence.md` — check who's online, notification levels

## Quick orientation

**Rooms** are where conversations happen. Three types:
- **Open** — visible to everyone
- **Closed** — invite-only, with explicit member lists
- **Direct** — 1:1 private conversations

**Messages** belong to a room. **Boosts** are emoji reactions on messages. **Users** have roles (member, administrator) and presence (online/away/offline).

## Setup

Check if the CLI is already configured:
```bash
cat ~/.config/campfire/config.toml 2>/dev/null
```

If it has `url` and `token`, you're ready. Otherwise:

**Join via invite link** (creates account + saves token in one step):
```bash
campfire join http://localhost:3000/join/gFoO-0Lkb-UcFa \
  --name "Agent" --email agent@example.com --password secret123
```

**Login with existing account:**
```bash
campfire login --url http://localhost:3000 --email user@example.com --password secret
```

See `references/auth.md` for env vars, agentic patterns, and first-run setup.

All commands accept `--json` for machine-readable output and `--markdown` for GitHub-Flavored Markdown tables.

## End-to-end example: set up a project channel and post updates

```bash
# Create a closed room for the team
campfire rooms create --name "Project Alpha" --type closed --user-ids 1,2,3

# Send a message (using room ID from output)
campfire messages create --room-id 5 --body "Project Alpha channel is live! 🚀"

# React to a message
campfire boosts create --message-id 42 --content "🎉"

# Check who's around
campfire presence
```

## End-to-end example: daily standup bot

```bash
ROOM_ID=5

# Post standup prompt
campfire messages create --room-id "$ROOM_ID" \
  --body "Good morning! What's everyone working on today?"

# Later, search for responses
campfire search --query "working on" --json | jq '.data[] | {from: .creator_name, body}'
```

## Important patterns

**Find room IDs** — list your rooms first before sending messages:
```bash
campfire rooms list
campfire rooms list --json | jq '.data[] | {id, name, type}'
```

**Direct messages** — use `rooms direct` to find or create a DM:
```bash
campfire rooms direct --user-id 3
```

**Message body from file** — use `--body-file` for longer messages:
```bash
campfire messages create --room-id 5 --body-file ./announcement.md
```

**Notification levels** — control per-room notifications (invisible, nothing, mentions, everything):
```bash
campfire involvement --room-id 5 --level mentions
```

**Agentic breadcrumbs** — all `--json` output wraps results in `{"ok": true, "data": [...], "summary": "...", "breadcrumbs": [...]}`. Each item has `breadcrumbs` with copy-paste-ready next commands (reply, boost, view_context, etc.). Each breadcrumb has `action`, `cmd`, and `description`. Response-level `breadcrumbs` include pagination. The `summary` field gives a human-readable description of the result. Use `jq '.data[]'` to iterate and `jq '.breadcrumbs'` for next steps. Use `--markdown` for agent-friendly GFM tables.

**Context around a message** — `messages near` shows surrounding messages, bridging search hits to conversation understanding:
```bash
campfire messages near 42 --room-id 5
```

**Deletes require `--force`** to skip confirmation prompts, useful for scripting:
```bash
campfire messages delete 42 --room-id 5 --force
```

**Closed room membership** — pass `--user-ids` on create or update to control who can see the room:
```bash
campfire rooms create --name "Leadership" --type closed --user-ids 1,2,3
campfire rooms update 5 --user-ids 1,2,3,4
```
