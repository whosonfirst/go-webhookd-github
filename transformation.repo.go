package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	gogithub "github.com/google/go-github/v48/github"
	"github.com/whosonfirst/go-webhookd/v3"
	"github.com/whosonfirst/go-webhookd/v3/transformation"
	"net/url"
	"strconv"
)

func init() {

	ctx := context.Background()
	err := transformation.RegisterTransformation(ctx, "githubrepo", NewGitHubRepoTransformation)

	if err != nil {
		panic(err)
	}
}

// see also: https://github.com/whosonfirst/go-whosonfirst-updated/issues/8

// GitHubRepoTransformation implements the `webhookd.WebhookTransformation` interface for transforming GitHub
// commit webhook messages in to the name of the repository where the commit occurred.
type GitHubRepoTransformation struct {
	webhookd.WebhookTransformation
	// ExcludeAdditions is a boolean flag to exclude newly added files from consideration.
	ExcludeAdditions bool
	// ExcludeModifications is a boolean flag to exclude updated (modified) files from consideration.
	ExcludeModifications bool
	// ExcludeDeletions is a boolean flag to exclude updated (modified) files from consideration.
	ExcludeDeletions bool
	prepend_message  bool
	prepend_author   bool
}

// NewGitHubRepoTransformation() creates a new `GitHubRepoTransformation` instance, configured by 'uri'
// which is expected to take the form of:
//
//	githubrepo://?{PARAMETERS}
//
// Where {PARAMTERS} is:
// * `?exclude_additions` An optional boolean value to exclude newly added files from consideration.
// * `?exclude_modifications` An optional boolean value to exclude update (modified) files from consideration.
// * `?exclude_deletions` An optional boolean value to exclude deleted files from consideration.
// * `?prepend_message` An optional boolean value to prepend the commit message to the final output. This takes the form of '#message {COMMIT_MESSAGE}'
// * `?prepend_author` An optional boolean value to prepend the name of the commit author to the final output. This takes the form of '#author {COMMIT_AUTHOR}'
func NewGitHubRepoTransformation(ctx context.Context, uri string) (webhookd.WebhookTransformation, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	str_additions := q.Get("exclude_additions")
	str_modifications := q.Get("exclude_modifications")
	str_deletions := q.Get("exclude_deletions")
	str_message := q.Get("prepend_message")
	str_author := q.Get("prepend_author")

	exclude_additions := false
	exclude_modifications := false
	exclude_deletions := false

	prepend_message := false
	prepend_author := false

	if str_additions != "" {

		v, err := strconv.ParseBool(str_additions)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", str_additions, err)
		}

		exclude_additions = v
	}

	if str_modifications != "" {

		v, err := strconv.ParseBool(str_modifications)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", str_modifications, err)
		}

		exclude_modifications = v
	}

	if str_deletions != "" {

		v, err := strconv.ParseBool(str_deletions)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", str_deletions, err)
		}

		exclude_deletions = v
	}

	if str_message != "" {

		v, err := strconv.ParseBool(str_message)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", str_message, err)
		}

		prepend_message = v
	}

	if str_author != "" {

		v, err := strconv.ParseBool(str_author)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", str_author, err)
		}

		prepend_author = v
	}

	p := GitHubRepoTransformation{
		ExcludeAdditions:     exclude_additions,
		ExcludeModifications: exclude_modifications,
		ExcludeDeletions:     exclude_deletions,
		prepend_message:      prepend_message,
		prepend_author:       prepend_author,
	}

	return &p, nil
}

// Transform() transforms 'body' (which is assumed to be a GitHub commit webhook message) in to name of the repository
// where the commit occurred.
func (p *GitHubRepoTransformation) Transform(ctx context.Context, body []byte) ([]byte, *webhookd.WebhookError) {

	select {
	case <-ctx.Done():
		return nil, nil
	default:
		// pass
	}

	var event gogithub.PushEvent

	err := json.Unmarshal(body, &event)

	if err != nil {
		err := &webhookd.WebhookError{Code: 999, Message: err.Error()}
		return nil, err
	}

	buf := new(bytes.Buffer)

	repo := event.Repo
	repo_name := *repo.Name

	has_updates := false

	for _, c := range event.Commits {

		if !p.ExcludeAdditions {

			if len(c.Added) > 0 {
				has_updates = true
			}
		}

		if !p.ExcludeModifications {

			if len(c.Modified) > 0 {
				has_updates = true
			}
		}

		if !p.ExcludeDeletions {

			if len(c.Removed) > 0 {
				has_updates = true
			}
		}
	}

	if has_updates {

		if p.prepend_message {
			msg := fmt.Sprintf("#message %s\n", *event.HeadCommit.Message)
			buf.WriteString(msg)
		}

		if p.prepend_author {
			msg := fmt.Sprintf("#author %s\n", *event.HeadCommit.Author.Name)
			buf.WriteString(msg)
		}

		buf.WriteString(repo_name)
	}

	return buf.Bytes(), nil
}
