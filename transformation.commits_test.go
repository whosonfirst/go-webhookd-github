package github

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"github.com/whosonfirst/go-webhookd/v3"
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

func TestGitHubCommitsTransformationWithPrepend(t *testing.T) {

	expected_hash := "1f7cff82034f23c836682db2102e2a79541dcc9afb64e2fe189b306c8591c004"
	expected_rows := 1608

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

	tr, err := transformation.NewTransformation(ctx, "githubcommits://?prepend_message=true")

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

	r = bytes.NewReader(data)
	csv_r = csv.NewReader(r)

	row, err := csv_r.Read()

	if err != nil {
		t.Fatalf("Failed to read row, %v", err)
	}

	if row[0] != "#message append SWIM data for 20200521" {
		t.Fatalf("Unexpected value for first row '%s'", row[0])
	}

}

func TestGitHubCommitsTransformationWithHalt(t *testing.T) {

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

	tr, err := transformation.NewTransformation(ctx, "githubcommits://?halt_on_message=SWIM")

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
