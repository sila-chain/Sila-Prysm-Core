package sync

import (
	"context"
	"fmt"

	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// skipcq: SCC-U1000
func (s *Service) syncCommitteeMessageSubscriber(_ context.Context, msg proto.Message) error {
	m, ok := msg.(*silapb.SyncCommitteeMessage)
	if !ok {
		return fmt.Errorf("message was not type *silapb.SyncCommitteeMessage, type=%T", msg)
	}

	if m == nil {
		return errors.New("nil sync committee message")
	}

	return s.cfg.syncCommsPool.SaveSyncCommitteeMessage(m)
}
