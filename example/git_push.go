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
	"strings"
	"time"

	"github.com/google/go-github/v39/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GitPushHandler struct {
	githubapp.ClientCreator

	preamble string
}

func (h *GitPushHandler) Handles() []string {
	return []string{"push"}
}

func (h *GitPushHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {

	logctx := zerolog.Ctx(ctx).With()
	logger := logctx.Logger()

	// Handle push events
	if eventType == "push" {
		var event github.PushEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return errors.Wrap(err, "failed to parse issue comment event payload")
		}
		logger.Info().Msg("Push event invoked")

		// Write the event in its own file
		file, _ := json.MarshalIndent(event, "", " ")
		_ = ioutil.WriteFile(h.GetFilenameDate("push"), file, 0644)

		h.GetAppInfo(ctx)
		h.GetAllCommits(ctx, event)
		h.ListRepoContent(ctx, event)
	}

	return nil
}

func (h *GitPushHandler) GetAppInfo(ctx context.Context) error {
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

func (h *GitPushHandler) GetAllCommits(ctx context.Context, event github.PushEvent) error {
	logctx := zerolog.Ctx(ctx).With()
	logger := logctx.Logger()

	client, err := h.NewInstallationClient(*event.Installation.ID)
	if err != nil {
		return err
	}

	opt := github.CommitsListOptions{
		//SHA: *event.After, // The SHA of the most recent commit on ref after the push
	}

	commits, _, err := client.Repositories.ListCommits(ctx, *event.Repo.Owner.Name, *event.Repo.Name, &opt)
	if err != nil {
		logger.Error().Msgf("ListCommits returned error: %v", err)
	}

	logger.Info().Msgf("Total commit : %d for SHA : %S", len(commits), *event.After)

	//list_opt := github.ListOptions{}
	for _, commit := range commits {
		logger.Info().Msgf("Author Name : %s, SHA : %s", commit.Author, *commit.SHA)
		logger.Info().Msg("List of file changed as part of this commit are : ")
		/*
			repoCommit, _, err := client.Repositories.GetCommit(ctx, *event.Repo.Owner.Name, *event.Repo.Name, *commit.SHA, &list_opt)
			if err != nil {
				logger.Error().Msgf("Error fetching files for commit SHA : %s", *commit.SHA)
			}

			for _, file := range repoCommit.Files {
				logger.Info().Msgf("File Name : %s, File URL : %s", file.GetFilename(), file.GetBlobURL())
				logger.Info().Msgf("File Name : %s, File URL : %s", *file.Filename, *file.BlobURL)
			}
		*/
	}

	return nil
}

func (h *GitPushHandler) ListRepoContent(ctx context.Context, event github.PushEvent) error {
	logctx := zerolog.Ctx(ctx).With()
	logger := logctx.Logger()

	client, err := h.NewInstallationClient(*event.Installation.ID)
	if err != nil {
		return err
	}

	opt := github.RepositoryContentGetOptions{
		Ref: *event.Ref,
	}

	_, directories, _, err := client.Repositories.GetContents(ctx, *event.Repo.Owner.Login, *event.Repo.Name, "/", &opt)
	if err != nil {
		logger.Error().Msgf("GetContent returned error: %v", err)
	}

	logger.Info().Msgf("Total files at root : %d", len(directories))
	h.listSubFiles(ctx, event, "", directories, logger)

	return nil
}

func (h *GitPushHandler) listSubFiles(ctx context.Context, event github.PushEvent, path string, directories []*github.RepositoryContent, logger zerolog.Logger) error {

	for _, dir := range directories {
		if strings.Compare(*dir.Type, "dir") == 0 {
			client, err := h.NewInstallationClient(*event.Installation.ID)
			if err != nil {
				return err
			}

			opt := github.RepositoryContentGetOptions{
				Ref: *event.Ref,
			}
			_, dirs, _, err := client.Repositories.GetContents(ctx, *event.Repo.Owner.Login, *event.Repo.Name, *dir.Path, &opt)

			if err != nil {
				logger.Error().Msgf("GetContent subdirectory returned error: %v", err)
			}

			h.listSubFiles(ctx, event, *dir.Path, dirs, logger)

		} else if strings.Compare(*dir.Type, "file") == 0 {
			logger.Info().Msgf("File Name : %s, Path : %s", dir.GetName(), dir.GetPath())
		}

	}

	return nil

}

// Get filename based on current time
func (*GitPushHandler) GetFilenameDate(eventType string) string {
	// Use layout string for time format.
	const layout = "01-02-2021"
	// Place now in the string.
	t := time.Now()
	return eventType + t.Format(layout) + ".json"
}
