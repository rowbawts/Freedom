package main

import (
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v54/github"
	"golang.org/x/net/context"
	"net/http"
)

func initGitHubClient() {
	// Wrap the shared transport for use with the integration ID and authenticating with installation ID.
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 381312, 41105280, "theopenestsource.2023-08-26.private-key.pem")

	if err != nil {
		// Handle error.
	}

	// Use installation transport with client.
	client := github.NewClient(&http.Client{Transport: itr})

	// Use client...
	//client.PullRequests.CreateComment(ctx, "rowbawts", "theopenestsource", 1, comment)
	ctx := context.Background()

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

	switch event := event.(type) {
	case *github.PullRequestEvent:
		processPullRequestEvent(event)
	case *github.IssueCommentEvent:
		processIssueCommentEvent(event)
	}
}

func processPullRequestEvent(event *github.PullRequestEvent) {
	fmt.Println(event.PullRequest.Comments)
}

func processIssueCommentEvent(event *github.IssueCommentEvent) {
	fmt.Println(event)
	// Wrap the shared transport for use with the integration ID and authenticating with installation ID.
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 381312, 41105280, "theopenestsource.2023-08-26.private-key.pem")

	if err != nil {
		// Handle error.
	}

	// Use installation transport with client.
	client := github.NewClient(&http.Client{Transport: itr})

	ctx := context.Background()

	s := "test from bot"

	comment := github.IssueComment{
		Body: &s,
	}

	client.Issues.CreateComment(ctx, event.GetRepo().GetOwner().GetName(), event.GetRepo().GetName(), event.GetIssue().GetNumber(), &comment)
}
