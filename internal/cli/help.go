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
		{Name: "help", Purpose: "List commands or show one command", Usage: "recs help [command]", Flags: []string{"--json"}, Examples: []string{"recs help", "recs help query --json"}, JSON: `{"ok":true,"commands":[{"name":"query","purpose":"..."}]}`},
		{Name: "init", Purpose: "Create workspace layout", Usage: "recs init [--root <dir>]", Flags: []string{"--root <dir>", "--json"}, Examples: []string{"recs init", "recs init --root ./ws --json"}, JSON: `{"ok":true,"root":"/path"}`},
		{Name: "create", Purpose: "Create a record. Types are open-ended: any type string is allowed. Known types use a named folder; inbox stays inbox.", Usage: "recs create <type> [--id ID] [--title T] [--body TEXT] [--set k=v]", Flags: []string{"--id <id>", "--title <text>", "--name <text>", "--body <text>", "--set k=v", "--json"}, Examples: []string{`recs create grant --title "Demo" --set status=researching`, "recs create note --id note_1 --json"}, JSON: `{"ok":true,"record":{"id":"...","type":"grant"}}`},
		{Name: "show", Purpose: "Show a record", Usage: "recs show <id>", Flags: []string{"--json"}, Examples: []string{"recs show grant_1", "recs show grant_1 --json"}, JSON: `{"ok":true,"record":{"id":"...","type":"grant","body":"..."}}`},
		{Name: "list", Purpose: "List records", Usage: "recs list [--type <type>]", Flags: []string{"--type <type>", "--json"}, Examples: []string{"recs list", "recs list --type grant --json"}, JSON: `{"ok":true,"records":[{"id":"...","type":"grant","status":"..."}]}`},
		{Name: "search", Purpose: "Full-text search. An empty query is a usage error.", Usage: "recs search <query>", Flags: []string{"--json"}, Examples: []string{"recs search acme", "recs search acme --json"}, JSON: `{"ok":true,"records":[{"id":"...","type":"..."}]}`},
		{Name: "query", Purpose: "Filter records. Operators: = != < > <= >= contains in. Space joins clauses with AND.", Usage: "recs query <expr>", Flags: []string{"--json"}, Examples: []string{"recs query 'type=grant'", "recs query 'status=open' --json"}, JSON: `{"ok":true,"records":[{"id":"...","type":"grant"}]}`},
		{Name: "set", Purpose: "Set one field", Usage: "recs set <id> <field> <value>", Flags: []string{"--json"}, Examples: []string{"recs set grant_1 status applied", "recs set grant_1 status applied --json"}, JSON: `{"ok":true,"record":"grant_1","changed":true,"version":"..."}`},
		{Name: "patch", Purpose: "Patch fields. --set or --body is required.", Usage: "recs patch <id> --set k=v [--if-version HASH]", Flags: []string{"--set k=v", "--if-version <hash>", "--json"}, Examples: []string{"recs patch <id> --set status=applied", "recs patch <id> --set status=applied --if-version sha256:abc --json"}, JSON: `{"ok":true,"record":"<id>","changed":true,"version":"..."}`},
		{Name: "board", Purpose: "List boards or show a board", Usage: "recs board [name] [--filter k=v]", Flags: []string{"--filter k=v", "--json"}, Examples: []string{"recs board", "recs board grants --filter status=applied --json"}, JSON: `{"ok":true,"board":"grants","columns":[{"id":"...","records":[]}]}`},
		{Name: "dashboard", Purpose: "List dashboards, show one, or write a new YAML view", Usage: "recs dashboard [id] | recs dashboard new <id>", Flags: []string{"--name <text>", "--json"}, Examples: []string{"recs dashboard", "recs dashboard workspace --json", "recs dashboard new extra --name Extra"}, JSON: `{"ok":true,"dashboards":["workspace"]}`},
		{Name: "move", Purpose: "Move a card by changing frontmatter", Usage: "recs move <id> <board> <column>", Flags: []string{"--json"}, Examples: []string{"recs move <id> <board> <column>", "recs move <id> <board> <column> --json"}, JSON: `{"ok":true,"record":{"id":"<id>"},"path":"..."}`},
		{Name: "next", Purpose: "List next actions. A row is produced for an open task or a record with due, next_action, or priority.", Usage: "recs next", Flags: []string{"--json"}, Examples: []string{"recs next", "recs next --json"}, JSON: `{"ok":true,"actions":[{"id":"...","title":"...","priority":"high"}]}`},
		{Name: "triage", Purpose: "List items that need a decision. Reasons: inbox, overdue, blocker, missing_metadata.", Usage: "recs triage", Flags: []string{"--json"}, Examples: []string{"recs triage", "recs triage --json"}, JSON: `{"ok":true,"items":[{"id":"...","reason":"inbox","title":"..."}]}`},
		{Name: "validate", Purpose: "Validate optional schemas", Usage: "recs validate", Flags: []string{"--json"}, Examples: []string{"recs validate", "recs validate --json"}, JSON: `{"ok":true,"schema_present":true,"violations":[]}`},
		{Name: "index", Purpose: "Rebuild disposable index", Usage: "recs index", Flags: []string{"--json"}, Examples: []string{"recs index", "recs index --json"}, JSON: `{"ok":true,"records":12}`},
		{Name: "context", Purpose: "Assemble related records", Usage: "recs context <id>", Flags: []string{"--md", "--json"}, Examples: []string{"recs context grant_1", "recs context grant_1 --json"}, JSON: `{"ok":true,"seed":{"id":"..."},"related":[]}`},
		{Name: "inbox", Purpose: "List unclassified inbox records", Usage: "recs inbox", Flags: []string{"--json"}, Examples: []string{"recs inbox", "recs inbox --json"}, JSON: `{"ok":true,"records":[{"id":"...","type":"note"}]}`},
		{Name: "serve", Purpose: "Local HTTP UI on :7777", Usage: "recs serve [--port N]", Flags: []string{"--port <n>"}, Examples: []string{"recs serve", "recs serve --port 8080"}, JSON: ""},
		{Name: "edit", Purpose: "Edit body or frontmatter, or open $EDITOR", Usage: "recs edit <id> [--body TEXT] [--set k=v] [--if-version HASH]", Flags: []string{"--body <text>", "--set k=v", "--if-version <hash>", "--json"}, Examples: []string{"recs edit grant_1 --body '# New'", "recs edit grant_1 --set status=applied --json"}, JSON: `{"ok":true,"record":"grant_1","changed":true,"version":"..."}`},
		{Name: "delete", Purpose: "Delete a record file", Usage: "recs delete <id>", Flags: []string{"--json"}, Examples: []string{"recs delete note_1", "recs delete note_1 --json"}, JSON: `{"ok":true,"record":"note_1"}`},
		{Name: "link", Purpose: "Write a canonical relation", Usage: "recs link <id> <target> --relation <type>", Flags: []string{"--relation <type>", "--json"}, Examples: []string{"recs link grant_1 org_acme --relation company", "recs link grant_1 org_acme --relation company --json"}, JSON: `{"ok":true,"record":"grant_1","relation":{"type":"company","target":"org_acme"}}`},
		{Name: "ingest", Purpose: "Create a record from provider-neutral JSON. Stdin is read only when the operand is -.", Usage: "recs ingest [email|record] [file|-]", Flags: []string{"--json"}, Examples: []string{"recs ingest email mail.json", "cat mail.json | recs ingest - --json", `{"type":"email","subject":"Hello","body":"..."}`}, JSON: `{"ok":true,"record":{"id":"...","type":"email"}}`},
		{Name: "export", Purpose: "Export records as JSON or CSV. CSV is a flat subset (id,type,title,name,status,owner,tags,body) and is not round-trip complete.", Usage: "recs export [--csv]", Flags: []string{"--csv", "--json"}, Examples: []string{"recs export --json", "recs export --csv"}, JSON: `{"ok":true,"records":[...]}`},
		{Name: "import", Purpose: "Import records from CSV", Usage: "recs import <file.csv> [--type <type>]", Flags: []string{"--type <type>", "--json"}, Examples: []string{"recs import customers.csv", "recs import customers.csv --type person --json"}, JSON: `{"ok":true,"records":["p1"],"count":1}`},
		{Name: "diff", Purpose: "Show git diff when a repo exists", Usage: "recs diff", Flags: []string{"--json"}, Examples: []string{"recs diff", "recs diff --json"}, JSON: `{"ok":true,"git":true,"output":"..."}`},
		{Name: "changed", Purpose: "List git-changed workspace files", Usage: "recs changed", Flags: []string{"--json"}, Examples: []string{"recs changed", "recs changed --json"}, JSON: `{"ok":true,"git":true,"changed":["records/a.md"]}`},
		{Name: "history", Purpose: "Show git history for a record", Usage: "recs history <id>", Flags: []string{"--json"}, Examples: []string{"recs history grant_1", "recs history grant_1 --json"}, JSON: `{"ok":true,"git":true,"history":["..."]}`},
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
			"ok": true, "tool": "recs",
			"purpose":  "file-native records stored in files",
			"commands": out,
			"flags":    []string{"--json", "--csv", "--md", "--root <dir>", "--set k=v", "--body text", "--type <type>", "--relation <type>", "--filter k=v", "--if-version <hash>", "--port <n>"},
			"errors":   map[string]any{"ok": false, "error": "unknown_command", "field": "command", "allowed": commandNames(), "message": "...", "next": "recs --help"},
			"next":     "recs help <cmd>",
		})
		return 0
	}
	var b strings.Builder
	b.WriteString("recs - file-native records stored in files\n\n")
	b.WriteString("Agents discover this tool from help and structured errors.\n")
	b.WriteString("There is no AGENTS.md, SKILL.md, or recs agent install.\n\n")
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
	b.WriteString("\nNext: recs help <cmd>\n")
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
			"message": "unknown command " + name, "next": "recs --help",
		}
		if jsonOut {
			_ = json.NewEncoder(w).Encode(payload)
			return 1
		}
		fmt.Fprintln(errw, "unknown command "+name)
		fmt.Fprintln(errw, "next: recs --help")
		return 1
	}
	if jsonOut {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true, "command": c.Name, "purpose": c.Purpose, "usage": c.Usage,
			"flags": c.Flags, "examples": c.Examples, "json": c.JSON,
			"global_flags": []string{"--root <dir>", "--json"},
			"errors": map[string]any{"ok": false, "error": "error", "field": "field", "allowed": []string{}, "message": "...", "next": "recs help " + c.Name},
		})
		return 0
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n%s\n\nUsage:\n  %s\n", c.Name, c.Purpose, c.Usage)
	if len(c.Flags) > 0 { //mcdc:ignore:defensive every catalog command declares at least one flag
		b.WriteString("\nFlags:\n")
		for _, f := range c.Flags {
			fmt.Fprintf(&b, "  %s\n", f)
		}
	}
	if len(c.Examples) > 0 { //mcdc:ignore:defensive every catalog command declares at least one example
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
	b.WriteString("Global flags apply to every command. Use --root <dir> to set the workspace.\n")
	fmt.Fprint(w, b.String())
	return 0
}

// Implements: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM SW-REQ-260821-E5V8
func nextHint(cmd, errMsg string) string {
	if strings.Contains(errMsg, "crm.yaml not found") {
		return "recs init --root <dir>"
	}
	if strings.Contains(errMsg, "unknown flag") {
		if cmd != "" && cmd != "help" {
			return "recs help " + cmd
		}
		return "recs --help"
	}
	if strings.HasPrefix(errMsg, "unknown command") {
		return "recs --help"
	}
	if strings.Contains(errMsg, "dashboard") && strings.Contains(errMsg, "not found") {
		return "recs board"
	}
	if strings.Contains(errMsg, "unknown column") {
		return "recs help move"
	}
	if strings.Contains(errMsg, "board") && strings.Contains(errMsg, "not found") {
		return "recs board"
	}
	if strings.Contains(errMsg, "not found") {
		return "recs list"
	}
	if strings.HasPrefix(errMsg, "usage:") || strings.Contains(errMsg, "EDITOR") {
		if cmd != "" && cmd != "help" { //mcdc:ignore:defensive fail() always has a non-empty command name
			return "recs help " + cmd
		}
		return "recs --help"
	}
	if cmd != "" && cmd != "help" { //mcdc:ignore:defensive fail() always has a non-empty command name
		return "recs help " + cmd
	}
	return "recs --help"
}
