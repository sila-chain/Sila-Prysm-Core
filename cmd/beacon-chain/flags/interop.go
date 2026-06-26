package flags

import (
	"github.com/urfave/cli/v2"
)

var (
	// InteropMockSilaDataVotesFlag enables mocking the silaexec proof-of-work chain data put into blocks by proposers.
	InteropMockSilaDataVotesFlag = &cli.BoolFlag{
		Name:  "interop-silaData-votes",
		Usage: "Enable mocking of silaexec data votes for proposers to package into blocks",
	}
)
