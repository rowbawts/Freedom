package main

import (
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v54/github"
	"golang.org/x/net/context"
	"log"
	"net/http"
)

// Wrap the shared transport for use with the integration ID and authenticating with installation ID.
var itr, _ = ghinstallation.NewKeyFromFile(http.DefaultTransport, 381312, 41105280, "theopenestsource.2023-08-26.private-key.pem")

// Use installation transport with client.
var client = github.NewClient(&http.Client{Transport: itr})
var ctx = context.Background()

func initGitHubClient() {
	readme, _, err := client.Repositories.GetReadme(ctx, "rowbawts", "theopenestsource", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	content, err := readme.GetContent()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(content)
}

func listenForWebhook() {
	fmt.Println("Listening on :3333......")

	http.HandleFunc("/", webhookHandler)

	err := http.ListenAndServe(":3333", nil)

	if err != nil {
		panic(err)
	}
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, nil)
	if err != nil {
		panic(err)
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(200)

	switch event := event.(type) {
	case *github.IssuesEvent:
		processIssuesEvent(event)
		break
	case *github.IssueCommentEvent:
		processIssueCommentEvent(event)
		break
	case *github.PullRequestEvent:
		processPullRequestEvent(event)
		break
	default:
		fmt.Println("Unhandled Event!")
		break
	}
}

func processIssuesEvent(event *github.IssuesEvent) {
	if event.GetAction() == "opened" {
		// Respond with a comment
		comment := &github.IssueComment{
			Body: github.String("Thanks for opening this issue!"),
		}

		_, _, err := client.Issues.CreateComment(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), event.GetIssue().GetNumber(), comment)
		if err != nil {
			log.Println("Error creating comment:", err)
		}
	}
}

func processIssueCommentEvent(event *github.IssueCommentEvent) {
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetIssue().GetNumber()
	reactionCount := 0

	if event.GetIssue().IsPullRequest() {
		comments, _, err := client.PullRequests.ListComments(ctx, owner, repo, prNumber, nil)
		if err != nil {
			log.Println("Error fetching reactions:", err)
			return
		}

		// Check if there are thumbs up (:+1:) reactions
		for _, comment := range comments {
			if *comment.Body == "+1" {
				reactionCount++

				if reactionCount >= 1 {
					// Merge the pull request
					merge := &github.PullRequestOptions{
						MergeMethod: "merge", // Change this as needed
					}

					_, _, err := client.PullRequests.Merge(ctx, owner, repo, prNumber, "Merging based on reactions", merge)
					if err != nil {
						log.Println("Error merging pull request:", err)
					} else {
						log.Println("Pull request merged successfully")
					}

					reactionCount = 0
					return
				}
			}
		}
	}
}

func processPullRequestEvent(event *github.PullRequestEvent) {
	if event.GetAction() == "opened" || event.GetAction() == "reopened" {
		// Respond with a comment
		comment := &github.IssueComment{
			Body: github.String("React to this comment with :+1: to vote for getting it merged!"),
		}

		_, _, err := client.Issues.CreateComment(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), event.GetPullRequest().GetNumber(), comment)
		if err != nil {
			log.Println("Error creating comment:", err)
		}
	}
}
