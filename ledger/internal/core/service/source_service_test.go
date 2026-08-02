package service

import (
	"context"
	"testing"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

type fakeSourceRepo struct {
	created []string
}

func (f *fakeSourceRepo) Create(_ context.Context, name, kind string) (*domain.Source, error) {
	f.created = append(f.created, name)
	return &domain.Source{ID: 1, Name: name, Kind: kind}, nil
}
func (f *fakeSourceRepo) GetByName(_ context.Context, name string) (*domain.Source, error) {
	return nil, nil
}
func (f *fakeSourceRepo) List(_ context.Context) ([]*domain.Source, error) { return nil, nil }

func TestSourceServiceRejectsInvalidKind(t *testing.T) {
	repo := &fakeSourceRepo{}
	svc := NewSourceService(repo)

	_, err := svc.Add(context.Background(), "국민카드", "creditcard")
	if err == nil {
		t.Fatal("want error for invalid kind, got nil")
	}
	if len(repo.created) != 0 {
		t.Errorf("repo should not be called on invalid kind")
	}
}

func TestSourceServiceAddValidKind(t *testing.T) {
	svc := NewSourceService(&fakeSourceRepo{})
	s, err := svc.Add(context.Background(), "국민카드", domain.SourceKindCard)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if s.Name != "국민카드" {
		t.Errorf("unexpected source: %+v", s)
	}
}
