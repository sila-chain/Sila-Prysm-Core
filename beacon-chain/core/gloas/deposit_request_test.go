package gloas

import (
	"bytes"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	stateTesting "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func TestProcessDepositRequests_EmptyAndNil(t *testing.T) {
	st := newGloasState(t, nil, nil)

	t.Run("empty requests continues", func(t *testing.T) {
		err := ProcessDepositRequests(t.Context(), st, []*silaenginev1.DepositRequest{}, nil)
		require.NoError(t, err)
	})

	t.Run("nil request errors", func(t *testing.T) {
		err := ProcessDepositRequests(t.Context(), st, []*silaenginev1.DepositRequest{nil}, nil)
		require.ErrorContains(t, "nil deposit request", err)
	})
}

func TestProcessDepositRequest_BuilderDepositAddsBuilder(t *testing.T) {
	sk, err := bls.RandKey()
	require.NoError(t, err)

	cred := builderWithdrawalCredentials()
	pd := stateTesting.GeneratePendingDeposit(t, sk, 1234, cred, 0)
	req := depositRequestFromPending(pd, 1)

	st := newGloasState(t, nil, nil)
	err = processDepositRequest(st, req, nil)
	require.NoError(t, err)

	idx, ok := st.BuilderIndexByPubkey(toBytes48(req.Pubkey))
	require.Equal(t, true, ok)

	builder, err := st.Builder(idx)
	require.NoError(t, err)
	require.NotNil(t, builder)
	require.DeepEqual(t, req.Pubkey, builder.Pubkey)
	require.DeepEqual(t, []byte{cred[0]}, builder.Version)
	require.DeepEqual(t, cred[12:], builder.ExecutionAddress)
	require.Equal(t, uint64(1234), uint64(builder.Balance))
	require.Equal(t, params.BeaconConfig().FarFutureEpoch, builder.WithdrawableEpoch)

	pending, err := st.PendingDeposits()
	require.NoError(t, err)
	require.Equal(t, 0, len(pending))
}

func TestProcessDepositRequest_ExistingBuilderIncreasesBalance(t *testing.T) {
	sk, err := bls.RandKey()
	require.NoError(t, err)

	pubkey := sk.PublicKey().Marshal()
	builders := []*silapb.Builder{
		{
			Pubkey:            pubkey,
			Version:           []byte{0},
			ExecutionAddress:  bytes.Repeat([]byte{0x11}, 20),
			Balance:           5,
			WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
		},
	}
	st := newGloasState(t, nil, builders)

	cred := validatorWithdrawalCredentials()
	pd := stateTesting.GeneratePendingDeposit(t, sk, 200, cred, 0)
	req := depositRequestFromPending(pd, 9)

	err = processDepositRequest(st, req, nil)
	require.NoError(t, err)

	idx, ok := st.BuilderIndexByPubkey(toBytes48(pubkey))
	require.Equal(t, true, ok)
	builder, err := st.Builder(idx)
	require.NoError(t, err)
	require.Equal(t, uint64(205), uint64(builder.Balance))

	pending, err := st.PendingDeposits()
	require.NoError(t, err)
	require.Equal(t, 0, len(pending))
}

func TestProcessDepositRequest_BuilderDepositWithExistingPendingDepositStaysPending(t *testing.T) {
	sk, err := bls.RandKey()
	require.NoError(t, err)

	validatorCred := validatorWithdrawalCredentials()
	builderCred := builderWithdrawalCredentials()
	existingPending := stateTesting.GeneratePendingDeposit(t, sk, 1234, validatorCred, 0)
	req := depositRequestFromPending(stateTesting.GeneratePendingDeposit(t, sk, 200, builderCred, 1), 9)

	st := newGloasState(t, nil, nil)
	require.NoError(t, st.SetPendingDeposits([]*silapb.PendingDeposit{existingPending}))

	err = processDepositRequest(st, req, nil)
	require.NoError(t, err)

	_, ok := st.BuilderIndexByPubkey(toBytes48(req.Pubkey))
	require.Equal(t, false, ok)

	pending, err := st.PendingDeposits()
	require.NoError(t, err)
	require.Equal(t, 2, len(pending))
	require.DeepEqual(t, existingPending.PublicKey, pending[0].PublicKey)
	require.DeepEqual(t, req.Pubkey, pending[1].PublicKey)
	require.DeepEqual(t, req.WithdrawalCredentials, pending[1].WithdrawalCredentials)
	require.Equal(t, req.Amount, pending[1].Amount)
}

func TestApplyDepositForBuilder_InvalidSignatureIgnoresDeposit(t *testing.T) {
	sk, err := bls.RandKey()
	require.NoError(t, err)

	cred := builderWithdrawalCredentials()
	st := newGloasState(t, nil, nil)
	err = applyDepositForNewBuilder(st, sk.PublicKey().Marshal(), cred[:], 100, make([]byte, 96), nil)
	require.NoError(t, err)

	_, ok := st.BuilderIndexByPubkey(toBytes48(sk.PublicKey().Marshal()))
	require.Equal(t, false, ok)
}

func TestPrefetchedDepositSigs_NilOrEmpty(t *testing.T) {
	require.Equal(t, true, prefetchedDepositSigs(nil) == nil)
	require.Equal(t, true, prefetchedDepositSigs(&silaenginev1.ExecutionRequests{}) == nil)
}

func TestPrefetchedDepositSigs_CacheMiss(t *testing.T) {
	rqs := &silaenginev1.ExecutionRequests{Deposits: []*silaenginev1.DepositRequest{{
		Pubkey:                make([]byte, 48),
		WithdrawalCredentials: make([]byte, 32),
		Signature:             make([]byte, 96),
	}}}
	require.Equal(t, true, prefetchedDepositSigs(rqs) == nil)
}

func TestPrefetchedDepositSigs_CacheHitAllValid(t *testing.T) {
	cache.DepositSig = cache.NewDepositSigCache()
	rqs := &silaenginev1.ExecutionRequests{Deposits: []*silaenginev1.DepositRequest{
		{Pubkey: make([]byte, 48), WithdrawalCredentials: make([]byte, 32), Signature: make([]byte, 96)},
		{Pubkey: make([]byte, 48), WithdrawalCredentials: make([]byte, 32), Signature: make([]byte, 96), Amount: 1},
	}}
	root, err := rqs.HashTreeRoot()
	require.NoError(t, err)
	cache.DepositSig.Put(root, []int{})

	got := prefetchedDepositSigs(rqs)
	require.DeepEqual(t, []bool{true, true}, got)
}

func TestPrefetchedDepositSigs_CacheHitMarksInvalid(t *testing.T) {
	cache.DepositSig = cache.NewDepositSigCache()
	rqs := &silaenginev1.ExecutionRequests{Deposits: []*silaenginev1.DepositRequest{
		{Pubkey: make([]byte, 48), WithdrawalCredentials: make([]byte, 32), Signature: make([]byte, 96)},
		{Pubkey: make([]byte, 48), WithdrawalCredentials: make([]byte, 32), Signature: make([]byte, 96), Amount: 1},
		{Pubkey: make([]byte, 48), WithdrawalCredentials: make([]byte, 32), Signature: make([]byte, 96), Amount: 2},
	}}
	root, err := rqs.HashTreeRoot()
	require.NoError(t, err)
	cache.DepositSig.Put(root, []int{1})

	got := prefetchedDepositSigs(rqs)
	require.DeepEqual(t, []bool{true, false, true}, got)
}

func TestPrefetchedDepositSigs_OutOfRangeIndexReturnsNil(t *testing.T) {
	cache.DepositSig = cache.NewDepositSigCache()
	rqs := &silaenginev1.ExecutionRequests{Deposits: []*silaenginev1.DepositRequest{
		{Pubkey: make([]byte, 48), WithdrawalCredentials: make([]byte, 32), Signature: make([]byte, 96)},
	}}
	root, err := rqs.HashTreeRoot()
	require.NoError(t, err)
	cache.DepositSig.Put(root, []int{5})

	require.Equal(t, true, prefetchedDepositSigs(rqs) == nil)
}

func TestProcessDepositRequests_PrefetchedInvalidSkipsBuilderAdd(t *testing.T) {
	sk, err := bls.RandKey()
	require.NoError(t, err)
	cred := builderWithdrawalCredentials()
	pd := stateTesting.GeneratePendingDeposit(t, sk, 1234, cred, 0)
	req := depositRequestFromPending(pd, 1)

	st := newGloasState(t, nil, nil)
	err = ProcessDepositRequests(t.Context(), st, []*silaenginev1.DepositRequest{req}, []bool{false})
	require.NoError(t, err)

	_, ok := st.BuilderIndexByPubkey(toBytes48(req.Pubkey))
	require.Equal(t, false, ok)
}

func TestProcessDepositRequests_PrefetchedValidBypassesBLS(t *testing.T) {
	sk, err := bls.RandKey()
	require.NoError(t, err)
	cred := builderWithdrawalCredentials()
	req := &silaenginev1.DepositRequest{
		Pubkey:                sk.PublicKey().Marshal(),
		WithdrawalCredentials: cred[:],
		Amount:                1234,
		Signature:             make([]byte, 96),
		Index:                 1,
	}

	st := newGloasState(t, nil, nil)
	err = ProcessDepositRequests(t.Context(), st, []*silaenginev1.DepositRequest{req}, []bool{true})
	require.NoError(t, err)

	_, ok := st.BuilderIndexByPubkey(toBytes48(req.Pubkey))
	require.Equal(t, true, ok)
}

func newGloasState(t *testing.T, validators []*silapb.Validator, builders []*silapb.Builder) state.BeaconState {
	t.Helper()

	st, err := state_native.InitializeFromProtoGloas(&silapb.BeaconStateGloas{
		DepositRequestsStartIndex: params.BeaconConfig().UnsetDepositRequestsStartIndex,
		Validators:                validators,
		Balances:                  make([]uint64, len(validators)),
		PendingDeposits:           []*silapb.PendingDeposit{},
		Builders:                  builders,
	})
	require.NoError(t, err)

	return st
}

func depositRequestFromPending(pd *silapb.PendingDeposit, index uint64) *silaenginev1.DepositRequest {
	return &silaenginev1.DepositRequest{
		Pubkey:                pd.PublicKey,
		WithdrawalCredentials: pd.WithdrawalCredentials,
		Amount:                pd.Amount,
		Signature:             pd.Signature,
		Index:                 index,
	}
}

func builderWithdrawalCredentials() [32]byte {
	var cred [32]byte
	cred[0] = params.BeaconConfig().BuilderWithdrawalPrefixByte
	copy(cred[12:], bytes.Repeat([]byte{0x22}, 20))
	return cred
}

func validatorWithdrawalCredentials() [32]byte {
	var cred [32]byte
	cred[0] = params.BeaconConfig().SilaExecutionAddressWithdrawalPrefixByte
	copy(cred[12:], bytes.Repeat([]byte{0x33}, 20))
	return cred
}

func toBytes48(b []byte) [48]byte {
	var out [48]byte
	copy(out[:], b)
	return out
}
