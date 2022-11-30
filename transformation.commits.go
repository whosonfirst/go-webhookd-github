package github

import (
	"bytes"
	"context"
	"encoding/csv"
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
	err := transformation.RegisterTransformation(ctx, "githubcommits", NewGitHubCommitsTransformation)

	if err != nil {
		panic(err)
	}
}

// see also: https://github.com/whosonfirst/go-whosonfirst-updated/issues/8

// GitHubCommitsTransformation implements the `webhookd.WebhookTransformation` interface for transforming GitHub
// commit webhook messages in to CSV data containing: the commit hash, the name of the repository and the path
// to the file commited.
type GitHubCommitsTransformation struct {
	webhookd.WebhookTransformation
	// ExcludeAdditions is a boolean flag to exclude newly added files from the final output.
	ExcludeAdditions bool
	// ExcludeModifications is a boolean flag to exclude updated (modified) files from the final output.
	ExcludeModifications bool
	// ExcludeDeletions is a boolean flag to exclude deleted files from the final output.
	ExcludeDeletions bool
	prepend_message  bool
	prepend_author   bool
}

// NewGitHubCommitsTransformation() creates a new `GitHubCommitsTransformation` instance, configured by 'uri'
// which is expected to take the form of:
//
//	githubcommits://?{PARAMETERS}
//
// Where {PARAMTERS} is:
// * `?exclude_additions` An optional boolean value to exclude newly added files from the final output.
// * `?exclude_modifications` An optional boolean value to exclude update (modified) files from the final output.
// * `?exclude_deletions` An optional boolean value to exclude deleted files from the final output.
// * `?prepend_message` An optional boolean value to prepend the commit message to the final output. This takes the form of '#message,{COMMIT_MESSAGE},'
// * `?prepend_author` An optional boolean value to prepend the name of the commit author to the final output. This takes the form of '#author,{COMMIT_AUTHOR},'
func NewGitHubCommitsTransformation(ctx context.Context, uri string) (webhookd.WebhookTransformation, error) {

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

	p := GitHubCommitsTransformation{
		ExcludeAdditions:     exclude_additions,
		ExcludeModifications: exclude_modifications,
		ExcludeDeletions:     exclude_deletions,
		prepend_message:      prepend_message,
		prepend_author:       prepend_author,
	}

	return &p, nil
}

// Transform() transforms 'body' (which is assumed to be a GitHub commit webhook message) in to CSV data containing:
// the commit hash, the name of the repository and the path to the file commited.
func (p *GitHubCommitsTransformation) Transform(ctx context.Context, body []byte) ([]byte, *webhookd.WebhookError) {

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
	wr := csv.NewWriter(buf)

	if p.prepend_message {
		wr.Write([]string{"#message", *event.HeadCommit.Message, ""})
	}

	if p.prepend_author {
		wr.Write([]string{"#author", *event.HeadCommit.Author.Name, ""})
	}

	repo := event.Repo
	repo_name := *repo.Name
	commit_hash := *event.HeadCommit.ID

	for _, c := range event.Commits {

		if !p.ExcludeAdditions {
			for _, path := range c.Added {
				commit := []string{commit_hash, repo_name, path}
				wr.Write(commit)
			}
		}

		if !p.ExcludeModifications {
			for _, path := range c.Modified {
				commit := []string{commit_hash, repo_name, path}
				wr.Write(commit)
			}
		}

		if !p.ExcludeDeletions {
			for _, path := range c.Removed {
				commit := []string{commit_hash, repo_name, path}
				wr.Write(commit)
			}
		}
	}

	wr.Flush()

	return buf.Bytes(), nil
}
