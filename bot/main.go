package main

var version = "v0.02"

func main() {
	initGitHubClient(version)
	listenForWebhook()
}
