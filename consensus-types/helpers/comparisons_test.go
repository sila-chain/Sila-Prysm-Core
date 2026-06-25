package helpers

import (
	"testing"

	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

func TestForksEqual(t *testing.T) {
	tests := []struct {
		name string
		s    *silapb.Fork
		t    *silapb.Fork
		want bool
	}{
		{
			name: "both nil",
			s:    nil,
			t:    nil,
			want: true,
		},
		{
			name: "first nil",
			s:    nil,
			t:    &silapb.Fork{Epoch: 1},
			want: false,
		},
		{
			name: "second nil",
			s:    &silapb.Fork{Epoch: 1},
			t:    nil,
			want: false,
		},
		{
			name: "equal forks",
			s: &silapb.Fork{
				Epoch:           100,
				PreviousVersion: []byte{1, 2, 3, 4},
				CurrentVersion:  []byte{5, 6, 7, 8},
			},
			t: &silapb.Fork{
				Epoch:           100,
				PreviousVersion: []byte{1, 2, 3, 4},
				CurrentVersion:  []byte{5, 6, 7, 8},
			},
			want: true,
		},
		{
			name: "different epoch",
			s: &silapb.Fork{
				Epoch:           100,
				PreviousVersion: []byte{1, 2, 3, 4},
				CurrentVersion:  []byte{5, 6, 7, 8},
			},
			t: &silapb.Fork{
				Epoch:           200,
				PreviousVersion: []byte{1, 2, 3, 4},
				CurrentVersion:  []byte{5, 6, 7, 8},
			},
			want: false,
		},
		{
			name: "different previous version",
			s: &silapb.Fork{
				Epoch:           100,
				PreviousVersion: []byte{1, 2, 3, 4},
				CurrentVersion:  []byte{5, 6, 7, 8},
			},
			t: &silapb.Fork{
				Epoch:           100,
				PreviousVersion: []byte{9, 10, 11, 12},
				CurrentVersion:  []byte{5, 6, 7, 8},
			},
			want: false,
		},
		{
			name: "different current version",
			s: &silapb.Fork{
				Epoch:           100,
				PreviousVersion: []byte{1, 2, 3, 4},
				CurrentVersion:  []byte{5, 6, 7, 8},
			},
			t: &silapb.Fork{
				Epoch:           100,
				PreviousVersion: []byte{1, 2, 3, 4},
				CurrentVersion:  []byte{9, 10, 11, 12},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ForksEqual(tt.s, tt.t); got != tt.want {
				t.Errorf("ForksEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockHeadersEqual(t *testing.T) {
	tests := []struct {
		name string
		s    *silapb.BeaconBlockHeader
		t    *silapb.BeaconBlockHeader
		want bool
	}{
		{
			name: "both nil",
			s:    nil,
			t:    nil,
			want: true,
		},
		{
			name: "first nil",
			s:    nil,
			t:    &silapb.BeaconBlockHeader{Slot: 1},
			want: false,
		},
		{
			name: "second nil",
			s:    &silapb.BeaconBlockHeader{Slot: 1},
			t:    nil,
			want: false,
		},
		{
			name: "equal headers",
			s: &silapb.BeaconBlockHeader{
				Slot:          100,
				ProposerIndex: 50,
				ParentRoot:    []byte{1, 2, 3, 4},
				StateRoot:     []byte{5, 6, 7, 8},
				BodyRoot:      []byte{9, 10, 11, 12},
			},
			t: &silapb.BeaconBlockHeader{
				Slot:          100,
				ProposerIndex: 50,
				ParentRoot:    []byte{1, 2, 3, 4},
				StateRoot:     []byte{5, 6, 7, 8},
				BodyRoot:      []byte{9, 10, 11, 12},
			},
			want: true,
		},
		{
			name: "different slot",
			s: &silapb.BeaconBlockHeader{
				Slot:          100,
				ProposerIndex: 50,
				ParentRoot:    []byte{1, 2, 3, 4},
				StateRoot:     []byte{5, 6, 7, 8},
				BodyRoot:      []byte{9, 10, 11, 12},
			},
			t: &silapb.BeaconBlockHeader{
				Slot:          200,
				ProposerIndex: 50,
				ParentRoot:    []byte{1, 2, 3, 4},
				StateRoot:     []byte{5, 6, 7, 8},
				BodyRoot:      []byte{9, 10, 11, 12},
			},
			want: false,
		},
		{
			name: "different proposer index",
			s: &silapb.BeaconBlockHeader{
				Slot:          100,
				ProposerIndex: 50,
				ParentRoot:    []byte{1, 2, 3, 4},
				StateRoot:     []byte{5, 6, 7, 8},
				BodyRoot:      []byte{9, 10, 11, 12},
			},
			t: &silapb.BeaconBlockHeader{
				Slot:          100,
				ProposerIndex: 75,
				ParentRoot:    []byte{1, 2, 3, 4},
				StateRoot:     []byte{5, 6, 7, 8},
				BodyRoot:      []byte{9, 10, 11, 12},
			},
			want: false,
		},
		{
			name: "different parent root",
			s: &silapb.BeaconBlockHeader{
				Slot:          100,
				ProposerIndex: 50,
				ParentRoot:    []byte{1, 2, 3, 4},
				StateRoot:     []byte{5, 6, 7, 8},
				BodyRoot:      []byte{9, 10, 11, 12},
			},
			t: &silapb.BeaconBlockHeader{
				Slot:          100,
				ProposerIndex: 50,
				ParentRoot:    []byte{13, 14, 15, 16},
				StateRoot:     []byte{5, 6, 7, 8},
				BodyRoot:      []byte{9, 10, 11, 12},
			},
			want: false,
		},
		{
			name: "different state root",
			s: &silapb.BeaconBlockHeader{
				Slot:          100,
				ProposerIndex: 50,
				ParentRoot:    []byte{1, 2, 3, 4},
				StateRoot:     []byte{5, 6, 7, 8},
				BodyRoot:      []byte{9, 10, 11, 12},
			},
			t: &silapb.BeaconBlockHeader{
				Slot:          100,
				ProposerIndex: 50,
				ParentRoot:    []byte{1, 2, 3, 4},
				StateRoot:     []byte{13, 14, 15, 16},
				BodyRoot:      []byte{9, 10, 11, 12},
			},
			want: false,
		},
		{
			name: "different body root",
			s: &silapb.BeaconBlockHeader{
				Slot:          100,
				ProposerIndex: 50,
				ParentRoot:    []byte{1, 2, 3, 4},
				StateRoot:     []byte{5, 6, 7, 8},
				BodyRoot:      []byte{9, 10, 11, 12},
			},
			t: &silapb.BeaconBlockHeader{
				Slot:          100,
				ProposerIndex: 50,
				ParentRoot:    []byte{1, 2, 3, 4},
				StateRoot:     []byte{5, 6, 7, 8},
				BodyRoot:      []byte{13, 14, 15, 16},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BlockHeadersEqual(tt.s, tt.t); got != tt.want {
				t.Errorf("BlockHeadersEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEth1DataEqual(t *testing.T) {
	tests := []struct {
		name string
		s    *silapb.Eth1Data
		t    *silapb.Eth1Data
		want bool
	}{
		{
			name: "both nil",
			s:    nil,
			t:    nil,
			want: true,
		},
		{
			name: "first nil",
			s:    nil,
			t:    &silapb.Eth1Data{DepositCount: 1},
			want: false,
		},
		{
			name: "second nil",
			s:    &silapb.Eth1Data{DepositCount: 1},
			t:    nil,
			want: false,
		},
		{
			name: "equal eth1 data",
			s: &silapb.Eth1Data{
				DepositRoot:  []byte{1, 2, 3, 4},
				DepositCount: 100,
				BlockHash:    []byte{5, 6, 7, 8},
			},
			t: &silapb.Eth1Data{
				DepositRoot:  []byte{1, 2, 3, 4},
				DepositCount: 100,
				BlockHash:    []byte{5, 6, 7, 8},
			},
			want: true,
		},
		{
			name: "different deposit root",
			s: &silapb.Eth1Data{
				DepositRoot:  []byte{1, 2, 3, 4},
				DepositCount: 100,
				BlockHash:    []byte{5, 6, 7, 8},
			},
			t: &silapb.Eth1Data{
				DepositRoot:  []byte{9, 10, 11, 12},
				DepositCount: 100,
				BlockHash:    []byte{5, 6, 7, 8},
			},
			want: false,
		},
		{
			name: "different deposit count",
			s: &silapb.Eth1Data{
				DepositRoot:  []byte{1, 2, 3, 4},
				DepositCount: 100,
				BlockHash:    []byte{5, 6, 7, 8},
			},
			t: &silapb.Eth1Data{
				DepositRoot:  []byte{1, 2, 3, 4},
				DepositCount: 200,
				BlockHash:    []byte{5, 6, 7, 8},
			},
			want: false,
		},
		{
			name: "different block hash",
			s: &silapb.Eth1Data{
				DepositRoot:  []byte{1, 2, 3, 4},
				DepositCount: 100,
				BlockHash:    []byte{5, 6, 7, 8},
			},
			t: &silapb.Eth1Data{
				DepositRoot:  []byte{1, 2, 3, 4},
				DepositCount: 100,
				BlockHash:    []byte{9, 10, 11, 12},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Eth1DataEqual(tt.s, tt.t); got != tt.want {
				t.Errorf("Eth1DataEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingDepositsEqual(t *testing.T) {
	tests := []struct {
		name string
		s    *silapb.PendingDeposit
		t    *silapb.PendingDeposit
		want bool
	}{
		{
			name: "both nil",
			s:    nil,
			t:    nil,
			want: true,
		},
		{
			name: "first nil",
			s:    nil,
			t:    &silapb.PendingDeposit{Amount: 1},
			want: false,
		},
		{
			name: "second nil",
			s:    &silapb.PendingDeposit{Amount: 1},
			t:    nil,
			want: false,
		},
		{
			name: "equal pending deposits",
			s: &silapb.PendingDeposit{
				PublicKey:             []byte{1, 2, 3, 4},
				WithdrawalCredentials: []byte{5, 6, 7, 8},
				Amount:                32000000000,
				Signature:             []byte{9, 10, 11, 12},
				Slot:                  100,
			},
			t: &silapb.PendingDeposit{
				PublicKey:             []byte{1, 2, 3, 4},
				WithdrawalCredentials: []byte{5, 6, 7, 8},
				Amount:                32000000000,
				Signature:             []byte{9, 10, 11, 12},
				Slot:                  100,
			},
			want: true,
		},
		{
			name: "different public key",
			s: &silapb.PendingDeposit{
				PublicKey:             []byte{1, 2, 3, 4},
				WithdrawalCredentials: []byte{5, 6, 7, 8},
				Amount:                32000000000,
				Signature:             []byte{9, 10, 11, 12},
				Slot:                  100,
			},
			t: &silapb.PendingDeposit{
				PublicKey:             []byte{13, 14, 15, 16},
				WithdrawalCredentials: []byte{5, 6, 7, 8},
				Amount:                32000000000,
				Signature:             []byte{9, 10, 11, 12},
				Slot:                  100,
			},
			want: false,
		},
		{
			name: "different withdrawal credentials",
			s: &silapb.PendingDeposit{
				PublicKey:             []byte{1, 2, 3, 4},
				WithdrawalCredentials: []byte{5, 6, 7, 8},
				Amount:                32000000000,
				Signature:             []byte{9, 10, 11, 12},
				Slot:                  100,
			},
			t: &silapb.PendingDeposit{
				PublicKey:             []byte{1, 2, 3, 4},
				WithdrawalCredentials: []byte{13, 14, 15, 16},
				Amount:                32000000000,
				Signature:             []byte{9, 10, 11, 12},
				Slot:                  100,
			},
			want: false,
		},
		{
			name: "different amount",
			s: &silapb.PendingDeposit{
				PublicKey:             []byte{1, 2, 3, 4},
				WithdrawalCredentials: []byte{5, 6, 7, 8},
				Amount:                32000000000,
				Signature:             []byte{9, 10, 11, 12},
				Slot:                  100,
			},
			t: &silapb.PendingDeposit{
				PublicKey:             []byte{1, 2, 3, 4},
				WithdrawalCredentials: []byte{5, 6, 7, 8},
				Amount:                16000000000,
				Signature:             []byte{9, 10, 11, 12},
				Slot:                  100,
			},
			want: false,
		},
		{
			name: "different signature",
			s: &silapb.PendingDeposit{
				PublicKey:             []byte{1, 2, 3, 4},
				WithdrawalCredentials: []byte{5, 6, 7, 8},
				Amount:                32000000000,
				Signature:             []byte{9, 10, 11, 12},
				Slot:                  100,
			},
			t: &silapb.PendingDeposit{
				PublicKey:             []byte{1, 2, 3, 4},
				WithdrawalCredentials: []byte{5, 6, 7, 8},
				Amount:                32000000000,
				Signature:             []byte{13, 14, 15, 16},
				Slot:                  100,
			},
			want: false,
		},
		{
			name: "different slot",
			s: &silapb.PendingDeposit{
				PublicKey:             []byte{1, 2, 3, 4},
				WithdrawalCredentials: []byte{5, 6, 7, 8},
				Amount:                32000000000,
				Signature:             []byte{9, 10, 11, 12},
				Slot:                  100,
			},
			t: &silapb.PendingDeposit{
				PublicKey:             []byte{1, 2, 3, 4},
				WithdrawalCredentials: []byte{5, 6, 7, 8},
				Amount:                32000000000,
				Signature:             []byte{9, 10, 11, 12},
				Slot:                  200,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PendingDepositsEqual(tt.s, tt.t); got != tt.want {
				t.Errorf("PendingDepositsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingPartialWithdrawalsEqual(t *testing.T) {
	tests := []struct {
		name string
		s    *silapb.PendingPartialWithdrawal
		t    *silapb.PendingPartialWithdrawal
		want bool
	}{
		{
			name: "both nil",
			s:    nil,
			t:    nil,
			want: true,
		},
		{
			name: "first nil",
			s:    nil,
			t:    &silapb.PendingPartialWithdrawal{Index: 1},
			want: false,
		},
		{
			name: "second nil",
			s:    &silapb.PendingPartialWithdrawal{Index: 1},
			t:    nil,
			want: false,
		},
		{
			name: "equal pending partial withdrawals",
			s: &silapb.PendingPartialWithdrawal{
				Index:             50,
				Amount:            1000000000,
				WithdrawableEpoch: 200,
			},
			t: &silapb.PendingPartialWithdrawal{
				Index:             50,
				Amount:            1000000000,
				WithdrawableEpoch: 200,
			},
			want: true,
		},
		{
			name: "different index",
			s: &silapb.PendingPartialWithdrawal{
				Index:             50,
				Amount:            1000000000,
				WithdrawableEpoch: 200,
			},
			t: &silapb.PendingPartialWithdrawal{
				Index:             75,
				Amount:            1000000000,
				WithdrawableEpoch: 200,
			},
			want: false,
		},
		{
			name: "different amount",
			s: &silapb.PendingPartialWithdrawal{
				Index:             50,
				Amount:            1000000000,
				WithdrawableEpoch: 200,
			},
			t: &silapb.PendingPartialWithdrawal{
				Index:             50,
				Amount:            2000000000,
				WithdrawableEpoch: 200,
			},
			want: false,
		},
		{
			name: "different withdrawable epoch",
			s: &silapb.PendingPartialWithdrawal{
				Index:             50,
				Amount:            1000000000,
				WithdrawableEpoch: 200,
			},
			t: &silapb.PendingPartialWithdrawal{
				Index:             50,
				Amount:            1000000000,
				WithdrawableEpoch: 300,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PendingPartialWithdrawalsEqual(tt.s, tt.t); got != tt.want {
				t.Errorf("PendingPartialWithdrawalsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingConsolidationsEqual(t *testing.T) {
	tests := []struct {
		name string
		s    *silapb.PendingConsolidation
		t    *silapb.PendingConsolidation
		want bool
	}{
		{
			name: "both nil",
			s:    nil,
			t:    nil,
			want: true,
		},
		{
			name: "first nil",
			s:    nil,
			t:    &silapb.PendingConsolidation{SourceIndex: 1},
			want: false,
		},
		{
			name: "second nil",
			s:    &silapb.PendingConsolidation{SourceIndex: 1},
			t:    nil,
			want: false,
		},
		{
			name: "equal pending consolidations",
			s: &silapb.PendingConsolidation{
				SourceIndex: 10,
				TargetIndex: 20,
			},
			t: &silapb.PendingConsolidation{
				SourceIndex: 10,
				TargetIndex: 20,
			},
			want: true,
		},
		{
			name: "different source index",
			s: &silapb.PendingConsolidation{
				SourceIndex: 10,
				TargetIndex: 20,
			},
			t: &silapb.PendingConsolidation{
				SourceIndex: 15,
				TargetIndex: 20,
			},
			want: false,
		},
		{
			name: "different target index",
			s: &silapb.PendingConsolidation{
				SourceIndex: 10,
				TargetIndex: 20,
			},
			t: &silapb.PendingConsolidation{
				SourceIndex: 10,
				TargetIndex: 25,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PendingConsolidationsEqual(tt.s, tt.t); got != tt.want {
				t.Errorf("PendingConsolidationsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilderPendingWithdrawalsEqual(t *testing.T) {
	tests := []struct {
		name string
		s    *silapb.BuilderPendingWithdrawal
		t    *silapb.BuilderPendingWithdrawal
		want bool
	}{
		{
			name: "both nil",
			s:    nil,
			t:    nil,
			want: true,
		},
		{
			name: "first nil",
			s:    nil,
			t:    &silapb.BuilderPendingWithdrawal{Amount: 1},
			want: false,
		},
		{
			name: "second nil",
			s:    &silapb.BuilderPendingWithdrawal{Amount: 1},
			t:    nil,
			want: false,
		},
		{
			name: "equal",
			s: &silapb.BuilderPendingWithdrawal{
				FeeRecipient: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				Amount:       1000,
				BuilderIndex: 5,
			},
			t: &silapb.BuilderPendingWithdrawal{
				FeeRecipient: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				Amount:       1000,
				BuilderIndex: 5,
			},
			want: true,
		},
		{
			name: "different amount",
			s: &silapb.BuilderPendingWithdrawal{
				FeeRecipient: make([]byte, 20),
				Amount:       1000,
				BuilderIndex: 5,
			},
			t: &silapb.BuilderPendingWithdrawal{
				FeeRecipient: make([]byte, 20),
				Amount:       2000,
				BuilderIndex: 5,
			},
			want: false,
		},
		{
			name: "different builder index",
			s: &silapb.BuilderPendingWithdrawal{
				FeeRecipient: make([]byte, 20),
				Amount:       1000,
				BuilderIndex: 5,
			},
			t: &silapb.BuilderPendingWithdrawal{
				FeeRecipient: make([]byte, 20),
				Amount:       1000,
				BuilderIndex: 10,
			},
			want: false,
		},
		{
			name: "different fee recipient",
			s: &silapb.BuilderPendingWithdrawal{
				FeeRecipient: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				Amount:       1000,
				BuilderIndex: 5,
			},
			t: &silapb.BuilderPendingWithdrawal{
				FeeRecipient: []byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				Amount:       1000,
				BuilderIndex: 5,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuilderPendingWithdrawalsEqual(tt.s, tt.t); got != tt.want {
				t.Errorf("BuilderPendingWithdrawalsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
