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

type BranchHandler struct {
	githubapp.ClientCreator

	preamble string
}

func (h *BranchHandler) Handles() []string {
	return []string{"create", "delete", "push"}
}

func (h *BranchHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {

	logctx := zerolog.Ctx(ctx).With()
	logger := logctx.Logger()

	// Handle creation of new branch
	if eventType == "create" {
		var event github.CreateEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return errors.Wrap(err, "failed to parse issue comment event payload")
		}
		logger.Info().Msg("Create event invoked")
		logger.Info().Msgf("Created new branch : %s inside repo : %s", *event.Ref, *event.GetRepo().Name)

		// Write the event in its own file
		file, _ := json.MarshalIndent(event, "", " ")
		_ = ioutil.WriteFile(GetFilenameDate("create-branch"), file, 0644)
	}

	// Handle deletion of a branch
	if eventType == "delete" {
		var event github.DeleteEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return errors.Wrap(err, "failed to parse issue comment event payload")
		}
		logger.Info().Msg("Delete event invoked")
		logger.Info().Msgf("Deleted branch : %s inside repo : %s", *event.Ref, *event.GetRepo().Name)

		// Write the event in its own file
		file, _ := json.MarshalIndent(event, "", " ")
		_ = ioutil.WriteFile(GetFilenameDate("delete-branch"), file, 0644)
	}

	// Handle push events
	if eventType == "push" {
		var event github.PushEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return errors.Wrap(err, "failed to parse issue comment event payload")
		}
		logger.Info().Msg("Push event invoked")

		// Write the event in its own file
		file, _ := json.MarshalIndent(event, "", " ")
		_ = ioutil.WriteFile(GetFilenameDate("push"), file, 0644)

		h.GetAppInfo(ctx)
		h.GetAllCommits(ctx, event)
	}

	return nil
}

func (h *BranchHandler) GetAppInfo(ctx context.Context) error {
	logctx := zerolog.Ctx(ctx).With()
	logger := logctx.Logger()

	client, err := h.NewAppClient()
	if err != nil {
		return err
	}
	app, _, err := client.Apps.Get(ctx, "")
	if err != nil {
		logger.Error().Msgf("Apps.Get returned error: %v", err)
	}
	logger.Info().Msgf("Installed App ID : %d, App Name : %s", *app.ID, *app.Name)
	return nil
}

func (h *BranchHandler) GetAllCommits(ctx context.Context, event github.PushEvent) error {
	logctx := zerolog.Ctx(ctx).With()
	logger := logctx.Logger()

	client, err := h.NewInstallationClient(*event.Installation.ID)
	if err != nil {
		return err
	}

	opt := github.CommitsListOptions{}
	commits, _, err := client.Repositories.ListCommits(ctx, *event.Repo.Owner.Name, *event.Repo.Name, &opt)
	if err != nil {
		logger.Error().Msgf("ListCommits returned error: %v", err)
	}

	logger.Info().Msgf("Total commit : %d", len(commits))

	for _, commit := range commits {
		logger.Info().Msgf("Author Name : %s, SHA : %s", commit.Author, *commit.SHA)
		logger.Info().Msg("List of file changed as part of this commit are : ")
		for _, file := range commit.Files {
			logger.Info().Msgf("File Name : %s, File URL : %s", file.Filename, file.BlobURL)
		}
	}

	return nil
}

// PreparePRContext adds information about a pull request to the logger in a
// context and returns the modified context and logger.
func PreparePRContext(ctx context.Context, installationID int64, repo *github.Repository, number int) (context.Context, zerolog.Logger) {
	logctx := zerolog.Ctx(ctx).With()

	logger := logctx.Logger()
	return logger.WithContext(ctx), logger
}

// Get filename based on current time
func GetFilenameDate(eventType string) string {
	// Use layout string for time format.
	const layout = "01-02-2021"
	// Place now in the string.
	t := time.Now()
	return eventType + t.Format(layout) + ".json"
}
