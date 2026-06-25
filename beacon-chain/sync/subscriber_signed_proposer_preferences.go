package sync

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed"
	opfeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/operation"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"google.golang.org/protobuf/proto"
)

func (s *Service) signedProposerPreferencesSubscriber(_ context.Context, msg proto.Message) error {
	signedPreferences, ok := msg.(*silapb.SignedProposerPreferences)
	if !ok {
		return errWrongMessage
	}
	if signedPreferences == nil || signedPreferences.Message == nil {
		return errNilMessage
	}
	s.cfg.operationNotifier.OperationFeed().Send(&feed.Event{
		Type: opfeed.ProposerPreferencesReceived,
		Data: &opfeed.ProposerPreferencesReceivedData{
			Data: signedPreferences,
		},
	})
	return nil
}
