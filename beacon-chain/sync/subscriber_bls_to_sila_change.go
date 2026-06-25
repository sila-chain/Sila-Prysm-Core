package sync

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed"
	opfeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/operation"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *Service) blsToSilaChangeSubscriber(_ context.Context, msg proto.Message) error {
	blsMsg, ok := msg.(*silapb.SignedBLSToSilaChange)
	if !ok {
		return errors.Errorf("incorrect type of message received, wanted %T but got %T", &silapb.SignedBLSToSilaChange{}, msg)
	}
	s.cfg.operationNotifier.OperationFeed().Send(&feed.Event{
		Type: opfeed.BLSToSilaChangeReceived,
		Data: &opfeed.BLSToSilaChangeReceivedData{
			Change: blsMsg,
		},
	})
	s.cfg.blsToExecPool.InsertBLSToExecChange(blsMsg)
	return nil
}
