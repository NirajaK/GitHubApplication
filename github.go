package deployment

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	token = "AccessToken"
	success = 1
	failure = -1
)

type GitHubInterface interface {
	CreatePullRequest(ctx context.Context, req *CreatePullRequest) (*CreatePullResponse, error)
}

// VSphereAdapter vcenter operations struct
type GitHubAdapter struct {
	client          *github.Client
}

// NewVSphereAdapter returns VSphereAdapter instance
func NewGitHubAdapter() GitHubInterface {
	return &GitHubAdapter{
		client:   nil,
	}
}

func (ga *GitHubAdapter) connect(ctx context.Context) error {

	if ga.client != nil {
		return nil
	}
	token, ok := ctx.Value(token).(string); if !ok {
		return errors.New("Access Token not found")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	ga.client = github.NewClient(tc)

	return nil
}



// Disconnect disconnect from github
func (ga *GitHubAdapter) disconnect() {
	if ga.client != nil {
		ga.client = nil
	}
}


func (ga *GitHubAdapter) getRef(ctx context.Context, req *CreatePullRequest) (ref *github.Reference, err error) {
	if ref, _, err = ga.client.Git.GetRef(ctx, req.SourceRepo.Owner, req.SourceRepo.Repo, "refs/heads/"+req.CommitBranch); err == nil {
		return ref, nil
	}


	if req.CommitBranch == req.BaseBranch {
		return nil, errors.New("the commit branch does not exist but `-base-branch` is the same as `-commit-branch`")
	}

	if req.BaseBranch == "" {
		return nil, errors.New("the `-base-branch` should not be set to an empty string when the branch specified by `-commit-branch` does not exists")
	}

	var baseRef *github.Reference
	if baseRef, _, err = ga.client.Git.GetRef(ctx, req.SourceRepo.Owner, req.SourceRepo.Repo, "refs/heads/"+req.BaseBranch); err != nil {
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String("refs/heads/" + req.CommitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = ga.client.Git.CreateRef(ctx, req.SourceRepo.Owner, req.SourceRepo.Repo, newRef)
	return ref, err
}


func (ga *GitHubAdapter) getTree(ctx context.Context, req *CreatePullRequest, ref *github.Reference) (tree *github.Tree, err error) {
	entries := []*github.TreeEntry{}

	for _, fileInfo := range req.ChangeSet {
		file, content, err := ga.getFileContent(ctx, fileInfo.Name)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &github.TreeEntry{Path: github.String(file), Type: github.String("blob"), Content: github.String(string(content)), Mode: github.String("100644")})
	}

	tree, _, err = ga.client.Git.CreateTree(ctx, req.SourceRepo.Owner, req.SourceRepo.Repo, *ref.Object.SHA, entries)
	return tree, err
}


func (ga *GitHubAdapter) getFileContent(ctx context.Context, fileArg string) (targetName string, modifiedContent []byte, err error) {
	targetName = fileArg
	modifiedContent = []byte("dummy modifications in source file")
	err = nil
	return
}


func (ga *GitHubAdapter) pushCommit(ctx context.Context, req *CreatePullRequest, ref *github.Reference, tree *github.Tree) (err error) {
	parent, _, err := ga.client.Repositories.GetCommit(ctx, req.SourceRepo.Owner, req.SourceRepo.Repo, *ref.Object.SHA, nil)
	if err != nil {
		return err
	}

	parent.Commit.SHA = parent.SHA

	author := &github.CommitAuthor{Name: &req.AuthorName, Email: &req.AuthorEmail}
	commit := &github.Commit{Author: author, Message: &req.CommitMessage, Tree: tree, Parents: []*github.Commit{parent.Commit}}
	newCommit, _, err := ga.client.Git.CreateCommit(ctx, req.SourceRepo.Owner, req.SourceRepo.Repo, commit)
	if err != nil {
		return err
	}

	ref.Object.SHA = newCommit.SHA
	_, _, err = ga.client.Git.UpdateRef(ctx, req.SourceRepo.Owner, req.SourceRepo.Repo, ref, false)
	return err
}


func (ga *GitHubAdapter) createPR(ctx context.Context, req *CreatePullRequest) (pr *github.PullRequest, err error) {
	var commitBranch, prRepoOwner, prRepo string
	if req.PRSubject == "" {
		return nil, errors.New("missing `-pr-title` flag; skipping PR creation")
	}
	commitBranch = req.CommitBranch
	if req.TargetRepo.Owner != "" && req.TargetRepo.Owner != req.SourceRepo.Owner {
		commitBranch = fmt.Sprintf("%s:%s", req.SourceRepo.Owner, req.CommitBranch)
	} else {
		prRepoOwner = req.SourceRepo.Owner
	}

	if req.TargetRepo.Owner == "" {
		prRepo = req.SourceRepo.Repo
	}

	newPR := &github.NewPullRequest{
		Title:               &req.PRSubject,
		Head:                &commitBranch,
		Base:                &req.TargetBranch,
		Body:                &req.PRDescription,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err = ga.client.PullRequests.Create(ctx, prRepoOwner, prRepo, newPR)
	if err != nil {
		return nil, err
	}

	return
}

func (ga *GitHubAdapter) CreatePullRequest(ctx context.Context, req *CreatePullRequest) (*CreatePullResponse, error) {

	res := &CreatePullResponse{Status: failure, URL: ""}

	err := ga.connect(ctx)
	if err != nil {
		res.Err = err
		return res, err
	}
	defer ga.disconnect()

	ref, err := ga.getRef(ctx, req)
	if err != nil {
		res.Err = err
		return res, err
	}
	if ref == nil {
		res.Err = errors.New("No error where returned but the reference is nil")
		return res, err
	}

	tree, err := ga.getTree(ctx, req, ref)
	if err != nil {
		res.Err = err
		return res, err
	}

	if err := ga.pushCommit(ctx, req, ref, tree); err != nil {
		res.Err = err
		return res, err
	}

	pr, err := ga.createPR(ctx, req)
	if err != nil {
		res.Err = err
		return res, err
	}
	res.Status = success
	res.URL = pr.GetHTMLURL()
	res.Err = nil
	return res, nil
}
