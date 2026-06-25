package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila/common/hexutil"
	"github.com/pkg/errors"
)

func (c *beaconApiValidatorClient) proposeExit(ctx context.Context, signedVoluntaryExit *silapb.SignedVoluntaryExit) (*silapb.ProposeExitResponse, error) {
	if signedVoluntaryExit == nil {
		return nil, errors.New("signed voluntary exit is nil")
	}

	if signedVoluntaryExit.Exit == nil {
		return nil, errors.New("exit is nil")
	}

	jsonSignedVoluntaryExit := structs.SignedVoluntaryExit{
		Message: &structs.VoluntaryExit{
			Epoch:          strconv.FormatUint(uint64(signedVoluntaryExit.Exit.Epoch), 10),
			ValidatorIndex: strconv.FormatUint(uint64(signedVoluntaryExit.Exit.ValidatorIndex), 10),
		},
		Signature: hexutil.Encode(signedVoluntaryExit.Signature),
	}

	marshalledSignedVoluntaryExit, err := json.Marshal(jsonSignedVoluntaryExit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal signed voluntary exit")
	}

	if err = c.handler.Post(
		ctx,
		"/sila/v1/beacon/pool/voluntary_exits",
		nil,
		bytes.NewBuffer(marshalledSignedVoluntaryExit),
		nil,
	); err != nil {
		return nil, err
	}

	exitRoot, err := signedVoluntaryExit.Exit.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to compute exit root")
	}

	return &silapb.ProposeExitResponse{
		ExitRoot: exitRoot[:],
	}, nil
}
