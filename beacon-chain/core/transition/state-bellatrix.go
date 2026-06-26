package transition

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/altair"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stateutil"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// GenesisBeaconStateBellatrix gets called when MinGenesisActiveValidatorCount count of
// full deposits were made to the sila deposit and the ChainStart log gets emitted.
//
// Spec pseudocode definition:
//
//	def initialize_beacon_state_from_silaexec(silaexec_block_hash: Bytes32,
//	                                    silaexec_timestamp: uint64,
//	                                    deposits: Sequence[Deposit]) -> BeaconState:
//	  fork = Fork(
//	      previous_version=GENESIS_FORK_VERSION,
//	      current_version=GENESIS_FORK_VERSION,
//	      epoch=GENESIS_EPOCH,
//	  )
//	  state = BeaconState(
//	      genesis_time=silaexec_timestamp + GENESIS_DELAY,
//	      fork=fork,
//	      sila_execution_data=SilaExecutionData(block_hash=silaexec_block_hash, deposit_count=uint64(len(deposits))),
//	      latest_block_header=BeaconBlockHeader(body_root=hash_tree_root(BeaconBlockBody())),
//	      randao_mixes=[silaexec_block_hash] * EPOCHS_PER_HISTORICAL_VECTOR,  # Seed RANDAO with SilaExecution entropy
//	  )
//
//	  # Process deposits
//	  leaves = list(map(lambda deposit: deposit.data, deposits))
//	  for index, deposit in enumerate(deposits):
//	      deposit_data_list = List[DepositData, 2**SILA_DEPOSIT_TREE_DEPTH](*leaves[:index + 1])
//	      state.sila_execution_data.deposit_root = hash_tree_root(deposit_data_list)
//	      process_deposit(state, deposit)
//
//	  # Process activations
//	  for index, validator in enumerate(state.validators):
//	      balance = state.balances[index]
//	      validator.effective_balance = min(balance - balance % EFFECTIVE_BALANCE_INCREMENT, MAX_EFFECTIVE_BALANCE)
//	      if validator.effective_balance == MAX_EFFECTIVE_BALANCE:
//	          validator.activation_eligibility_epoch = GENESIS_EPOCH
//	          validator.activation_epoch = GENESIS_EPOCH
//
//	  # Set genesis validators root for domain separation and chain versioning
//	  state.genesis_validators_root = hash_tree_root(state.validators)
//
//	  return state
//
// This method differs from the spec so as to process deposits beforehand instead of the end of the function.
func GenesisBeaconStateBellatrix(ctx context.Context, deposits []*silapb.Deposit, genesisTime uint64, silaexecData *silapb.SilaExecutionData, ep *silaenginev1.SilaPayload) (state.BeaconState, error) {
	st, err := EmptyGenesisStateBellatrix()
	if err != nil {
		return nil, err
	}

	// Process initial deposits.
	st, err = helpers.UpdateGenesisSilaExecutionData(st, deposits, silaexecData)
	if err != nil {
		return nil, err
	}

	st, err = altair.ProcessPreGenesisDeposits(ctx, st, deposits)
	if err != nil {
		return nil, errors.Wrap(err, "could not process validator deposits")
	}

	// After deposits have been processed, overwrite silaExecutionData to what is passed in. This allows us to "pre-mine" validators
	// without the deposit root and count mismatching the real sila deposit.
	if err := st.SetSilaExecutionData(silaexecData); err != nil {
		return nil, err
	}
	if err := st.SetSilaExecutionDepositIndex(silaexecData.DepositCount); err != nil {
		return nil, err
	}

	return OptimizedGenesisBeaconStateBellatrix(genesisTime, st, st.SilaExecutionData(), ep)
}

// OptimizedGenesisBeaconStateBellatrix is used to create a state that has already processed deposits. This is to efficiently
// create a mainnet state at chainstart.
func OptimizedGenesisBeaconStateBellatrix(genesisTime uint64, preState state.BeaconState, silaexecData *silapb.SilaExecutionData, ep *silaenginev1.SilaPayload) (state.BeaconState, error) {
	if silaexecData == nil {
		return nil, errors.New("no silaExecutionData provided for genesis state")
	}

	randaoMixes := make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector)
	for i := range randaoMixes {
		h := make([]byte, 32)
		copy(h, silaexecData.BlockHash)
		randaoMixes[i] = h
	}

	zeroHash := params.BeaconConfig().ZeroHash[:]

	activeIndexRoots := make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector)
	for i := range activeIndexRoots {
		activeIndexRoots[i] = zeroHash
	}

	blockRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := range blockRoots {
		blockRoots[i] = zeroHash
	}

	stateRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := range stateRoots {
		stateRoots[i] = zeroHash
	}

	slashings := make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector)

	compactValidators := stateutil.CompactValidatorsFromProto(preState.Validators())
	genesisValidatorsRoot, err := stateutil.ValidatorRegistryRoot(compactValidators)
	if err != nil {
		return nil, errors.Wrapf(err, "could not hash tree root genesis validators %v", err)
	}

	scores, err := preState.InactivityScores()
	if err != nil {
		return nil, err
	}
	scoresMissing := len(preState.Validators()) - len(scores)
	if scoresMissing > 0 {
		for range scoresMissing {
			scores = append(scores, 0)
		}
	}
	wep, err := blocks.WrappedSilaPayload(ep)
	if err != nil {
		return nil, err
	}
	eph, err := blocks.PayloadToHeader(wep)
	if err != nil {
		return nil, err
	}
	st := &silapb.BeaconStateBellatrix{
		// Misc fields.
		Slot:                  0,
		GenesisTime:           genesisTime,
		GenesisValidatorsRoot: genesisValidatorsRoot[:],

		Fork: &silapb.Fork{
			PreviousVersion: params.BeaconConfig().AltairForkVersion,
			CurrentVersion:  params.BeaconConfig().BellatrixForkVersion,
			Epoch:           0,
		},

		// Validator registry fields.
		Validators: preState.Validators(),
		Balances:   preState.Balances(),

		// Randomness and committees.
		RandaoMixes: randaoMixes,

		// Finality.
		PreviousJustifiedCheckpoint: &silapb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		JustificationBits: []byte{0},
		FinalizedCheckpoint: &silapb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},

		HistoricalRoots: [][]byte{},
		BlockRoots:      blockRoots,
		StateRoots:      stateRoots,
		Slashings:       slashings,

		// SilaExecution data.
		SilaExecutionData:                     silaexecData,
		SilaExecutionDataVotes:                []*silapb.SilaExecutionData{},
		SilaExecutionDepositIndex:             preState.SilaExecutionDepositIndex(),
		LatestSilaPayloadHeader: eph,
		InactivityScores:             scores,
	}

	bodyRoot, err := (&silapb.BeaconBlockBodyBellatrix{
		RandaoReveal: make([]byte, 96),
		SilaExecutionData: &silapb.SilaExecutionData{
			DepositRoot: make([]byte, 32),
			BlockHash:   make([]byte, 32),
		},
		Graffiti: make([]byte, 32),
		SyncAggregate: &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		},
		SilaPayload: &silaenginev1.SilaPayload{
			ParentHash:    make([]byte, 32),
			FeeRecipient:  make([]byte, 20),
			StateRoot:     make([]byte, 32),
			ReceiptsRoot:  make([]byte, 32),
			LogsBloom:     make([]byte, 256),
			PrevRandao:    make([]byte, 32),
			ExtraData:     make([]byte, 0),
			BaseFeePerGas: make([]byte, 32),
			BlockHash:     make([]byte, 32),
			Transactions:  make([][]byte, 0),
		},
	}).HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "could not hash tree root empty block body")
	}

	st.LatestBlockHeader = &silapb.BeaconBlockHeader{
		ParentRoot: zeroHash,
		StateRoot:  zeroHash,
		BodyRoot:   bodyRoot[:],
	}

	ist, err := state_native.InitializeFromProtoUnsafeBellatrix(st)
	if err != nil {
		return nil, err
	}
	sc, err := altair.NextSyncCommittee(context.Background(), ist)
	if err != nil {
		return nil, err
	}
	if err := ist.SetNextSyncCommittee(sc); err != nil {
		return nil, err
	}
	if err := ist.SetCurrentSyncCommittee(sc); err != nil {
		return nil, err
	}
	return ist, nil
}

// EmptyGenesisStateBellatrix returns an empty beacon state object.
func EmptyGenesisStateBellatrix() (state.BeaconState, error) {
	st := &silapb.BeaconStateBellatrix{
		// Misc fields.
		Slot: 0,
		Fork: &silapb.Fork{
			PreviousVersion: params.BeaconConfig().AltairForkVersion,
			CurrentVersion:  params.BeaconConfig().BellatrixForkVersion,
			Epoch:           0,
		},
		// Validator registry fields.
		Validators: []*silapb.Validator{},
		Balances:   []uint64{},

		JustificationBits: []byte{0},
		HistoricalRoots:   [][]byte{},

		// SilaExecution data.
		SilaExecutionData:         &silapb.SilaExecutionData{},
		SilaExecutionDataVotes:    []*silapb.SilaExecutionData{},
		SilaExecutionDepositIndex: 0,
		LatestSilaPayloadHeader: &silaenginev1.SilaPayloadHeader{
			ParentHash:       make([]byte, 32),
			FeeRecipient:     make([]byte, 20),
			StateRoot:        make([]byte, 32),
			ReceiptsRoot:     make([]byte, 32),
			LogsBloom:        make([]byte, 256),
			PrevRandao:       make([]byte, 32),
			ExtraData:        make([]byte, 0),
			BaseFeePerGas:    make([]byte, 32),
			BlockHash:        make([]byte, 32),
			TransactionsRoot: make([]byte, 32),
		},
	}

	return state_native.InitializeFromProtoUnsafeBellatrix(st)
}
