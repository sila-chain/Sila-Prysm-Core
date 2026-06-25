package synccommittee

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/container/queue"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// SaveSyncCommitteeMessage saves a sync committee message in to a priority queue.
// The priority queue capped at syncCommitteeMaxQueueSize contributions.
func (s *Store) SaveSyncCommitteeMessage(msg *silapb.SyncCommitteeMessage) error {
	if msg == nil {
		return errNilMessage
	}

	s.messageLock.Lock()
	defer s.messageLock.Unlock()

	item, err := s.messageCache.PopByKey(syncCommitteeKey(msg.Slot))
	if err != nil {
		return err
	}

	copied := silapb.CopySyncCommitteeMessage(msg)
	// Messages exist in the queue. Append instead of insert new.
	if item != nil {
		messages, ok := item.Value.([]*silapb.SyncCommitteeMessage)
		if !ok {
			return errors.New("not typed []silapb.SyncCommitteeMessage")
		}

		idx := -1
		for i, msg := range messages {
			if msg.ValidatorIndex == copied.ValidatorIndex {
				idx = i
				break
			}
		}
		if idx >= 0 {
			// Override the existing messages with a new one
			messages[idx] = copied
		} else {
			// Append the new message
			messages = append(messages, copied)
			savedSyncCommitteeMessageTotal.Inc()
		}

		return s.messageCache.Push(&queue.Item{
			Key:      syncCommitteeKey(msg.Slot),
			Value:    messages,
			Priority: int64(msg.Slot),
		})
	}

	// Message does not exist. Insert new.
	if err := s.messageCache.Push(&queue.Item{
		Key:      syncCommitteeKey(msg.Slot),
		Value:    []*silapb.SyncCommitteeMessage{copied},
		Priority: int64(msg.Slot),
	}); err != nil {
		return err
	}
	savedSyncCommitteeMessageTotal.Inc()

	// Trim messages in queue down to syncCommitteeMaxQueueSize.
	if s.messageCache.Len() > syncCommitteeMaxQueueSize {
		if _, err := s.messageCache.Pop(); err != nil {
			return err
		}
	}

	return nil
}

// SyncCommitteeMessages returns sync committee messages by slot from the priority queue.
// When calling this method a copy is avoided as the caller is assumed to be only reading the
// messages from the store rather than modifying it.
func (s *Store) SyncCommitteeMessages(slot primitives.Slot) ([]*silapb.SyncCommitteeMessage, error) {
	s.messageLock.RLock()
	defer s.messageLock.RUnlock()

	item := s.messageCache.RetrieveByKey(syncCommitteeKey(slot))
	if item == nil {
		return nil, nil
	}

	messages, ok := item.Value.([]*silapb.SyncCommitteeMessage)
	if !ok {
		return nil, errors.New("not typed []silapb.SyncCommitteeMessage")
	}

	return messages, nil
}
