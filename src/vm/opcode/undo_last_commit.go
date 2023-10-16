package opcode

import "github.com/git-town/git-town/v9/src/vm/shared"

type UndoLastCommit struct {
	undeclaredOpcodeMethods
}

func (self *UndoLastCommit) Run(args shared.RunArgs) error {
	return args.Runner.Frontend.UndoLastCommit()
}
