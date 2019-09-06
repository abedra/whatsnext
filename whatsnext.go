package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type config struct {
	Users              []string
	Orgs               []string
	Whitelist          []string
	Blacklist          []string
	ConstrainAssignees []string `yaml:"constrain_assignees"`
	Url                string
	Token              string
}

func contains(list []string, s *string) bool {
	for _, item := range list {
		if item == *s {
			return true
		}
	}
	return false
}

func usersContains(a []*github.User, b []string) bool {
	for _, item := range a {
		if contains(b, item.Login) {
			return true
		}
	}
	return false
}

func printIssue(issue *github.Issue) {
	if issue.IsPullRequest() {
		return
	}

	if *issue.Number < 10 {
		fmt.Printf("    [%d]  - %s - %s\n", *issue.Number, *issue.Title, *issue.HTMLURL)
	} else {
		fmt.Printf("    [%d] - %s - %s\n", *issue.Number, *issue.Title, *issue.HTMLURL)
	}
}

func printPullRequest(pr *github.PullRequest) {
	fmt.Printf("    [PR] - %s - %s\n", *pr.Title, *pr.HTMLURL)
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

	if config.Url == "" {
		client := github.NewClient(tc)
		return client
	} else {
		client, err := github.NewEnterpriseClient(config.Url, config.Url, tc)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return client
	}
}

func processRepositories(client *github.Client, ctx context.Context, entity string, repos []*github.Repository, config config) {
	fmt.Printf("Open issues for %s\n", entity)
	for _, repo := range repos {
		if config.Whitelist != nil && !contains(config.Whitelist, repo.Name) {
			continue
		}

		if config.Blacklist != nil && contains(config.Blacklist, repo.Name) {
			continue
		}

		if *repo.Owner.Login != entity {
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
					if pr.Assignees != nil && pr.User != nil && contains(config.ConstrainAssignees, pr.User.Login) || usersContains(pr.RequestedReviewers, config.ConstrainAssignees) {
						printPullRequest(pr)
					}
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
					if issue.Assignees != nil && usersContains(issue.Assignees, config.ConstrainAssignees) {
						printIssue(issue)
					}
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
			processRepositories(client, ctx, org, repos, config)
		}
	}

	for _, user := range config.Users {
		opt := &github.RepositoryListOptions{Type: "all", ListOptions: github.ListOptions{PerPage: 1000}}
		repos, _, err := client.Repositories.List(ctx, user, opt)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			processRepositories(client, ctx, user, repos, config)
		}
	}
}

func defaultConfigFile() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return usr.HomeDir + "/.whatsnext/config.yml"
}

func main() {
	configPtr := flag.String("config", defaultConfigFile(), "Path to configuration yaml file")
	flag.Parse()
	config := readConfig(*configPtr)
	client := buildClient(config)
	processEntities(config, client)
}
