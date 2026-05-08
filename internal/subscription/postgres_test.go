package subscription

import "testing"

func TestPostgresRepositoryImplementsRepository(t *testing.T) {
	var _ Repository = (*PostgresRepository)(nil)
}

func TestNewPostgresRepository(t *testing.T) {
	repo := NewPostgresRepository(nil)
	if repo == nil {
		t.Fatal("NewPostgresRepository() returned nil")
	}
}
