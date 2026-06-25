package migration

import (
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaapi/v1"
	eth "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

func V1Alpha1ConnectionStateToV1(connState eth.ConnectionState) silapb.ConnectionState {
	alphaString := connState.String()
	v1Value := silapb.ConnectionState_value[alphaString]
	return silapb.ConnectionState(v1Value)
}

func V1Alpha1PeerDirectionToV1(peerDirection eth.PeerDirection) (silapb.PeerDirection, error) {
	alphaString := peerDirection.String()
	if alphaString == eth.PeerDirection_UNKNOWN.String() {
		return 0, errors.New("peer direction unknown")
	}
	v1Value := silapb.PeerDirection_value[alphaString]
	return silapb.PeerDirection(v1Value), nil
}
