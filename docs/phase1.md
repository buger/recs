<!-- Documents: STK-REQ-260820-4255 STK-REQ-260821-QTPP STK-REQ-260821-NTWY SYS-REQ-260821-JYEJ SYS-REQ-260821-QF1J SYS-REQ-260821-AFPN SYS-REQ-260821-8FKR SW-REQ-260821-8C2C SW-REQ-260821-82BA SW-REQ-260821-AC3S SW-REQ-260821-FCGM INT-REQ-260821-8HAC INT-REQ-260821-MRGW INT-REQ-260821-BSH3 -->
# Phase 1 file-native CRM

This document explains Phase 1 behavior for the local `crm` binary.

## Stakeholders

- STK-REQ-260820-KAGT file_user: Markdown files stay the canonical store.
- STK-REQ-260820-V5ZD agent_operator: the CLI emits stable JSON.
- STK-REQ-260820-Y3Q4 web_operator: `crm serve` shows the same records.
- STK-REQ-260820-T8AZ maintainer: one local binary, no required Node or database.

## System behavior

- SYS-REQ-260820-KJ34 / SW-REQ-260820-MQF2: `crm init` writes the workspace layout.
- SYS-REQ-260820-9J7C / SW-REQ-260820-N02Y: create, show, and list records with stable ids.
- SYS-REQ-260820-2SQZ / SW-REQ-260820-Q3C4: `crm set` and `crm patch` update frontmatter atomically.
- SYS-REQ-260820-ZTC3 / SW-REQ-260820-6EVX: `crm query` filters with simple operators.
- SYS-REQ-260820-HJPH / SW-REQ-260820-X37F: `crm search` scans frontmatter, body, and filenames.
- SYS-REQ-260820-4628 / SW-REQ-260820-NBGR: boards are matcher views.
- SYS-REQ-260820-BVBE / SW-REQ-260820-EX7Q: `crm move` changes frontmatter only.
- SYS-REQ-260820-5C9D / SW-REQ-260820-ZKCV: `crm next` lists due actions.
- SYS-REQ-260820-DCG4 / SW-REQ-260820-D5WE: `crm triage` lists decision items.
- SYS-REQ-260820-YWV4 / SW-REQ-260820-8PMR: `crm validate` uses optional schemas.
- SYS-REQ-260820-Q8GR / SW-REQ-260820-BNR7: `crm index` rebuilds disposable cache.
- SYS-REQ-260820-0TQX / SW-REQ-260820-V48V: `crm context` assembles related records.
- SYS-REQ-260820-9W1S / SW-REQ-260820-8ZS7: `crm serve` binds localhost and serves the Kanban UI.
- SYS-REQ-260820-456X / SW-REQ-260820-NA06 / SW-REQ-260820-EJVT: YAML dashboards, gallery previews, and a 2x2 widget view.
- SYS-REQ-260820-PG9C / SW-REQ-260820-YB5C: `--json` emits stable machine output.
- SYS-REQ-260821-8FKR / SW-REQ-260821-FCGM / INT-REQ-260821-BSH3: help and structured errors are the agent interface.
- SYS-REQ-260820-7WT4 / SW-REQ-260820-9C5Z: status lives in frontmatter, not folders.

## Interfaces

- INT-REQ-260820-JC9M: CLI uses the shared application layer.
- INT-REQ-260820-AHKR: HTTP uses the same application layer as the CLI.
- INT-REQ-260820-NHBY: HTTP dashboard routes project through the same application layer.
- INT-REQ-260820-JRWN: board move writes frontmatter through the store.
- INT-REQ-260820-2JKK: index rebuild scans store records and does not mutate canonical files.
- SYS-REQ-260821-JYEJ / SW-REQ-260821-8C2C / INT-REQ-260821-8HAC: edit, delete, link, ingest, export, import, and git companion commands.
- SYS-REQ-260821-QF1J / SW-REQ-260821-82BA / INT-REQ-260821-MRGW: record view, editor, search page, board filters, attachments, and wikilinks.
- SYS-REQ-260821-AFPN / SW-REQ-260821-AC3S: Cosmopolitan APE wrapper (`make ape`).

## Run

```
go test ./...
go build -o crm ./cmd/crm
./crm init
./crm create grant --title "Demo" --set status=researching
./crm query 'type=grant' --json
./crm board grants
./crm dashboard
./crm serve
```

## Agent contract

- Agents discover commands from `crm --help`, `crm help <cmd>`, and structured JSON errors. The binary does not write AGENTS.md or SKILL.md.
- `crm inbox` lists unclassified records and files under `inbox/`.
- `crm create` applies `templates/<type>.md` when the body is empty.
- `crm set` and `crm patch` return structured `invalid_enum` errors when a schema enum rejects a value.

- `crm patch --if-version` returns structured `conflict` with expected_version and current_version.
- `crm set`/`crm patch` expand `now` to an RFC3339 timestamp.
- `crm triage` treats `deadline` as an overdue date in addition to `due` and `next_action.date`.
- `crm edit` applies `--body` / `--set`, or opens `$EDITOR` on a temp copy and writes back atomically.
- `crm delete` removes the record file.
- `crm link <id> <target> --relation <type>` writes `relations:` on the source record.
- `crm ingest [email] [file|-]` creates a record from provider-neutral JSON.
- `crm export --json|--csv` and `crm import <file.csv> [--type]` interchange workspace records.
- `crm diff`, `crm changed`, and `crm history <id>` delegate to git when a repo exists.
- `crm serve` record routes are `#/r/<id>` and `#/search`.
- `make ape` builds `crm.com` (APE wrapper + host Go blob). See docs/distribution.md.
