package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// CommandHelp is the agent-facing contract for one command.
type CommandHelp struct {
	Name     string   `json:"name"`
	Purpose  string   `json:"purpose"`
	Usage    string   `json:"usage"`
	Flags    []string `json:"flags"`
	Examples []string `json:"examples"`
	JSON     string   `json:"json"`
}

// Implements: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM INT-REQ-260821-BSH3
func commandCatalog() []CommandHelp {
	return []CommandHelp{
		{Name: "help", Purpose: "List commands or show one command", Usage: "crm help [command]", Flags: []string{"--json"}, Examples: []string{"crm help", "crm help query --json"}, JSON: `{"ok":true,"commands":[{"name":"query","purpose":"..."}]}`},
		{Name: "init", Purpose: "Create workspace layout", Usage: "crm init [--root <dir>]", Flags: []string{"--root <dir>", "--json"}, Examples: []string{"crm init", "crm init --root ./ws --json"}, JSON: `{"ok":true,"root":"/path"}`},
		{Name: "create", Purpose: "Create a record", Usage: "crm create <type> [--id ID] [--title T] [--body TEXT] [--set k=v]", Flags: []string{"--id <id>", "--title <text>", "--name <text>", "--body <text>", "--set k=v", "--json"}, Examples: []string{`crm create grant --title "Demo" --set status=researching`, "crm create note --id note_1 --json"}, JSON: `{"ok":true,"record":{"id":"...","type":"grant"}}`},
		{Name: "show", Purpose: "Show a record", Usage: "crm show <id>", Flags: []string{"--json"}, Examples: []string{"crm show grant_1", "crm show grant_1 --json"}, JSON: `{"ok":true,"record":{"id":"...","type":"grant","body":"..."}}`},
		{Name: "list", Purpose: "List records", Usage: "crm list [--type <type>]", Flags: []string{"--type <type>", "--json"}, Examples: []string{"crm list", "crm list --type grant --json"}, JSON: `{"ok":true,"records":[{"id":"...","type":"grant","status":"..."}]}`},
		{Name: "search", Purpose: "Full-text search", Usage: "crm search <query>", Flags: []string{"--json"}, Examples: []string{"crm search acme", "crm search acme --json"}, JSON: `{"ok":true,"records":[{"id":"...","type":"..."}]}`},
		{Name: "query", Purpose: "Filter records", Usage: "crm query <expr>", Flags: []string{"--json"}, Examples: []string{"crm query 'type=grant'", "crm query 'status=open' --json"}, JSON: `{"ok":true,"records":[{"id":"...","type":"grant"}]}`},
		{Name: "set", Purpose: "Set one field", Usage: "crm set <id> <field> <value>", Flags: []string{"--json"}, Examples: []string{"crm set grant_1 status applied", "crm set grant_1 status applied --json"}, JSON: `{"ok":true,"record":"grant_1","changed":true,"version":"..."}`},
		{Name: "patch", Purpose: "Patch fields", Usage: "crm patch <id> --set k=v [--if-version HASH]", Flags: []string{"--set k=v", "--if-version <hash>", "--json"}, Examples: []string{"crm patch grant_1 --set status=applied", "crm patch grant_1 --set status=applied --if-version sha256:abc --json"}, JSON: `{"ok":true,"record":"grant_1","changed":true,"version":"..."}`},
		{Name: "board", Purpose: "List boards or show a board", Usage: "crm board [name] [--filter k=v]", Flags: []string{"--filter k=v", "--json"}, Examples: []string{"crm board", "crm board grants --filter status=applied --json"}, JSON: `{"ok":true,"board":"grants","columns":[{"id":"...","records":[]}]}`},
		{Name: "dashboard", Purpose: "List dashboards, show one, or write a new YAML view", Usage: "crm dashboard [id] | crm dashboard new <id>", Flags: []string{"--name <text>", "--json"}, Examples: []string{"crm dashboard", "crm dashboard workspace --json", "crm dashboard new extra --name Extra"}, JSON: `{"ok":true,"dashboards":["workspace"]}`},
		{Name: "move", Purpose: "Move a card by changing frontmatter", Usage: "crm move <id> <board> <column>", Flags: []string{"--json"}, Examples: []string{"crm move grant_1 grants applied", "crm move grant_1 grants applied --json"}, JSON: `{"ok":true,"record":{"id":"grant_1"},"path":"..."}`},
		{Name: "next", Purpose: "List next actions", Usage: "crm next", Flags: []string{"--json"}, Examples: []string{"crm next", "crm next --json"}, JSON: `{"ok":true,"actions":[{"id":"...","title":"...","priority":"high"}]}`},
		{Name: "triage", Purpose: "List items that need a decision", Usage: "crm triage", Flags: []string{"--json"}, Examples: []string{"crm triage", "crm triage --json"}, JSON: `{"ok":true,"items":[{"id":"...","reason":"inbox","title":"..."}]}`},
		{Name: "validate", Purpose: "Validate optional schemas", Usage: "crm validate", Flags: []string{"--json"}, Examples: []string{"crm validate", "crm validate --json"}, JSON: `{"ok":true,"schema_present":true,"violations":[]}`},
		{Name: "index", Purpose: "Rebuild disposable index", Usage: "crm index", Flags: []string{"--json"}, Examples: []string{"crm index", "crm index --json"}, JSON: `{"ok":true,"records":12}`},
		{Name: "context", Purpose: "Assemble related records", Usage: "crm context <id>", Flags: []string{"--md", "--json"}, Examples: []string{"crm context grant_1", "crm context grant_1 --json"}, JSON: `{"ok":true,"seed":{"id":"..."},"related":[]}`},
		{Name: "inbox", Purpose: "List unclassified inbox records", Usage: "crm inbox", Flags: []string{"--json"}, Examples: []string{"crm inbox", "crm inbox --json"}, JSON: `{"ok":true,"records":[{"id":"...","type":"note"}]}`},
		{Name: "serve", Purpose: "Local HTTP UI on :7777", Usage: "crm serve [--port N]", Flags: []string{"--port <n>"}, Examples: []string{"crm serve", "crm serve --port 8080"}, JSON: ""},
		{Name: "edit", Purpose: "Edit body or frontmatter, or open $EDITOR", Usage: "crm edit <id> [--body TEXT] [--set k=v] [--if-version HASH]", Flags: []string{"--body <text>", "--set k=v", "--if-version <hash>", "--json"}, Examples: []string{"crm edit grant_1 --body '# New'", "crm edit grant_1 --set status=applied --json"}, JSON: `{"ok":true,"record":"grant_1","changed":true,"version":"..."}`},
		{Name: "delete", Purpose: "Delete a record file", Usage: "crm delete <id>", Flags: []string{"--json"}, Examples: []string{"crm delete note_1", "crm delete note_1 --json"}, JSON: `{"ok":true,"record":"note_1"}`},
		{Name: "link", Purpose: "Write a canonical relation", Usage: "crm link <id> <target> --relation <type>", Flags: []string{"--relation <type>", "--json"}, Examples: []string{"crm link grant_1 org_acme --relation company", "crm link grant_1 org_acme --relation company --json"}, JSON: `{"ok":true,"record":"grant_1","relation":{"type":"company","target":"org_acme"}}`},
		{Name: "ingest", Purpose: "Create a record from provider-neutral JSON", Usage: "crm ingest [email|record] [file|-]", Flags: []string{"--json"}, Examples: []string{"crm ingest email mail.json", "cat mail.json | crm ingest --json"}, JSON: `{"ok":true,"record":{"id":"...","type":"email"}}`},
		{Name: "export", Purpose: "Export records as JSON or CSV", Usage: "crm export [--csv]", Flags: []string{"--csv", "--json"}, Examples: []string{"crm export --json", "crm export --csv"}, JSON: `{"ok":true,"records":[...]}`},
		{Name: "import", Purpose: "Import records from CSV", Usage: "crm import <file.csv> [--type <type>]", Flags: []string{"--type <type>", "--json"}, Examples: []string{"crm import customers.csv", "crm import customers.csv --type person --json"}, JSON: `{"ok":true,"records":["p1"],"count":1}`},
		{Name: "diff", Purpose: "Show git diff when a repo exists", Usage: "crm diff", Flags: []string{"--json"}, Examples: []string{"crm diff", "crm diff --json"}, JSON: `{"ok":true,"git":true,"output":"..."}`},
		{Name: "changed", Purpose: "List git-changed workspace files", Usage: "crm changed", Flags: []string{"--json"}, Examples: []string{"crm changed", "crm changed --json"}, JSON: `{"ok":true,"git":true,"changed":["records/a.md"]}`},
		{Name: "history", Purpose: "Show git history for a record", Usage: "crm history <id>", Flags: []string{"--json"}, Examples: []string{"crm history grant_1", "crm history grant_1 --json"}, JSON: `{"ok":true,"git":true,"history":["..."]}`},
	}
}

// Implements: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM
func commandNames() []string {
	cats := commandCatalog()
	names := make([]string, 0, len(cats))
	for _, c := range cats {
		if c.Name == "help" {
			continue
		}
		names = append(names, c.Name)
	}
	sort.Strings(names)
	return names
}

// Implements: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM
func lookupCommand(name string) (CommandHelp, bool) {
	for _, c := range commandCatalog() {
		if c.Name == name {
			return c, true
		}
	}
	return CommandHelp{}, false
}

// Implements: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM INT-REQ-260821-BSH3
func printGlobalHelp(w io.Writer, jsonOut bool) int {
	cats := commandCatalog()
	if jsonOut {
		out := make([]map[string]string, 0, len(cats))
		for _, c := range cats {
			out = append(out, map[string]string{"name": c.Name, "purpose": c.Purpose, "usage": c.Usage})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true, "tool": "crm",
			"purpose":  "agentic CRM stored in files",
			"commands": out,
			"flags":    []string{"--json", "--csv", "--md", "--root <dir>", "--set k=v", "--body text", "--type <type>", "--relation <type>", "--filter k=v", "--if-version <hash>", "--port <n>"},
			"errors":   map[string]any{"ok": false, "error": "unknown_command", "field": "command", "allowed": commandNames(), "message": "...", "next": "crm --help"},
			"next":     "crm help <cmd>",
		})
		return 0
	}
	var b strings.Builder
	b.WriteString("crm - agentic CRM stored in files\n\n")
	b.WriteString("Agents discover this tool from help and structured errors.\n")
	b.WriteString("There is no AGENTS.md, SKILL.md, or crm agent install.\n\n")
	b.WriteString("Commands:\n")
	for _, c := range cats {
		fmt.Fprintf(&b, "  %-18s %s\n", c.Usage, c.Purpose)
	}
	b.WriteString("\nFlags:\n")
	b.WriteString("  --json               Stable machine output\n")
	b.WriteString("  --csv                CSV export format\n")
	b.WriteString("  --md                 Markdown context output\n")
	b.WriteString("  --root <dir>         Workspace root\n")
	b.WriteString("  --set k=v            Field assignment\n")
	b.WriteString("  --body text          Markdown body\n")
	b.WriteString("  --type <type>        Type filter or import default\n")
	b.WriteString("  --relation <type>    Relation type for link\n")
	b.WriteString("  --filter k=v         Board filter\n")
	b.WriteString("  --if-version <hash>  Optimistic concurrency\n")
	b.WriteString("  --port <n>           serve port (default 7777)\n")
	b.WriteString("\nNext: crm help <cmd>\n")
	fmt.Fprint(w, b.String())
	return 0
}

// Implements: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM INT-REQ-260821-BSH3
func printCommandHelp(w, errw io.Writer, name string, jsonOut bool) int {
	c, ok := lookupCommand(name)
	if !ok {
		payload := map[string]any{
			"ok": false, "error": "unknown_command", "field": "command",
			"value": name, "allowed": commandNames(),
			"message": "unknown command " + name, "next": "crm --help",
		}
		if jsonOut {
			_ = json.NewEncoder(w).Encode(payload)
			return 1
		}
		fmt.Fprintln(errw, "unknown command "+name)
		fmt.Fprintln(errw, "next: crm --help")
		return 1
	}
	if jsonOut {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true, "command": c.Name, "purpose": c.Purpose, "usage": c.Usage,
			"flags": c.Flags, "examples": c.Examples, "json": c.JSON,
			"errors": map[string]any{"ok": false, "error": "error", "field": "field", "allowed": []string{}, "message": "...", "next": "crm help " + c.Name},
		})
		return 0
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n%s\n\nUsage:\n  %s\n", c.Name, c.Purpose, c.Usage)
	if len(c.Flags) > 0 {
		b.WriteString("\nFlags:\n")
		for _, f := range c.Flags {
			fmt.Fprintf(&b, "  %s\n", f)
		}
	}
	if len(c.Examples) > 0 {
		b.WriteString("\nExamples:\n")
		for _, e := range c.Examples {
			fmt.Fprintf(&b, "  %s\n", e)
		}
	}
	if c.JSON != "" {
		b.WriteString("\nJSON shape:\n  ")
		b.WriteString(c.JSON)
		b.WriteString("\n")
	}
	b.WriteString("\nErrors use {ok, error, field, allowed, message, next}.\n")
	fmt.Fprint(w, b.String())
	return 0
}

// Implements: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM
func nextHint(cmd, errMsg string) string {
	if strings.HasPrefix(errMsg, "unknown command") {
		return "crm --help"
	}
	if strings.HasPrefix(errMsg, "usage:") || strings.Contains(errMsg, "EDITOR") {
		if cmd != "" && cmd != "help" {
			return "crm help " + cmd
		}
		return "crm --help"
	}
	if cmd != "" && cmd != "help" {
		return "crm help " + cmd
	}
	return "crm --help"
}
