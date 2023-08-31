package main

var version = "v0.09"

func main() {
	initGitHubClient(version)
	listenForWebhook()
}
