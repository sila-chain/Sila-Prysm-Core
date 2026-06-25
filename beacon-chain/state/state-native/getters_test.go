package state_native

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	testtmpl "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/testing"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

func TestBeaconState_SlotDataRace_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateSlotDataRace(t, func() (state.BeaconState, error) {
		return InitializeFromProtoPhase0(&silapb.BeaconState{Slot: 1})
	})
}

func TestBeaconState_SlotDataRace_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateSlotDataRace(t, func() (state.BeaconState, error) {
		return InitializeFromProtoAltair(&silapb.BeaconStateAltair{Slot: 1})
	})
}

func TestBeaconState_SlotDataRace_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateSlotDataRace(t, func() (state.BeaconState, error) {
		return InitializeFromProtoBellatrix(&silapb.BeaconStateBellatrix{Slot: 1})
	})
}

func TestBeaconState_SlotDataRace_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateSlotDataRace(t, func() (state.BeaconState, error) {
		return InitializeFromProtoCapella(&silapb.BeaconStateCapella{Slot: 1})
	})
}

func TestBeaconState_SlotDataRace_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStateSlotDataRace(t, func() (state.BeaconState, error) {
		return InitializeFromProtoDeneb(&silapb.BeaconStateDeneb{Slot: 1})
	})
}

func TestBeaconState_MatchCurrentJustifiedCheckpt_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchCurrentJustifiedCheckptNative(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoPhase0(&silapb.BeaconState{CurrentJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_MatchCurrentJustifiedCheckpt_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchCurrentJustifiedCheckptNative(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoAltair(&silapb.BeaconStateAltair{CurrentJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_MatchCurrentJustifiedCheckpt_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchCurrentJustifiedCheckptNative(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoBellatrix(&silapb.BeaconStateBellatrix{CurrentJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_MatchCurrentJustifiedCheckpt_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchCurrentJustifiedCheckptNative(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoCapella(&silapb.BeaconStateCapella{CurrentJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_MatchPreviousJustifiedCheckpt_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchPreviousJustifiedCheckptNative(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoPhase0(&silapb.BeaconState{PreviousJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_MatchPreviousJustifiedCheckpt_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchPreviousJustifiedCheckptNative(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoAltair(&silapb.BeaconStateAltair{PreviousJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_MatchPreviousJustifiedCheckpt_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchPreviousJustifiedCheckptNative(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoBellatrix(&silapb.BeaconStateBellatrix{PreviousJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_MatchPreviousJustifiedCheckpt_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchPreviousJustifiedCheckptNative(
		t,
		func(cp *silapb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoCapella(&silapb.BeaconStateCapella{PreviousJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_ValidatorByPubkey_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorByPubkey(t, func() (state.BeaconState, error) {
		return InitializeFromProtoPhase0(&silapb.BeaconState{})
	})
}

func TestBeaconState_ValidatorByPubkey_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorByPubkey(t, func() (state.BeaconState, error) {
		return InitializeFromProtoAltair(&silapb.BeaconStateAltair{})
	})
}

func TestBeaconState_ValidatorByPubkey_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorByPubkey(t, func() (state.BeaconState, error) {
		return InitializeFromProtoBellatrix(&silapb.BeaconStateBellatrix{})
	})
}

func TestBeaconState_ValidatorByPubkey_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorByPubkey(t, func() (state.BeaconState, error) {
		return InitializeFromProtoCapella(&silapb.BeaconStateCapella{})
	})
}
