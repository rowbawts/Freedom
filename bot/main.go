package main

var version = "v0.06"

func main() {
	initGitHubClient(version)
	listenForWebhook()
}
