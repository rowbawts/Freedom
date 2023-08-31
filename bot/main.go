package main

var version = "v0.08"

func main() {
	initGitHubClient(version)
	listenForWebhook()
}
