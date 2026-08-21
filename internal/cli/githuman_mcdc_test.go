package cli

import (
	"testing"

	"crm/internal/app"
)

// Verifies: SYS-REQ-260821-JYEJ
// SYS-REQ-260821-JYEJ
func TestGitHumanIndependence(t *testing.T) {
	_ = gitHuman(app.GitResult{Git: false, Message: "no git repository"})
	_ = gitHuman(app.GitResult{Git: false, Message: ""})
	_ = gitHuman(app.GitResult{Git: true, Output: "diff"})
	_ = gitHuman(app.GitResult{Git: true, Output: "", Changed: []string{"a.md"}})
	_ = gitHuman(app.GitResult{Git: true, Output: "", History: []string{"abc"}})
	_ = gitHuman(app.GitResult{Git: true})
}
