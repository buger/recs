# Agent contract

Files are the database. Prefer `crm` CLI mutations for structured fields.

Use `--json` for machine output.

Canonical state lives in Markdown plus YAML frontmatter.

`.crm/index` is disposable. Rebuild with `crm index`.

Do not treat folder location as semantic status.

## Repository structure

- `crm.yaml` holds name, default port, and optional type schemas.
- `records/` holds canonical Markdown records.
- `boards/` holds matcher views.
- `inbox/` holds unclassified incoming files.
- `templates/` holds create templates.
- `.crm/` holds disposable index, cache, and locks.

## Record schema

Every record is a Markdown file. YAML frontmatter stores machine state. The Markdown body stores context.

Required machine fields are `id` and `type`. Other fields are open.

Stable ids look like `person_alice` or `company_acme`. Relations store those ids.

## Search and query

- `crm search "MCDC Solana"` scans frontmatter, body, and filename.
- `crm query 'type=grant status=preparing'` filters with `= != < > <= >= contains in`.

Both commands accept `--json`.

## Mutations

Use CLI mutations for structured fields:

```
crm set grant_x status applied
crm patch grant_x --set status=applied --if-version sha256:...
crm move grant_x grants applied
```

Direct Markdown editing remains valid. Run `crm validate` after a manual edit.

## Boards

Boards are views. A move updates the column field or `on_drop` sets. The file path stays.

## Validation

`crm validate` uses optional schemas in `crm.yaml`. No schema means skip.

## CLI versus files

Use the CLI when you need validation, atomic writes, or a version check. Edit files when you write narrative notes.

See README.md and SKILL.md.
