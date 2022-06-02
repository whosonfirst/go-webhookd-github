package github

import (
	"io"
	"os"
	"testing"
)

func TestGenerateSignature(t *testing.T) {

	secret := "s33kret"
	expected_sig := "sha1=3f9dc35631573aa328f0ea9c09849883859151c8"

	msg := "fixtures/events/push.json"
	fh, err := os.Open(msg)

	if err != nil {
		t.Fatalf("Failed to open %s, %v", msg, err)
	}

	defer fh.Close()

	body, err := io.ReadAll(fh)

	if err != nil {
		t.Fatalf("Failed to read %s, %v", msg, err)
	}

	sig, err := GenerateSignature(string(body), secret)

	if err != nil {
		t.Fatalf("Failed to generate signature, %v", err)
	}

	if sig != expected_sig {
		t.Fatalf("Unexpected signature: %s", sig)
	}

}

func TestUnmarshalEvent(t *testing.T) {

	msg := "fixtures/events/push.json"
	fh, err := os.Open(msg)

	if err != nil {
		t.Fatalf("Failed to open %s, %v", msg, err)
	}

	defer fh.Close()

	body, err := io.ReadAll(fh)

	if err != nil {
		t.Fatalf("Failed to read %s, %v", msg, err)
	}

	_, err = UnmarshalEvent("push", body)

	if err != nil {
		t.Fatalf("Unable to unmarshal push event, %v", err)
	}
}
