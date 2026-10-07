//go:build ignore

// Generate the ADR-0037 migration from the Go model (Ref: ADR-0004).
package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
)

func column(field reflect.StructField) string {
	return strings.Split(field.Tag.Get("db"), ",")[0] + " " + field.Tag.Get("sql")
}

func main() {
	var sql strings.Builder
	sql.WriteString("-- Generated from pkg/model by scripts/gen-proposal-migration.go. DO NOT EDIT.\n-- Ref: ADR-0004, ADR-0036, ADR-0037\n\n")
	memberField, _ := reflect.TypeOf(model.Member{}).FieldByName("MetadataRevision")
	versionField, _ := reflect.TypeOf(model.MasterVersion{}).FieldByName("DataRevision")
	fmt.Fprintf(&sql, "ALTER TABLE members ADD COLUMN %s;\n", column(memberField))
	fmt.Fprintf(&sql, "ALTER TABLE master_versions ADD COLUMN %s;\n\n", column(versionField))
	sql.WriteString("CREATE TABLE metadata_edit_proposals (\n")
	proposal := reflect.TypeOf(model.MetadataEditProposal{})
	for i := 0; i < proposal.NumField(); i++ {
		fmt.Fprintf(&sql, "    %s,\n", column(proposal.Field(i)))
	}
	fmt.Fprintf(&sql, "    CHECK (status IN ('%s', '%s', '%s')),\n", model.ProposalPending, model.ProposalApproved, model.ProposalRejected)
	fmt.Fprintf(&sql, `    CHECK (
        (status = '%s' AND approved_at IS NULL AND rejected_at IS NULL AND approver_user_id IS NULL AND rejection_reason IS NULL)
        OR (status = '%s' AND approved_at IS NOT NULL AND rejected_at IS NULL AND rejection_reason IS NULL)
        OR (status = '%s' AND approved_at IS NULL AND rejected_at IS NOT NULL)
    )
);

CREATE INDEX idx_metadata_edit_proposals_status_submitted
    ON metadata_edit_proposals(status, submitted_at, id);
CREATE INDEX idx_metadata_edit_proposals_member_status
    ON metadata_edit_proposals(member_id, status, submitted_at, id);
`, model.ProposalPending, model.ProposalApproved, model.ProposalRejected)
	if err := os.WriteFile("migrations/000004_add_metadata_edit_proposals.up.sql", []byte(sql.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
