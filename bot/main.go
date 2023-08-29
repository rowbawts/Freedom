package main

var version = "v0.01"

func main() {
	initGitHubClient(version)
	listenForWebhook()
}
