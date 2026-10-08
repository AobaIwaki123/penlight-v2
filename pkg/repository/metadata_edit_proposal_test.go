package repository_test

import (
	"context"
	"testing"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
	"github.com/aobaiwaki/penlight-v2/seeds"
)

func TestADR0037_ProposalStorageBoundaryAndConstraints(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()
	seedSQL, err := seeds.FS.ReadFile("seed.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.DB().Exec(string(seedSQL)); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	members, err := repo.ListMembers(ctx)
	if err != nil || len(members) == 0 {
		t.Fatalf("members: %v, %v", members, err)
	}
	memberID := members[0].ID
	before, err := repo.GetMember(ctx, memberID)
	if err != nil || before.MetadataRevision != 1 {
		t.Fatalf("initial member revision: %+v, %v", before, err)
	}
	versionBefore, err := repo.GetMasterVersion(ctx)
	if err != nil || versionBefore.DataRevision != 1 {
		t.Fatalf("initial master revision: %+v, %v", versionBefore, err)
	}
	insert := `INSERT INTO metadata_edit_proposals
		(id, member_id, base_revision, changes_json, status, submitted_at, approved_at, rejected_at)
		VALUES (?, ?, ?, ?, ?, '2026-10-08T00:00:00Z', ?, ?)`
	if _, err := repo.DB().Exec(insert, "prp_pending", memberID, 1,
		`{"generation":{"before":1,"after":2}}`, model.ProposalPending, nil, nil); err != nil {
		t.Fatal(err)
	}
	after, err := repo.GetMember(ctx, memberID)
	if err != nil || after.Generation != before.Generation || after.MetadataRevision != before.MetadataRevision {
		t.Fatalf("pending proposal changed public member: %+v, %v", after, err)
	}
	versionAfter, err := repo.GetMasterVersion(ctx)
	if err != nil || *versionAfter != *versionBefore {
		t.Fatalf("pending proposal changed public master version: %+v, %v", versionAfter, err)
	}

	tests := []struct {
		name               string
		memberID           model.ID
		revision           int
		json               string
		status             model.MetadataEditProposalStatus
		approved, rejected any
	}{
		{"approved_without_timestamp", memberID, 1, `{}`, model.ProposalApproved, nil, nil},
		{"rejected_without_timestamp", memberID, 1, `{}`, model.ProposalRejected, nil, nil},
		{"pending_with_timestamp", memberID, 1, `{}`, model.ProposalPending, "2026-10-08T00:00:00Z", nil},
		{"both_decisions", memberID, 1, `{}`, model.ProposalApproved, "2026-10-08T00:00:00Z", "2026-10-08T00:00:00Z"},
		{"unknown_status", memberID, 1, `{}`, "unknown", nil, nil},
		{"zero_revision", memberID, 0, `{}`, model.ProposalPending, nil, nil},
		{"invalid_json", memberID, 1, `{`, model.ProposalPending, nil, nil},
		{"missing_member", "mem_missing", 1, `{}`, model.ProposalPending, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := repo.DB().Exec(insert, "prp_"+tt.name, tt.memberID, tt.revision, tt.json, tt.status, tt.approved, tt.rejected); err == nil {
				t.Fatal("expected database constraint violation")
			}
		})
	}
	for _, status := range []model.MetadataEditProposalStatus{model.ProposalApproved, model.ProposalRejected} {
		var approved, rejected any
		if status == model.ProposalApproved {
			approved = "2026-10-08T00:00:00Z"
		} else {
			rejected = "2026-10-08T00:00:00Z"
		}
		if _, err := repo.DB().Exec(insert, "prp_"+string(status), memberID, 1, `{}`, status, approved, rejected); err != nil {
			t.Fatalf("valid %s decision rejected: %v", status, err)
		}
	}
	if _, err := repo.DB().Exec("DELETE FROM members WHERE id = ?", memberID); err == nil {
		t.Fatal("proposal history must prevent member deletion")
	}
}
