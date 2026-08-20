File-Native Agent CRM

1. Product idea

A lightweight CRM where plain Markdown files are the canonical database.

There is no required server, SaaS backend, PostgreSQL, SQLite, or proprietary storage format.

The CRM is designed primarily for:

AI agents such as Claude Code, Codex, Gemini CLI, etc.

humans working directly with files and Git

optional local visualization through a simple web interface

automation through a native CLI

distribution as a single Cosmopolitan APE executable

The core principle is:

Files are the database. The CLI provides deterministic operations. The agent provides intelligence. The web interface provides visualization.

A repository should remain fully understandable and usable even if the CRM executable disappears.

2. Core architecture

                    ┌──────────────────────┐                     │     AI Agent         │                     │ Claude / Codex / etc │                     └──────────┬───────────┘                                │                        CLI / filesystem                                │                     ┌──────────▼───────────┐                     │       crm.com        │                     │                      │                     │ query / edit / lint  │                     │ board / search       │                     │ ingest / serve       │                     └──────────┬───────────┘                                │              ┌─────────────────┼─────────────────┐              │                 │                 │              ▼                 ▼                 ▼         Markdown files     Derived indexes    Local Web UI         source of truth    disposable cache   visualization

Canonical state:

.crm/ crm.yaml  records/ boards/ inbox/ attachments/

Generated state:

.crm/ ├── index/ ├── cache/ └── runtime/

Everything under .crm/ should be rebuildable.

3. Distribution

The application should be built using Cosmopolitan libc / APE.

The target user experience is:

curl .../crm.com -o crm.com chmod +x crm.com  ./crm.com init ./crm.com serve

Ideally the same binary works on:

Linux x86-64

Linux ARM64

macOS x86-64

macOS ARM64

Windows x86-64

supported BSD targets

Cosmopolitan currently supports APE binaries across these families, with some architecture-specific limitations; Windows ARM64 should not initially be treated as a guaranteed native target. (⁠GitHub (https://github.com/jart/cosmopolitan/blob/master/ape/specification.md?utm_source=chatgpt.com))

The binary may embed:

default templates

JSON schemas

static web UI

agent instructions

migrations

default board definitions

so distribution remains a single executable.

4. Storage model

Markdown + YAML frontmatter

Every CRM entity is a Markdown file.

Example:

---
id: person_alice
type: person
name: Alice Smith

emails:
  - alice@example.com

company: company_acme

status: active

tags:
  - prospect
  - web3
  - grants

owner: leonid
created_at: 2026-08-20T10:30:00Z
updated_at: 2026-08-20T13:50:00Z
---

# Alice Smith

Met through the Solana ecosystem.

## Context

Alice is interested in better verification tooling for Rust/BPF programs.

## Notes

Potentially useful introduction to the security team.

## Next steps

Discuss grant opportunity once prototype is ready.

The distinction is intentional:

Frontmatter contains machine-queryable state.

Markdown contains context and memory.

5. Generic record model

The CRM should not hard-code everything around Customer.

Instead, everything is a record.

Core fields:

id: grant_solana_mcdc
type: grant

title: Solana MC/DC tooling
status: research

tags:
  - solana
  - rust
  - verification

owner: leonid

created_at: ...
updated_at: ...

Common optional fields:

priority: high

people:
  - person_alice

companies:
  - company_solana_foundation

links:
  website: ...
  application: ...

dates:
  deadline: 2026-09-20
  next_action: 2026-08-25

amount:
  requested: 75000
  currency: USD

Arbitrary metadata must be permitted:

metadata:
  chain: solana
  language: rust
  proposal_version: 3
  funding_type: grant

The system must not require schema changes every time the user wants another property.

6. Record types

Built-in recommended types:

person
company
customer
lead
deal
grant
application
onboarding
email
interaction
task
note
project

But record types should be open-ended.

For example:

type: investor

or:

type: conference

should work without recompiling the CRM.

Schemas can optionally be defined in configuration.

7. Filesystem layout

Recommended layout:

crm/
├── crm.yaml
│
├── records/
│   ├── people/
│   ├── companies/
│   ├── customers/
│   ├── grants/
│   ├── applications/
│   ├── onboarding/
│   ├── emails/
│   ├── tasks/
│   └── misc/
│
├── boards/
│   ├── inbox.yaml
│   ├── onboarding.yaml
│   ├── grants.yaml
│   └── sales.yaml
│
├── inbox/
├── attachments/
├── templates/
└── .crm/
    ├── index/
    ├── cache/
    └── runtime/

Folder location must not determine semantic state.

For example, moving a card from Research to Applied must not require moving files between directories.

State lives in metadata:

status: applied

This avoids Git churn and broken links.

8. References between records

Relationships should use stable IDs.

company: company_acme

contacts:
  - person_alice
  - person_bob

Markdown may additionally use human-readable wikilinks:

Talked with [[Alice Smith]] about [[Acme]] onboarding.

The CLI should resolve links to IDs where possible.

Example:

crm link person_alice company_acme --relation works_at

Potential canonical representation:

relations:
  - type: works_at
    target: company_acme

  - type: introduced_by
    target: person_bob

9. Configurable Kanban boards

Boards are views over records, not separate storage.

This is one of the central concepts.

Example:

# boards/grants.yaml

name: Grants
description: Grant applications and opportunities

match:
  type:
    - grant
    - application

columns:
  - id: discovered
    title: Discovered

  - id: researching
    title: Researching

  - id: preparing
    title: Preparing

  - id: applied
    title: Applied

  - id: waiting
    title: Waiting

  - id: won
    title: Won

  - id: rejected
    title: Rejected

column:
  field: status

Moving a card simply changes:

status: researching

to:

status: preparing

10. Matcher-based boards

Boards should support more sophisticated matching.

Example:

match:
  all:
    - type: customer
    - status:
        not_in:
          - archived
          - churned

Or:

match:
  any:
    - tags:
        contains: grant
    - type: grant
    - type: application

Possible matcher operators:

eq
neq
in
not_in
contains
contains_any
contains_all
exists
missing
before
after
lt
lte
gt
gte
matches

Nested boolean logic:

match:
  all:
    - type: customer

    - any:
        - onboarding.status: pending
        - onboarding.status: blocked

    - priority:
        in:
          - high
          - critical

11. Filtered boards

A board can define static filters:

filters:
  archived: false
  owner: leonid

And interactive filters exposed by the UI:

filter_controls:
  - field: owner
    type: select

  - field: tags
    type: multiselect

  - field: priority
    type: select

  - field: deadline
    type: date_range

The CLI should provide the same model:

crm board grants \
  --filter owner=leonid \
  --filter tags=solana

12. Multiple boards over the same record

A major requirement is that one record can appear on many boards simultaneously.

Example customer:

type: customer

sales_stage: closed_won
onboarding_stage: integration
support_status: healthy

It could appear on:

Sales
Customer onboarding
Account health
Enterprise customers

Each board can project a different field.

Sales:

column:
  field: sales_stage

Onboarding:

column:
  field: onboarding_stage

This is much more powerful than treating Kanban columns as physical containers.

13. Computed columns

Boards should also support computed membership.

For example an email triage board:

columns:

  - id: unread
    title: New
    match:
      triage_status: new

  - id: needs_reply
    title: Needs reply
    match:
      requires_reply: true

  - id: waiting
    title: Waiting on customer
    match:
      waiting_for: customer

  - id: done
    title: Done
    match:
      triage_status: done

Some columns therefore modify a field when cards are moved:

on_drop:
  set:
    triage_status: done

14. Board actions

Dropping a record on a column can execute deterministic mutations.

Example:

- id: applied
  title: Applied

  on_drop:
    set:
      status: applied
      applied_at: $now

Or:

- id: rejected

  on_drop:
    set:
      status: rejected
    remove:
      - next_action

Actions should initially be limited to safe declarative transformations rather than arbitrary shell commands.

15. Card configuration

Each board controls how records are rendered.

Example:

card:
  title: "{{title}}"

  subtitle: "{{organization.name}}"

  fields:
    - deadline
    - amount.requested
    - owner

  badges:
    - tags
    - priority

For onboarding:

card:
  title: "{{company.name}}"

  fields:
    - technical_owner
    - target_go_live
    - blockers

16. Customer email triage use case

Incoming mail can become records.

Example:

---
id: email_20260820_01
type: email

from:
  name: Alice
  email: alice@acme.com

company: company_acme
customer: customer_acme

subject: Problems configuring gateway

received_at: 2026-08-20T10:20:00Z

triage_status: new
priority: high
requires_reply: true

thread_id: gmail_...
---

# Problems configuring gateway

Original email text...

## Agent summary

Customer is blocked configuring authentication.

## Suggested actions

- Check environment details.
- Ask for Gateway version.

Possible board:

New
Needs triage
Needs reply
Waiting on customer
Resolved

17. Email ingestion

The core should not require Gmail integration.

Instead it should define an ingestion protocol.

For example:

crm ingest email email.json

or:

cat email.json | crm ingest

This allows external systems to supply mail from:

Gmail

IMAP

Outlook

webhook automation

agent harness

MCP tools

The CRM remains provider-neutral.

18. Inbox

Unclassified records can enter:

inbox/

or:

status: inbox

The agent can run:

crm inbox

and perform triage.

Typical agent workflow:

incoming item
      ↓
classify
      ↓
link person/company
      ↓
extract metadata
      ↓
decide board/state
      ↓
create follow-up

19. Customer onboarding use case

Example:

---
id: onboarding_acme
type: onboarding

customer: customer_acme

status: integration

owner: leonid

started_at: 2026-08-12
target_go_live: 2026-09-10

technical_owner: person_alice

blockers:
  - waiting_for_vpn

health: at_risk
---

# Acme onboarding

## Goal

Deploy self-managed Gateway into their production environment.

## Current state

VPN connectivity remains the major blocker.

## Decisions

Customer decided to use Kubernetes rather than VMs.

## Next actions

- Follow up regarding VPN.
- Schedule architecture validation.

Kanban:

New
Kickoff
Integration
Validation
Production
Complete
Blocked

20. Grants use case

Example:

---
id: grant_arbitrum_verification

type: grant
program: Arbitrum Foundation

status: preparing

ecosystem:
  - arbitrum
  - ethereum

technology:
  - solidity
  - formal-verification
  - mcdc

deadline: 2026-09-15

amount:
  target: 100000
  currency: USD

likelihood: medium

url: ...

contacts:
  - person_x

next_action:
  date: 2026-08-25
  action: Finish technical proposal
---

# Arbitrum verification tooling

## Grant

Arbitrum Foundation grant for developer/security tooling.

## Proposal

Build MC/DC-style coverage and requirement verification tooling for smart contracts.

## Why us

Existing verification and coverage infrastructure can be adapted to Solidity and Stylus.

## Requested funding

Potential scope...

## Research

...

## Application draft

...

## Correspondence

...

Possible board:

Discovered
Researching
Qualified
Preparing
Applied
Waiting
Follow-up
Won
Rejected

21. Notes and history

Records should support narrative history naturally:

## Timeline

### 2026-08-20

Talked with Alice.

She confirmed the security team is interested.

### 2026-08-14

Initial introduction.

However important events should optionally be separate records:

type: interaction
subject: Call with Alice
related:
  - person_alice
  - company_acme
  - grant_arbitrum_verification

That gives the system both:

rich Markdown context

queryable activity history

22. Tasks and next actions

Tasks should be first-class records or embedded next-actions.

Simple case:

next_action:
  date: 2026-08-25
  action: Follow up with Alice

Complex case:

type: task

title: Finish Arbitrum proposal

status: open
due: 2026-08-25

related:
  - grant_arbitrum_verification

Global command:

crm next

Example output:

TODAY

[HIGH] Follow up with Acme
       customer_acme

[MED]  Finish Arbitrum proposal
       grant_arbitrum_verification

23. CLI

The CLI is a first-class product surface.

Core operations:

crm init

crm list
crm show <id>

crm create <type>
crm edit <id>
crm delete <id>

crm search <query>
crm query <expression>

crm board
crm board <name>

crm move <id> <board> <column>

crm inbox
crm next

crm validate
crm index

crm serve

24. Structured output

Every meaningful CLI command should support:

--json

Example:

crm query 'type=grant status=preparing' --json

Output:

{
  "records": [
    {
      "id": "grant_arbitrum_verification",
      "type": "grant",
      "status": "preparing"
    }
  ]
}

This is essential for agents.

CLI output should therefore have two layers:

Human-friendly default output

Machine-stable JSON output

25. Query language

Simple syntax should work first:

crm query 'type=grant'

crm query 'type=grant status=waiting'

crm query 'deadline<2026-09-01'

crm query 'tags contains solana'

More complex queries can use expressions:

type == "grant" && status in ["researching", "preparing"] && deadline < today + 30d

The implementation must remain deterministic.

26. Full-text search

Search must cover:

YAML metadata

Markdown prose

filenames

optionally attachments

Example:

crm search "MCDC Solana"

The first version can simply scan files.

Generated indexes can later accelerate this.

The canonical files must never depend on the index existing.

27. AI-native harness

The project should ship an agent contract such as:

AGENTS.md
SKILL.md

embedded in the executable and installable with:

crm agent install

This teaches agents:

repository structure

record schema

how to search

how to mutate records

how relationships work

how boards work

validation rules

when to use CLI versus direct file editing

28. Agent workflow API

Agents should preferably use deterministic commands for structured mutations.

Instead of manually modifying:

status: applied

the agent may call:

crm set grant_arbitrum status applied

or:

crm patch grant_arbitrum \
  --set status=applied \
  --set applied_at=now

Reasons:

validation

atomic writes

schema enforcement

normalization

easier audit

fewer malformed YAML documents

Direct Markdown editing must still remain supported.

29. Agent-safe mutation protocol

Commands should return explicit results:

{
  "ok": true,
  "record": "grant_arbitrum",
  "changed": {
    "status": {
      "from": "preparing",
      "to": "applied"
    }
  }
}

Errors should be equally structured:

{
  "ok": false,
  "error": "invalid_enum",
  "field": "status",
  "value": "foo",
  "allowed": [
    "preparing",
    "applied",
    "won",
    "rejected"
  ]
}

This makes the CLI itself an excellent AI tool interface without requiring MCP.

30. AI context commands

Provide opinionated context extraction:

crm context company_acme

This can deterministically assemble:

company
contacts
active deals
onboarding
recent emails
recent interactions
open tasks
relevant notes

Agent invocation:

crm context company_acme --json

or:

crm context company_acme --md

This avoids making every agent rediscover relationship traversal.

31. Triage command

A useful concept:

crm triage

This presents records requiring decisions:

New emails
Unlinked contacts
Records missing metadata
Overdue next actions
Customer blockers
Grant deadlines

Machine output:

crm triage --json

Agents can use this as their default starting point.

32. Web interface

The web interface should remain intentionally lightweight.

Run:

crm serve

Then:

http://localhost:7777

No separate Node.js runtime.

No frontend development server.

No database.

Static web assets should be embedded in crm.com.

33. Web UI capabilities

Initial UI:

Boards

Kanban visualization.

Researching      Preparing       Applied
┌──────────┐    ┌──────────┐    ┌──────────┐
│ Solana   │    │ Arbitrum │    │ Ethereum │
│ MC/DC    │    │ verifier │    │ grant    │
└──────────┘    └──────────┘    └──────────┘

Drag/drop invokes the same deterministic board mutation engine as:

crm move ...

Record view

Display:

metadata

rendered Markdown

relationships

backlinks

activity

attachments

Search

Global full-text search.

Filters

Dynamic filters based on board definitions.

Editor

Basic Markdown/frontmatter editing.

34. Web architecture

Prefer:

Browser
   │
   HTTP
   │
crm.com
   │
filesystem

Rather than introducing a client-side database.

The local HTTP API can mirror CLI functionality:

GET  /api/records
GET  /api/records/:id
POST /api/records
PATCH /api/records/:id

GET  /api/boards
GET  /api/boards/:id

POST /api/boards/:id/move

The CLI and HTTP API should use the same internal application layer.

35. Concurrency

Because files are canonical, concurrency requires care.

For local use:

acquire record lock

read file

verify modification timestamp/hash

apply patch

write a temp file

atomic rename

release lock

For agents, optionally support optimistic concurrency:

crm patch customer_acme \
  --if-version 01JD...

Conflict:

{
  "ok": false,
  "error": "conflict",
  "expected_version": "...",
  "current_version": "..."
}

36. Versioning

Every record can expose a computed version:

_version: sha256:...

But preferably _version is derived rather than persisted.

Git provides long-term history.

The CRM itself only needs enough versioning for safe concurrent edits.

37. Git-native behavior

Git should be treated as a natural companion, not mandatory infrastructure.

Useful commands:

crm diff

crm changed

crm history customer_acme

These may delegate to Git when available.

The CRM itself should continue working in a directory with no Git repository.

38. Schemas

Schemas should be optional.

Example:

types:

  grant:
    required:
      - title
      - status
    fields:
      status:
        type: string
        enum:
          - discovered
          - researching
          - qualified
          - preparing
          - applied
          - waiting
          - won
          - rejected
      deadline:
        type: date
      amount.target:
        type: number

Validation:

crm validate

Agents should run validation after mutations.

39. Templates

Record templates:

templates/
├── person.md
├── customer.md
├── grant.md
└── onboarding.md

Command:

crm create grant

could open or generate:

---
type: grant
status: discovered
created_at: ...
---

# {{title}}

## Opportunity

## Why this fits

## Proposal idea

## Contacts

## Next action

40. Derived indexes

To keep querying fast, .crm/index/ may contain generated structures.

For example:

.crm/index/
├── records.json
├── by-type.json
├── by-tag.json
├── backlinks.json
├── dates.json
└── search.idx

These must always be disposable.

rm -rf .crm/index
crm index

must reconstruct everything.

41. Attachments

Files can reference local attachments.

attachments:
  - attachments/arbitrum/program-guidelines.pdf

The CRM does not need to understand every attachment format initially.

Agents may operate on them separately.

The web UI should at least expose/download/open them.

42. Import/export

Because files are canonical, export is naturally simple.

Still support:

crm export --json
crm export --csv

and:

crm import customers.csv

Potential integrations later:

HubSpot
Salesforce
Notion
Airtable
Linear
Gmail
Google Calendar
Outlook
Slack

These should remain adapters around the core filesystem model.

43. Hooks

Eventually support declarative lifecycle hooks:

hooks:

  after_move:
    - board: grants
      column: applied
      action: create_task
      args:
        due: +7d
        title: Follow up on {{record.title}}

Avoid arbitrary executable shell hooks in the first version.

Prefer deterministic built-in actions.

44. Views beyond Kanban

The board/query architecture should eventually support several projections:

kanban
table
list
calendar
timeline

Same records, different views.

Example:

name: Grant Deadlines

view: calendar

match:
  type: grant

date_field: deadline

No new data representation is required.

45. Dashboard

A configurable dashboard might contain:

widgets:

  - type: count
    title: Grants waiting
    query:
      type: grant
      status: waiting

  - type: count
    title: Blocked onboardings
    query:
      type: onboarding
      status: blocked

  - type: list
    title: Actions due this week
    query:
      next_action.date:
        before: +7d

  - type: board
    board: inbox

This can power both CLI summaries and web UI.

46. Recommended initial commands

For the first usable version:

crm init

crm create
crm show
crm list
crm search
crm query

crm set
crm patch

crm board
crm move

crm next
crm triage

crm validate
crm index

crm context

crm serve

That is enough to make the system genuinely useful.

47. MVP scope

Phase 1

Build only:

Markdown/frontmatter parser

record ID model

relationships

configuration

query engine

configurable Kanban

CLI

JSON CLI output

validation

basic search

local HTTP server

simple web Kanban

Cosmopolitan build

embedded web assets

AGENTS.md / SKILL.md

Do not initially build:

authentication

cloud sync

multi-user permissions

email provider integrations

elaborate plugin systems

vector database

embeddings

internal LLM calls

workflow engine

SQLite fallback

server deployment infrastructure

The CRM should deliberately not contain an AI model.

The AI already lives outside it.

48. Important architectural principle

The system should distinguish:

Knowledge
   ↓
Markdown body

State
   ↓
YAML frontmatter

Relations
   ↓
stable IDs

Views
   ↓
board configs

Queries
   ↓
deterministic query engine

Intelligence
   ↓
external AI agent

This avoids turning the CRM into another opaque AI application.

49. Example complete workspace

my-crm/
│
├── crm.yaml
│
├── records/
│   ├── people/
│   │   ├── alice-smith.md
│   │   └── john-doe.md
│   │
│   ├── companies/
│   │   ├── acme.md
│   │   └── solana-foundation.md
│   │
│   ├── customers/
│   │   └── acme.md
│   │
│   ├── onboarding/
│   │   └── acme.md
│   │
│   ├── grants/
│   │   ├── solana-mcdc.md
│   │   └── arbitrum-verification.md
│   │
│   ├── emails/
│   │   └── 2026/
│   │       └── 08/
│   │           └── ...
│   │
│   └── tasks/
│
├── boards/
│   ├── inbox.yaml
│   ├── customers.yaml
│   ├── onboarding.yaml
│   └── grants.yaml
│
├── templates/
├── attachments/
├── AGENTS.md
└── .crm/
    ├── index/
    ├── cache/
    └── runtime/

50. Product philosophy

The CRM should feel less like Salesforce and more like:

Git + Markdown + jq + Kanban + an agent harness

The filesystem itself becomes the API.

The CLI adds deterministic semantics where plain file editing becomes dangerous.

The web UI is merely a projection of that filesystem.

And AI agents don’t need a special SDK, database driver, GraphQL schema, or MCP server to understand the CRM.

They can simply:

crm triage --json
crm context customer_acme --md
crm query 'type=grant status=preparing' --json
crm move grant_arbitrum grants applied

or directly inspect:

cat records/grants/arbitrum-verification.md
rg "MCDC" records/

That should remain the central design constraint.
