package synccommittee

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

var _ = Pool(&Store{})

// Pool defines the necessary methods for Sila sync pool to serve
// validators. In the current design, aggregated attestations
// are used by proposers and sync committee messages are used by
// sync aggregators.
type Pool interface {
	// Methods for Sync Contributions.
	SaveSyncCommitteeContribution(contr *silapb.SyncCommitteeContribution) error
	SyncCommitteeContributions(slot primitives.Slot) ([]*silapb.SyncCommitteeContribution, error)

	// Methods for Sync Committee Messages.
	SaveSyncCommitteeMessage(sig *silapb.SyncCommitteeMessage) error
	SyncCommitteeMessages(slot primitives.Slot) ([]*silapb.SyncCommitteeMessage, error)
}

// NewPool returns the sync committee store fulfilling the pool interface.
func NewPool() Pool {
	return NewStore()
}
