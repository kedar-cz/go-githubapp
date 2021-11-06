// Copyright 2018 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/google/go-github/v39/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type PullRequestHandler struct {
	githubapp.ClientCreator

	preamble string
}

func (h *PullRequestHandler) Handles() []string {
	return []string{"pull_request", "pull_request_review", "pull_request_review_comment"}
}

func (h *PullRequestHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {

	logctx := zerolog.Ctx(ctx).With()
	logger := logctx.Logger()

	// Handle PR

	if eventType == "pull_request" {
		var event github.PullRequestEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return errors.Wrap(err, "failed to parse issue comment event payload")
		}
		logger.Info().Msgf("Pull request event invoked. Action : %s", event.GetAction())
		logger.Info().Msgf("Pull Request : %s Title : %s", event.PullRequest.GetID(), event.PullRequest.GetTitle)

		// Write the event in its own file
		file, _ := json.MarshalIndent(event, "", " ")
		_ = ioutil.WriteFile(GetPRFilename("pull-request"), file, 0644)
		h.ListPRFiles(ctx, event)
	}

	// Handle PR review
	if eventType == "pull_request_review" {
		var event github.PullRequestReview
		if err := json.Unmarshal(payload, &event); err != nil {
			return errors.Wrap(err, "failed to parse issue comment event payload")
		}
		logger.Info().Msg("PR review event")
		logger.Info().Msgf("Review : %s", event.GetBody())

		// Write the event in its own file
		file, _ := json.MarshalIndent(event, "", " ")
		_ = ioutil.WriteFile(GetFilenameDate("pr-review"), file, 0644)
	}

	// Handle PR review comment
	if eventType == "pull_request_review_comment" {
		var event github.PullRequestReviewCommentEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return errors.Wrap(err, "failed to parse issue comment event payload")
		}
		logger.Info().Msg("PR Review comment")
		logger.Info().Msgf("Comment : %s", *event.Comment.Body)

		// Write the event in its own file
		file, _ := json.MarshalIndent(event, "", " ")
		_ = ioutil.WriteFile(GetFilenameDate("pr-review-comment"), file, 0644)
	}

	return nil
}

func (h *PullRequestHandler) ListPRFiles(ctx context.Context, event github.PullRequestEvent) error {
	logctx := zerolog.Ctx(ctx).With()
	logger := logctx.Logger()

	client, err := h.NewInstallationClient(*event.Installation.ID)
	if err != nil {
		return err
	}

	opt := github.ListOptions{}
	files, _, err := client.PullRequests.ListFiles(ctx, *event.Repo.Owner.Login, *event.Repo.Name, *event.Number, &opt)
	if err != nil {
		logger.Error().Msgf("PR - ListFiles returned error: %v", err)
	}

	logger.Info().Msgf("Total commit : %d", len(files))

	for _, file := range files {
		logger.Info().Msgf("File Name : %s, SHA : %s", *file.Filename, file.GetRawURL())
	}

	return nil
}

// Get filename based on current time
func GetPRFilename(eventType string) string {
	// Use layout string for time format.
	const layout = "01-02-2021"
	// Place now in the string.
	t := time.Now()
	return eventType + t.Format(layout) + ".json"
}
