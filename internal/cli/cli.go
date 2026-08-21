package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"crm/internal/app"
	"crm/internal/board"
	"crm/internal/dashboard"
	"crm/internal/record"
	"crm/internal/serve"
	"crm/internal/store"
)

// Input is stdin for ingest; tests may replace it.
var Input io.Reader = os.Stdin

// Main is the CLI entry.
// Implements: SYS-REQ-260820-PG9C SW-REQ-260820-YB5C INT-REQ-260820-JC9M SYS-REQ-260821-8FKR SW-REQ-260821-FCGM INT-REQ-260821-BSH3
func Main(args []string, stdout, stderr io.Writer) int {
	jsonOut := false
	root := ""
	port := 0
	filtered := make([]string, 0, len(args))
	sets := map[string]any{}
	filters := map[string]string{}
	ifVersion := ""
	typFilter := ""
	mdOut := false
	csvOut := false
	relation := ""
	helpFlag := false
	if len(args) == 0 {
		return printGlobalHelp(stdout, false)
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--json":
			jsonOut = true
		case a == "--md":
			mdOut = true
		case a == "--help" || a == "-h":
			helpFlag = true
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
		case a == "--csv":
			csvOut = true
		case a == "--relation" && i+1 < len(args):
			i++
			relation = args[i]
		case strings.HasPrefix(a, "--relation="):
			relation = strings.TrimPrefix(a, "--relation=")
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
		return printGlobalHelp(stdout, jsonOut)
	}
	cmd := filtered[0]
	rest := filtered[1:]
	if cmd == "help" {
		if len(rest) == 0 {
			return printGlobalHelp(stdout, jsonOut)
		}
		return printCommandHelp(stdout, stderr, rest[0], jsonOut)
	}
	if helpFlag {
		return printCommandHelp(stdout, stderr, cmd, jsonOut)
	}
	write := func(v any, human string) int {
		if jsonOut {
			enc := json.NewEncoder(stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(v)
			return 0
		}
		fmt.Fprint(stdout, human)
		if human != "" && !strings.HasSuffix(human, "\n") { //mcdc:ignore:defensive every CLI human writer already terminates with a newline
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
		var enumErr *store.EnumError
		if errors.As(err, &enumErr) {
			code = "invalid_enum"
			if jsonOut {
				_ = json.NewEncoder(stdout).Encode(map[string]any{
					"ok": false, "error": code, "field": enumErr.Field,
					"value": enumErr.Value, "allowed": enumErr.Allowed,
					"message": err.Error(), "next": nextHint(cmd, err.Error()),
				})
				return 1
			}
		}
		var confErr *store.ConflictError
		if errors.As(err, &confErr) {
			if jsonOut {
				_ = json.NewEncoder(stdout).Encode(map[string]any{
					"ok": false, "error": "conflict",
					"expected_version": confErr.Expected, "current_version": confErr.Current,
					"message": err.Error(), "next": "crm show " + strings.Join(rest, " "),
				})
				return 1
			}
		}
		payload := map[string]any{"ok": false, "error": code, "message": err.Error()}
		if strings.HasPrefix(err.Error(), "unknown command") {
			payload["error"] = "unknown_command"
			payload["field"] = "command"
			parts := strings.Split(err.Error(), " ")
			if len(parts) >= 3 {
				payload["value"] = parts[2]
			}
			payload["allowed"] = commandNames()
		}
		if n := nextHint(cmd, err.Error()); n != "" {
			payload["next"] = n
		}
		if jsonOut {
			_ = json.NewEncoder(stdout).Encode(payload)
		} else {
			fmt.Fprintln(stderr, err.Error())
			if n, ok := payload["next"].(string); ok && n != "" {
				fmt.Fprintln(stderr, "next: "+n)
			}
		}
		return 1
	}

	if _, ok := lookupCommand(cmd); !ok {
		return fail(fmt.Errorf("unknown command %s", cmd))
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
	case "dashboard":
		if len(rest) == 0 {
			dashboards, err := a.ListDashboards()
			if err != nil {
				return fail(err)
			}
			var names []string
			var b strings.Builder
			for _, d := range dashboards {
				names = append(names, d.ID)
				fmt.Fprintf(&b, "%s\t%s\t%d widgets\n", d.ID, d.Name, len(d.Widgets))
			}
			return write(map[string]any{"ok": true, "dashboards": names}, b.String())
		}
		if rest[0] == "new" {
			if len(rest) < 2 {
				return fail(fmt.Errorf("usage: crm dashboard new <id>"))
			}
			id := rest[1]
			name, _ := sets["name"].(string)
			layout, _ := sets["layout"].(string)
			desc, _ := sets["description"].(string)
			var widgets []dashboard.Widget
			if typ, ok := sets["type"].(string); ok && typ != "" {
				widgets = []dashboard.Widget{{ID: "main", Type: typ, Title: name, Query: sets["query"]}}
			}
			d, err := a.CreateDashboard(id, name, layout, desc, widgets)
			if err != nil {
				return fail(err)
			}
			return write(map[string]any{"ok": true, "dashboard": d.ID, "path": d.File}, "created "+d.ID+"\n")
		}
		proj, err := a.ProjectDashboard(rest[0])
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "dashboard": proj}, dashboardHuman(proj))
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
		if err := serve.Listen(a, port); err != nil { //mcdc:ignore:defensive http.Serve always returns a non-nil error after a successful bind
			return fail(err)
		}
		return 0
	case "inbox":
		recs, err := a.Inbox()
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "records": publicList(recs)}, listHuman(recs))
	case "edit":
		if len(rest) < 1 {
			return fail(fmt.Errorf("usage: crm edit <id> [--body ...] [--set k=v]"))
		}
		body, hasBody := sets["_body"].(string)
		delete(sets, "_body")
		if hasBody || len(sets) > 0 {
			var bp *string
			if hasBody {
				bp = &body
			}
			res, err := a.Edit(rest[0], sets, bp, ifVersion)
			if err != nil {
				return fail(err)
			}
			return write(map[string]any{"ok": true, "record": res.Record.ID, "changed": res.Changed, "version": res.Version}, res.Record.ID+"\n")
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			return fail(fmt.Errorf("set EDITOR or pass --body / --set"))
		}
		res, err := a.EditWithEditor(rest[0], editor)
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "record": res.Record.ID, "changed": res.Changed, "version": res.Version}, res.Record.ID+"\n")
	case "delete":
		if len(rest) < 1 {
			return fail(fmt.Errorf("usage: crm delete <id>"))
		}
		if err := a.Delete(rest[0]); err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "record": rest[0]}, "deleted "+rest[0]+"\n")
	case "link":
		if len(rest) < 2 || relation == "" {
			return fail(fmt.Errorf("usage: crm link <id> <target> --relation <type>"))
		}
		res, err := a.Link(rest[0], rest[1], relation)
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "record": res.Record.ID, "relation": map[string]string{"type": relation, "target": rest[1]}, "version": res.Version}, rest[0]+" -> "+rest[1]+" ("+relation+")\n")
	case "ingest":
		kind := ""
		src := ""
		if len(rest) == 1 {
			if rest[0] == "email" || rest[0] == "record" {
				kind = rest[0]
			} else {
				src = rest[0]
			}
		} else if len(rest) >= 2 {
			kind = rest[0]
			src = rest[1]
		}
		var data []byte
		var err error
		if src == "" || src == "-" {
			data, err = io.ReadAll(Input)
		} else {
			data, err = os.ReadFile(src)
		}
		if err != nil {
			return fail(err)
		}
		rec, err := a.Ingest(kind, data)
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "record": public(rec)}, rec.ID+"\n")
	case "export":
		if csvOut {
			text, err := a.ExportCSV()
			if err != nil {
				return fail(err)
			}
			if jsonOut {
				_ = json.NewEncoder(stdout).Encode(map[string]any{"ok": true, "format": "csv", "csv": text})
				return 0
			}
			fmt.Fprint(stdout, text)
			return 0
		}
		recs, err := a.ExportJSON()
		if err != nil {
			return fail(err)
		}
		return write(map[string]any{"ok": true, "records": recs}, exportHuman(recs))
	case "import":
		if len(rest) < 1 {
			return fail(fmt.Errorf("usage: crm import <file.csv> [--type <type>]"))
		}
		f, err := os.Open(rest[0])
		if err != nil {
			return fail(err)
		}
		defer f.Close()
		created, err := a.ImportCSV(f, typFilter)
		if err != nil {
			return fail(err)
		}
		ids := make([]string, 0, len(created))
		for _, rec := range created {
			ids = append(ids, rec.ID)
		}
		return write(map[string]any{"ok": true, "records": ids, "count": len(ids)}, strings.Join(ids, "\n")+"\n")
	case "diff":
		res := a.Diff()
		if !res.OK {
			return fail(fmt.Errorf("%s", res.Message))
		}
		return write(res, gitHuman(res))
	case "changed":
		res := a.Changed()
		if !res.OK {
			return fail(fmt.Errorf("%s", res.Message))
		}
		return write(res, gitHuman(res))
	case "history":
		if len(rest) < 1 {
			return fail(fmt.Errorf("usage: crm history <id>"))
		}
		res := a.History(rest[0])
		if !res.OK {
			return fail(fmt.Errorf("%s", res.Message))
		}
		return write(res, gitHuman(res))
	default:
		return fail(fmt.Errorf("unknown command %s", cmd)) //mcdc:ignore:defensive unknown commands are rejected before openApp
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

// Implements: SYS-REQ-260820-456X
func dashboardHuman(proj *dashboard.Projection) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\t%s\t%d components\n", proj.ID, proj.Name, proj.WidgetCount)
	for _, w := range proj.Widgets {
		kind := w.Type
		if w.Placeholder {
			kind = "placeholder"
		}
		fmt.Fprintf(&b, "  %s\t%s\t%s\n", w.ID, kind, w.Title)
	}
	return b.String()
}

// Implements: SYS-REQ-260821-JYEJ
func exportHuman(recs []map[string]any) string {
	var b strings.Builder
	for _, rec := range recs {
		fmt.Fprintf(&b, "%v\t%v\t%v\n", rec["id"], rec["type"], rec["title"])
	}
	return b.String()
}

// Implements: SYS-REQ-260821-JYEJ
func gitHuman(res app.GitResult) string {
	if !res.Git {
		if res.Message != "" {
			return res.Message + "\n"
		}
		return ""
	}
	if res.Output != "" {
		return res.Output
	}
	if len(res.Changed) > 0 {
		return strings.Join(res.Changed, "\n") + "\n"
	}
	if len(res.History) > 0 {
		return strings.Join(res.History, "\n") + "\n"
	}
	return ""
}
