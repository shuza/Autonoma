package store

import (
	"context"
	"testing"
)

func TestNewPostgresStoreRequiresDatabaseURL(t *testing.T) {
	t.Parallel()

	_, err := NewPostgresStore(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty database url")
	}
}
