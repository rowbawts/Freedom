package main

var version = "v0.03"

func main() {
	initGitHubClient(version)
	listenForWebhook()
}
