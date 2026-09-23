package service

import (
	"context"
	"testing"

	"github.com/linkasu/linka.type-backend/internal/models"
	"github.com/linkasu/linka.type-backend/internal/store"
)

type replaceStore struct {
	store.Store
	categories   []models.Category
	statements   []models.Statement
	replaceCalls int
	changes      []models.ChangeEvent
}

func (s *replaceStore) ListCategories(_ context.Context, _ string) ([]models.Category, error) {
	return append([]models.Category(nil), s.categories...), nil
}

func (s *replaceStore) ListStatements(_ context.Context, _, _ string) ([]models.Statement, error) {
	return append([]models.Statement(nil), s.statements...), nil
}

func (s *replaceStore) ReplaceStatements(_ context.Context, _, _ string, statements []models.Statement) error {
	s.replaceCalls++
	s.statements = append([]models.Statement(nil), statements...)
	return nil
}

func (s *replaceStore) AppendChange(_ context.Context, _ string, change models.ChangeEvent) error {
	s.changes = append(s.changes, change)
	return nil
}

func TestNormalizeStatementLines(t *testing.T) {
	got, duplicates := normalizeStatementLines(" first\r\n\r second \rfirst\nFIRST\r\n")
	want := []string{"first", "second", "FIRST"}
	if duplicates != 1 || len(got) != len(want) {
		t.Fatalf("got %#v, duplicates %d", got, duplicates)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestReplaceStatementsConfirmationNoOpAndRace(t *testing.T) {
	ctx := context.Background()
	backend := &replaceStore{categories: []models.Category{{ID: "cat"}}, statements: []models.Statement{
		{ID: "one", CategoryID: "cat", Text: "one"},
		{ID: "two", CategoryID: "cat", Text: "two"},
	}}
	svc := &Service{Store: backend}

	confirmed, err := svc.ReplaceStatements(ctx, "user", "cat", "two\nthree", "")
	if err != nil || confirmed.Applied || confirmed.Summary.Removed != 1 || confirmed.ConfirmationToken == "" {
		t.Fatalf("confirmation = %#v, err = %v", confirmed, err)
	}

	applied, err := svc.ReplaceStatements(ctx, "user", "cat", "two\nthree", confirmed.ConfirmationToken)
	if err != nil || !applied.Applied || backend.replaceCalls != 1 {
		t.Fatalf("applied = %#v, calls = %d, err = %v", applied, backend.replaceCalls, err)
	}
	if len(applied.Statements) != 2 || applied.Statements[0].Text != "two" || applied.Statements[1].Text != "three" {
		t.Fatalf("statements = %#v", applied.Statements)
	}
	if len(backend.changes) != 1 || backend.changes[0].Op != "statements_replace" {
		t.Fatalf("changes = %#v", backend.changes)
	}

	noOp, err := svc.ReplaceStatements(ctx, "user", "cat", "two\nthree", "")
	if err != nil || !noOp.Applied || backend.replaceCalls != 1 {
		t.Fatalf("no-op = %#v, calls = %d, err = %v", noOp, backend.replaceCalls, err)
	}

	raceBackend := &replaceStore{categories: []models.Category{{ID: "cat"}}, statements: []models.Statement{
		{ID: "one", CategoryID: "cat", Text: "one"},
		{ID: "two", CategoryID: "cat", Text: "two"},
	}}
	raceService := &Service{Store: raceBackend}
	raceConfirmation, err := raceService.ReplaceStatements(ctx, "user", "cat", "two\nthree", "")
	if err != nil || raceConfirmation.Applied {
		t.Fatalf("race confirmation = %#v, err = %v", raceConfirmation, err)
	}
	raceBackend.statements = append(raceBackend.statements, models.Statement{ID: "four", CategoryID: "cat", Text: "four"})
	stale, err := raceService.ReplaceStatements(ctx, "user", "cat", "two\nthree", raceConfirmation.ConfirmationToken)
	if err != nil || stale.Applied || stale.ConfirmationToken == "" || stale.ConfirmationToken == raceConfirmation.ConfirmationToken || raceBackend.replaceCalls != 0 {
		t.Fatalf("stale = %#v, calls = %d, err = %v", stale, raceBackend.replaceCalls, err)
	}
}

func TestReplaceStatementsRejectsMissingCategory(t *testing.T) {
	backend := &replaceStore{}
	result, err := (&Service{Store: backend}).ReplaceStatements(context.Background(), "user", "missing", "one", "")
	if err != store.ErrNotFound || result.Applied || backend.replaceCalls != 0 {
		t.Fatalf("result = %#v, calls = %d, err = %v", result, backend.replaceCalls, err)
	}
}
