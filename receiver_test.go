package github

import (
	"bytes"
	"context"
	"fmt"
	"github.com/whosonfirst/go-webhookd/v3/receiver"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
)

func TestGitHubReceiver(t *testing.T) {

	secret := "s33kret"

	receiver_uri := fmt.Sprintf("github://?secret=%s", secret)

	ctx := context.Background()

	r, err := receiver.NewReceiver(ctx, receiver_uri)

	if err != nil {
		t.Fatalf("Failed to create new receiver, %v", err)
	}

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

	req, err := http.NewRequest("POST", "http://localhost:8080/github", bytes.NewReader(body))

	if err != nil {
		t.Fatalf("Failed to create new request, %v", err)
	}

	req.Header.Set("X-GitHub-Event", "debug")
	req.Header.Set("X-Hub-Signature", sig)

	req.Header.Add("Content-Length", strconv.Itoa(len(body)))

	body2, err2 := r.Receive(ctx, req)

	if err2 != nil {
		t.Fatalf("Failed to receive message, %v", err)
	}

	if !bytes.Equal(body2, body) {
		t.Fatalf("Unexpected output '%s'", string(body2))
	}
}
