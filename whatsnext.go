package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type config struct {
	Users     []string
	Orgs      []string
	Whitelist []string
	Token     string
}

func contains(list []string, s *string) bool {
	for _, item := range list {
		if item == *s {
			return true
		}
	}
	return false
}

func printIssue(issue *github.Issue) {
	if *issue.Number < 10 {
		fmt.Printf("    [%d]  - %s - %s\n", *issue.Number, *issue.Title, *issue.URL)
	} else {
		fmt.Printf("    [%d] - %s - %s\n", *issue.Number, *issue.Title, *issue.URL)
	}
}

func printPullRequest(pr *github.PullRequest) {
	fmt.Printf("    Pull request: %s (%s)", *pr.Title, *pr.CommitsURL)
}

func readConfig(file string) config {
	var config config

	raw, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(raw, &config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return config
}

func buildClient(config config) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	return client
}

func processRepositories(client *github.Client, ctx context.Context, entity string, repos []*github.Repository, whitelist []string) {
	fmt.Printf("Open issues for %s\n", entity)
	for _, repo := range repos {
		if whitelist != nil && !contains(whitelist, repo.Name) {
			continue
		}

		fmt.Printf("Repository: %s\n", *repo.Name)
		prOpt := &github.PullRequestListOptions{ListOptions: github.ListOptions{PerPage: 100}}
		prs, _, err := client.PullRequests.List(ctx, entity, *repo.Name, prOpt)

		if err != nil {
			fmt.Println("Error with PR's:", err)
		} else {
			if len(prs) != 0 {
				for _, pr := range prs {
					printPullRequest(pr)
				}
			}
		}

		issueOpt := &github.IssueListByRepoOptions{ListOptions: github.ListOptions{PerPage: 1000}}
		issues, _, err := client.Issues.ListByRepo(ctx, entity, *repo.Name, issueOpt)
		if err != nil {
			fmt.Println("Error with Issues:", err)
		} else {
			if len(issues) != 0 {
				for _, issue := range issues {
					printIssue(issue)
				}
			}
		}
	}
}

func processEntities(config config, client *github.Client) {
	ctx := context.Background()
	for _, org := range config.Orgs {
		opt := &github.RepositoryListByOrgOptions{Type: "all", ListOptions: github.ListOptions{PerPage: 1000}}
		repos, _, err := client.Repositories.ListByOrg(ctx, org, opt)

		if err != nil {
			fmt.Println("Error:", err)
		} else {
			processRepositories(client, ctx, org, repos, config.Whitelist)
		}
	}

	for _, user := range config.Users {
		opt := &github.RepositoryListOptions{Type: "all", ListOptions: github.ListOptions{PerPage: 1000}}
		repos, _, err := client.Repositories.List(ctx, user, opt)

		if err != nil {
			fmt.Println("Error:", err)
		} else {
			processRepositories(client, ctx, user, repos, config.Whitelist)
		}
	}
}

func main() {
	configPtr := flag.String("config", "config.yml", "Path to configuration yaml file")
	flag.Parse()
	config := readConfig(*configPtr)
	client := buildClient(config)
	processEntities(config, client)
}
