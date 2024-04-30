package gitlab

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/git-town/git-town/v14/src/cli/print"
	"github.com/git-town/git-town/v14/src/config/configdomain"
	"github.com/git-town/git-town/v14/src/git/gitdomain"
	"github.com/git-town/git-town/v14/src/git/giturl"
	. "github.com/git-town/git-town/v14/src/gohacks/prelude"
	"github.com/git-town/git-town/v14/src/hosting/hostingdomain"
	"github.com/git-town/git-town/v14/src/messages"
	"github.com/xanzy/go-gitlab"
)

// Connector provides standardized connectivity for the given repository (gitlab.com/owner/repo)
// via the GitLab API.
type Connector struct {
	client *gitlab.Client
	Data
	log print.Logger
}

func (self Connector) FindProposal(branch, target gitdomain.LocalBranchName) (Option[hostingdomain.Proposal], error) {
	opts := &gitlab.ListProjectMergeRequestsOptions{
		State:        gitlab.Ptr("opened"),
		SourceBranch: gitlab.Ptr(branch.String()),
		TargetBranch: gitlab.Ptr(target.String()),
	}
	mergeRequests, _, err := self.client.MergeRequests.ListProjectMergeRequests(self.projectPath(), opts)
	if err != nil {
		return None[hostingdomain.Proposal](), err
	}
	if len(mergeRequests) == 0 {
		return None[hostingdomain.Proposal](), nil
	}
	if len(mergeRequests) > 1 {
		return None[hostingdomain.Proposal](), fmt.Errorf(messages.ProposalMultipleFound, len(mergeRequests), branch, target)
	}
	proposal := parseMergeRequest(mergeRequests[0])
	return Some(proposal), nil
}

func (self Connector) SquashMergeProposal(number int, message gitdomain.CommitMessage) error {
	if number <= 0 {
		return errors.New(messages.ProposalNoNumberGiven)
	}
	self.log.Start(messages.HostingGitlabMergingViaAPI, number)
	// the GitLab API wants the full commit message in the body
	_, _, err := self.client.MergeRequests.AcceptMergeRequest(self.projectPath(), number, &gitlab.AcceptMergeRequestOptions{
		SquashCommitMessage: gitlab.Ptr(message.String()),
		Squash:              gitlab.Ptr(true),
		// the branch will be deleted by Git Town
		ShouldRemoveSourceBranch: gitlab.Ptr(false),
	})
	if err != nil {
		self.log.Failed(err)
		return err
	}
	self.log.Success()
	return nil
}

func (self Connector) UpdateProposalTarget(number int, target gitdomain.LocalBranchName) error {
	self.log.Start(messages.HostingGitlabUpdateMRViaAPI, number, target)
	_, _, err := self.client.MergeRequests.UpdateMergeRequest(self.projectPath(), number, &gitlab.UpdateMergeRequestOptions{
		TargetBranch: gitlab.Ptr(target.String()),
	})
	if err != nil {
		self.log.Failed(err)
		return err
	}
	self.log.Success()
	return nil
}

// NewGitlabConfig provides GitLab configuration data if the current repo is hosted on GitLab,
// otherwise nil.
func NewConnector(args NewConnectorArgs) (Connector, error) {
	gitlabData := Data{
		APIToken: args.APIToken,
		Data: hostingdomain.Data{
			Hostname:     args.OriginURL.Host,
			Organization: args.OriginURL.Org,
			Repository:   args.OriginURL.Repo,
		},
	}
	clientOptFunc := gitlab.WithBaseURL(gitlabData.baseURL())
	httpClient := gitlab.WithHTTPClient(&http.Client{})
	client, err := gitlab.NewOAuthClient(gitlabData.APIToken.String(), httpClient, clientOptFunc)
	if err != nil {
		return Connector{}, err
	}
	connector := Connector{
		Data:   gitlabData,
		client: client,
		log:    args.Log,
	}
	return connector, nil
}

type NewConnectorArgs struct {
	APIToken  Option[configdomain.GitLabToken]
	Log       print.Logger
	OriginURL giturl.Parts
}

func parseMergeRequest(mergeRequest *gitlab.MergeRequest) hostingdomain.Proposal {
	return hostingdomain.Proposal{
		Number:       mergeRequest.IID,
		Target:       gitdomain.NewLocalBranchName(mergeRequest.TargetBranch),
		Title:        mergeRequest.Title,
		MergeWithAPI: true,
	}
}
