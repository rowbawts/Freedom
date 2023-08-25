package main

import (
	"fmt"
	"golang.org/x/net/context"
	"net/http"

	"github.com/google/go-github/v54/github"

	"github.com/bradleyfalzon/ghinstallation/v2"
)

func main() {
	// Wrap the shared transport for use with the integration ID 1 authenticating with installation ID 99.
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 381312, 41105280, "theopenestsource.2023-08-25.private-key.pem")
	// secret = 57d9b2f565aedc5a5d658b190555ff379701a86c

	// Or for endpoints that require JWT authentication
	// itr, err := ghinstallation.NewAppsTransportKeyFromFile(http.DefaultTransport, 1, "2016-10-19.private-key.pem")

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
