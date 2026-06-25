package depositsnapshot

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func TestPrunePendingDeposits_ZeroMerkleIndex(t *testing.T) {
	dc := Cache{}

	dc.pendingDeposits = []*silapb.DepositContainer{
		{Eth1BlockHeight: 2, Index: 2},
		{Eth1BlockHeight: 4, Index: 4},
		{Eth1BlockHeight: 6, Index: 6},
		{Eth1BlockHeight: 8, Index: 8},
		{Eth1BlockHeight: 10, Index: 10},
		{Eth1BlockHeight: 12, Index: 12},
	}

	dc.PrunePendingDeposits(t.Context(), 0)
	expected := []*silapb.DepositContainer{
		{Eth1BlockHeight: 2, Index: 2},
		{Eth1BlockHeight: 4, Index: 4},
		{Eth1BlockHeight: 6, Index: 6},
		{Eth1BlockHeight: 8, Index: 8},
		{Eth1BlockHeight: 10, Index: 10},
		{Eth1BlockHeight: 12, Index: 12},
	}
	assert.DeepEqual(t, expected, dc.pendingDeposits)
}

func TestPrunePendingDeposits_OK(t *testing.T) {
	dc := Cache{}

	dc.pendingDeposits = []*silapb.DepositContainer{
		{Eth1BlockHeight: 2, Index: 2},
		{Eth1BlockHeight: 4, Index: 4},
		{Eth1BlockHeight: 6, Index: 6},
		{Eth1BlockHeight: 8, Index: 8},
		{Eth1BlockHeight: 10, Index: 10},
		{Eth1BlockHeight: 12, Index: 12},
	}

	dc.PrunePendingDeposits(t.Context(), 6)
	expected := []*silapb.DepositContainer{
		{Eth1BlockHeight: 6, Index: 6},
		{Eth1BlockHeight: 8, Index: 8},
		{Eth1BlockHeight: 10, Index: 10},
		{Eth1BlockHeight: 12, Index: 12},
	}

	assert.DeepEqual(t, expected, dc.pendingDeposits)

	dc.pendingDeposits = []*silapb.DepositContainer{
		{Eth1BlockHeight: 2, Index: 2},
		{Eth1BlockHeight: 4, Index: 4},
		{Eth1BlockHeight: 6, Index: 6},
		{Eth1BlockHeight: 8, Index: 8},
		{Eth1BlockHeight: 10, Index: 10},
		{Eth1BlockHeight: 12, Index: 12},
	}

	dc.PrunePendingDeposits(t.Context(), 10)
	expected = []*silapb.DepositContainer{
		{Eth1BlockHeight: 10, Index: 10},
		{Eth1BlockHeight: 12, Index: 12},
	}

	assert.DeepEqual(t, expected, dc.pendingDeposits)
}

func TestPruneAllPendingDeposits(t *testing.T) {
	dc := Cache{}

	dc.pendingDeposits = []*silapb.DepositContainer{
		{Eth1BlockHeight: 2, Index: 2},
		{Eth1BlockHeight: 4, Index: 4},
		{Eth1BlockHeight: 6, Index: 6},
		{Eth1BlockHeight: 8, Index: 8},
		{Eth1BlockHeight: 10, Index: 10},
		{Eth1BlockHeight: 12, Index: 12},
	}

	dc.PruneAllPendingDeposits(t.Context())
	expected := []*silapb.DepositContainer{}

	assert.DeepEqual(t, expected, dc.pendingDeposits)
}

func TestPruneProofs_Ok(t *testing.T) {
	dc, err := New()
	require.NoError(t, err)

	deposits := []struct {
		blkNum  uint64
		deposit *silapb.Deposit
		index   int64
	}{
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk0"), 48)}},
			index: 0,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk1"), 48)}},
			index: 1,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk2"), 48)}},
			index: 2,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk3"), 48)}},
			index: 3,
		},
	}

	for _, ins := range deposits {
		assert.NoError(t, dc.InsertDeposit(t.Context(), ins.deposit, ins.blkNum, ins.index, [32]byte{}))
	}

	require.NoError(t, dc.PruneProofs(t.Context(), 1))

	assert.DeepEqual(t, [][]byte(nil), dc.deposits[0].Deposit.Proof)
	assert.DeepEqual(t, [][]byte(nil), dc.deposits[1].Deposit.Proof)
	assert.NotNil(t, dc.deposits[2].Deposit.Proof)
	assert.NotNil(t, dc.deposits[3].Deposit.Proof)
}

func TestPruneProofs_SomeAlreadyPruned(t *testing.T) {
	dc, err := New()
	require.NoError(t, err)

	deposits := []struct {
		blkNum  uint64
		deposit *silapb.Deposit
		index   int64
	}{
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: nil, Data: &silapb.Deposit_Data{
				PublicKey: bytesutil.PadTo([]byte("pk0"), 48)}},
			index: 0,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: nil, Data: &silapb.Deposit_Data{
				PublicKey: bytesutil.PadTo([]byte("pk1"), 48)}}, index: 1,
		},
		{
			blkNum:  0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(), Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk2"), 48)}},
			index:   2,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk3"), 48)}},
			index: 3,
		},
	}

	for _, ins := range deposits {
		assert.NoError(t, dc.InsertDeposit(t.Context(), ins.deposit, ins.blkNum, ins.index, [32]byte{}))
	}

	require.NoError(t, dc.PruneProofs(t.Context(), 2))

	assert.DeepEqual(t, [][]byte(nil), dc.deposits[2].Deposit.Proof)
}

func TestPruneProofs_PruneAllWhenDepositIndexTooBig(t *testing.T) {
	dc, err := New()
	require.NoError(t, err)

	deposits := []struct {
		blkNum  uint64
		deposit *silapb.Deposit
		index   int64
	}{
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk0"), 48)}},
			index: 0,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk1"), 48)}},
			index: 1,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk2"), 48)}},
			index: 2,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk3"), 48)}},
			index: 3,
		},
	}

	for _, ins := range deposits {
		assert.NoError(t, dc.InsertDeposit(t.Context(), ins.deposit, ins.blkNum, ins.index, [32]byte{}))
	}

	require.NoError(t, dc.PruneProofs(t.Context(), 99))

	assert.DeepEqual(t, [][]byte(nil), dc.deposits[0].Deposit.Proof)
	assert.DeepEqual(t, [][]byte(nil), dc.deposits[1].Deposit.Proof)
	assert.DeepEqual(t, [][]byte(nil), dc.deposits[2].Deposit.Proof)
	assert.DeepEqual(t, [][]byte(nil), dc.deposits[3].Deposit.Proof)
}

func TestPruneProofs_CorrectlyHandleLastIndex(t *testing.T) {
	dc, err := New()
	require.NoError(t, err)

	deposits := []struct {
		blkNum  uint64
		deposit *silapb.Deposit
		index   int64
	}{
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk0"), 48)}},
			index: 0,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk1"), 48)}},
			index: 1,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk2"), 48)}},
			index: 2,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk3"), 48)}},
			index: 3,
		},
	}

	for _, ins := range deposits {
		assert.NoError(t, dc.InsertDeposit(t.Context(), ins.deposit, ins.blkNum, ins.index, [32]byte{}))
	}

	require.NoError(t, dc.PruneProofs(t.Context(), 4))

	assert.DeepEqual(t, [][]byte(nil), dc.deposits[0].Deposit.Proof)
	assert.DeepEqual(t, [][]byte(nil), dc.deposits[1].Deposit.Proof)
	assert.DeepEqual(t, [][]byte(nil), dc.deposits[2].Deposit.Proof)
	assert.DeepEqual(t, [][]byte(nil), dc.deposits[3].Deposit.Proof)
}

func TestPruneAllProofs(t *testing.T) {
	dc, err := New()
	require.NoError(t, err)

	deposits := []struct {
		blkNum  uint64
		deposit *silapb.Deposit
		index   int64
	}{
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk0"), 48)}},
			index: 0,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk1"), 48)}},
			index: 1,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk2"), 48)}},
			index: 2,
		},
		{
			blkNum: 0,
			deposit: &silapb.Deposit{Proof: makeDepositProof(),
				Data: &silapb.Deposit_Data{PublicKey: bytesutil.PadTo([]byte("pk3"), 48)}},
			index: 3,
		},
	}

	for _, ins := range deposits {
		assert.NoError(t, dc.InsertDeposit(t.Context(), ins.deposit, ins.blkNum, ins.index, [32]byte{}))
	}

	dc.PruneAllProofs(t.Context())

	assert.DeepEqual(t, [][]byte(nil), dc.deposits[0].Deposit.Proof)
	assert.DeepEqual(t, [][]byte(nil), dc.deposits[1].Deposit.Proof)
	assert.DeepEqual(t, [][]byte(nil), dc.deposits[2].Deposit.Proof)
	assert.DeepEqual(t, [][]byte(nil), dc.deposits[3].Deposit.Proof)
}
