package main

import (
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v54/github"
	"golang.org/x/net/context"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Wrap the shared transport for use with the integration ID and authenticating with installation ID.
var privateKey = os.Getenv("privateKey")
var host = os.Getenv("host")
var port = os.Getenv("port")
var itr, _ = ghinstallation.New(http.DefaultTransport, 381312, 41105280, []byte(privateKey))

// Use installation transport with client.
var client = github.NewClient(&http.Client{Transport: itr})
var ctx = context.Background()

func initGitHubClient(v string) {
	log.Println("Initializing......", v)

	if privateKey != "" {
		log.Println("Private key loaded from env!")
	} else {
		log.Println("No private key specified in env!")
		os.Exit(0)
	}

	if port != "" {
		log.Println("Port configured from env!")
	} else {
		log.Println("No port specified in env!")
		os.Exit(0)
	}
}

func listenForWebhook() {
	log.Printf("Listening on :%s......\n", port)

	http.HandleFunc("/", webHandle)
	http.HandleFunc("/webhook", webhookHandler)

	address := net.JoinHostPort(host, port)

	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}
}

func webHandle(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "Turn back!")
	if err != nil {
		return
	}

	log.Println("Request received to / endpoint:", r)
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
		log.Println("Received Issues Event: processing now!")
		processIssuesEvent(event)
		break
	case *github.IssueCommentEvent:
		userType := strings.ToLower(event.GetComment().GetUser().GetType())
		commentBody := event.GetComment().GetBody()

		log.Println("User type:", userType, "Comment body:", commentBody)

		if !strings.Contains(userType, "bot") && strings.Contains(commentBody, ":+1:") {
			log.Println("Received Issue Comment Event: processing now!")
			processIssueCommentEvent(event)
			break
		}

		break
	case *github.PullRequestEvent:
		log.Println("Received Pull Request Event: processing now!")
		processPullRequestEvent(event)
		break
	default:
		log.Println("Received Unhandled Event!")
		break
	}
}

func processIssuesEvent(event *github.IssuesEvent) {
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	issueNumber := event.GetIssue().GetNumber()

	if event.GetAction() == "opened" {
		commentText := "Thanks for opening this issue! Someone will be responding soon! :smile:"

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
	eventSender := event.GetSender().GetLogin()

	reactionCountGoal := 5
	approvals := map[string]int{}

	if event.GetIssue().IsPullRequest() {
		comments, _, err := client.Issues.ListComments(ctx, owner, repo, prNumber, nil)
		if err != nil {
			log.Println("Error fetching comments:", err)
			return
		}

		// Check if there are thumbs up (:+1:) reactions
		for _, comment := range comments {
			userType := strings.ToLower(event.GetComment().GetUser().GetType())
			commentAuthor := comment.GetUser().GetLogin()
			commentBody := comment.GetBody()

			if !strings.Contains(userType, "bot") && strings.Contains(commentBody, ":+1:") {
				_, exists := approvals[commentAuthor]
				if !exists {
					approvals[commentAuthor] = 1
				} else {
					approvals[commentAuthor]++
				}
			}
		}

		if len(approvals) >= reactionCountGoal {
			// Merge the pull request
			merge := &github.PullRequestOptions{
				MergeMethod: "merge", // Change this as needed
			}

			// Respond with a comment
			comment := &github.IssueComment{
				Body: github.String("Merging based on reactions :heavy_check_mark: :rocket:"),
			}

			_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
			if err != nil {
				log.Println("Error creating comment:", err)
			}

			_, _, err = client.PullRequests.Merge(ctx, owner, repo, prNumber, "Merging based on reactions!", merge)
			if err != nil {
				log.Println("Error merging pull request:", err)
			} else {
				log.Println("Pull request #", prNumber, "merged successfully!")
			}

			return
		}

		if approvals[eventSender] > 1 {
			commentText := "@(#{commentAuthor}) your vote has already been counted :x:"
			commentText = strings.Replace(commentText, "(#{commentAuthor})", eventSender, 1)

			// Respond with a comment
			comment := &github.IssueComment{
				Body: github.String(commentText),
			}

			_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
			if err != nil {
				log.Println("Error creating comment:", err)
			}
		} else {
			commentText := "@(#{commentAuthor}) voted! :tada:\n" + "Votes: (#{reactionCount})/(#{reactionCountGoal})"
			commentText = strings.Replace(commentText, "(#{commentAuthor})", eventSender, 1)
			commentText = strings.Replace(commentText, "(#{reactionCount})", strconv.Itoa(len(approvals)), 1)
			commentText = strings.Replace(commentText, "(#{reactionCountGoal})", strconv.Itoa(reactionCountGoal), 1)

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
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	prNumber := event.GetPullRequest().GetNumber()

	if event.GetAction() == "opened" || event.GetAction() == "reopened" {
		// Respond with a comment
		comment := &github.IssueComment{
			Body: github.String("Comment on this PR with :+1: to vote for getting it merged!"),
		}

		_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
		if err != nil {
			log.Println("Error creating comment:", err)
		}
	}
}
