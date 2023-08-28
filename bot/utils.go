package main

import (
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v54/github"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Wrap the shared transport for use with the integration ID and authenticating with installation ID.
var itr, _ = ghinstallation.NewKeyFromFile(http.DefaultTransport, 381312, 41105280, "theopenestsource.2023-08-26.private-key.pem")

// Use installation transport with client.
var client = github.NewClient(&http.Client{Transport: itr})
var ctx = context.Background()

func initGitHubClient() {
	log.Println("Initializing......")
}

func listenForWebhook() {
	log.Println("Listening on :3333......")

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
		if !strings.Contains(event.GetComment().GetUser().GetLogin(), "bot") {
			processIssueCommentEvent(event)
			break
		}
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
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	issueNumber := event.GetIssue().GetNumber()

	if event.GetAction() == "opened" {
		commentText := "Thanks for opening this issue!"

		// Respond with a comment
		comment := &github.IssueComment{
			Body: github.String(commentText),
		}

		_, _, err := client.Issues.CreateComment(ctx, owner, repo, issueNumber, comment)
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
	reactionCountGoal := 5
	approvers := []string{}

	if event.GetIssue().IsPullRequest() {
		comments, _, err := client.Issues.ListComments(ctx, owner, repo, prNumber, nil)
		if err != nil {
			log.Println("Error fetching reactions:", err)
			return
		}

		// Check if there are thumbs up (:+1:) reactions
		for _, comment := range comments {
			if strings.Contains(comment.GetBody(), "+1") && !strings.Contains(comment.GetUser().GetLogin(), "bot") {
				for _, name := range approvers {
					if !strings.Contains(name, comment.GetUser().GetLogin()) {
						reactionCount++
						approvers = append(approvers, comment.GetUser().GetLogin())
					} else {
						// Respond with a comment
						comment := &github.IssueComment{
							Body: github.String("You've already voted!"),
						}

						_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
						if err != nil {
							log.Println("Error creating comment:", err)
						}
					}
				}
			}
		}

		if reactionCount >= reactionCountGoal {
			// Merge the pull request
			merge := &github.PullRequestOptions{
				MergeMethod: "merge", // Change this as needed
			}

			// Respond with a comment
			comment := &github.IssueComment{
				Body: github.String("Merging based on reactions :fireworks:"),
			}

			_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
			if err != nil {
				log.Println("Error creating comment:", err)
			}

			_, _, err = client.PullRequests.Merge(ctx, owner, repo, prNumber, "Merging based on reactions", merge)
			if err != nil {
				log.Println("Error merging pull request:", err)
			} else {
				log.Println("Pull request #", prNumber, "merged successfully")
			}

			return
		} else {
			commentText := "Current :+1: count is (#{reactionCount}) need (#{reactionRemainingCount}) more to merge"
			commentText = strings.Replace(commentText, "(#{reactionCount})", strconv.Itoa(reactionCount), 1)
			commentText = strings.Replace(commentText, "(#{reactionRemainingCount})", strconv.Itoa(reactionCountGoal-reactionCount), 1)

			// Respond with a comment
			comment := &github.IssueComment{
				Body: github.String(commentText),
			}

			_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
			if err != nil {
				log.Println("Error creating comment:", err)
			}
		}
	}
}

func processPullRequestEvent(event *github.PullRequestEvent) {
	if event.GetAction() == "opened" || event.GetAction() == "reopened" {
		// Respond with a comment
		comment := &github.IssueComment{
			Body: github.String("Comment on this PR with :+1: to vote for getting it merged!"),
		}

		_, _, err := client.Issues.CreateComment(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), event.GetPullRequest().GetNumber(), comment)
		if err != nil {
			log.Println("Error creating comment:", err)
		}
	}
}
