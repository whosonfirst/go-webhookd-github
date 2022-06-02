package github

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"github.com/whosonfirst/go-webhookd/v3/transformation"
	"io"
	"os"
	"testing"
)

func TestGitHubCommitsTransformation(t *testing.T) {

	expected_hash := "696a396febbe79310b6f54e576c753a28e97ccbbfa1614e72de050d041cc81c5"
	expected_rows := 1607

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

	tr, err := transformation.NewTransformation(ctx, "githubcommits://")

	if err != nil {
		t.Fatalf("Failed to create new transformation, %v", err)
	}

	data, err2 := tr.Transform(ctx, body)

	if err2 != nil {
		t.Fatalf("Failed to transform message, %v", err2)
	}

	sum := sha256.Sum256([]byte(data))
	hash := fmt.Sprintf("%x", sum)

	if hash != expected_hash {
		t.Fatalf("Unexpected hash of commit data: %s", hash)
	}

	r := bytes.NewReader(data)

	csv_r := csv.NewReader(r)

	rows, err := csv_r.ReadAll()

	if err != nil {
		t.Fatalf("Failed to read CSV data, %v", err)
	}

	if len(rows) != expected_rows {
		t.Fatalf("Unexpected row coutn: %d", len(rows))
	}
}
