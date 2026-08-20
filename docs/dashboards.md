<!-- Documents: SYS-REQ-260820-456X SW-REQ-260820-NA06 SW-REQ-260820-EJVT INT-REQ-260820-NHBY -->

# File-native dashboards

Dashboards are views. Canonical state stays in Markdown records and YAML configs.

## Files

`crm init` writes `dashboards/`:

- `prospects.yaml` — Prospect CRM metrics card
- `workspace.yaml` — list, count, board, and notes
- `inbox.yaml` — inbox list and watch list

Add a dashboard by writing `dashboards/<id>.yaml`. Agents write YAML. There is no in-app chat.

```yaml
id: workspace
name: Workspace
layout: 2x2
theme: light
widgets:
  - id: grants
    type: list
    title: Grant pipeline
    query: 'type=grant'
```

## Widget types

- `count` — numeric rollup from a query
- `list` — grouped records with status pills
- `notes` — title, body excerpt, owner
- `watch` — name, blurb, action pill
- `pipeline` — numbered items and optional progress
- `metrics` — big stats and overdue reminders
- `board` — embed an existing kanban board
- `markdown` — render a workspace `.md` file

Widgets get data from `query:`, `source:`, or `board:`.

`layout: 2x2` is the default. Missing slots render as empty placeholders.

## Commands

```
crm dashboard
crm dashboard workspace --json
crm dashboard new myboard --name "My board"
crm serve
```

Open `http://localhost:7777`. The home view is the gallery. Boards stay at `#/boards`.
