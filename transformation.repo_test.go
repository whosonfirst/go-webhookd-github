package github

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/whosonfirst/go-webhookd/v3"
	"github.com/whosonfirst/go-webhookd/v3/transformation"	
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

func TestGitHubRepoTransformationWithPrepend(t *testing.T) {

	expected_repo := []byte(`#author sfomuseumbot
sfomuseum-data-flights-2020-05`)

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

	tr, err := transformation.NewTransformation(ctx, "githubrepo://?prepend_author=true")

	if err != nil {
		t.Fatalf("Failed to create new transformation, %v", err)
	}

	output, err2 := tr.Transform(ctx, body)

	if err2 != nil {
		t.Fatalf("Failed to transform message, %v", err2)
	}

	if !bytes.Equal(output, expected_repo) {
		t.Fatalf("Unexpected output: '%s'", string(output))
	}

}

func TestGitHubRepoTransformationWithHalt(t *testing.T) {

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

	tr, err := transformation.NewTransformation(ctx, "githubrepo://?halt_on_author=sfomuseumbot")

	if err != nil {
		t.Fatalf("Failed to create new transformation, %v", err)
	}

	_, err2 := tr.Transform(ctx, body)

	if err2 == nil {
		t.Fatalf("Expected error transforming message")
	}

	if err2.Code != webhookd.HaltEvent {
		t.Fatalf("Expected halt event")
	}
}
