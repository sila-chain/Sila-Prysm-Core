package monitor

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

func TestProcessExitsFromBlockTrackedIndices(t *testing.T) {
	hook := logTest.NewGlobal()
	s := &Service{
		TrackedValidators: map[primitives.ValidatorIndex]bool{
			1: true,
			2: true,
		},
	}

	exits := []*silapb.SignedVoluntaryExit{
		{
			Exit: &silapb.VoluntaryExit{
				ValidatorIndex: 3,
				Epoch:          1,
			},
		},
		{
			Exit: &silapb.VoluntaryExit{
				ValidatorIndex: 2,
				Epoch:          0,
			},
		},
	}

	block := &silapb.BeaconBlock{
		Body: &silapb.BeaconBlockBody{
			VoluntaryExits: exits,
		},
	}

	wb, err := blocks.NewBeaconBlock(block)
	require.NoError(t, err)
	s.processExitsFromBlock(wb)
	require.LogsContain(t, hook, "\"Voluntary exit was included\" package=beacon-chain/monitor slot=0 validatorIndex=2")
}

func TestProcessExitsFromBlockUntrackedIndices(t *testing.T) {
	hook := logTest.NewGlobal()
	s := &Service{
		TrackedValidators: map[primitives.ValidatorIndex]bool{
			1: true,
			2: true,
		},
	}

	exits := []*silapb.SignedVoluntaryExit{
		{
			Exit: &silapb.VoluntaryExit{
				ValidatorIndex: 3,
				Epoch:          1,
			},
		},
		{
			Exit: &silapb.VoluntaryExit{
				ValidatorIndex: 4,
				Epoch:          0,
			},
		},
	}

	block := &silapb.BeaconBlock{
		Body: &silapb.BeaconBlockBody{
			VoluntaryExits: exits,
		},
	}

	wb, err := blocks.NewBeaconBlock(block)
	require.NoError(t, err)
	s.processExitsFromBlock(wb)
	require.LogsDoNotContain(t, hook, "\"Voluntary exit was included\"")
}

func TestProcessExitP2PTrackedIndices(t *testing.T) {
	hook := logTest.NewGlobal()
	s := &Service{
		TrackedValidators: map[primitives.ValidatorIndex]bool{
			1: true,
			2: true,
		},
	}

	exit := &silapb.SignedVoluntaryExit{
		Exit: &silapb.VoluntaryExit{
			ValidatorIndex: 1,
			Epoch:          1,
		},
		Signature: make([]byte, 96),
	}
	s.processExit(exit)
	require.LogsContain(t, hook, "\"Voluntary exit was processed\" package=beacon-chain/monitor validatorIndex=1")
}

func TestProcessExitP2PUntrackedIndices(t *testing.T) {
	hook := logTest.NewGlobal()
	s := &Service{
		TrackedValidators: map[primitives.ValidatorIndex]bool{
			1: true,
			2: true,
		},
	}

	exit := &silapb.SignedVoluntaryExit{
		Exit: &silapb.VoluntaryExit{
			ValidatorIndex: 3,
			Epoch:          1,
		},
	}
	s.processExit(exit)
	require.LogsDoNotContain(t, hook, "\"Voluntary exit was processed\"")
}
