package server

import (
	"path/filepath"
	"testing"

	"github.com/aobaiwaki/penlight-v2/migrations"
	"github.com/aobaiwaki/penlight-v2/pkg/repository"
	"github.com/aobaiwaki/penlight-v2/seeds"
)

func TestADR0037_MigrateExistingDatabaseAndRestart(t *testing.T) {
	repo, err := repository.NewSQLiteRepository(filepath.Join(t.TempDir(), "upgrade.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	for _, name := range []string{"000001_init.up.sql", "000002_add_series_and_songs.up.sql", "000003_add_member_verified_at.up.sql"} {
		data, err := migrations.FS.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repo.DB().Exec(string(data)); err != nil {
			t.Fatal(err)
		}
	}
	seed, err := seeds.FS.ReadFile("seed.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.DB().Exec(string(seed)); err != nil {
		t.Fatal(err)
	}
	// Simulate the migration history of an existing server.
	if _, err := repo.DB().Exec(`CREATE TABLE schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL);
		INSERT INTO schema_migrations VALUES ('000001_init.up.sql', '2026-10-08'), ('000002_add_series_and_songs.up.sql', '2026-10-08'), ('000003_add_member_verified_at.up.sql', '2026-10-08');`); err != nil {
		t.Fatal(err)
	}
	var before int
	if err := repo.DB().QueryRow("SELECT COUNT(*) FROM members").Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := applyMigrations(repo.DB()); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.DB().Exec("UPDATE members SET metadata_revision = 7; UPDATE master_versions SET data_revision = 9;"); err != nil {
		t.Fatal(err)
	}
	if err := applyMigrations(repo.DB()); err != nil {
		t.Fatalf("restart must not reapply migration: %v", err)
	}
	var count, minRevision, maxRevision, dataRevision int
	if err := repo.DB().QueryRow("SELECT COUNT(*), MIN(metadata_revision), MAX(metadata_revision) FROM members").Scan(&count, &minRevision, &maxRevision); err != nil {
		t.Fatal(err)
	}
	if err := repo.DB().QueryRow("SELECT data_revision FROM master_versions WHERE id = 'current'").Scan(&dataRevision); err != nil {
		t.Fatal(err)
	}
	if count != before || minRevision != 7 || maxRevision != 7 || dataRevision != 9 {
		t.Fatalf("restart changed master: count=%d (before=%d), revision=%d..%d, data_revision=%d", count, before, minRevision, maxRevision, dataRevision)
	}
}
