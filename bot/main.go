package main

var version = "v0.05"

func main() {
	initGitHubClient(version)
	listenForWebhook()
}
