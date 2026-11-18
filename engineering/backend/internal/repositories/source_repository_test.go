package repositories_test

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/testutil"
)

// loadSourceUnscoped reads a source including soft-deleted rows.
func loadSourceUnscoped(t *testing.T, db *gorm.DB, id uint) models.Source {
	t.Helper()
	var src models.Source
	if err := db.Unscoped().First(&src, id).Error; err != nil {
		t.Fatalf("failed to load source %d: %v", id, err)
	}
	return src
}

func countRowsUnscoped(t *testing.T, db *gorm.DB, model interface{}) int64 {
	t.Helper()
	var n int64
	if err := db.Unscoped().Model(model).Count(&n).Error; err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	return n
}

// CreateSource

func TestStore_CreateSource(t *testing.T) {
	tests := []struct {
		name string
		// seed prepares existing rows; it returns the id of a row the case
		// expects to be reused (0 if a new row should be created).
		seed       func(t *testing.T, db *gorm.DB) uint
		input      models.Source
		wantErr    bool
		wantActive bool
		wantType   models.SourceType
		wantRows   int64 // rows in source after the call, deleted ones included
	}{
		{
			name:       "creates a new active source",
			input:      models.Source{Name: "Daily News", Type: models.SourceTypeManual, IsActive: true},
			wantActive: true,
			wantType:   models.SourceTypeManual,
			wantRows:   1,
		},
		{
			// default:true would turn a plain Create of false into true.
			name:       "creates a new inactive source",
			input:      models.Source{Name: "Daily News", Type: models.SourceTypeManual, IsActive: false},
			wantActive: false,
			wantType:   models.SourceTypeManual,
			wantRows:   1,
		},
		{
			name: "a live source with the same name is rejected",
			seed: func(t *testing.T, db *gorm.DB) uint {
				seedSource(t, db, "Rooster")
				return 0
			},
			input:    models.Source{Name: "Rooster", Type: models.SourceTypeAuto, IsActive: true},
			wantErr:  true,
			wantRows: 1,
		},
		{
			name: "a deleted source with the same name is restored with the new settings",
			seed: func(t *testing.T, db *gorm.DB) uint {
				id := seedSourceOf(t, db, "Daily News", models.SourceTypeManual, true)
				if err := db.Delete(&models.Source{}, id).Error; err != nil {
					t.Fatalf("failed to soft-delete source: %v", err)
				}
				return id
			},
			input:      models.Source{Name: "Daily News", Type: models.SourceTypeAuto, IsActive: false},
			wantActive: false,
			wantType:   models.SourceTypeAuto,
			wantRows:   1,
		},
		{
			name: "a different deleted name does not affect creation",
			seed: func(t *testing.T, db *gorm.DB) uint {
				id := seedSource(t, db, "Rooster")
				if err := db.Delete(&models.Source{}, id).Error; err != nil {
					t.Fatalf("failed to soft-delete source: %v", err)
				}
				return 0
			},
			input:      models.Source{Name: "Ikman", Type: models.SourceTypeAuto, IsActive: true},
			wantActive: true,
			wantType:   models.SourceTypeAuto,
			wantRows:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.SetupTestDB(t)
			store := repositories.NewStore(db)

			var reuseID uint
			if tt.seed != nil {
				reuseID = tt.seed(t, db)
			}

			src := tt.input
			err := store.CreateSource(context.Background(), &src)

			if n := countRowsUnscoped(t, db, &models.Source{}); n != tt.wantRows {
				t.Errorf("expected %d source row(s), got %d", tt.wantRows, n)
			}
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if reuseID != 0 && src.ID != reuseID {
				t.Errorf("expected the deleted row %d to be reused, got id %d", reuseID, src.ID)
			}

			saved := loadSourceUnscoped(t, db, src.ID)
			if saved.DeletedAt.Valid {
				t.Error("expected the source not to be deleted")
			}
			if saved.IsActive != tt.wantActive {
				t.Errorf("expected is_active=%t, got %t", tt.wantActive, saved.IsActive)
			}
			if saved.Type != tt.wantType {
				t.Errorf("expected type %q, got %q", tt.wantType, saved.Type)
			}
			if src.IsActive != tt.wantActive {
				t.Errorf("returned source: expected is_active=%t, got %t", tt.wantActive, src.IsActive)
			}
		})
	}
}

// A restored source keeps its id, so jobs saved before the delete still point at it.
func TestStore_CreateSource_RestoreKeepsJobsLinked(t *testing.T) {
	db := testutil.SetupTestDB(t)
	store := repositories.NewStore(db)
	ctx := context.Background()

	id := seedSourceOf(t, db, "Daily News", models.SourceTypeManual, true)
	runID := seedCrawlerRun(t, db)
	jobID := seedActiveJob(t, db, "old job", 1, &id, &runID, "bucket_a")
	if err := store.Sources().Delete(ctx, id); err != nil {
		t.Fatalf("failed to delete source: %v", err)
	}

	src := models.Source{Name: "Daily News", Type: models.SourceTypeManual, IsActive: true}
	if err := store.CreateSource(ctx, &src); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := loadMeta(t, db, jobID)
	if meta.SourceID == nil || *meta.SourceID != src.ID {
		t.Errorf("expected the job to stay linked to source %d, got %v", src.ID, meta.SourceID)
	}
	if _, err := store.GetSourceByName(ctx, "Daily News"); err != nil {
		t.Errorf("expected the restored source to be found by name, got %v", err)
	}
}

// Delete (generic soft delete)

func TestStore_Sources_Delete(t *testing.T) {
	tests := []struct {
		name    string
		missing bool // delete an id that was never created
		twice   bool // delete the same source a second time
		wantErr error
	}{
		{name: "soft-deletes a source"},
		{name: "unknown id is not found", missing: true, wantErr: gorm.ErrRecordNotFound},
		{name: "already deleted is not found", twice: true, wantErr: gorm.ErrRecordNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.SetupTestDB(t)
			store := repositories.NewStore(db)
			ctx := context.Background()

			id := seedSource(t, db, "Rooster")
			target := id
			if tt.missing {
				target = 9999
			}
			if tt.twice {
				if err := store.Sources().Delete(ctx, id); err != nil {
					t.Fatalf("first delete failed: %v", err)
				}
			}

			err := store.Sources().Delete(ctx, target)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !loadSourceUnscoped(t, db, id).DeletedAt.Valid {
				t.Error("expected deleted_at to be set")
			}

			// Hidden from every normal lookup.
			if _, err := store.Sources().GetByID(ctx, id); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Errorf("GetByID: expected not found, got %v", err)
			}
			if _, err := store.GetSourceByName(ctx, "Rooster"); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Errorf("GetSourceByName: expected not found, got %v", err)
			}
			if list, err := store.ListSources(ctx, repositories.SourceFilter{}); err != nil || len(list) != 0 {
				t.Errorf("ListSources: expected no sources, got %d (err %v)", len(list), err)
			}
			if _, err := store.Sources().Update(ctx, id, map[string]interface{}{"is_active": false}); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Errorf("Update: expected not found, got %v", err)
			}
		})
	}
}
