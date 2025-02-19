package opcodes

import (
	"github.com/git-town/git-town/v14/src/git/gitdomain"
	"github.com/git-town/git-town/v14/src/vm/shared"
)

// CheckoutParent checks out the parent branch of the current branch.
type CheckoutParent struct {
	CurrentBranch           gitdomain.LocalBranchName
	undeclaredOpcodeMethods `exhaustruct:"optional"`
}

func (self *CheckoutParent) Run(args shared.RunArgs) error {
	parent, hasParent := args.Lineage.Parent(self.CurrentBranch).Get()
	if !hasParent || parent == self.CurrentBranch {
		return nil
	}
	return args.Frontend.CheckoutBranch(parent, false)
}
