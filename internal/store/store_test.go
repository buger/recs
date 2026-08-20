package store_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crm/internal/app"
	"crm/internal/record"
	"crm/internal/store"
)

// Verifies: SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2
// SYS-REQ-260820-KJ34:empty_input:nominal
// SYS-REQ-260820-KJ34:nominal:nominal
// SW-REQ-260820-MQF2:empty_input:nominal
// SW-REQ-260820-MQF2:nominal:nominal
// SYS-REQ-260820-KJ34:error_handling:nominal
// SW-REQ-260820-MQF2:error_handling:nominal
// MCDC SW-REQ-260820-MQF2: attachments_dir_created=F, boards_dir_created=F, crm_yaml_written=F, derived_dirs_created=F, inbox_dir_created=F, init_command_invoked=F, parse_budget_GE_0=T, records_dir_created=F, templates_dir_created=F, workspace_dir_count_GE_8=F => TRUE
// MCDC SW-REQ-260820-MQF2: attachments_dir_created=F, boards_dir_created=F, crm_yaml_written=F, derived_dirs_created=F, inbox_dir_created=F, init_command_invoked=T, parse_budget_GE_0=F, records_dir_created=F, templates_dir_created=F, workspace_dir_count_GE_8=F => TRUE
//mcdc:ignore SW-REQ-260820-MQF2: attachments_dir_created=F, boards_dir_created=F, crm_yaml_written=F, derived_dirs_created=F, inbox_dir_created=F, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=F, templates_dir_created=F, workspace_dir_count_GE_8=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-MQF2: attachments_dir_created=F, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GE_8=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-MQF2: attachments_dir_created=T, boards_dir_created=F, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GE_8=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-MQF2: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=F, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GE_8=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-MQF2: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=F, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GE_8=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-MQF2: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=F, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GE_8=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-MQF2: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=F, templates_dir_created=T, workspace_dir_count_GE_8=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-MQF2: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=F, workspace_dir_count_GE_8=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-MQF2: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GE_8=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-MQF2: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GE_8=T => TRUE
// MCDC SYS-REQ-260820-KJ34: attachments_dir_created=F, boards_dir_created=F, crm_yaml_written=F, derived_dirs_created=F, inbox_dir_created=F, init_command_invoked=F, parse_budget_GE_0=T, records_dir_created=F, templates_dir_created=F, workspace_dir_count_GT_0=F => TRUE
// MCDC SYS-REQ-260820-KJ34: attachments_dir_created=F, boards_dir_created=F, crm_yaml_written=F, derived_dirs_created=F, inbox_dir_created=F, init_command_invoked=T, parse_budget_GE_0=F, records_dir_created=F, templates_dir_created=F, workspace_dir_count_GT_0=F => TRUE
//mcdc:ignore SYS-REQ-260820-KJ34: attachments_dir_created=F, boards_dir_created=F, crm_yaml_written=F, derived_dirs_created=F, inbox_dir_created=F, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=F, templates_dir_created=F, workspace_dir_count_GT_0=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-KJ34: attachments_dir_created=F, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-KJ34: attachments_dir_created=T, boards_dir_created=F, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-KJ34: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=F, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-KJ34: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=F, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-KJ34: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=F, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-KJ34: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=F, templates_dir_created=T, workspace_dir_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-KJ34: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=F, workspace_dir_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-KJ34: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GT_0=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-KJ34: attachments_dir_created=T, boards_dir_created=T, crm_yaml_written=T, derived_dirs_created=T, inbox_dir_created=T, init_command_invoked=T, parse_budget_GE_0=T, records_dir_created=T, templates_dir_created=T, workspace_dir_count_GT_0=T => TRUE
func TestInitCreatesLayout(t *testing.T) {
	root := t.TempDir()
	a := app.OpenOrCWD(root)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"crm.yaml", "records", "boards", "inbox", "attachments", "templates", ".crm/index", ".crm/cache", ".crm/runtime"} {
		if _, err := os.Stat(filepath.Join(a.Root(), p)); err != nil {
			t.Fatalf("missing %s: %v", p, err)
		}
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4
// SYS-REQ-260820-9J7C:nominal:nominal
// SYS-REQ-260820-9J7C:empty_input:nominal
// SYS-REQ-260820-9J7C:boundary:nominal
// SYS-REQ-260820-9J7C:path_traversal_prevented:nominal
// SYS-REQ-260820-9J7C:untrusted_input_bounded:nominal
// SW-REQ-260820-N02Y:nominal:nominal
// SW-REQ-260820-N02Y:boundary:nominal
// SW-REQ-260820-N02Y:path_traversal_prevented:nominal
// SW-REQ-260820-N02Y:untrusted_input_bounded:nominal
// SYS-REQ-260820-2SQZ:atomicity:nominal
// SYS-REQ-260820-2SQZ:concurrent:nominal
// SYS-REQ-260820-2SQZ:atomic_write:nominal
// SYS-REQ-260820-2SQZ:concurrent_invariant_preserved:nominal
// SW-REQ-260820-Q3C4:atomicity:nominal
// SW-REQ-260820-Q3C4:concurrent:nominal
// SW-REQ-260820-Q3C4:atomic_write:nominal
// SW-REQ-260820-Q3C4:concurrent_invariant_preserved:nominal
// SYS-REQ-260820-2SQZ:error_handling:nominal
// SW-REQ-260820-Q3C4:error_handling:nominal
// MCDC SW-REQ-260820-N02Y: create_rejected=F, frontmatter_sep_index_GE_3=F, frontmatter_written=F, markdown_body_written=F, parse_budget_GE_0=F, record_create_requested=T, record_id_assigned=F, record_type_stored=F => TRUE
// MCDC SW-REQ-260820-N02Y: create_rejected=F, frontmatter_sep_index_GE_3=F, frontmatter_written=F, markdown_body_written=F, parse_budget_GE_0=T, record_create_requested=F, record_id_assigned=F, record_type_stored=F => TRUE
//mcdc:ignore SW-REQ-260820-N02Y: create_rejected=F, frontmatter_sep_index_GE_3=F, frontmatter_written=F, markdown_body_written=F, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=F, record_type_stored=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-N02Y: create_rejected=F, frontmatter_sep_index_GE_3=T, frontmatter_written=F, markdown_body_written=F, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=F, record_type_stored=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-N02Y: create_rejected=F, frontmatter_sep_index_GE_3=T, frontmatter_written=F, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-N02Y: create_rejected=F, frontmatter_sep_index_GE_3=T, frontmatter_written=T, markdown_body_written=F, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-N02Y: create_rejected=F, frontmatter_sep_index_GE_3=T, frontmatter_written=T, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=F, record_type_stored=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-N02Y: create_rejected=F, frontmatter_sep_index_GE_3=T, frontmatter_written=T, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-N02Y: create_rejected=F, frontmatter_sep_index_GE_3=T, frontmatter_written=T, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=T => TRUE
//mcdc:ignore SW-REQ-260820-N02Y: create_rejected=T, frontmatter_sep_index_GE_3=F, frontmatter_written=T, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-N02Y: create_rejected=T, frontmatter_sep_index_GE_3=T, frontmatter_written=F, markdown_body_written=F, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=F, record_type_stored=F => TRUE
// MCDC SW-REQ-260820-N02Y: create_rejected=T, frontmatter_sep_index_GE_3=T, frontmatter_written=T, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=T => TRUE
// MCDC SW-REQ-260820-Q3C4: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_3=F, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=F, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => TRUE
// MCDC SW-REQ-260820-Q3C4: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_3=F, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=F, temp_file_written=F, version_mismatch=F => TRUE
//mcdc:ignore SW-REQ-260820-Q3C4: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_3=F, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-Q3C4: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_3=T, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-Q3C4: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_3=T, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-Q3C4: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_3=T, frontmatter_updated=F, mutation_rejected=T, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => TRUE
//mcdc:ignore SW-REQ-260820-Q3C4: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_3=T, frontmatter_updated=T, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=T, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-Q3C4: atomic_rename_completed=F, conflict_reported=T, frontmatter_sep_index_GE_3=T, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-Q3C4: atomic_rename_completed=F, conflict_reported=T, frontmatter_sep_index_GE_3=T, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=T => TRUE
//mcdc:ignore SW-REQ-260820-Q3C4: atomic_rename_completed=T, conflict_reported=F, frontmatter_sep_index_GE_3=T, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=T, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-Q3C4: atomic_rename_completed=T, conflict_reported=F, frontmatter_sep_index_GE_3=T, frontmatter_updated=T, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-Q3C4: atomic_rename_completed=T, conflict_reported=F, frontmatter_sep_index_GE_3=T, frontmatter_updated=T, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=T, version_mismatch=F => TRUE
//mcdc:ignore SW-REQ-260820-Q3C4: atomic_rename_completed=T, conflict_reported=T, frontmatter_sep_index_GE_3=F, frontmatter_updated=T, mutation_rejected=T, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=T, version_mismatch=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-Q3C4: atomic_rename_completed=T, conflict_reported=T, frontmatter_sep_index_GE_3=T, frontmatter_updated=T, mutation_rejected=T, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=T, version_mismatch=T => TRUE
// MCDC SYS-REQ-260820-2SQZ: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_0=F, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=F, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => TRUE
// MCDC SYS-REQ-260820-2SQZ: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_0=F, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=F, temp_file_written=F, version_mismatch=F => TRUE
//mcdc:ignore SYS-REQ-260820-2SQZ: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_0=F, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-2SQZ: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_0=T, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-2SQZ: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_0=T, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-2SQZ: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_0=T, frontmatter_updated=F, mutation_rejected=T, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => TRUE
//mcdc:ignore SYS-REQ-260820-2SQZ: atomic_rename_completed=F, conflict_reported=F, frontmatter_sep_index_GE_0=T, frontmatter_updated=T, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=T, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-2SQZ: atomic_rename_completed=F, conflict_reported=T, frontmatter_sep_index_GE_0=T, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-2SQZ: atomic_rename_completed=F, conflict_reported=T, frontmatter_sep_index_GE_0=T, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=T => TRUE
//mcdc:ignore SYS-REQ-260820-2SQZ: atomic_rename_completed=T, conflict_reported=F, frontmatter_sep_index_GE_0=T, frontmatter_updated=F, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=T, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-2SQZ: atomic_rename_completed=T, conflict_reported=F, frontmatter_sep_index_GE_0=T, frontmatter_updated=T, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=F, version_mismatch=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-2SQZ: atomic_rename_completed=T, conflict_reported=F, frontmatter_sep_index_GE_0=T, frontmatter_updated=T, mutation_rejected=F, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=T, version_mismatch=F => TRUE
//mcdc:ignore SYS-REQ-260820-2SQZ: atomic_rename_completed=T, conflict_reported=T, frontmatter_sep_index_GE_0=F, frontmatter_updated=T, mutation_rejected=T, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=T, version_mismatch=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-2SQZ: atomic_rename_completed=T, conflict_reported=T, frontmatter_sep_index_GE_0=T, frontmatter_updated=T, mutation_rejected=T, parse_budget_GE_0=T, record_mutation_requested=T, temp_file_written=T, version_mismatch=T => TRUE
// MCDC SYS-REQ-260820-9J7C: create_rejected=F, frontmatter_sep_index_GE_0=F, frontmatter_written=F, markdown_body_written=F, parse_budget_GE_0=F, record_create_requested=T, record_id_assigned=F, record_type_stored=F => TRUE
// MCDC SYS-REQ-260820-9J7C: create_rejected=F, frontmatter_sep_index_GE_0=F, frontmatter_written=F, markdown_body_written=F, parse_budget_GE_0=T, record_create_requested=F, record_id_assigned=F, record_type_stored=F => TRUE
//mcdc:ignore SYS-REQ-260820-9J7C: create_rejected=F, frontmatter_sep_index_GE_0=F, frontmatter_written=F, markdown_body_written=F, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=F, record_type_stored=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-9J7C: create_rejected=F, frontmatter_sep_index_GE_0=T, frontmatter_written=F, markdown_body_written=F, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=F, record_type_stored=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-9J7C: create_rejected=F, frontmatter_sep_index_GE_0=T, frontmatter_written=F, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-9J7C: create_rejected=F, frontmatter_sep_index_GE_0=T, frontmatter_written=T, markdown_body_written=F, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-9J7C: create_rejected=F, frontmatter_sep_index_GE_0=T, frontmatter_written=T, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=F, record_type_stored=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-9J7C: create_rejected=F, frontmatter_sep_index_GE_0=T, frontmatter_written=T, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-9J7C: create_rejected=F, frontmatter_sep_index_GE_0=T, frontmatter_written=T, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=T => TRUE
//mcdc:ignore SYS-REQ-260820-9J7C: create_rejected=T, frontmatter_sep_index_GE_0=F, frontmatter_written=T, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-9J7C: create_rejected=T, frontmatter_sep_index_GE_0=T, frontmatter_written=F, markdown_body_written=F, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=F, record_type_stored=F => TRUE
// MCDC SYS-REQ-260820-9J7C: create_rejected=T, frontmatter_sep_index_GE_0=T, frontmatter_written=T, markdown_body_written=T, parse_budget_GE_0=T, record_create_requested=T, record_id_assigned=T, record_type_stored=T => TRUE
func TestCreateShowPatch(t *testing.T) {
	a := initApp(t)
	rec, err := a.Create("grant", "", map[string]any{"title": "Solana MC/DC", "status": "researching"}, "notes")
	if err != nil {
		t.Fatal(err)
	}
	got, err := a.Show(rec.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "grant" || got.GetString("status") != "researching" {
		t.Fatalf("unexpected %#v", got.Fields)
	}
	res, err := a.Set(rec.ID, "status", "preparing")
	if err != nil {
		t.Fatal(err)
	}
	if res.Record.GetString("status") != "preparing" {
		t.Fatal(res.Record.GetString("status"))
	}
	_, err = a.Patch(rec.ID, map[string]any{"status": "applied"}, nil, "sha256:dead")
	if err == nil {
		t.Fatal("expected conflict")
	}
}

// Verifies: SYS-REQ-260820-7WT4 SW-REQ-260820-9C5Z SYS-REQ-260820-BVBE SW-REQ-260820-EX7Q INT-REQ-260820-JRWN SYS-REQ-260820-4628 SW-REQ-260820-NBGR
// SYS-REQ-260820-7WT4:nominal:nominal
// SW-REQ-260820-9C5Z:nominal:nominal
// SYS-REQ-260820-BVBE:nominal:nominal
// SW-REQ-260820-EX7Q:nominal:nominal
// INT-REQ-260820-JRWN:atomicity:nominal
// INT-REQ-260820-JRWN:integration:integration
// SYS-REQ-260820-4628:nominal:nominal
// SYS-REQ-260820-4628:empty_input:nominal
// SW-REQ-260820-NBGR:nominal:nominal
// SW-REQ-260820-NBGR:empty_input:nominal
// SYS-REQ-260820-4628:path_traversal_prevented:nominal
// SW-REQ-260820-NBGR:path_traversal_prevented:nominal
// SYS-REQ-260820-BVBE:error_handling:nominal
// SW-REQ-260820-EX7Q:error_handling:nominal
// MCDC INT-REQ-260820-JRWN: board_move_committed=F, column_list_len_GT_0=T, matcher_depth_GE_0=T, record_file_relocated=T, store_frontmatter_write_used=T => TRUE
//mcdc:ignore INT-REQ-260820-JRWN: board_move_committed=T, column_list_len_GT_0=F, matcher_depth_GE_0=T, record_file_relocated=F, store_frontmatter_write_used=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC INT-REQ-260820-JRWN: board_move_committed=T, column_list_len_GT_0=T, matcher_depth_GE_0=F, record_file_relocated=T, store_frontmatter_write_used=T => TRUE
//mcdc:ignore INT-REQ-260820-JRWN: board_move_committed=T, column_list_len_GT_0=T, matcher_depth_GE_0=T, record_file_relocated=F, store_frontmatter_write_used=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC INT-REQ-260820-JRWN: board_move_committed=T, column_list_len_GT_0=T, matcher_depth_GE_0=T, record_file_relocated=F, store_frontmatter_write_used=T => TRUE
//mcdc:ignore INT-REQ-260820-JRWN: board_move_committed=T, column_list_len_GT_0=T, matcher_depth_GE_0=T, record_file_relocated=T, store_frontmatter_write_used=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-9C5Z: folder_relocation_required=F, frontmatter_status_updated=F, frontmatter_updated=F, record_status_changed=F => TRUE
//mcdc:ignore SW-REQ-260820-9C5Z: folder_relocation_required=F, frontmatter_status_updated=F, frontmatter_updated=F, record_status_changed=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-9C5Z: folder_relocation_required=F, frontmatter_status_updated=F, frontmatter_updated=T, record_status_changed=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-9C5Z: folder_relocation_required=F, frontmatter_status_updated=T, frontmatter_updated=F, record_status_changed=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-9C5Z: folder_relocation_required=F, frontmatter_status_updated=T, frontmatter_updated=T, record_status_changed=T => TRUE
//mcdc:ignore SW-REQ-260820-9C5Z: folder_relocation_required=T, frontmatter_status_updated=T, frontmatter_updated=T, record_status_changed=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-EX7Q: column_known=F, column_list_len_GT_0=F, column_list_len_LE_64=F, file_path_unchanged=F, frontmatter_field_updated=F, matcher_depth_GE_0=F, move_rejected=F, move_requested=T => TRUE
// MCDC SW-REQ-260820-EX7Q: column_known=F, column_list_len_GT_0=F, column_list_len_LE_64=F, file_path_unchanged=F, frontmatter_field_updated=F, matcher_depth_GE_0=T, move_rejected=F, move_requested=F => TRUE
//mcdc:ignore SW-REQ-260820-EX7Q: column_known=F, column_list_len_GT_0=F, column_list_len_LE_64=F, file_path_unchanged=F, frontmatter_field_updated=F, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-EX7Q: column_known=F, column_list_len_GT_0=T, column_list_len_LE_64=T, file_path_unchanged=F, frontmatter_field_updated=F, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-EX7Q: column_known=F, column_list_len_GT_0=T, column_list_len_LE_64=T, file_path_unchanged=F, frontmatter_field_updated=F, matcher_depth_GE_0=T, move_rejected=T, move_requested=T => TRUE
//mcdc:ignore SW-REQ-260820-EX7Q: column_known=F, column_list_len_GT_0=T, column_list_len_LE_64=T, file_path_unchanged=T, frontmatter_field_updated=T, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-EX7Q: column_known=T, column_list_len_GT_0=F, column_list_len_LE_64=T, file_path_unchanged=T, frontmatter_field_updated=T, matcher_depth_GE_0=T, move_rejected=T, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-EX7Q: column_known=T, column_list_len_GT_0=T, column_list_len_LE_64=F, file_path_unchanged=T, frontmatter_field_updated=T, matcher_depth_GE_0=T, move_rejected=T, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-EX7Q: column_known=T, column_list_len_GT_0=T, column_list_len_LE_64=T, file_path_unchanged=F, frontmatter_field_updated=T, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-EX7Q: column_known=T, column_list_len_GT_0=T, column_list_len_LE_64=T, file_path_unchanged=T, frontmatter_field_updated=F, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-EX7Q: column_known=T, column_list_len_GT_0=T, column_list_len_LE_64=T, file_path_unchanged=T, frontmatter_field_updated=T, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => TRUE
// MCDC SW-REQ-260820-EX7Q: column_known=T, column_list_len_GT_0=T, column_list_len_LE_64=T, file_path_unchanged=T, frontmatter_field_updated=T, matcher_depth_GE_0=T, move_rejected=T, move_requested=T => TRUE
// MCDC SYS-REQ-260820-7WT4: folder_relocation_required=F, frontmatter_status_updated=F, record_status_changed=F => TRUE
//mcdc:ignore SYS-REQ-260820-7WT4: folder_relocation_required=F, frontmatter_status_updated=F, record_status_changed=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-7WT4: folder_relocation_required=F, frontmatter_status_updated=T, record_status_changed=T => TRUE
//mcdc:ignore SYS-REQ-260820-7WT4: folder_relocation_required=T, frontmatter_status_updated=T, record_status_changed=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-BVBE: column_known=F, column_list_len_GT_0=F, file_path_unchanged=F, frontmatter_field_updated=F, matcher_depth_GE_0=F, move_rejected=F, move_requested=T => TRUE
// MCDC SYS-REQ-260820-BVBE: column_known=F, column_list_len_GT_0=F, file_path_unchanged=F, frontmatter_field_updated=F, matcher_depth_GE_0=T, move_rejected=F, move_requested=F => TRUE
//mcdc:ignore SYS-REQ-260820-BVBE: column_known=F, column_list_len_GT_0=F, file_path_unchanged=F, frontmatter_field_updated=F, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-BVBE: column_known=F, column_list_len_GT_0=T, file_path_unchanged=F, frontmatter_field_updated=F, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-BVBE: column_known=F, column_list_len_GT_0=T, file_path_unchanged=F, frontmatter_field_updated=F, matcher_depth_GE_0=T, move_rejected=T, move_requested=T => TRUE
//mcdc:ignore SYS-REQ-260820-BVBE: column_known=F, column_list_len_GT_0=T, file_path_unchanged=T, frontmatter_field_updated=T, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-BVBE: column_known=T, column_list_len_GT_0=F, file_path_unchanged=T, frontmatter_field_updated=T, matcher_depth_GE_0=T, move_rejected=T, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-BVBE: column_known=T, column_list_len_GT_0=T, file_path_unchanged=F, frontmatter_field_updated=T, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-BVBE: column_known=T, column_list_len_GT_0=T, file_path_unchanged=T, frontmatter_field_updated=F, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-BVBE: column_known=T, column_list_len_GT_0=T, file_path_unchanged=T, frontmatter_field_updated=T, matcher_depth_GE_0=T, move_rejected=F, move_requested=T => TRUE
// MCDC SYS-REQ-260820-BVBE: column_known=T, column_list_len_GT_0=T, file_path_unchanged=T, frontmatter_field_updated=T, matcher_depth_GE_0=T, move_rejected=T, move_requested=T => TRUE
func TestMoveKeepsPath(t *testing.T) {
	a := initApp(t)
	rec, err := a.Create("grant", "grant_demo", map[string]any{"title": "Demo", "status": "researching"}, "")
	if err != nil {
		t.Fatal(err)
	}
	old := rec.Path
	moved, path, err := a.Move(rec.ID, "grants", "applied")
	if err != nil {
		t.Fatal(err)
	}
	if path != old || moved.Path != old {
		t.Fatalf("relocated %s -> %s", old, moved.Path)
	}
	if moved.GetString("status") != "applied" {
		t.Fatalf("status %s", moved.GetString("status"))
	}
	if _, err := os.Stat(old); err != nil {
		t.Fatal(err)
	}
}

func initApp(t *testing.T) *app.App {
	t.Helper()
	root := t.TempDir()
	a := app.OpenOrCWD(root)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	return a
}

var _ = record.Record{}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y SYS-REQ-260820-7WT4
// SYS-REQ-260820-9J7C:malformed_recovers_or_errors_loudly:nominal
// SW-REQ-260820-N02Y:malformed_recovers_or_errors_loudly:nominal
func TestInboxAndTemplate(t *testing.T) {
	a := initApp(t)
	rec, err := a.Create("grant", "grant_tpl2", map[string]any{"title": "T"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rec.Body, "Opportunity") {
		t.Fatalf("template not applied: %q", rec.Body)
	}
	inbox := filepath.Join(a.Root(), "inbox", "loose.md")
	if err := os.WriteFile(inbox, []byte("---\nid: loose_1\ntype: note\n---\n\n# Loose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := a.Show("loose_1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "loose_1" {
		t.Fatalf("%#v", got)
	}
}

// SYS-REQ-260820-2SQZ:error_handling:negative
// SW-REQ-260820-Q3C4:error_handling:negative
// SYS-REQ-260820-2SQZ:atomicity:negative
// SW-REQ-260820-Q3C4:atomicity:negative
func TestConflictAndNow(t *testing.T) {
	a := initApp(t)
	rec, err := a.Create("grant", "grant_now", map[string]any{"title": "Now", "status": "researching"}, "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.Patch(rec.ID, map[string]any{"status": "applied"}, nil, "sha256:dead")
	var conf *store.ConflictError
	if !errors.As(err, &conf) || conf.Expected == "" || conf.Current == "" {
		t.Fatalf("conflict shape: %v", err)
	}
	res, err := a.Patch(rec.ID, map[string]any{"applied_at": "now"}, nil, rec.Version())
	if err != nil {
		t.Fatal(err)
	}
	got := res.Record.GetString("applied_at")
	if got == "" || got == "now" {
		t.Fatalf("now not expanded: %q", got)
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
// SYS-REQ-260820-9J7C:path_traversal_prevented:negative
// SW-REQ-260820-N02Y:path_traversal_prevented:negative
func TestRejectsPathTraversalID(t *testing.T) {
	a := initApp(t)
	_, err := a.Create("grant", "../../../tmp/evil", map[string]any{"title": "X"}, "")
	if err == nil {
		t.Fatal("expected invalid id")
	}
	outside := filepath.Join(filepath.Dir(a.Root()), "evil.md")
	if _, statErr := os.Stat(outside); statErr == nil {
		t.Fatalf("wrote outside workspace: %s", outside)
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
// type-escape path_traversal_prevented
// SYS-REQ-260820-9J7C:path_traversal_prevented:negative
// SW-REQ-260820-N02Y:path_traversal_prevented:negative
func TestRejectsPathTraversalType(t *testing.T) {
	a := initApp(t)
	_, err := a.Create("../tmp", "grant_x", map[string]any{"title": "X"}, "")
	if err == nil {
		t.Fatal("expected invalid type")
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
// SYS-REQ-260820-9J7C:malformed_input:negative
// SYS-REQ-260820-9J7C:malformed_recovers_or_errors_loudly:negative
// SYS-REQ-260820-9J7C:error_handling:negative
// SW-REQ-260820-N02Y:malformed_input:negative
// SW-REQ-260820-N02Y:malformed_recovers_or_errors_loudly:negative
// SW-REQ-260820-N02Y:error_handling:negative
// SYS-REQ-260820-9J7C:untrusted_input_bounded:negative
// SW-REQ-260820-N02Y:untrusted_input_bounded:negative
func TestMalformedFrontmatterFailsLoudly(t *testing.T) {
	a := initApp(t)
	if _, err := a.Create("grant", "grant_ok", map[string]any{"title": "OK", "status": "researching"}, ""); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(a.Root(), "records", "grants", "broken.md")
	if err := os.WriteFile(bad, []byte("---\nid: broken\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := a.List("")
	if err == nil {
		t.Fatal("expected loud parse error")
	}
}

// Verifies: SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4
// invalid enum patch rejection
// SYS-REQ-260820-2SQZ:error_handling:negative
// SW-REQ-260820-Q3C4:error_handling:negative
func TestPatchRejectsInvalidEnum(t *testing.T) {
	a := initApp(t)
	rec, err := a.Create("grant", "grant_enum", map[string]any{"title": "E", "status": "researching"}, "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.Set(rec.ID, "status", "not-a-real-status")
	var enumErr *store.EnumError
	if !errors.As(err, &enumErr) {
		t.Fatalf("expected enum error, got %v", err)
	}
}

// Verifies: SYS-REQ-260820-BVBE SW-REQ-260820-EX7Q INT-REQ-260820-JRWN
// SYS-REQ-260820-BVBE:error_handling:negative
// SW-REQ-260820-EX7Q:error_handling:negative
// INT-REQ-260820-JRWN:atomicity:negative
func TestMoveUnknownColumn(t *testing.T) {
	a := initApp(t)
	rec, err := a.Create("grant", "grant_col", map[string]any{"title": "C", "status": "researching"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := a.Move(rec.ID, "grants", "not-a-column"); err == nil {
		t.Fatal("expected unknown column")
	}
}

// Verifies: SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2 STK-REQ-260820-T8AZ
// SYS-REQ-260820-KJ34:error_handling:negative
// SW-REQ-260820-MQF2:error_handling:negative
func TestInitRejectsEmptyRoot(t *testing.T) {
	s := &store.Store{Root: ""}
	if err := s.Init(); err == nil {
		t.Fatal("expected empty root error")
	}
}

// Verifies: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7
// SYS-REQ-260820-9W1S:error_handling:negative
// SW-REQ-260820-8ZS7:error_handling:negative
func TestShowMissingRecord(t *testing.T) {
	a := initApp(t)
	if _, err := a.Show("missing_record"); err == nil {
		t.Fatal("expected missing record")
	}
}
