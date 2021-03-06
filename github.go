package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"net/http"
	"os"
	"strings"
)

type Release struct {
	Key       string
	ReleaseId string
	Commits   []Commit
}

type Commit struct {
	Key     string
	Sha     string
	Message string
	Author  string
}

func getCommitComparison(client *github.Client, userName string, repositoryName string, branch1 string, branch2 string) ([]Commit, error) {
	repos, _, err := client.Repositories.CompareCommits(userName, repositoryName, branch1, branch2)

	if err != nil {
		return nil, err
	}

	var commits []Commit
	for _, value := range repos.Commits {
		message := *value.Commit.Message
		commits = append(commits, Commit{*value.SHA, *value.SHA, message, *value.Commit.Author.Name})
	}
	return commits, nil
}

func getRepositories(client *github.Client) {
	repos, _, _ := client.Repositories.List("", nil)

	var s []string
	var owner []string
	for _, value := range repos {
		s = append(s, *value.FullName)
		owner = append(owner, *value.Owner.Login)
	}
	fmt.Println(owner)
}

func getAuthenticatedGitHubClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_CLIENT_TOKEN")},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	return client
}

func getRepositoryTags(client *github.Client, userName string, repositoryName string) []github.RepositoryTag {

	repos, _, _ := client.Repositories.ListTags(userName, repositoryName, nil)

	var s []string
	for _, value := range repos {
		s = append(s, *value.Name)
	}
	return repos
}

func getReleases(userName string, repositoryName string) []Release {
	client := getAuthenticatedGitHubClient()
	releases := []Release{}

	tags := getRepositoryTags(client, userName, repositoryName)
	tagCounter := 0
	var lastTagName string

	for _, value := range tags {
		tagCounter++

		if tagCounter == 1 {
			lastTagName = *value.Name
			continue
		}

		//Remove this later
		if tagCounter == 6 {
			lastTagName = *value.Name
			break
		}

		releaseNames := []string{*value.Name, lastTagName}
		releaseName := strings.Join(releaseNames, "-")

		commits, err := getCommitComparison(client, userName, repositoryName, *value.Name, lastTagName)
		if err != nil {
			fmt.Println("Houston we have an error")
			continue
		}

		releases = append(releases, Release{releaseName, releaseName, commits})
		lastTagName = *value.Name
	}
	return releases
}

func releaseEndpointHandler(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	userName := urlParams["user"]
	repositoryName := urlParams["repo"]

	var releases []Release
	releases = getReleases(userName, repositoryName)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(releases)
}
