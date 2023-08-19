package domain

import (
	"fmt"

	"github.com/git-town/git-town/v9/src/messages"
)

// BranchInfos contains the BranchInfos for all branches in a repo.
// Tracking branches on the origin remote don't get their own entry,
// they are listed in the `TrackingBranch` property of the local branch they track.
type BranchInfos []BranchInfo

// IsKnown indicates whether the given local branch is already known to this BranchesSyncStatus instance.
func (bs BranchInfos) HasLocalBranch(localBranch LocalBranchName) bool {
	for _, branch := range bs {
		if branch.Name == localBranch {
			return true
		}
	}
	return false
}

// HasMatchingRemoteBranchFor indicates whether there is already a remote branch matching the given local branch.
func (bs BranchInfos) HasMatchingRemoteBranchFor(localBranch LocalBranchName) bool {
	remoteName := localBranch.RemoteName()
	for _, branch := range bs {
		if branch.RemoteName == remoteName {
			return true
		}
	}
	return false
}

// LocalBranches provides only the branches that exist on the local machine.
func (bs BranchInfos) LocalBranches() BranchInfos {
	result := BranchInfos{}
	for _, branch := range bs {
		if branch.IsLocal() {
			result = append(result, branch)
		}
	}
	return result
}

// LocalBranchesWithDeletedTrackingBranches provides only the branches that exist locally and have a deleted tracking branch.
func (bs BranchInfos) LocalBranchesWithDeletedTrackingBranches() BranchInfos {
	result := BranchInfos{}
	for _, branch := range bs {
		if branch.SyncStatus == SyncStatusDeletedAtRemote {
			result = append(result, branch)
		}
	}
	return result
}

// LookupLocalBranch provides the branch with the given name if one exists.
// TODO: rename to FindLocalBranch.
func (bs BranchInfos) LookupLocalBranch(branchName LocalBranchName) *BranchInfo {
	for bi, branch := range bs {
		if branch.Name == branchName {
			return &bs[bi]
		}
	}
	return nil
}

// LookupLocalBranchWithTracking provides the local branch that has the given remote branch as its tracking branch
// or nil if no such branch exists.
// TODO: rename to FindLocalBranchWithTracking.
func (bs BranchInfos) LookupLocalBranchWithTracking(remoteBranch RemoteBranchName) *BranchInfo {
	for b, branch := range bs {
		if branch.RemoteName == remoteBranch {
			return &bs[b]
		}
	}
	return nil
}

// Names provides the names of all local branches in this BranchesSyncStatus instance.
func (bs BranchInfos) Names() LocalBranchNames {
	result := make(LocalBranchNames, 0, len(bs))
	for _, branch := range bs {
		if !branch.Name.IsEmpty() {
			result = append(result, branch.Name)
		}
	}
	return result
}

func (bs BranchInfos) Remove(branchName LocalBranchName) BranchInfos {
	result := BranchInfos{}
	for _, branch := range bs {
		if branch.Name != branchName {
			result = append(result, branch)
		}
	}
	return result
}

// Select provides the BranchSyncStatus elements with the given names.
func (bs BranchInfos) Select(names []LocalBranchName) (BranchInfos, error) {
	result := make(BranchInfos, len(names))
	for n, name := range names {
		branch := bs.LookupLocalBranch(name)
		if branch == nil {
			return result, fmt.Errorf(messages.BranchDoesntExist, name)
		}
		result[n] = *branch
	}
	return result, nil
}
