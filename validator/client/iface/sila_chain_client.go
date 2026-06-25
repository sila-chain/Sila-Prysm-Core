package iface

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/validator"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

var ErrNotSupported = errors.New("endpoint not supported")

type ValidatorCount struct {
	Status string
	Count  uint64
}

// SilaChainClient defines an interface required to implement all the sila specific custom endpoints.
type SilaChainClient interface {
	ValidatorCount(context.Context, string, []validator.Status) ([]ValidatorCount, error)
	ValidatorPerformance(context.Context, *silapb.ValidatorPerformanceRequest) (*silapb.ValidatorPerformanceResponse, error)
}
