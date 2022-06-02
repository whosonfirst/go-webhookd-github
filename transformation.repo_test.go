package github

import (
	"bytes"
	"context"
	"github.com/whosonfirst/go-webhookd/v3/transformation"
	"io"
	"os"
	"testing"
)

func TestGitHubRepoTransformation(t *testing.T) {

	expected_repo := []byte("sfomuseum-data-flights-2020-05")

	msg := "fixtures/events/flights.json"
	fh, err := os.Open(msg)

	if err != nil {
		t.Fatalf("Failed to open %s, %v", msg, err)
	}

	defer fh.Close()

	body, err := io.ReadAll(fh)

	if err != nil {
		t.Fatalf("Failed to read %s, %v", msg, err)
	}

	ctx := context.Background()

	tr, err := transformation.NewTransformation(ctx, "githubrepo://")

	if err != nil {
		t.Fatalf("Failed to create new transformation, %v", err)
	}

	repo_name, err2 := tr.Transform(ctx, body)

	if err2 != nil {
		t.Fatalf("Failed to transform message, %v", err2)
	}

	if !bytes.Equal(repo_name, expected_repo) {
		t.Fatalf("Unexpected repo: %s", string(repo_name))
	}

}
