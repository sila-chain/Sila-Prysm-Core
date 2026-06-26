package kv

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila/common"
)

func TestStore_SilaDeposit(t *testing.T) {
	db := setupDB(t)
	ctx := t.Context()
	contractAddress := common.Address{1, 2, 3}
	retrieved, err := db.SilaDepositAddress(ctx)
	require.NoError(t, err)
	assert.DeepEqual(t, []uint8(nil), retrieved, "Expected nil contract address")
	require.NoError(t, db.SaveSilaDepositAddress(ctx, contractAddress))
	retrieved, err = db.SilaDepositAddress(ctx)
	require.NoError(t, err)
	assert.Equal(t, contractAddress, common.BytesToAddress(retrieved), "Unexpected address")
	otherAddress := common.Address{4, 5, 6}
	err = db.SaveSilaDepositAddress(ctx, otherAddress)
	want := "cannot override sila deposit address"
	assert.ErrorContains(t, want, err, "Should not have been able to override old sila deposit address")
}
