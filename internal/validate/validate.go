package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buger/recs/internal/record"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Types map[string]TypeSchema `yaml:"types"`
}

type TypeSchema struct {
	Required []string               `yaml:"required"`
	Fields   map[string]FieldSchema `yaml:"fields"`
}

type FieldSchema struct {
	Type string   `yaml:"type"`
	Enum []string `yaml:"enum"`
}

type Violation struct {
	Record string `json:"record"`
	Field  string `json:"field,omitempty"`
	Error  string `json:"error"`
	Value  any    `json:"value,omitempty"`
	Allowed []string `json:"allowed,omitempty"`
}

type Result struct {
	SchemaPresent bool         `json:"schema_present"`
	OK            bool         `json:"ok"`
	Violations    []Violation  `json:"violations,omitempty"`
}

// Implements: SYS-REQ-260820-YWV4
func LoadConfig(root string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(root, "crm.yaml"))
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Check validates records against optional schemas.
// Implements: SYS-REQ-260820-YWV4 SW-REQ-260820-8PMR
func Check(root string, recs []*record.Record) (*Result, error) {
	cfg, err := LoadConfig(root)
	if err != nil {
		if os.IsNotExist(err) {
			return &Result{SchemaPresent: false, OK: true}, nil
		}
		return nil, err
	}
	if len(cfg.Types) == 0 {
		return &Result{SchemaPresent: false, OK: true}, nil
	}
	res := &Result{SchemaPresent: true, OK: true}
	for _, rec := range recs {
		schema, ok := cfg.Types[rec.Type]
		if !ok {
			continue
		}
		for _, req := range schema.Required {
			if rec.GetString(req) == "" {
				res.OK = false
				res.Violations = append(res.Violations, Violation{
					Record: rec.ID, Field: req, Error: "missing_required",
				})
			}
		}
		for field, fs := range schema.Fields {
			v := rec.Get(field)
			if v == nil || rec.GetString(field) == "" {
				continue
			}
			if len(fs.Enum) > 0 {
				got := rec.GetString(field)
				found := false
				for _, e := range fs.Enum {
					if strings.EqualFold(e, got) {
						found = true
						break
					}
				}
				if !found {
					res.OK = false
					res.Violations = append(res.Violations, Violation{
						Record: rec.ID, Field: field, Error: "invalid_enum",
						Value: got, Allowed: fs.Enum,
					})
				}
			}
		}
	}
	if res.Violations == nil {
		res.Violations = []Violation{}
	}
	return res, nil
}

// Implements: SYS-REQ-260820-YWV4
func (v Violation) String() string {
	if v.Field != "" {
		return fmt.Sprintf("%s %s: %s", v.Record, v.Field, v.Error)
	}
	return fmt.Sprintf("%s: %s", v.Record, v.Error)
}
