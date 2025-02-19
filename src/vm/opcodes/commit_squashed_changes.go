package opcodes

import (
	"github.com/git-town/git-town/v14/src/git/gitdomain"
	"github.com/git-town/git-town/v14/src/vm/shared"
)

// CommitOpenChanges commits all open changes as a new commit.
type CommitSquashedChanges struct {
	Message                 gitdomain.CommitMessage
	undeclaredOpcodeMethods `exhaustruct:"optional"`
}

func (self *CommitSquashedChanges) Run(args shared.RunArgs) error {
	return args.Frontend.Commit(self.Message, "")
}
