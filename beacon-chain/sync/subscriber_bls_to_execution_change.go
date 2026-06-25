package sync

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed"
	opfeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/operation"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *Service) blsToExecutionChangeSubscriber(_ context.Context, msg proto.Message) error {
	blsMsg, ok := msg.(*silapb.SignedBLSToExecutionChange)
	if !ok {
		return errors.Errorf("incorrect type of message received, wanted %T but got %T", &silapb.SignedBLSToExecutionChange{}, msg)
	}
	s.cfg.operationNotifier.OperationFeed().Send(&feed.Event{
		Type: opfeed.BLSToExecutionChangeReceived,
		Data: &opfeed.BLSToExecutionChangeReceivedData{
			Change: blsMsg,
		},
	})
	s.cfg.blsToExecPool.InsertBLSToExecChange(blsMsg)
	return nil
}
