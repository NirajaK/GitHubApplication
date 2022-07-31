# GitHubApplication
Go based Rest Server to interact and invoke GitHub apis. Exposed CreatePullRequest which connects to Git using OAuth2, creates branch(if not exists), 
creates changeset with dummy text and raises Pull Request for desitination branch.


Application
		  |---- githubapplication.go   -> code for rest server
		  |---- github.go              -> business logic for interacting with github
		  |---- spec.go                -> request, response structures
		  |---- transport.go           -> handler specific code
      |---- Dockerfile
