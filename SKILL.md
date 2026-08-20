# CRM skill

Use this skill to operate a file-native CRM workspace.

## Rules

1. Markdown files are the database. Prefer `crm` for structured field changes.
2. Use `--json` for machine output. Use `--md` only for `crm context`.
3. Status lives in YAML frontmatter. Do not move files to change a board column.
4. `.crm/index` is disposable. Rebuild it with `crm index`.
5. Direct file edits are allowed. Run `crm validate` after a manual edit.

## Layout

- `crm.yaml` — workspace config and optional type schemas
- `records/` — canonical Markdown records
- `boards/` — Kanban views over records
- `dashboards/` — YAML dashboard views
- `inbox/` — unclassified incoming files
- `templates/` — record templates
- `.crm/` — derived index, cache, and locks

## Commands

```
crm init
crm create <type> --title "..." --set k=v
crm show <id>
crm list [--type <type>]
crm search <query>
crm query 'type=grant status=preparing'
crm set <id> <field> <value>
crm patch <id> --set k=v [--if-version sha256:...]
crm board [name] [--filter k=v]
crm move <id> <board> <column>
crm inbox
crm next
crm triage
crm context <id> [--json|--md]
crm validate
crm index
crm serve [--port 7777]
crm agent install
```

## Mutations

Prefer `crm set` and `crm patch`. They write a temp file, rename it, and return `ok`, `changed`, and `version`.

On conflict:

```
{"ok": false, "error": "conflict"}
```

On a bad enum:

```
{"ok": false, "error": "invalid_enum", "field": "status", "allowed": ["preparing"]}
```

## Boards

A board is a matcher view. `crm move` updates frontmatter only.

## Dashboards

Write `dashboards/<id>.yaml`. `crm dashboard` lists them. `crm serve` shows the gallery.
