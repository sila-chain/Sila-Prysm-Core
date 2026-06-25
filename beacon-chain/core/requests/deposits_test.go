package requests_test

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/requests"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	eth "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func TestProcessDepositRequests(t *testing.T) {
	st, _ := util.DeterministicGenesisStateElectra(t, 1)
	sk, err := bls.RandKey()
	require.NoError(t, err)
	require.NoError(t, st.SetDepositRequestsStartIndex(1))

	t.Run("empty requests continues", func(t *testing.T) {
		newSt, err := requests.ProcessDepositRequests(t.Context(), st, []*silaenginev1.DepositRequest{})
		require.NoError(t, err)
		require.DeepEqual(t, newSt, st)
	})
	t.Run("nil request errors", func(t *testing.T) {
		_, err = requests.ProcessDepositRequests(t.Context(), st, []*silaenginev1.DepositRequest{nil})
		require.ErrorContains(t, "nil deposit request", err)
	})

	vals := st.Validators()
	vals[0].PublicKey = sk.PublicKey().Marshal()
	vals[0].WithdrawalCredentials[0] = params.BeaconConfig().SilaExecutionAddressWithdrawalPrefixByte
	require.NoError(t, st.SetValidators(vals))
	bals := st.Balances()
	bals[0] = params.BeaconConfig().MinActivationBalance + 2000
	require.NoError(t, st.SetBalances(bals))
	require.NoError(t, st.SetPendingDeposits(make([]*eth.PendingDeposit, 0))) // reset pbd as the deterministic state populates this already
	withdrawalCred := make([]byte, 32)
	withdrawalCred[0] = params.BeaconConfig().CompoundingWithdrawalPrefixByte
	depositMessage := &eth.DepositMessage{
		PublicKey:             sk.PublicKey().Marshal(),
		Amount:                1000,
		WithdrawalCredentials: withdrawalCred,
	}
	domain, err := signing.ComputeDomain(params.BeaconConfig().DomainDeposit, nil, nil)
	require.NoError(t, err)
	sr, err := signing.ComputeSigningRoot(depositMessage, domain)
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	reqs := []*silaenginev1.DepositRequest{
		{
			Pubkey:                depositMessage.PublicKey,
			Index:                 0,
			WithdrawalCredentials: depositMessage.WithdrawalCredentials,
			Amount:                depositMessage.Amount,
			Signature:             sig.Marshal(),
		},
	}
	st, err = requests.ProcessDepositRequests(t.Context(), st, reqs)
	require.NoError(t, err)

	pbd, err := st.PendingDeposits()
	require.NoError(t, err)
	require.Equal(t, 1, len(pbd))
	require.Equal(t, uint64(1000), pbd[0].Amount)
	require.DeepEqual(t, bytesutil.SafeCopyBytes(reqs[0].Pubkey), pbd[0].PublicKey)
}
