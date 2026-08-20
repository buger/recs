package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"crm/internal/app"
	"crm/internal/board"
	"crm/internal/record"
	"crm/internal/serve"
	"crm/internal/store"
)

// Main is the CLI entry.
// Implements: SYS-REQ-260820-PG9C SW-REQ-260820-YB5C INT-REQ-260820-JC9M
func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}
	jsonOut := false
	root := ""
	port := 0
	filtered := make([]string, 0, len(args))
	sets := map[string]any{}
	filters := map[string]string{}
	ifVersion := ""
	typFilter := ""
	mdOut := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--json":
			jsonOut = true
		case a == "--md":
			mdOut = true
		case a == "--help" || a == "-h":
			printHelp(stdout)
			return 0
		case a == "--root" && i+1 < len(args):
			i++
			root = args[i]
		case strings.HasPrefix(a, "--root="):
			root = strings.TrimPrefix(a, "--root=")
		case a == "--port" && i+1 < len(args):
			i++
			port, _ = strconv.Atoi(args[i])
		case strings.HasPrefix(a, "--port="):
			port, _ = strconv.Atoi(strings.TrimPrefix(a, "--port="))
		case a == "--set" && i+1 < len(args):
			i++
			k, v, ok := strings.Cut(args[i], "=")
			if ok {
				sets[k] = v
			}
		case a == "--filter" && i+1 < len(args):
			i++
			k, v, ok := strings.Cut(args[i], "=")
			if ok {
				filters[k] = v
			}
		case a == "--if-version" && i+1 < len(args):
			i++
			ifVersion = args[i]
		case a == "--type" && i+1 < len(args):
			i++
			typFilter = args[i]
		case a == "--id" && i+1 < len(args):
			i++
			sets["id"] = args[i]
		case a == "--title" && i+1 < len(args):
			i++
			sets["title"] = args[i]
		case a == "--name" && i+1 < len(args):
			i++
			sets["name"] = args[i]
		case a == "--body" && i+1 < len(args):
			i++
			sets["_body"] = args[i]
		default:
			filtered = append(filtered, a)
		}
	}
	if len(filtered) == 0 {
		printHelp(stdout)
		return 0
	}
	cmd := filtered[0]
	rest := filtered[1:]
	write := func(v any, human string) int {
		if jsonOut {
			enc := json.NewEncoder(stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(v)
			return 0
		}
		fmt.Fprint(stdout, human)
		if human != "" && !strings.HasSuffix(human, "\n") {
			fmt.Fprintln(stdout)
		}
		return 0
	}
	fail := func(err error) int {
		code := "error"
		if errors.Is(err, store.ErrNotFound) {
			code = "not_found"
		} else if errors.Is(err, store.ErrConflict) {
			code = "conflict"
		} else if errors.Is(err, store.ErrExists) {
			code = "exists"
		}
		if jsonOut {
			_ = json.NewEncoder(stdout).Encode(map[string]any{"ok": false, "error": code, "message": err.Error()})
		} else {
			fmt.Fprintln(stderr, err.Error())
		}
		return 1
	}

	if cmd == "init" {
		a := app.OpenOrCWD(root)
		if err := a.Init(); err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "root": a.Root()}, "initialized "+a.Root()+"\n")
	}

	a, err := openApp(cmd, root)
	if err != nil {
		return fail(err)
	}

	switch cmd {
	case "create":
		if len(rest) < 1 {
			return fail(fmt.Errorf("usage: crm create <type>"))
		}
		typ := rest[0]
		id, _ := sets["id"].(string)
		body, _ := sets["_body"].(string)
		delete(sets, "id")
		delete(sets, "_body")
		rec, err := a.Create(typ, id, sets, body)
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "record": public(rec)}, rec.ID+"\n")
	case "show":
		if len(rest) < 1 {
			return fail(fmt.Errorf("usage: crm show <id>"))
		}
		rec, err := a.Show(rest[0])
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "record": public(rec)}, string(rec.Bytes()))
	case "list":
		recs, err := a.List(typFilter)
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "records": publicList(recs)}, listHuman(recs))
	case "search":
		q := strings.Join(rest, " ")
		recs, err := a.Search(q)
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "records": publicList(recs)}, listHuman(recs))
	case "query":
		expr := strings.Join(rest, " ")
		recs, err := a.Query(expr)
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "records": publicList(recs)}, listHuman(recs))
	case "set":
		if len(rest) < 3 {
			return fail(fmt.Errorf("usage: crm set <id> <field> <value>"))
		}
		res, err := a.Set(rest[0], rest[1], strings.Join(rest[2:], " "))
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "record": res.Record.ID, "changed": res.Changed, "version": res.Version}, res.Record.ID+"\n")
	case "patch":
		if len(rest) < 1 {
			return fail(fmt.Errorf("usage: crm patch <id> --set k=v"))
		}
		res, err := a.Patch(rest[0], sets, nil, ifVersion)
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "record": res.Record.ID, "changed": res.Changed, "version": res.Version}, res.Record.ID+"\n")
	case "board":
		if len(rest) == 0 {
			boards, err := a.ListBoards()
			if err != nil {
				return fail(err)
			}
			var names []string
			var b strings.Builder
			for _, bd := range boards {
				names = append(names, bd.ID)
				fmt.Fprintf(&b, "%s\t%s\n", bd.ID, bd.Name)
			}
			return write(map[string]any{"ok": true, "boards": names}, b.String())
		}
		view, err := a.Board(rest[0], filters)
		if err != nil {
			return fail(err)
		}
		return write(boardJSON(view), boardHuman(view))
	case "move":
		if len(rest) < 3 {
			return fail(fmt.Errorf("usage: crm move <id> <board> <column>"))
		}
		rec, path, err := a.Move(rest[0], rest[1], rest[2])
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "record": public(rec), "path": path}, rec.ID+" -> "+rest[2]+"\n")
	case "next":
		items, err := a.Next()
		if err != nil {
			return fail(err)
		}
		var b strings.Builder
		for _, it := range items {
			fmt.Fprintf(&b, "[%s] %s\n       %s\n", strings.ToUpper(it.Priority), it.Title, it.ID)
		}
		return write(map[string]any{"ok": true, "actions": items}, b.String())
	case "triage":
		items, err := a.Triage()
		if err != nil {
			return fail(err)
		}
		var b strings.Builder
		for _, it := range items {
			fmt.Fprintf(&b, "%s\t%s\t%s\n", it.Reason, it.ID, it.Title)
		}
		return write(map[string]any{"ok": true, "items": items}, b.String())
	case "validate":
		res, err := a.Validate()
		if err != nil {
			return fail(err)
		}
		var b strings.Builder
		if !res.SchemaPresent {
			b.WriteString("no schemas; skipped\n")
		} else if res.OK {
			b.WriteString("ok\n")
		} else {
			for _, v := range res.Violations {
				fmt.Fprintln(&b, v.String())
			}
		}
		code := 0
		if !res.OK {
			code = 1
		}
		payload := map[string]any{"ok": res.OK, "schema_present": res.SchemaPresent, "violations": res.Violations}
		if jsonOut {
			_ = json.NewEncoder(stdout).Encode(payload)
			return code
		}
		fmt.Fprint(stdout, b.String())
		return code
	case "index":
		snap, err := a.RebuildIndex()
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "records": len(snap.Records)}, fmt.Sprintf("indexed %d records\n", len(snap.Records)))
	case "context":
		if len(rest) < 1 {
			return fail(fmt.Errorf("usage: crm context <id>"))
		}
		bundle, err := a.Context(rest[0])
		if err != nil {
			return fail(err)
		}
		if mdOut && !jsonOut {
			fmt.Fprintf(stdout, "# %s\n\n%s\n", record.DisplayName(bundle.Seed), bundle.Seed.Body)
			for _, rel := range bundle.Related {
				fmt.Fprintf(stdout, "## %s (%s)\n\n", record.DisplayName(rel), rel.ID)
			}
			return 0
		}
		return write(map[string]any{"ok": true, "seed": public(bundle.Seed), "related": publicList(bundle.Related)}, contextHuman(bundle.Seed, bundle.Related))
	case "serve":
		if port == 0 {
			port = 7777
		}
		fmt.Fprintf(stdout, "serving http://localhost:%d\n", port)
		if err := serve.Listen(a, port); err != nil {
			return fail(err)
		}
		return 0
	default:
		return fail(fmt.Errorf("unknown command %s", cmd))
	}
}

// Implements: SYS-REQ-260820-PG9C
func openApp(cmd, root string) (*app.App, error) {
	if cmd == "init" {
		return app.OpenOrCWD(root), nil
	}
	return app.Open(root)
}

// Implements: SYS-REQ-260820-PG9C
func public(rec *record.Record) map[string]any {
	out := map[string]any{"id": rec.ID, "type": rec.Type, "path": rec.Path, "version": rec.Version()}
	for k, v := range rec.Fields {
		out[k] = v
	}
	out["body"] = rec.Body
	return out
}

// Implements: SYS-REQ-260820-PG9C
func publicList(recs []*record.Record) []map[string]any {
	out := make([]map[string]any, 0, len(recs))
	for _, rec := range recs {
		item := map[string]any{"id": rec.ID, "type": rec.Type, "status": rec.GetString("status")}
		if t := rec.GetString("title"); t != "" {
			item["title"] = t
		}
		if n := rec.GetString("name"); n != "" {
			item["name"] = n
		}
		out = append(out, item)
	}
	return out
}

// Implements: SYS-REQ-260820-PG9C
func listHuman(recs []*record.Record) string {
	var b strings.Builder
	for _, rec := range recs {
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\n", rec.ID, rec.Type, rec.GetString("status"), record.DisplayName(rec))
	}
	return b.String()
}

// Implements: SYS-REQ-260820-PG9C
func boardJSON(view *board.View) map[string]any {
	cols := make([]map[string]any, 0, len(view.Columns))
	for _, c := range view.Columns {
		cols = append(cols, map[string]any{
			"id": c.Column.ID, "title": c.Column.Title, "records": publicList(c.Records),
		})
	}
	return map[string]any{"ok": true, "board": view.Board.ID, "name": view.Board.Name, "columns": cols}
}

// Implements: SYS-REQ-260820-PG9C
func boardHuman(view *board.View) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", view.Board.Name)
	for _, c := range view.Columns {
		fmt.Fprintf(&b, "\n[%s]\n", c.Column.Title)
		for _, rec := range c.Records {
			fmt.Fprintf(&b, "  %s\t%s\n", rec.ID, record.DisplayName(rec))
		}
	}
	return b.String()
}

// Implements: SYS-REQ-260820-PG9C
func contextHuman(seed *record.Record, related []*record.Record) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\t%s\n", seed.ID, record.DisplayName(seed))
	for _, rel := range related {
		fmt.Fprintf(&b, "  %s\t%s\n", rel.ID, record.DisplayName(rel))
	}
	return b.String()
}

// Implements: SYS-REQ-260820-PG9C
func printHelp(w io.Writer) {
	fmt.Fprint(w, `crm - file-native agent CRM

Commands:
  init                 Create workspace layout
  create <type>        Create a record
  show <id>            Show a record
  list                 List records
  search <query>       Full-text search
  query <expr>         Filter records
  set <id> <f> <v>     Set one field
  patch <id> --set k=v Patch fields
  board [name]         List boards or show a board
  move <id> <b> <col>  Move a card
  next                 List next actions
  triage               List items that need a decision
  validate             Validate optional schemas
  index                Rebuild disposable index
  context <id>         Assemble related records
  serve                Local HTTP UI on :7777

Flags:
  --json               Stable machine output
  --root <dir>         Workspace root
  --set k=v            Field assignment
  --filter k=v         Board filter
  --if-version <hash>  Optimistic concurrency
`)
}
