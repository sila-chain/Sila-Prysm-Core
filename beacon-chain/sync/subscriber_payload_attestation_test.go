package sync

import (
	"errors"
	"testing"

	mock "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/payloadattestation"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func TestPayloadAttestationSubscriber_WrongMessage(t *testing.T) {
	s := &Service{
		payloadAttestationCache: &cache.PayloadAttestationCache{},
		cfg:                     &config{chain: &mock.ChainService{}},
	}
	err := s.payloadAttestationSubscriber(t.Context(), &silapb.SignedVoluntaryExit{})
	require.ErrorIs(t, err, errWrongMessage)
}

func TestPayloadAttestationSubscriber_NilData(t *testing.T) {
	s := &Service{
		payloadAttestationCache: &cache.PayloadAttestationCache{},
		cfg:                     &config{chain: &mock.ChainService{}},
	}
	err := s.payloadAttestationSubscriber(t.Context(), &silapb.PayloadAttestationMessage{})
	require.ErrorIs(t, err, errNilMessage)
}

func TestPayloadAttestationSubscriber_NoPool(t *testing.T) {
	st, _ := util.DeterministicGenesisStateGloas(t, 64)
	ptc, err := st.PayloadCommitteeReadOnly(0)
	require.NoError(t, err)
	require.NotEmpty(t, ptc)

	s := &Service{
		payloadAttestationCache: &cache.PayloadAttestationCache{},
		cfg: &config{
			chain:                  &mock.ChainService{State: st},
			payloadAttestationPool: payloadattestation.NewPool(),
			operationNotifier:      &mock.MockOperationNotifier{},
		},
	}
	msg := &silapb.PayloadAttestationMessage{
		ValidatorIndex: ptc[0],
		Data: &silapb.PayloadAttestationData{
			BeaconBlockRoot: make([]byte, 32),
			Slot:            0,
		},
		Signature: make([]byte, 96),
	}
	require.NoError(t, s.payloadAttestationSubscriber(t.Context(), msg))
}

func TestPayloadAttestationSubscriber_HeadStateError(t *testing.T) {
	headErr := errors.New("head state unavailable")
	s := &Service{
		payloadAttestationCache: &cache.PayloadAttestationCache{},
		cfg: &config{
			chain: &mock.ChainService{
				HeadStateErr: headErr,
			},
			payloadAttestationPool: payloadattestation.NewPool(),
			operationNotifier:      &mock.MockOperationNotifier{},
		},
	}
	msg := &silapb.PayloadAttestationMessage{
		ValidatorIndex: 0,
		Data: &silapb.PayloadAttestationData{
			BeaconBlockRoot: make([]byte, 32),
			Slot:            0,
		},
		Signature: make([]byte, 96),
	}
	require.ErrorIs(t, s.payloadAttestationSubscriber(t.Context(), msg), headErr)
}

func TestPayloadAttestationSubscriber_ValidatorInPTC(t *testing.T) {
	st, _ := util.DeterministicGenesisStateGloas(t, 64)
	ptc, err := st.PayloadCommitteeReadOnly(0)
	require.NoError(t, err)
	require.NotEmpty(t, ptc)

	pool := payloadattestation.NewPool()
	s := &Service{
		payloadAttestationCache: &cache.PayloadAttestationCache{},
		cfg: &config{
			chain:                  &mock.ChainService{State: st},
			payloadAttestationPool: pool,
			operationNotifier:      &mock.MockOperationNotifier{},
		},
	}
	msg := &silapb.PayloadAttestationMessage{
		ValidatorIndex: ptc[0],
		Data: &silapb.PayloadAttestationData{
			BeaconBlockRoot: make([]byte, 32),
			Slot:            0,
		},
		Signature: make([]byte, 96),
	}
	require.NoError(t, s.payloadAttestationSubscriber(t.Context(), msg))
	require.Equal(t, 1, len(pool.PendingPayloadAttestations(0)))
}

func TestPayloadAttestationSubscriber_ValidatorNotInPTC(t *testing.T) {
	st, _ := util.DeterministicGenesisStateGloas(t, 64)
	ptc, err := st.PayloadCommitteeReadOnly(0)
	require.NoError(t, err)

	ptcSet := make(map[primitives.ValidatorIndex]bool, len(ptc))
	for _, idx := range ptc {
		ptcSet[idx] = true
	}
	var notInPTC primitives.ValidatorIndex
	for i := range primitives.ValidatorIndex(64) {
		if !ptcSet[i] {
			notInPTC = i
			break
		}
	}

	s := &Service{
		payloadAttestationCache: &cache.PayloadAttestationCache{},
		cfg: &config{
			chain:                  &mock.ChainService{State: st},
			payloadAttestationPool: payloadattestation.NewPool(),
			operationNotifier:      &mock.MockOperationNotifier{},
		},
	}
	msg := &silapb.PayloadAttestationMessage{
		ValidatorIndex: notInPTC,
		Data: &silapb.PayloadAttestationData{
			BeaconBlockRoot: make([]byte, 32),
			Slot:            0,
		},
		Signature: make([]byte, 96),
	}
	require.ErrorContains(t, "not in PTC", s.payloadAttestationSubscriber(t.Context(), msg))
}
