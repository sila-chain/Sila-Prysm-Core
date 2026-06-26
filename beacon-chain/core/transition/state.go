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
	"github.com/sila-chain/Sila-Consensus-Core/v7/container/trie"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// GenesisBeaconState gets called when MinGenesisActiveValidatorCount count of
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
func GenesisBeaconState(ctx context.Context, deposits []*silapb.Deposit, genesisTime uint64, silaexecData *silapb.SilaExecutionData) (state.BeaconState, error) {
	st, err := EmptyGenesisState()
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

	return OptimizedGenesisBeaconState(genesisTime, st, st.SilaExecutionData())
}

// PreminedGenesisBeaconState works almost exactly like GenesisBeaconState, except that it assumes that genesis deposits
// are not represented in the sila deposit and are only found in the genesis state validator registry. In order
// to ensure the deposit root and count match the empty sila deposit deployed in a testnet genesis block, the root
// of an empty deposit trie is computed and used as SilaExecutionData.deposit_root, and the deposit count is set to 0.
func PreminedGenesisBeaconState(ctx context.Context, deposits []*silapb.Deposit, genesisTime uint64, silaexecData *silapb.SilaExecutionData) (state.BeaconState, error) {
	st, err := EmptyGenesisState()
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

	t, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	if err != nil {
		return nil, err
	}
	dr, err := t.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	if err := st.SetSilaExecutionData(&silapb.SilaExecutionData{DepositRoot: dr[:], BlockHash: silaexecData.BlockHash}); err != nil {
		return nil, err
	}
	if err := st.SetSilaExecutionDepositIndex(0); err != nil {
		return nil, err
	}
	return OptimizedGenesisBeaconState(genesisTime, st, st.SilaExecutionData())
}

// OptimizedGenesisBeaconState is used to create a state that has already processed deposits. This is to efficiently
// create a mainnet state at chainstart.
func OptimizedGenesisBeaconState(genesisTime uint64, preState state.BeaconState, silaexecData *silapb.SilaExecutionData) (state.BeaconState, error) {
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

	st := &silapb.BeaconState{
		// Misc fields.
		Slot:                  0,
		GenesisTime:           genesisTime,
		GenesisValidatorsRoot: genesisValidatorsRoot[:],

		Fork: &silapb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
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

		HistoricalRoots:           [][]byte{},
		BlockRoots:                blockRoots,
		StateRoots:                stateRoots,
		Slashings:                 slashings,
		CurrentEpochAttestations:  []*silapb.PendingAttestation{},
		PreviousEpochAttestations: []*silapb.PendingAttestation{},

		// SilaExecution data.
		SilaExecutionData:         silaexecData,
		SilaExecutionDataVotes:    []*silapb.SilaExecutionData{},
		SilaExecutionDepositIndex: preState.SilaExecutionDepositIndex(),
	}

	bodyRoot, err := (&silapb.BeaconBlockBody{
		RandaoReveal: make([]byte, 96),
		SilaExecutionData: &silapb.SilaExecutionData{
			DepositRoot: make([]byte, 32),
			BlockHash:   make([]byte, 32),
		},
		Graffiti: make([]byte, 32),
	}).HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "could not hash tree root empty block body")
	}

	st.LatestBlockHeader = &silapb.BeaconBlockHeader{
		ParentRoot: zeroHash,
		StateRoot:  zeroHash,
		BodyRoot:   bodyRoot[:],
	}

	return state_native.InitializeFromProtoUnsafePhase0(st)
}

// EmptyGenesisState returns an empty beacon state object.
func EmptyGenesisState() (state.BeaconState, error) {
	blockRoots := make([][]byte, fieldparams.BlockRootsLength)
	for i := range blockRoots {
		blockRoots[i] = make([]byte, fieldparams.RootLength)
	}
	stateRoots := make([][]byte, fieldparams.StateRootsLength)
	for i := range stateRoots {
		stateRoots[i] = make([]byte, fieldparams.RootLength)
	}
	mixes := make([][]byte, fieldparams.RandaoMixesLength)
	for i := range mixes {
		mixes[i] = make([]byte, fieldparams.RootLength)
	}
	st := &silapb.BeaconState{
		// Misc fields.
		Slot: 0,
		Fork: &silapb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			Epoch:           0,
		},
		BlockRoots:  blockRoots,
		StateRoots:  stateRoots,
		RandaoMixes: mixes,
		// Validator registry fields.
		Validators: []*silapb.Validator{},
		Balances:   []uint64{},

		JustificationBits:         []byte{0},
		HistoricalRoots:           [][]byte{},
		CurrentEpochAttestations:  []*silapb.PendingAttestation{},
		PreviousEpochAttestations: []*silapb.PendingAttestation{},

		// SilaExecution data.
		SilaExecutionData:         &silapb.SilaExecutionData{},
		SilaExecutionDataVotes:    []*silapb.SilaExecutionData{},
		SilaExecutionDepositIndex: 0,
	}
	return state_native.InitializeFromProtoUnsafePhase0(st)
}

// IsValidGenesisState gets called whenever there's a deposit event,
// it checks whether there's enough effective balance to trigger and
// if the minimum genesis time arrived already.
//
// Spec pseudocode definition:
//
//	def is_valid_genesis_state(state: BeaconState) -> bool:
//	   if state.genesis_time < MIN_GENESIS_TIME:
//	       return False
//	   if len(get_active_validator_indices(state, GENESIS_EPOCH)) < MIN_GENESIS_ACTIVE_VALIDATOR_COUNT:
//	       return False
//	   return True
//
// This method has been modified from the spec to allow whole states not to be saved
// but instead only cache the relevant information.
func IsValidGenesisState(chainStartDepositCount, currentTime uint64) bool {
	if currentTime < params.BeaconConfig().MinGenesisTime {
		return false
	}
	if chainStartDepositCount < params.BeaconConfig().MinGenesisActiveValidatorCount {
		return false
	}
	return true
}
