package usecase

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/config"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
)

type editionStoreFake struct {
	stored *domain.StoredEditionConfig
	err    error
}

func (f *editionStoreFake) GetStoredEditionConfig(context.Context) (*domain.StoredEditionConfig, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.stored, nil
}

func (f *editionStoreFake) UpdateStoredEditionConfig(_ context.Context, stored *domain.StoredEditionConfig) error {
	if f.err != nil {
		return f.err
	}
	f.stored = stored
	return nil
}

func newEditionTestUsecase(store EditionStore) *EditionUsecase {
	return NewEditionUsecase(store, log.NewLogger(&config.Config{}))
}

func TestEditionFallbacks(t *testing.T) {
	tests := []struct {
		name   string
		stored *domain.StoredEditionConfig
		err    error
	}{
		{name: "missing", err: gorm.ErrRecordNotFound},
		{name: "corrupt", stored: nil, err: errors.New("invalid json")},
		{name: "schema", stored: &domain.StoredEditionConfig{SchemaVersion: 99, EditionID: domain.EditionLegal}},
		{name: "edition", stored: &domain.StoredEditionConfig{SchemaVersion: 1, EditionID: "unknown"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newEditionTestUsecase(&editionStoreFake{stored: tt.stored, err: tt.err}).Get(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if got.EditionID != domain.EditionCommon {
				t.Fatalf("got %q, want common", got.EditionID)
			}
		})
	}
}

func TestEditionOldVersionAndOverrides(t *testing.T) {
	name := "Legacy Legal Wiki"
	store := &editionStoreFake{stored: &domain.StoredEditionConfig{
		SchemaVersion: 1, EditionID: domain.EditionLegal, EditionVersion: "0.1.0",
		Overrides: domain.EditionOverrides{ProductName: &name},
	}}
	got, err := newEditionTestUsecase(store).Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.EditionID != domain.EditionLegal || got.EditionVersion != "1.0.0" || got.ProductName != name {
		t.Fatalf("unexpected resolved config: %+v", got)
	}
}

func TestEditionRejectsInvalidOverrides(t *testing.T) {
	tooLong := make([]byte, 16001)
	for _, req := range []*domain.UpdateEditionReq{
		{EditionID: domain.EditionLegal, Overrides: domain.EditionOverrides{EnabledFeatures: []string{"admin"}}},
		{EditionID: domain.EditionLegal, Overrides: domain.EditionOverrides{Terminology: map[string]string{"secret": "x"}}},
		{EditionID: domain.EditionLegal, Overrides: domain.EditionOverrides{DefaultPrompts: &domain.EditionPrompts{Chat: string(tooLong)}}},
	} {
		if err := validateEditionRequest(req); err == nil {
			t.Fatalf("request was accepted: %+v", req)
		}
	}
}

func TestEditionUpdateFailureDoesNotReplaceStoredValue(t *testing.T) {
	old := &domain.StoredEditionConfig{SchemaVersion: 1, EditionID: domain.EditionCommon, EditionVersion: "1.0.0"}
	store := &editionStoreFake{stored: old, err: errors.New("write failed")}
	_, err := newEditionTestUsecase(store).Update(context.Background(), 1001, &domain.UpdateEditionReq{EditionID: domain.EditionLegal})
	if err == nil {
		t.Fatal("expected write failure")
	}
	if store.stored != old {
		t.Fatal("stored value changed after failed update")
	}
}
