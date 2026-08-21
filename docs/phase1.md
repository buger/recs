<!-- Documents: STK-REQ-260820-4255 STK-REQ-260821-QTPP STK-REQ-260821-NTWY SYS-REQ-260821-JYEJ SYS-REQ-260821-QF1J SYS-REQ-260821-8FKR SW-REQ-260821-8C2C SW-REQ-260821-82BA SW-REQ-260821-FCGM INT-REQ-260821-8HAC INT-REQ-260821-MRGW INT-REQ-260821-BSH3 INT-REQ-260821-5BJJ SW-REQ-260821-9657 SW-REQ-260821-9737 SW-REQ-260821-AY8F SW-REQ-260821-CR08 SW-REQ-260821-E5V8 SW-REQ-260821-MFR2 SW-REQ-260821-T9AY -->
# Phase 1 file-native CRM

This document explains Phase 1 behavior for the local `recs` binary.

## Stakeholders

- STK-REQ-260820-KAGT file_user: Markdown files stay the canonical store.
- STK-REQ-260820-V5ZD agent_operator: the CLI emits stable JSON.
- STK-REQ-260820-Y3Q4 web_operator: `recs serve` shows the same records.
- STK-REQ-260820-T8AZ maintainer: one local binary from go build, no required Node or database.

## System behavior

- SYS-REQ-260820-KJ34 / SW-REQ-260820-MQF2: `recs init` writes the workspace layout.
- SYS-REQ-260820-9J7C / SW-REQ-260820-N02Y: create, show, and list records with stable ids.
- SYS-REQ-260820-2SQZ / SW-REQ-260820-Q3C4: `recs set` and `recs patch` update frontmatter atomically.
- SYS-REQ-260820-ZTC3 / SW-REQ-260820-6EVX: `recs query` filters with simple operators.
- SYS-REQ-260820-HJPH / SW-REQ-260820-X37F: `recs search` scans frontmatter, body, and filenames.
- SYS-REQ-260820-4628 / SW-REQ-260820-NBGR: boards are matcher views.
- SYS-REQ-260820-BVBE / SW-REQ-260820-EX7Q: `recs move` changes frontmatter only.
- SYS-REQ-260820-5C9D / SW-REQ-260820-ZKCV: `recs next` lists due actions.
- SYS-REQ-260820-DCG4 / SW-REQ-260820-D5WE: `recs triage` lists decision items.
- SYS-REQ-260820-YWV4 / SW-REQ-260820-8PMR: `recs validate` uses optional schemas.
- SYS-REQ-260820-Q8GR / SW-REQ-260820-BNR7: `recs index` rebuilds disposable cache.
- SYS-REQ-260820-0TQX / SW-REQ-260820-V48V: `recs context` assembles related records.
- SYS-REQ-260820-9W1S / SW-REQ-260820-8ZS7: `recs serve` binds localhost and serves the Kanban UI.
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
- INT-REQ-260821-5BJJ: concurrent patches serialize through the store lock and report version conflicts.
- SYS-REQ-260821-JYEJ / SW-REQ-260821-8C2C / INT-REQ-260821-8HAC: edit, delete, link, ingest, export, import, and git companion commands.
- SYS-REQ-260821-QF1J / SW-REQ-260821-82BA / INT-REQ-260821-MRGW: record view, editor, search page, board filters, attachments, and wikilinks.

## Run

```
go test ./...
go build -o recs ./cmd/recs
./recs init
./recs create grant --title "Demo" --set status=researching
./recs query 'type=grant' --json
./recs board grants
./recs dashboard
./recs serve
```

## Agent contract

- Agents discover commands from `recs --help`, `recs help <cmd>`, and structured JSON errors. The binary does not write AGENTS.md or SKILL.md.
- `recs inbox` lists unclassified records and files under `inbox/`.
- `recs create` applies `templates/<type>.md` when the body is empty.
- `recs set` and `recs patch` return structured `invalid_enum` errors when a schema enum rejects a value.

- `recs patch --if-version` returns structured `conflict` with expected_version and current_version.
- `recs set`/`recs patch` expand `now` to an RFC3339 timestamp.
- `recs triage` treats `deadline` as an overdue date in addition to `due` and `next_action.date`.
- `recs edit` applies `--body` / `--set`, or opens `$EDITOR` on a temp copy and writes back atomically.
- `recs delete` removes the record file.
- `recs link <id> <target> --relation <type>` writes `relations:` on the source record.
- `recs ingest [email] [file|-]` creates a record from provider-neutral JSON.
- `recs export --json|--csv` and `recs import <file.csv> [--type]` interchange workspace records.
- `recs diff`, `recs changed`, and `recs history <id>` delegate to git when a repo exists.
- `recs serve` record routes are `#/r/<id>` and `#/search`.

## CLI contract

These software requirements refine the agent-facing CLI so help and JSON stay parseable without sidecar files.

- SW-REQ-260821-9657: empty JSON collections (`next.actions`, `triage.items`, board column records) encode as `[]`, never `null`, so agents can iterate recovery lists.
- SW-REQ-260821-9737: `--set` without `=` or with an empty key is rejected; `--set` with no value is rejected; `patch` without `--set` or `--body` is rejected; `search` with empty query text is rejected.
- SW-REQ-260821-AY8F: an unknown flag exits non-zero and sets JSON `error` to `unknown_flag`.
- SW-REQ-260821-CR08: every JSON timestamp uses RFC3339, not Go `Time.String`.
- SW-REQ-260821-E5V8: when `crm.yaml` is missing the message names `--root`; workspace-not-found and `not_found` JSON set `next` to a list or board command.
- SW-REQ-260821-MFR2: ingest with no file and a TTY prints usage and exits 1; stdin is read only when the operand is `-`.
- SW-REQ-260821-T9AY: empty `next` or `triage` prints one explicit empty-result line (`no next actions` / `no triage items`) instead of blank success.
