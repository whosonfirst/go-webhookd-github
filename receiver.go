package github

// https://developer.github.com/webhooks/
// https://developer.github.com/webhooks/#payloads
// https://developer.github.com/v3/activity/events/types/#pushevent
// https://developer.github.com/v3/repos/hooks/#ping-a-hook

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	gogithub "github.com/google/go-github/github"
	"github.com/whosonfirst/go-webhookd/v3"
	"github.com/whosonfirst/go-webhookd/v3/receiver"
	"io/ioutil"
	_ "log"
	"net/http"
	"net/url"
)

func init() {

	ctx := context.Background()
	err := receiver.RegisterReceiver(ctx, "github", NewGitHubReceiver)

	if err != nil {
		panic(err)
	}
}

type GitHubReceiver struct {
	webhookd.WebhookReceiver
	secret string
	ref    string
}

func NewGitHubReceiver(ctx context.Context, uri string) (webhookd.WebhookReceiver, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	secret := q.Get("secret")
	ref := q.Get("ref")

	wh := GitHubReceiver{
		secret: secret,
		ref:    ref,
	}

	return wh, nil
}

func (wh GitHubReceiver) Receive(ctx context.Context, req *http.Request) ([]byte, *webhookd.WebhookError) {

	select {
	case <-ctx.Done():
		return nil, nil
	default:
		// pass
	}

	if req.Method != "POST" {

		code := http.StatusMethodNotAllowed
		message := "Method not allowed"

		err := &webhookd.WebhookError{Code: code, Message: message}
		return nil, err
	}

	event_type := req.Header.Get("X-GitHub-Event")

	if event_type == "" {

		code := http.StatusBadRequest
		message := "Bad Request - Missing X-GitHub-Event Header"

		err := &webhookd.WebhookError{Code: code, Message: message}
		return nil, err
	}

	sig := req.Header.Get("X-Hub-Signature")

	if sig == "" {

		code := http.StatusForbidden
		message := "Missing X-Hub-Signature required for HMAC verification"

		err := &webhookd.WebhookError{Code: code, Message: message}
		return nil, err
	}

	if event_type == "ping" {
		err := &webhookd.WebhookError{Code: -1, Message: "ping message is a no-op"}
		return nil, err
	}

	// remember that you want to configure GitHub to send webhooks as 'application/json'
	// or all this code will get confused (20190212/thisisaaronland)

	body, err := ioutil.ReadAll(req.Body)

	if err != nil {

		code := http.StatusInternalServerError
		message := err.Error()

		err := &webhookd.WebhookError{Code: code, Message: message}
		return nil, err
	}

	expectedSig, _ := GenerateSignature(string(body), wh.secret)

	if !hmac.Equal([]byte(expectedSig), []byte(sig)) {

		code := http.StatusForbidden
		message := "HMAC verification failed"

		err := &webhookd.WebhookError{Code: code, Message: message}
		return nil, err
	}

	if wh.ref != "" {

		var event gogithub.PushEvent

		err := json.Unmarshal(body, &event)

		if err != nil {
			err := &webhookd.WebhookError{Code: 999, Message: err.Error()}
			return nil, err
		}

		if wh.ref != *event.Ref {

			msg := "Invalid ref for commit"
			err := &webhookd.WebhookError{Code: 666, Message: msg}
			return nil, err
		}
	}

	/*

		So here's a thing that's not awesome: the event_type is passed in the header
		rather than anywhere in the payload body. So I don't know... maybe we need to
		change the signature of Receive method to be something like this:
		       { Payload: []byte, Extras: map[string]string }

		Which is not something that makes me "happy"... (20161016/thisisaaronland)

	*/

	return body, nil
}
