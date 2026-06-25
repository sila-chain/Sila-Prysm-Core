package debug

import (
	"testing"

	mockP2p "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p/testing"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/golang/protobuf/ptypes/empty"
)

func TestDebugServer_GetPeer(t *testing.T) {
	peersProvider := &mockP2p.MockPeersProvider{}
	mP2P := mockP2p.NewTestP2P(t)
	ds := &Server{
		PeersFetcher: peersProvider,
		PeerManager:  &mockP2p.MockPeerManager{BHost: mP2P.BHost},
	}
	firstPeer := peersProvider.Peers().All()[0]

	res, err := ds.GetPeer(t.Context(), &silapb.PeerRequest{PeerId: firstPeer.String()})
	require.NoError(t, err)
	require.Equal(t, firstPeer.String(), res.PeerId, "Unexpected peer ID")

	assert.Equal(t, int(silapb.PeerDirection_INBOUND), int(res.Direction), "Expected 1st peer to be an inbound connection")
	assert.Equal(t, silapb.ConnectionState_CONNECTED, res.ConnectionState, "Expected peer to be connected")
}

func TestDebugServer_ListPeers(t *testing.T) {
	peersProvider := &mockP2p.MockPeersProvider{}
	mP2P := mockP2p.NewTestP2P(t)
	ds := &Server{
		PeersFetcher: peersProvider,
		PeerManager:  &mockP2p.MockPeerManager{BHost: mP2P.BHost},
	}

	res, err := ds.ListPeers(t.Context(), &empty.Empty{})
	require.NoError(t, err)
	assert.Equal(t, 2, len(res.Responses))

	direction1 := res.Responses[0].Direction
	direction2 := res.Responses[1].Direction
	assert.Equal(t,
		true,
		direction1 == silapb.PeerDirection_INBOUND || direction2 == silapb.PeerDirection_INBOUND,
		"Expected an inbound peer")
	assert.Equal(t,
		true,
		direction1 == silapb.PeerDirection_OUTBOUND || direction2 == silapb.PeerDirection_OUTBOUND,
		"Expected an outbound peer")
	if len(res.Responses[0].ListeningAddresses) == 0 {
		t.Errorf("Expected 1st peer to have a multiaddress, instead they have no addresses")
	}
	if len(res.Responses[1].ListeningAddresses) == 0 {
		t.Errorf("Expected 2nd peer to have a multiaddress, instead they have no addresses")
	}
}
