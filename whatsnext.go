package main

import (
        "fmt"
	"io/ioutil"
	"flag"
        "github.com/google/go-github/github"
        "golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	User  string
	Token string
}

func printIssue(issue github.Issue) {
        if *issue.Number < 10 {
                fmt.Printf("    [%d]  - %s\n", *issue.Number, *issue.Title)
        } else {
                fmt.Printf("    [%d] - %s\n", *issue.Number, *issue.Title)
        }
}

func main() {
	configPtr := flag.String("config", "config.yml", "Path to configuration yaml file")
	flag.Parse()

        var config Config
        raw, err := ioutil.ReadFile(*configPtr)
        if err != nil {
                fmt.Println(err)
        }

        err = yaml.Unmarshal(raw, &config)
        if err != nil {
                fmt.Println(err)
        }

        ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token},)
        tc := oauth2.NewClient(oauth2.NoContext, ts)
        client := github.NewClient(tc)
        repos, _, err := client.Repositories.List(config.User, nil)

        if err != nil {
                fmt.Println("Error:", err)
        } else {
                fmt.Printf("%s repositories\n", config.User)
                for _, repo := range repos {
                        fmt.Printf("  %s\n", *repo.Name)
                        issues, _, err := client.Issues.ListByRepo(config.User, *repo.Name, nil)
                        if err != nil {
                                fmt.Println("Error:", err)
                        } else {
                                if len(issues) == 0 {
                                        fmt.Printf("    No Issues\n")
                                } else {
                                        for _, issue := range issues {
                                                printIssue(issue)
                                        }
                                }
                        }
                }
        }
}
