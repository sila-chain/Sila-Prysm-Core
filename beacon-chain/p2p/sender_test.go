package p2p

import (
	"sync"
	"testing"
	"time"

	testp2p "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
)

func TestService_Send(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	p1 := testp2p.NewTestP2P(t)
	p2 := testp2p.NewTestP2P(t)
	p1.Connect(p2)

	svc := &Service{
		host: p1.BHost,
		cfg:  &Config{},
	}

	msg := &silapb.Fork{
		CurrentVersion:  []byte("fooo"),
		PreviousVersion: []byte("barr"),
		Epoch:           55,
	}

	// Register external listener which will repeat the message back.
	var wg sync.WaitGroup
	wg.Add(1)
	topic := "/testing/1"
	RPCTopicMappings[topic] = new(silapb.Fork)
	defer func() {
		delete(RPCTopicMappings, topic)
	}()
	p2.SetStreamHandler(topic+"/ssz_snappy", func(stream network.Stream) {
		rcvd := &silapb.Fork{}
		require.NoError(t, svc.Encoding().DecodeWithMaxLength(stream, rcvd))
		_, err := svc.Encoding().EncodeWithMaxLength(stream, rcvd)
		require.NoError(t, err)
		assert.NoError(t, stream.Close())
		wg.Done()
	})

	stream, err := svc.Send(t.Context(), msg, "/testing/1", p2.BHost.ID())
	require.NoError(t, err)

	util.WaitTimeout(&wg, 1*time.Second)

	rcvd := &silapb.Fork{}
	require.NoError(t, svc.Encoding().DecodeWithMaxLength(stream, rcvd))
	if !proto.Equal(rcvd, msg) {
		t.Errorf("Expected identical message to be received. got %v want %v", rcvd, msg)
	}
}
