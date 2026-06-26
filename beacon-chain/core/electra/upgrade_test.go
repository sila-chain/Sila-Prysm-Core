package electra_test

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/electra"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/time"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
)

func TestUpgradeToElectra(t *testing.T) {
	st, _ := util.DeterministicGenesisStateDeneb(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	require.NoError(t, st.SetHistoricalRoots([][]byte{{1}}))
	vals := st.Validators()
	vals[0].ActivationEpoch = params.BeaconConfig().FarFutureEpoch
	vals[1].WithdrawalCredentials = []byte{params.BeaconConfig().CompoundingWithdrawalPrefixByte}
	require.NoError(t, st.SetValidators(vals))
	bals := st.Balances()
	bals[1] = params.BeaconConfig().MinActivationBalance + 1000
	require.NoError(t, st.SetBalances(bals))

	preForkState := st.Copy()
	mSt, err := electra.UpgradeToElectra(t.Context(), st)
	require.NoError(t, err)

	require.Equal(t, preForkState.GenesisTime(), mSt.GenesisTime())
	require.DeepSSZEqual(t, preForkState.GenesisValidatorsRoot(), mSt.GenesisValidatorsRoot())
	require.Equal(t, preForkState.Slot(), mSt.Slot())
	require.DeepSSZEqual(t, preForkState.LatestBlockHeader(), mSt.LatestBlockHeader())
	require.DeepSSZEqual(t, preForkState.BlockRoots(), mSt.BlockRoots())
	require.DeepSSZEqual(t, preForkState.StateRoots(), mSt.StateRoots())
	require.DeepSSZEqual(t, preForkState.Validators()[2:], mSt.Validators()[2:])
	require.DeepSSZEqual(t, preForkState.Balances()[2:], mSt.Balances()[2:])
	require.DeepSSZEqual(t, preForkState.SilaData(), mSt.SilaData())
	require.DeepSSZEqual(t, preForkState.SilaDataVotes(), mSt.SilaDataVotes())
	require.DeepSSZEqual(t, preForkState.SilaExecutionDepositIndex(), mSt.SilaExecutionDepositIndex())
	require.DeepSSZEqual(t, preForkState.RandaoMixes(), mSt.RandaoMixes())
	require.DeepSSZEqual(t, preForkState.Slashings(), mSt.Slashings())
	require.DeepSSZEqual(t, preForkState.JustificationBits(), mSt.JustificationBits())
	require.DeepSSZEqual(t, preForkState.PreviousJustifiedCheckpoint(), mSt.PreviousJustifiedCheckpoint())
	require.DeepSSZEqual(t, preForkState.CurrentJustifiedCheckpoint(), mSt.CurrentJustifiedCheckpoint())
	require.DeepSSZEqual(t, preForkState.FinalizedCheckpoint(), mSt.FinalizedCheckpoint())

	require.Equal(t, len(preForkState.Validators()), len(mSt.Validators()))

	preVal, err := preForkState.ValidatorAtIndex(0)
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxEffectiveBalance, preVal.EffectiveBalance)

	preVal2, err := preForkState.ValidatorAtIndex(1)
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxEffectiveBalance, preVal2.EffectiveBalance)

	mVal, err := mSt.ValidatorAtIndex(0)
	require.NoError(t, err)
	require.Equal(t, uint64(0), mVal.EffectiveBalance)

	mVal2, err := mSt.ValidatorAtIndex(1)
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MinActivationBalance, mVal2.EffectiveBalance)

	numValidators := mSt.NumValidators()
	p, err := mSt.PreviousEpochParticipation()
	require.NoError(t, err)
	require.DeepSSZEqual(t, make([]byte, numValidators), p)
	p, err = mSt.CurrentEpochParticipation()
	require.NoError(t, err)
	require.DeepSSZEqual(t, make([]byte, numValidators), p)
	s, err := mSt.InactivityScores()
	require.NoError(t, err)
	require.DeepSSZEqual(t, make([]uint64, numValidators), s)

	hr1 := preForkState.HistoricalRoots()
	hr2 := mSt.HistoricalRoots()
	require.DeepEqual(t, hr1, hr2)

	f := mSt.Fork()
	require.DeepSSZEqual(t, &silapb.Fork{
		PreviousVersion: st.Fork().CurrentVersion,
		CurrentVersion:  params.BeaconConfig().ElectraForkVersion,
		Epoch:           time.CurrentEpoch(st),
	}, f)
	csc, err := mSt.CurrentSyncCommittee()
	require.NoError(t, err)
	psc, err := preForkState.CurrentSyncCommittee()
	require.NoError(t, err)
	require.DeepSSZEqual(t, psc, csc)
	nsc, err := mSt.NextSyncCommittee()
	require.NoError(t, err)
	psc, err = preForkState.NextSyncCommittee()
	require.NoError(t, err)
	require.DeepSSZEqual(t, psc, nsc)

	header, err := mSt.LatestSilaPayloadHeader()
	require.NoError(t, err)
	protoHeader, ok := header.Proto().(*silaenginev1.SilaPayloadHeaderDeneb)
	require.Equal(t, true, ok)
	prevHeader, err := preForkState.LatestSilaPayloadHeader()
	require.NoError(t, err)
	txRoot, err := prevHeader.TransactionsRoot()
	require.NoError(t, err)

	wdRoot, err := prevHeader.WithdrawalsRoot()
	require.NoError(t, err)
	wanted := &silaenginev1.SilaPayloadHeaderDeneb{
		ParentHash:       prevHeader.ParentHash(),
		FeeRecipient:     prevHeader.FeeRecipient(),
		StateRoot:        prevHeader.StateRoot(),
		ReceiptsRoot:     prevHeader.ReceiptsRoot(),
		LogsBloom:        prevHeader.LogsBloom(),
		PrevRandao:       prevHeader.PrevRandao(),
		BlockNumber:      prevHeader.BlockNumber(),
		GasLimit:         prevHeader.GasLimit(),
		GasUsed:          prevHeader.GasUsed(),
		Timestamp:        prevHeader.Timestamp(),
		ExtraData:        prevHeader.ExtraData(),
		BaseFeePerGas:    prevHeader.BaseFeePerGas(),
		BlockHash:        prevHeader.BlockHash(),
		TransactionsRoot: txRoot,
		WithdrawalsRoot:  wdRoot,
	}
	require.DeepEqual(t, wanted, protoHeader)

	nwi, err := mSt.NextWithdrawalIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(0), nwi)

	lwvi, err := mSt.NextWithdrawalValidatorIndex()
	require.NoError(t, err)
	require.Equal(t, primitives.ValidatorIndex(0), lwvi)

	summaries, err := mSt.HistoricalSummaries()
	require.NoError(t, err)
	require.Equal(t, 0, len(summaries))

	startIndex, err := mSt.DepositRequestsStartIndex()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().UnsetDepositRequestsStartIndex, startIndex)

	balance, err := mSt.DepositBalanceToConsume()
	require.NoError(t, err)
	require.Equal(t, primitives.Gwei(0), balance)

	tab, err := helpers.TotalActiveBalance(t.Context(), mSt)
	require.NoError(t, err)

	ebtc, err := mSt.ExitBalanceToConsume()
	require.NoError(t, err)
	require.Equal(t, helpers.ActivationExitChurnLimit(primitives.Gwei(tab)), ebtc)

	eee, err := mSt.EarliestExitEpoch()
	require.NoError(t, err)
	require.Equal(t, helpers.ActivationExitEpoch(primitives.Epoch(1)), eee)

	cbtc, err := mSt.ConsolidationBalanceToConsume()
	require.NoError(t, err)
	require.Equal(t, helpers.ConsolidationChurnLimit(primitives.Gwei(tab)), cbtc)

	earliestConsolidationEpoch, err := mSt.EarliestConsolidationEpoch()
	require.NoError(t, err)
	require.Equal(t, helpers.ActivationExitEpoch(slots.ToEpoch(preForkState.Slot())), earliestConsolidationEpoch)

	pendingDeposits, err := mSt.PendingDeposits()
	require.NoError(t, err)
	require.Equal(t, 2, len(pendingDeposits))
	require.Equal(t, uint64(1000), pendingDeposits[1].Amount)

	numPendingPartialWithdrawals, err := mSt.NumPendingPartialWithdrawals()
	require.NoError(t, err)
	require.Equal(t, uint64(0), numPendingPartialWithdrawals)

	consolidations, err := mSt.PendingConsolidations()
	require.NoError(t, err)
	require.Equal(t, 0, len(consolidations))

}
