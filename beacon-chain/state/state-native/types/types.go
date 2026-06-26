package types

import (
	"fmt"

	consensus_types "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types"
	"github.com/pkg/errors"
)

// DataType signifies the data type of the field.
type DataType int

// List of current data types the state supports.
const (
	// BasicArray represents a simple array type for a field.
	BasicArray DataType = iota
	// CompositeArray represents a variable length array with
	// a non primitive type.
	CompositeArray
	// CompressedArray represents a variable length array which
	// can pack multiple elements into a leaf of the underlying
	// trie.
	CompressedArray
)

// FieldIndex represents the relevant field position in the
// state struct for a field.
type FieldIndex int

// String returns the name of the field index.
func (f FieldIndex) String() string {
	switch f {
	case GenesisTime:
		return "genesisTime"
	case GenesisValidatorsRoot:
		return "genesisValidatorsRoot"
	case Slot:
		return "slot"
	case Fork:
		return "fork"
	case LatestBlockHeader:
		return "latestBlockHeader"
	case BlockRoots:
		return "blockRoots"
	case StateRoots:
		return "stateRoots"
	case HistoricalRoots:
		return "historicalRoots"
	case SilaData:
		return "silaexecData"
	case SilaDataVotes:
		return "silaDataVotes"
	case SilaExecutionDepositIndex:
		return "silaExecutionDepositIndex"
	case Validators:
		return "validators"
	case Balances:
		return "balances"
	case RandaoMixes:
		return "randaoMixes"
	case Slashings:
		return "slashings"
	case PreviousEpochAttestations:
		return "previousEpochAttestations"
	case CurrentEpochAttestations:
		return "currentEpochAttestations"
	case PreviousEpochParticipationBits:
		return "previousEpochParticipationBits"
	case CurrentEpochParticipationBits:
		return "currentEpochParticipationBits"
	case JustificationBits:
		return "justificationBits"
	case PreviousJustifiedCheckpoint:
		return "previousJustifiedCheckpoint"
	case CurrentJustifiedCheckpoint:
		return "currentJustifiedCheckpoint"
	case FinalizedCheckpoint:
		return "finalizedCheckpoint"
	case InactivityScores:
		return "inactivityScores"
	case CurrentSyncCommittee:
		return "currentSyncCommittee"
	case NextSyncCommittee:
		return "nextSyncCommittee"
	case LatestSilaPayloadHeader:
		return "latestSilaPayloadHeader"
	case LatestSilaPayloadHeaderCapella:
		return "latestSilaPayloadHeaderCapella"
	case LatestSilaPayloadHeaderDeneb:
		return "latestSilaPayloadHeaderDeneb"
	case LatestSilaPayloadBid:
		return "latestSilaPayloadBid"
	case NextWithdrawalIndex:
		return "nextWithdrawalIndex"
	case NextWithdrawalValidatorIndex:
		return "nextWithdrawalValidatorIndex"
	case HistoricalSummaries:
		return "historicalSummaries"
	case DepositRequestsStartIndex:
		return "depositRequestsStartIndex"
	case DepositBalanceToConsume:
		return "depositBalanceToConsume"
	case ExitBalanceToConsume:
		return "exitBalanceToConsume"
	case EarliestExitEpoch:
		return "earliestExitEpoch"
	case ConsolidationBalanceToConsume:
		return "consolidationBalanceToConsume"
	case EarliestConsolidationEpoch:
		return "earliestConsolidationEpoch"
	case PendingDeposits:
		return "pendingDeposits"
	case PendingPartialWithdrawals:
		return "pendingPartialWithdrawals"
	case PendingConsolidations:
		return "pendingConsolidations"
	case ProposerLookahead:
		return "proposerLookahead"
	case Builders:
		return "builders"
	case NextWithdrawalBuilderIndex:
		return "nextWithdrawalBuilderIndex"
	case SilaPayloadAvailability:
		return "silaPayloadAvailability"
	case BuilderPendingPayments:
		return "builderPendingPayments"
	case BuilderPendingWithdrawals:
		return "builderPendingWithdrawals"
	case LatestBlockHash:
		return "latestBlockHash"
	case PayloadExpectedWithdrawals:
		return "payloadExpectedWithdrawals"
	case PTCWindow:
		return "ptcWindow"
	default:
		return fmt.Sprintf("unknown field index number: %d", f)
	}
}

// RealPosition denotes the position of the field in the beacon state.
// The value might differ for different state versions.
func (f FieldIndex) RealPosition() int {
	switch f {
	case GenesisTime:
		return 0
	case GenesisValidatorsRoot:
		return 1
	case Slot:
		return 2
	case Fork:
		return 3
	case LatestBlockHeader:
		return 4
	case BlockRoots:
		return 5
	case StateRoots:
		return 6
	case HistoricalRoots:
		return 7
	case SilaData:
		return 8
	case SilaDataVotes:
		return 9
	case SilaExecutionDepositIndex:
		return 10
	case Validators:
		return 11
	case Balances:
		return 12
	case RandaoMixes:
		return 13
	case Slashings:
		return 14
	case PreviousEpochAttestations, PreviousEpochParticipationBits:
		return 15
	case CurrentEpochAttestations, CurrentEpochParticipationBits:
		return 16
	case JustificationBits:
		return 17
	case PreviousJustifiedCheckpoint:
		return 18
	case CurrentJustifiedCheckpoint:
		return 19
	case FinalizedCheckpoint:
		return 20
	case InactivityScores:
		return 21
	case CurrentSyncCommittee:
		return 22
	case NextSyncCommittee:
		return 23
	case LatestSilaPayloadHeader, LatestSilaPayloadHeaderCapella, LatestSilaPayloadHeaderDeneb, LatestBlockHash:
		return 24
	case NextWithdrawalIndex:
		return 25
	case NextWithdrawalValidatorIndex:
		return 26
	case HistoricalSummaries:
		return 27
	case DepositRequestsStartIndex:
		return 28
	case DepositBalanceToConsume:
		return 29
	case ExitBalanceToConsume:
		return 30
	case EarliestExitEpoch:
		return 31
	case ConsolidationBalanceToConsume:
		return 32
	case EarliestConsolidationEpoch:
		return 33
	case PendingDeposits:
		return 34
	case PendingPartialWithdrawals:
		return 35
	case PendingConsolidations:
		return 36
	case ProposerLookahead:
		return 37
	case Builders:
		return 38
	case NextWithdrawalBuilderIndex:
		return 39
	case SilaPayloadAvailability:
		return 40
	case BuilderPendingPayments:
		return 41
	case BuilderPendingWithdrawals:
		return 42
	case LatestSilaPayloadBid:
		return 43
	case PayloadExpectedWithdrawals:
		return 44
	case PTCWindow:
		return 45
	default:
		return -1
	}
}

// ElemsInChunk returns the number of elements in the chunk (number of
// elements that are able to be packed).
func (f FieldIndex) ElemsInChunk() (uint64, error) {
	switch f {
	case Balances:
		return 4, nil
	default:
		return 0, errors.Errorf("field %d doesn't support element compression", f)
	}
}

// Below we define a set of useful enum values for the field
// indices of the beacon state. For example, genesisTime is the
// 0th field of the beacon state. This is helpful when we are
// updating the Merkle branches up the trie representation
// of the beacon state. The below field indexes correspond
// to the state.
const (
	GenesisTime FieldIndex = iota
	GenesisValidatorsRoot
	Slot
	Fork
	LatestBlockHeader
	BlockRoots
	StateRoots
	HistoricalRoots
	SilaData
	SilaDataVotes
	SilaExecutionDepositIndex
	Validators
	Balances
	RandaoMixes
	Slashings
	PreviousEpochAttestations
	CurrentEpochAttestations
	PreviousEpochParticipationBits
	CurrentEpochParticipationBits
	JustificationBits
	PreviousJustifiedCheckpoint
	CurrentJustifiedCheckpoint
	FinalizedCheckpoint
	InactivityScores
	CurrentSyncCommittee
	NextSyncCommittee
	LatestSilaPayloadHeader
	LatestSilaPayloadHeaderCapella
	LatestSilaPayloadHeaderDeneb
	LatestSilaPayloadBid // Gloas: SIP-7732
	NextWithdrawalIndex
	NextWithdrawalValidatorIndex
	HistoricalSummaries
	DepositRequestsStartIndex     // Electra: SIP-6110
	DepositBalanceToConsume       // Electra: SIP-7251
	ExitBalanceToConsume          // Electra: SIP-7251
	EarliestExitEpoch             // Electra: SIP-7251
	ConsolidationBalanceToConsume // Electra: SIP-7251
	EarliestConsolidationEpoch    // Electra: SIP-7251
	PendingDeposits               // Electra: SIP-7251
	PendingPartialWithdrawals     // Electra: SIP-7251
	PendingConsolidations         // Electra: SIP-7251
	ProposerLookahead             // Fulu: SIP-7917
	Builders                      // Gloas: SIP-7732
	NextWithdrawalBuilderIndex    // Gloas: SIP-7732
	SilaPayloadAvailability  // Gloas: SIP-7732
	BuilderPendingPayments        // Gloas: SIP-7732
	BuilderPendingWithdrawals     // Gloas: SIP-7732
	LatestBlockHash               // Gloas: SIP-7732
	PayloadExpectedWithdrawals    // Gloas: SIP-7732
	PTCWindow                     // Gloas: SIP-7732
)

// Enumerator keeps track of the number of states created since the node's start.
var Enumerator = &consensus_types.ThreadSafeEnumerator{}
