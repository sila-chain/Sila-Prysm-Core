package monitor

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sirupsen/logrus"
)

// processExitsFromBlock logs the event when a tracked validators' exit was included in a block
func (s *Service) processExitsFromBlock(blk interfaces.ReadOnlyBeaconBlock) {
	s.RLock()
	defer s.RUnlock()
	for _, exit := range blk.Body().VoluntaryExits() {
		idx := exit.Exit.ValidatorIndex
		if s.trackedIndex(idx) {
			log.WithFields(logrus.Fields{
				"validatorIndex": idx,
				"slot":           blk.Slot(),
			}).Info("Voluntary exit was included")
		}
	}
}

// processExit logs the event when tracked validators' exit was processed
func (s *Service) processExit(exit *silapb.SignedVoluntaryExit) {
	idx := exit.Exit.ValidatorIndex
	s.RLock()
	defer s.RUnlock()
	if s.trackedIndex(idx) {
		log.WithFields(logrus.Fields{
			"validatorIndex": idx,
		}).Info("Voluntary exit was processed")
	}
}
