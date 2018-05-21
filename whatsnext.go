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

type Config struct {
	Users []string
	Token string
}

func printIssue(issue *github.Issue) {
	if *issue.Number < 10 {
		fmt.Printf("    [%d]  - %s\n", *issue.Number, *issue.Title)
	} else {
		fmt.Printf("    [%d] - %s\n", *issue.Number, *issue.Title)
	}
}

func ReadConfig(file string) Config {
	var config Config

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

func BuildClient(config Config) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	return client
}

func ProcessRepositories(config Config, client *github.Client) {
	ctx := context.Background()
	for _, user := range config.Users {
		opt := &github.RepositoryListByOrgOptions{Type: "all", ListOptions: github.ListOptions{PerPage: 1000}}
		repos, _, err := client.Repositories.ListByOrg(ctx, user, opt)

		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("Open issues for %s\n", user)
			for _, repo := range repos {
				fmt.Println(*repo.Name)
				issueOpt := &github.IssueListByRepoOptions{ListOptions: github.ListOptions{PerPage: 1000}}
				issues, _, err := client.Issues.ListByRepo(ctx, user, *repo.Name, issueOpt)
				if err != nil {
					fmt.Println("Error:", err)
				} else {
					if len(issues) != 0 {
						fmt.Printf("  %s\n", *repo.Name)
						for _, issue := range issues {
							printIssue(issue)
						}
					}
				}
			}
		}
	}
}

func main() {
	configPtr := flag.String("config", "config.yml", "Path to configuration yaml file")
	flag.Parse()
	config := ReadConfig(*configPtr)
	client := BuildClient(config)
	ProcessRepositories(config, client)
}
