package payloadattribute

import (
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
)

type Attributer interface {
	Version() int
	PrevRandao() []byte
	Timestamp() uint64
	SuggestedFeeRecipient() []byte
	Withdrawals() ([]*silaenginev1.Withdrawal, error)
	ParentBeaconBlockRoot() ([]byte, error)
	PbV1() (*silaenginev1.PayloadAttributes, error)
	PbV2() (*silaenginev1.PayloadAttributesV2, error)
	PbV3() (*silaenginev1.PayloadAttributesV3, error)
	PbV4() (*silaenginev1.PayloadAttributesV4, error)
	IsEmpty() bool
}
