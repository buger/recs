package query

import (
	"fmt"
	"regexp"
	"strings"

	"crm/internal/record"
)

var clauseRe = regexp.MustCompile(`(?i)([A-Za-z0-9_.]+)\s*(!=|<=|>=|=|<|>|contains|in)\s*("[^"]+"|'[^']+'|\S+)`)

// Clause is one field operator value predicate.
type Clause struct {
	Field  string
	Op     string
	Value  string
	Values []string
}

// Parse reads a simple query.
// Implements: SYS-REQ-260820-ZTC3 SW-REQ-260820-6EVX
func Parse(expr string) ([]Clause, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty query")
	}
	matches := clauseRe.FindAllStringSubmatch(expr, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid query %q", expr)
	}
	out := make([]Clause, 0, len(matches))
	for _, m := range matches {
		c := Clause{Field: m[1], Value: strings.Trim(m[3], `"'`)}
		switch strings.ToLower(m[2]) {
		case "=":
			c.Op = "eq"
		case "!=":
			c.Op = "neq"
		case "<":
			c.Op = "lt"
		case ">":
			c.Op = "gt"
		case "<=":
			c.Op = "lte"
		case ">=":
			c.Op = "gte"
		case "contains":
			c.Op = "contains"
		case "in":
			c.Op = "in"
			c.Values = splitList(c.Value)
		default:
			return nil, fmt.Errorf("unknown operator %q", m[2])
		}
		out = append(out, c)
	}
	return out, nil
}

// Implements: SYS-REQ-260820-ZTC3
func splitClauses(expr string) []string {
	var parts []string
	var cur strings.Builder
	inQuote := false
	for i := 0; i < len(expr); i++ {
		ch := expr[i]
		if ch == '"' {
			inQuote = !inQuote
			cur.WriteByte(ch)
			continue
		}
		if !inQuote && ch == ' ' {
			if cur.Len() > 0 {
				parts = append(parts, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteByte(ch)
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

// Implements: SYS-REQ-260820-ZTC3
func parseClause(p string) (Clause, error) {
	p = strings.TrimSpace(p)
	ops := []string{"!=", "<=", ">=", " contains ", " in ", "contains", "in", "=", "<", ">"}
	lower := p
	for _, op := range ops {
		idx := strings.Index(strings.ToLower(lower), op)
		if idx == -1 {
			continue
		}
		field := strings.TrimSpace(p[:idx])
		raw := strings.TrimSpace(p[idx+len(op):])
		raw = strings.Trim(raw, `"'`)
		c := Clause{Field: field, Value: raw}
		switch strings.TrimSpace(op) {
		case "=":
			c.Op = "eq"
		case "!=":
			c.Op = "neq"
		case "<":
			c.Op = "lt"
		case ">":
			c.Op = "gt"
		case "<=":
			c.Op = "lte"
		case ">=":
			c.Op = "gte"
		case "contains":
			c.Op = "contains"
		case "in":
			c.Op = "in"
			c.Values = splitList(raw)
		}
		if field == "" {
			return c, fmt.Errorf("missing field in %q", p)
		}
		return c, nil
	}
	return Clause{}, fmt.Errorf("invalid clause %q", p)
}

// Implements: SYS-REQ-260820-ZTC3
func splitList(s string) []string {
	s = strings.Trim(s, "[]")
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.Trim(p, `"'`))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Match reports whether rec satisfies every clause.
// Implements: SYS-REQ-260820-ZTC3
func Match(rec *record.Record, clauses []Clause) bool {
	for _, c := range clauses {
		if !matchClause(rec, c) {
			return false
		}
	}
	return true
}

// Implements: SYS-REQ-260820-ZTC3
func matchClause(rec *record.Record, c Clause) bool {
	v := rec.Get(c.Field)
	switch c.Op {
	case "eq":
		return equalish(v, c.Value)
	case "neq":
		return !equalish(v, c.Value)
	case "contains":
		if sl := record.StringSlice(v); sl != nil {
			for _, s := range sl {
				if strings.EqualFold(s, c.Value) || strings.Contains(strings.ToLower(s), strings.ToLower(c.Value)) {
					return true
				}
			}
			return false
		}
		return strings.Contains(strings.ToLower(fmt.Sprint(v)), strings.ToLower(c.Value))
	case "in":
		for _, want := range c.Values {
			if equalish(v, want) {
				return true
			}
			if sl := record.StringSlice(v); sl != nil {
				for _, s := range sl {
					if strings.EqualFold(s, want) {
						return true
					}
				}
			}
		}
		return false
	case "lt":
		return record.CompareValues(v, c.Value) < 0
	case "gt":
		return record.CompareValues(v, c.Value) > 0
	case "lte":
		return record.CompareValues(v, c.Value) <= 0
	case "gte":
		return record.CompareValues(v, c.Value) >= 0
	default:
		return false
	}
}

// Implements: SYS-REQ-260820-ZTC3
func equalish(v any, want string) bool {
	if sl := record.StringSlice(v); len(sl) == 1 && sl[0] == fmt.Sprint(v) {
		return strings.EqualFold(fmt.Sprint(v), want)
	}
	if sl := record.StringSlice(v); sl != nil {
		for _, s := range sl {
			if strings.EqualFold(s, want) {
				return true
			}
		}
	}
	return strings.EqualFold(strings.TrimSpace(fmt.Sprint(v)), want)
}

// Filter keeps records that match expr.
// Implements: SYS-REQ-260820-ZTC3
func Filter(recs []*record.Record, expr string) ([]*record.Record, error) {
	clauses, err := Parse(expr)
	if err != nil {
		return nil, err
	}
	var out []*record.Record
	for _, rec := range recs {
		if Match(rec, clauses) {
			out = append(out, rec)
		}
	}
	return out, nil
}
