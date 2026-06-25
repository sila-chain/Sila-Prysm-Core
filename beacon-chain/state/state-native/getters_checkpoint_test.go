package state_native

import (
	"testing"

	"github.com/sila-chain/go-bitfield"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	testtmpl "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/testing"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

func TestBeaconState_PreviousJustifiedCheckpointNil_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStatePreviousJustifiedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
		})
}

func TestBeaconState_PreviousJustifiedCheckpointNil_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStatePreviousJustifiedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeAltair(&silapb.BeaconStateAltair{})
		})
}

func TestBeaconState_PreviousJustifiedCheckpointNil_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStatePreviousJustifiedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeBellatrix(&silapb.BeaconStateBellatrix{})
		})
}

func TestBeaconState_PreviousJustifiedCheckpointNil_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStatePreviousJustifiedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeCapella(&silapb.BeaconStateCapella{})
		})
}

func TestBeaconState_PreviousJustifiedCheckpointNil_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStatePreviousJustifiedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeDeneb(&silapb.BeaconStateDeneb{})
		})
}

func TestBeaconState_PreviousJustifiedCheckpoint_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStatePreviousJustifiedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafePhase0(&silapb.BeaconState{PreviousJustifiedCheckpoint: cp})
		})
}

func TestBeaconState_PreviousJustifiedCheckpoint_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStatePreviousJustifiedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeAltair(&silapb.BeaconStateAltair{PreviousJustifiedCheckpoint: cp})
		})
}

func TestBeaconState_PreviousJustifiedCheckpoint_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStatePreviousJustifiedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeBellatrix(&silapb.BeaconStateBellatrix{PreviousJustifiedCheckpoint: cp})
		})
}

func TestBeaconState_PreviousJustifiedCheckpoint_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStatePreviousJustifiedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeCapella(&silapb.BeaconStateCapella{PreviousJustifiedCheckpoint: cp})
		})
}

func TestBeaconState_PreviousJustifiedCheckpoint_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStatePreviousJustifiedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeDeneb(&silapb.BeaconStateDeneb{PreviousJustifiedCheckpoint: cp})
		})
}

func TestBeaconState_CurrentJustifiedCheckpointNil_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateCurrentJustifiedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
		})
}

func TestBeaconState_CurrentJustifiedCheckpointNil_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateCurrentJustifiedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeAltair(&silapb.BeaconStateAltair{})
		})
}

func TestBeaconState_CurrentJustifiedCheckpointNil_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateCurrentJustifiedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeBellatrix(&silapb.BeaconStateBellatrix{})
		})
}

func TestBeaconState_CurrentJustifiedCheckpointNil_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateCurrentJustifiedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeCapella(&silapb.BeaconStateCapella{})
		})
}

func TestBeaconState_CurrentJustifiedCheckpointNil_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStateCurrentJustifiedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeDeneb(&silapb.BeaconStateDeneb{})
		})
}

func TestBeaconState_CurrentJustifiedCheckpoint_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateCurrentJustifiedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafePhase0(&silapb.BeaconState{CurrentJustifiedCheckpoint: cp})
		})
}

func TestBeaconState_CurrentJustifiedCheckpoint_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateCurrentJustifiedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeAltair(&silapb.BeaconStateAltair{CurrentJustifiedCheckpoint: cp})
		})
}

func TestBeaconState_CurrentJustifiedCheckpoint_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateCurrentJustifiedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeBellatrix(&silapb.BeaconStateBellatrix{CurrentJustifiedCheckpoint: cp})
		})
}

func TestBeaconState_CurrentJustifiedCheckpoint_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateCurrentJustifiedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeCapella(&silapb.BeaconStateCapella{CurrentJustifiedCheckpoint: cp})
		})
}

func TestBeaconState_CurrentJustifiedCheckpoint_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStateCurrentJustifiedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeDeneb(&silapb.BeaconStateDeneb{CurrentJustifiedCheckpoint: cp})
		})
}

func TestBeaconState_FinalizedCheckpointNil_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateFinalizedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
		})
}

func TestBeaconState_FinalizedCheckpointNil_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateFinalizedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeAltair(&silapb.BeaconStateAltair{})
		})
}

func TestBeaconState_FinalizedCheckpointNil_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateFinalizedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeBellatrix(&silapb.BeaconStateBellatrix{})
		})
}

func TestBeaconState_FinalizedCheckpointNil_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateFinalizedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeCapella(&silapb.BeaconStateCapella{})
		})
}

func TestBeaconState_FinalizedCheckpointNil_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStateFinalizedCheckpointNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeDeneb(&silapb.BeaconStateDeneb{})
		})
}

func TestBeaconState_FinalizedCheckpoint_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateFinalizedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafePhase0(&silapb.BeaconState{FinalizedCheckpoint: cp})
		})
}

func TestBeaconState_FinalizedCheckpoint_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateFinalizedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeAltair(&silapb.BeaconStateAltair{FinalizedCheckpoint: cp})
		})
}

func TestBeaconState_FinalizedCheckpoint_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateFinalizedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeBellatrix(&silapb.BeaconStateBellatrix{FinalizedCheckpoint: cp})
		})
}

func TestBeaconState_FinalizedCheckpoint_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateFinalizedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeCapella(&silapb.BeaconStateCapella{FinalizedCheckpoint: cp})
		})
}

func TestBeaconState_FinalizedCheckpoint_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStateFinalizedCheckpoint(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeDeneb(&silapb.BeaconStateDeneb{FinalizedCheckpoint: cp})
		})
}

func TestBeaconState_JustificationBitsNil_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateJustificationBitsNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
		})
}

func TestBeaconState_JustificationBitsNil_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateJustificationBitsNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeAltair(&silapb.BeaconStateAltair{})
		})
}

func TestBeaconState_JustificationBitsNil_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateJustificationBitsNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeBellatrix(&silapb.BeaconStateBellatrix{})
		})
}

func TestBeaconState_JustificationBitsNil_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateJustificationBitsNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeCapella(&silapb.BeaconStateCapella{})
		})
}

func TestBeaconState_JustificationBitsNil_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStateJustificationBitsNil(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoUnsafeDeneb(&silapb.BeaconStateDeneb{})
		})
}

func TestBeaconState_JustificationBits_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateJustificationBits(
		t,
		func(bits bitfield.Bitvector4) (state.BeaconState, error) {
			return InitializeFromProtoUnsafePhase0(&silapb.BeaconState{JustificationBits: bits})
		})
}

func TestBeaconState_JustificationBits_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateJustificationBits(
		t,
		func(bits bitfield.Bitvector4) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeAltair(&silapb.BeaconStateAltair{JustificationBits: bits})
		})
}

func TestBeaconState_JustificationBits_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateJustificationBits(
		t,
		func(bits bitfield.Bitvector4) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeBellatrix(&silapb.BeaconStateBellatrix{JustificationBits: bits})
		})
}

func TestBeaconState_JustificationBits_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateJustificationBits(
		t,
		func(bits bitfield.Bitvector4) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeCapella(&silapb.BeaconStateCapella{JustificationBits: bits})
		})
}

func TestBeaconState_JustificationBits_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStateJustificationBits(
		t,
		func(bits bitfield.Bitvector4) (state.BeaconState, error) {
			return InitializeFromProtoUnsafeDeneb(&silapb.BeaconStateDeneb{JustificationBits: bits})
		})
}
