package main

var version = "v0.07"

func main() {
	initGitHubClient(version)
	listenForWebhook()
}
