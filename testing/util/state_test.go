package util

import (
	"testing"

	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func TestNewBeaconState(t *testing.T) {
	st, err := NewBeaconState()
	require.NoError(t, err)
	b, err := st.MarshalSSZ()
	require.NoError(t, err)
	got := &silapb.BeaconState{}
	require.NoError(t, got.UnmarshalSSZ(b))
	assert.DeepEqual(t, st.ToProtoUnsafe(), got)
}

func TestNewBeaconStateAltair(t *testing.T) {
	st, err := NewBeaconStateAltair()
	require.NoError(t, err)
	b, err := st.MarshalSSZ()
	require.NoError(t, err)
	got := &silapb.BeaconStateAltair{}
	require.NoError(t, got.UnmarshalSSZ(b))
	assert.DeepEqual(t, st.ToProtoUnsafe(), got)
}

func TestNewBeaconStateBellatrix(t *testing.T) {
	st, err := NewBeaconStateBellatrix()
	require.NoError(t, err)
	b, err := st.MarshalSSZ()
	require.NoError(t, err)
	got := &silapb.BeaconStateBellatrix{}
	require.NoError(t, got.UnmarshalSSZ(b))
	assert.DeepEqual(t, st.ToProtoUnsafe(), got)
}

func TestNewBeaconStateCapella(t *testing.T) {
	st, err := NewBeaconStateCapella()
	require.NoError(t, err)
	b, err := st.MarshalSSZ()
	require.NoError(t, err)
	got := &silapb.BeaconStateCapella{}
	require.NoError(t, got.UnmarshalSSZ(b))
	assert.DeepEqual(t, st.ToProtoUnsafe(), got)
}

func TestNewBeaconStateDeneb(t *testing.T) {
	st, err := NewBeaconStateDeneb()
	require.NoError(t, err)
	b, err := st.MarshalSSZ()
	require.NoError(t, err)
	got := &silapb.BeaconStateDeneb{}
	require.NoError(t, got.UnmarshalSSZ(b))
	assert.DeepEqual(t, st.ToProtoUnsafe(), got)
}

func TestNewBeaconStateElectra(t *testing.T) {
	st, err := NewBeaconStateElectra()
	require.NoError(t, err)
	b, err := st.MarshalSSZ()
	require.NoError(t, err)
	got := &silapb.BeaconStateElectra{}
	require.NoError(t, got.UnmarshalSSZ(b))
	assert.DeepEqual(t, st.ToProtoUnsafe(), got)
}

func TestNewBeaconStateFulu(t *testing.T) {
	st, err := NewBeaconStateFulu()
	require.NoError(t, err)
	b, err := st.MarshalSSZ()
	require.NoError(t, err)
	got := &silapb.BeaconStateFulu{}
	require.NoError(t, got.UnmarshalSSZ(b))
	assert.DeepEqual(t, st.ToProtoUnsafe(), got)
}

func TestNewBeaconStateGloas(t *testing.T) {
	st, err := NewBeaconStateGloas()
	require.NoError(t, err)
	b, err := st.MarshalSSZ()
	require.NoError(t, err)
	got := &silapb.BeaconStateGloas{}
	require.NoError(t, got.UnmarshalSSZ(b))
	assert.DeepEqual(t, st.ToProtoUnsafe(), got)
}

func TestNewBeaconState_HashTreeRoot(t *testing.T) {
	st, err := NewBeaconState()
	require.NoError(t, err)
	_, err = st.HashTreeRoot(t.Context())
	require.NoError(t, err)
	st, err = NewBeaconStateAltair()
	require.NoError(t, err)
	_, err = st.HashTreeRoot(t.Context())
	require.NoError(t, err)
	st, err = NewBeaconStateBellatrix()
	require.NoError(t, err)
	_, err = st.HashTreeRoot(t.Context())
	require.NoError(t, err)
	st, err = NewBeaconStateCapella()
	require.NoError(t, err)
	_, err = st.HashTreeRoot(t.Context())
	require.NoError(t, err)
	st, err = NewBeaconStateDeneb()
	require.NoError(t, err)
	_, err = st.HashTreeRoot(t.Context())
	require.NoError(t, err)
	st, err = NewBeaconStateElectra()
	require.NoError(t, err)
	_, err = st.HashTreeRoot(t.Context())
	require.NoError(t, err)
}
