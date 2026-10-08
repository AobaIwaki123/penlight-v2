package repository_test

import (
	"context"
	"sync"
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

func TestADR0037_ApprovalConcurrencyConflictAndRollback(t *testing.T) {
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
	before := members[0]
	changes := model.MetadataEditChanges{
		Generation: &model.GenerationChange{Before: before.Generation, After: before.Generation + 1},
	}
	proposal := model.MetadataEditProposal{
		ID:           "prp_concurrent_approval",
		MemberID:     before.ID,
		BaseRevision: before.MetadataRevision,
		Changes:      changes,
		Status:       model.ProposalPending,
	}
	if _, err := repo.CreateMetadataEditProposal(ctx, proposal); err != nil {
		t.Fatalf("create proposal: %v", err)
	}

	var wait sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			_, errs[index] = repo.ApproveMetadataEditProposal(context.Background(), proposal.ID, nil)
		}(i)
	}
	wait.Wait()
	for i, approveErr := range errs {
		if approveErr != nil {
			t.Fatalf("concurrent approval %d failed: %v", i, approveErr)
		}
	}
	approvedMember, err := repo.GetMember(ctx, before.ID)
	if err != nil {
		t.Fatal(err)
	}
	if approvedMember.MetadataRevision != before.MetadataRevision+1 || approvedMember.Generation != before.Generation+1 {
		t.Fatalf("concurrent approval was not applied once: before=%+v after=%+v", before, approvedMember)
	}
	version, err := repo.GetMasterVersion(ctx)
	if err != nil || version.DataRevision != 2 {
		t.Fatalf("concurrent approval changed master revision incorrectly: %+v, %v", version, err)
	}

	// A proposal remains pending when a public before value no longer matches.
	conflictChanges := model.MetadataEditChanges{
		Status: &model.MemberStatusChange{Before: approvedMember.Status, After: model.StatusGraduated},
	}
	conflictProposal := model.MetadataEditProposal{
		ID:           "prp_revision_conflict",
		MemberID:     approvedMember.ID,
		BaseRevision: approvedMember.MetadataRevision,
		Changes:      conflictChanges,
		Status:       model.ProposalPending,
	}
	if _, err := repo.CreateMetadataEditProposal(ctx, conflictProposal); err != nil {
		t.Fatalf("create conflict proposal: %v", err)
	}
	if _, err := repo.DB().Exec(`UPDATE members SET status = 'hiatus' WHERE id = ?`, string(approvedMember.ID)); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ApproveMetadataEditProposal(ctx, conflictProposal.ID, nil); err == nil {
		t.Fatal("expected approval conflict")
	}
	pending, err := repo.ListMetadataEditProposals(ctx, func() *model.MetadataEditProposalStatus {
		status := model.ProposalPending
		return &status
	}())
	if err != nil || len(pending) != 1 || pending[0].ID != conflictProposal.ID {
		t.Fatalf("conflict proposal should remain pending: %+v, %v", pending, err)
	}

	// A failure after the member write rolls back both the member and proposal state.
	rollbackMember, err := repo.GetMember(ctx, approvedMember.ID)
	if err != nil {
		t.Fatal(err)
	}
	rollbackProposal := model.MetadataEditProposal{
		ID:           "prp_approval_rollback",
		MemberID:     rollbackMember.ID,
		BaseRevision: rollbackMember.MetadataRevision,
		Changes: model.MetadataEditChanges{
			Generation: &model.GenerationChange{Before: rollbackMember.Generation, After: rollbackMember.Generation + 1},
		},
		Status: model.ProposalPending,
	}
	if _, err := repo.CreateMetadataEditProposal(ctx, rollbackProposal); err != nil {
		t.Fatalf("create rollback proposal: %v", err)
	}
	if _, err := repo.DB().Exec(`DELETE FROM master_versions`); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ApproveMetadataEditProposal(ctx, rollbackProposal.ID, nil); err == nil {
		t.Fatal("expected approval rollback failure")
	}
	rolledBackMember, err := repo.GetMember(ctx, rollbackMember.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rolledBackMember.Generation != rollbackMember.Generation || rolledBackMember.MetadataRevision != rollbackMember.MetadataRevision {
		t.Fatalf("failed approval partially changed member: before=%+v after=%+v", rollbackMember, rolledBackMember)
	}
	if _, err := repo.DB().Exec(`INSERT INTO master_versions (id, version, updated_at, data_revision) VALUES ('current', ?, ?, ?)`, model.CurrentMasterVersion, "2026-10-08T00:00:00Z", version.DataRevision); err != nil {
		t.Fatal(err)
	}
	rollbackStatus := model.ProposalPending
	rollbackList, err := repo.ListMetadataEditProposals(ctx, &rollbackStatus)
	if err != nil {
		t.Fatal(err)
	}
	foundRollback := false
	for _, item := range rollbackList {
		if item.ID == rollbackProposal.ID {
			foundRollback = true
		}
	}
	if !foundRollback {
		t.Fatal("rollback proposal was not left pending")
	}
}
