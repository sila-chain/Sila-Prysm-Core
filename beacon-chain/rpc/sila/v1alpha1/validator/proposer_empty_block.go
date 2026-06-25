package validator

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getEmptyBlock(slot primitives.Slot) (interfaces.SignedBeaconBlock, error) {
	var sBlk interfaces.SignedBeaconBlock
	var err error
	epoch := slots.ToEpoch(slot)
	switch {
	case epoch >= params.BeaconConfig().GloasForkEpoch:
		sBlk, err = blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockGloas{Block: &silapb.BeaconBlockGloas{Body: &silapb.BeaconBlockBodyGloas{}}})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not initialize block for proposal: %v", err)
		}
	case epoch >= params.BeaconConfig().FuluForkEpoch:
		sBlk, err = blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockFulu{Block: &silapb.BeaconBlockElectra{Body: &silapb.BeaconBlockBodyElectra{}}})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not initialize block for proposal: %v", err)
		}
	case epoch >= params.BeaconConfig().ElectraForkEpoch:
		sBlk, err = blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockElectra{Block: &silapb.BeaconBlockElectra{Body: &silapb.BeaconBlockBodyElectra{}}})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not initialize block for proposal: %v", err)
		}
	case epoch >= params.BeaconConfig().DenebForkEpoch:
		sBlk, err = blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockDeneb{Block: &silapb.BeaconBlockDeneb{Body: &silapb.BeaconBlockBodyDeneb{}}})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not initialize block for proposal: %v", err)
		}
	case epoch >= params.BeaconConfig().CapellaForkEpoch:
		sBlk, err = blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockCapella{Block: &silapb.BeaconBlockCapella{Body: &silapb.BeaconBlockBodyCapella{}}})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not initialize block for proposal: %v", err)
		}
	case epoch >= params.BeaconConfig().BellatrixForkEpoch:
		sBlk, err = blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockBellatrix{Block: &silapb.BeaconBlockBellatrix{Body: &silapb.BeaconBlockBodyBellatrix{}}})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not initialize block for proposal: %v", err)
		}
	case epoch >= params.BeaconConfig().AltairForkEpoch:
		sBlk, err = blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockAltair{Block: &silapb.BeaconBlockAltair{Body: &silapb.BeaconBlockBodyAltair{}}})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not initialize block for proposal: %v", err)
		}
	default:
		sBlk, err = blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlock{Block: &silapb.BeaconBlock{Body: &silapb.BeaconBlockBody{}}})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not initialize block for proposal: %v", err)
		}
	}
	return sBlk, err
}
