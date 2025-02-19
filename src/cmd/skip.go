package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/git-town/git-town/v14/src/cli/dialog/components"
	"github.com/git-town/git-town/v14/src/cli/flags"
	"github.com/git-town/git-town/v14/src/cli/print"
	"github.com/git-town/git-town/v14/src/cmd/cmdhelpers"
	"github.com/git-town/git-town/v14/src/execute"
	"github.com/git-town/git-town/v14/src/hosting"
	"github.com/git-town/git-town/v14/src/hosting/hostingdomain"
	"github.com/git-town/git-town/v14/src/messages"
	"github.com/git-town/git-town/v14/src/skip"
	"github.com/git-town/git-town/v14/src/validate"
	"github.com/git-town/git-town/v14/src/vm/statefile"
	"github.com/spf13/cobra"
)

const skipDesc = "Restart the last run Git Town command by skipping the current branch"

func skipCmd() *cobra.Command {
	addVerboseFlag, readVerboseFlag := flags.Verbose()
	cmd := cobra.Command{
		Use:     "skip",
		GroupID: "errors",
		Args:    cobra.NoArgs,
		Short:   skipDesc,
		Long:    cmdhelpers.Long(skipDesc),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeSkip(readVerboseFlag(cmd))
		},
	}
	addVerboseFlag(&cmd)
	return &cmd
}

func executeSkip(verbose bool) error {
	repo, err := execute.OpenRepo(execute.OpenRepoArgs{
		DryRun:           false,
		OmitBranchNames:  false,
		PrintCommands:    true,
		ValidateGitRepo:  true,
		ValidateIsOnline: false,
		Verbose:          verbose,
	})
	if err != nil {
		return err
	}
	dialogTestInputs := components.LoadTestInputs(os.Environ())
	repoStatus, err := repo.Backend.RepoStatus()
	if err != nil {
		return err
	}
	initialBranchesSnapshot, _, exit, err := execute.LoadRepoSnapshot(execute.LoadRepoSnapshotArgs{
		Config:                repo.Config,
		DialogTestInputs:      dialogTestInputs,
		Fetch:                 false,
		HandleUnfinishedState: false,
		Repo:                  repo,
		RepoStatus:            repoStatus,
		ValidateNoOpenChanges: false,
		Verbose:               verbose,
	})
	if err != nil || exit {
		return err
	}
	localBranches := initialBranchesSnapshot.Branches.LocalBranches().Names()
	repo.Config, exit, err = validate.Config(validate.ConfigArgs{
		Backend:            repo.Backend,
		BranchesToValidate: localBranches,
		FinalMessages:      repo.FinalMessages,
		LocalBranches:      localBranches,
		TestInputs:         dialogTestInputs,
		Unvalidated:        repo.Config,
	})
	if err != nil || exit {
		return err
	}
	runStateOpt, err := statefile.Load(repo.RootDir)
	if err != nil {
		return fmt.Errorf(messages.RunstateLoadProblem, err)
	}
	runState, hasRunState := runStateOpt.Get()
	if !hasRunState || runState.IsFinished() {
		return errors.New(messages.SkipNothingToDo)
	}
	if !runState.UnfinishedDetails.CanSkip {
		return errors.New(messages.SkipBranchHasConflicts)
	}
	var connector hostingdomain.Connector
	if originURL, hasOriginURL := repo.Config.OriginURL().Get(); hasOriginURL {
		connector, err = hosting.NewConnector(hosting.NewConnectorArgs{
			FullConfig:      repo.Config.Config,
			HostingPlatform: repo.Config.Config.HostingPlatform,
			Log:             print.Logger{},
			OriginURL:       originURL,
		})
		if err != nil {
			return err
		}
	}
	return skip.Execute(skip.ExecuteArgs{
		Backend:         repo.Backend,
		CommandsCounter: repo.CommandsCounter,
		Config:          repo.Config,
		Connector:       connector,
		CurrentBranch:   initialBranchesSnapshot.Active,
		FinalMessages:   repo.FinalMessages,
		Frontend:        repo.Frontend,
		HasOpenChanges:  repoStatus.OpenChanges,
		RootDir:         repo.RootDir,
		RunState:        runState,
		TestInputs:      dialogTestInputs,
		Verbose:         verbose,
	})
}
