package blocks

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	v "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/validators"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/fuzz"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	gofuzz "github.com/google/gofuzz"
)

func TestFuzzProcessAttestationNoVerify_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	ctx := t.Context()
	state := &silapb.BeaconState{}
	att := &silapb.Attestation{}

	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(att)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		_, err = ProcessAttestationNoVerifySignature(ctx, s, att)
		_ = err
		fuzz.FreeMemory(i)
	}
}

func TestFuzzProcessBlockHeader_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	block := &silapb.SignedBeaconBlock{}

	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(block)

		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		if block.Block == nil || block.Block.Body == nil || block.Block.Body.Eth1Data == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(block)
		require.NoError(t, err)
		_, err = ProcessBlockHeader(t.Context(), s, wsb)
		_ = err
		fuzz.FreeMemory(i)
	}
}

func TestFuzzverifyDepositDataSigningRoot_10000(_ *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	var ba []byte
	var pubkey [fieldparams.BLSPubkeyLength]byte
	var sig [96]byte
	var domain [4]byte
	var p []byte
	var s []byte
	var d []byte
	for range 10000 {
		fuzzer.Fuzz(&ba)
		fuzzer.Fuzz(&pubkey)
		fuzzer.Fuzz(&sig)
		fuzzer.Fuzz(&domain)
		fuzzer.Fuzz(&p)
		fuzzer.Fuzz(&s)
		fuzzer.Fuzz(&d)
		err := verifySignature(ba, pubkey[:], sig[:], domain[:])
		_ = err
		err = verifySignature(ba, p, s, d)
		_ = err
	}
}

func TestFuzzProcessEth1DataInBlock_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	e := &silapb.Eth1Data{}
	state, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	for range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(e)
		s, err := ProcessEth1DataInBlock(t.Context(), state, e)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for state: %v and eth1data: %v", s, err, state, e)
		}
	}
}

func TestFuzzareEth1DataEqual_10000(_ *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	eth1data := &silapb.Eth1Data{}
	eth1data2 := &silapb.Eth1Data{}

	for range 10000 {
		fuzzer.Fuzz(eth1data)
		fuzzer.Fuzz(eth1data2)
		AreEth1DataEqual(eth1data, eth1data2)
		AreEth1DataEqual(eth1data, eth1data)
	}
}

func TestFuzzEth1DataHasEnoughSupport_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	eth1data := &silapb.Eth1Data{}
	var stateVotes []*silapb.Eth1Data
	for i := range 100000 {
		fuzzer.Fuzz(eth1data)
		fuzzer.Fuzz(&stateVotes)
		s, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Eth1DataVotes: stateVotes,
		})
		require.NoError(t, err)
		_, err = Eth1DataHasEnoughSupport(s, eth1data)
		_ = err
		fuzz.FreeMemory(i)
	}

}

func TestFuzzProcessBlockHeaderNoVerify_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	block := &silapb.BeaconBlock{}

	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(block)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		_, err = ProcessBlockHeaderNoVerify(t.Context(), s, block.Slot, block.ProposerIndex, block.ParentRoot, []byte{})
		_ = err
		fuzz.FreeMemory(i)
	}
}

func TestFuzzProcessRandao_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	b := &silapb.SignedBeaconBlock{}

	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(b)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		if b.Block == nil || b.Block.Body == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(b)
		require.NoError(t, err)
		r, err := ProcessRandao(t.Context(), s, wsb)
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and block: %v", r, err, state, b)
		}
		fuzz.FreeMemory(i)
	}
}

func TestFuzzProcessRandaoNoVerify_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	blockBody := &silapb.BeaconBlockBody{}

	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(blockBody)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		r, err := ProcessRandaoNoVerify(s, blockBody.RandaoReveal)
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and block: %v", r, err, state, blockBody)
		}
		fuzz.FreeMemory(i)
	}
}

func TestFuzzProcessProposerSlashings_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	p := &silapb.ProposerSlashing{}
	ctx := t.Context()
	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(p)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		r, err := ProcessProposerSlashings(ctx, s, []*silapb.ProposerSlashing{p}, v.ExitInformation(s))
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and slashing: %v", r, err, state, p)
		}
		fuzz.FreeMemory(i)
	}
}

func TestFuzzVerifyProposerSlashing_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	proposerSlashing := &silapb.ProposerSlashing{}
	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(proposerSlashing)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		err = VerifyProposerSlashing(s, proposerSlashing)
		_ = err
		fuzz.FreeMemory(i)
	}
}

func TestFuzzProcessAttesterSlashings_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	a := &silapb.AttesterSlashing{}
	ctx := t.Context()
	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(a)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		r, err := ProcessAttesterSlashings(ctx, s, []silapb.AttSlashing{a}, v.ExitInformation(s))
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and slashing: %v", r, err, state, a)
		}
		fuzz.FreeMemory(i)
	}
}

func TestFuzzVerifyAttesterSlashing_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	attesterSlashing := &silapb.AttesterSlashing{}
	ctx := t.Context()
	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(attesterSlashing)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		err = VerifyAttesterSlashing(ctx, s, attesterSlashing)
		_ = err
		fuzz.FreeMemory(i)
	}
}

func TestFuzzIsSlashableAttestationData_10000(_ *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	attestationData := &silapb.AttestationData{}
	attestationData2 := &silapb.AttestationData{}

	for range 10000 {
		fuzzer.Fuzz(attestationData)
		fuzzer.Fuzz(attestationData2)
		IsSlashableAttestationData(attestationData, attestationData2)
	}
}

func TestFuzzslashableAttesterIndices_10000(_ *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	attesterSlashing := &silapb.AttesterSlashing{}

	for range 10000 {
		fuzzer.Fuzz(attesterSlashing)
		SlashableAttesterIndices(attesterSlashing)
	}
}

func TestFuzzProcessAttestationsNoVerify_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	b := &silapb.SignedBeaconBlock{}
	ctx := t.Context()
	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(b)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		if b.Block == nil || b.Block.Body == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(b)
		require.NoError(t, err)
		r, err := ProcessAttestationsNoVerifySignature(ctx, s, wsb.Block())
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and block: %v", r, err, state, b)
		}
		fuzz.FreeMemory(i)
	}
}

func TestFuzzVerifyIndexedAttestationn_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	idxAttestation := &silapb.IndexedAttestation{}
	ctx := t.Context()
	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(idxAttestation)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		err = VerifyIndexedAttestation(ctx, s, idxAttestation)
		_ = err
		fuzz.FreeMemory(i)
	}
}

func TestFuzzverifyDeposit_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	deposit := &silapb.Deposit{}
	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(deposit)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		err = helpers.VerifyDeposit(s, deposit)
		_ = err
		fuzz.FreeMemory(i)
	}
}

func TestFuzzProcessVoluntaryExits_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	e := &silapb.SignedVoluntaryExit{}
	ctx := t.Context()
	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(e)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		r, err := ProcessVoluntaryExits(ctx, s, []*silapb.SignedVoluntaryExit{e}, v.ExitInformation(s))
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and exit: %v", r, err, state, e)
		}
		fuzz.FreeMemory(i)
	}
}

func TestFuzzProcessVoluntaryExitsNoVerify_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	e := &silapb.SignedVoluntaryExit{}
	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(e)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		r, err := ProcessVoluntaryExits(t.Context(), s, []*silapb.SignedVoluntaryExit{e}, v.ExitInformation(s))
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and block: %v", r, err, state, e)
		}
		fuzz.FreeMemory(i)
	}
}

func TestFuzzVerifyExit_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	ve := &silapb.SignedVoluntaryExit{}
	rawVal := &silapb.Validator{}
	fork := &silapb.Fork{}
	var slot primitives.Slot

	for i := range 10000 {
		fuzzer.Fuzz(ve)
		fuzzer.Fuzz(rawVal)
		fuzzer.Fuzz(fork)
		fuzzer.Fuzz(&slot)

		state := &silapb.BeaconState{
			Slot:                  slot,
			Fork:                  fork,
			GenesisValidatorsRoot: params.BeaconConfig().ZeroHash[:],
		}
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)

		val, err := state_native.NewValidator(&silapb.Validator{})
		_ = err
		err = VerifyExitAndSignature(val, s, ve)
		_ = err
		fuzz.FreeMemory(i)
	}
}
