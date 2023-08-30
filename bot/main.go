package main

var version = "v0.04"

func main() {
	initGitHubClient(version)
	listenForWebhook()
}
