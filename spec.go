package deployment

// CreatePullRequest ...
type CreatePullRequest struct {
	SourceRepo	*RepoDetails `json:"source_repo"`
	TargetRepo *RepoDetails  `json:"target_repo"`
	CommitMessage string `json:"commit_msg"`
	CommitBranch string `json:"commit_branch"`
	BaseBranch string `json:"base_branch"`
	TargetBranch string `json:"target_branch"`
	PRSubject string `json:"pr_subject"`
	PRDescription string `json:"pr_description"`
	ChangeSet []*FileInfo `json:"changed_files"`
	AuthorName string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
}

// RepoDetails ...
type RepoDetails struct {
	Owner string `json:"owner"`
	Repo string `json:"repo_name"`
}

// FileInfo ...
type FileInfo struct {
	Name string `json:"file_name"`
}

// CreatePullResponse ...
type CreatePullResponse struct {
	Status int32
	URL string
	Err error
}
