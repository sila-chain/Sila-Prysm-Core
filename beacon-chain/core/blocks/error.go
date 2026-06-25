package blocks

import "github.com/pkg/errors"

var errNilSignedWithdrawalMessage = errors.New("nil SignedBLSToSilaChange message")
var errNilWithdrawalMessage = errors.New("nil BLSToSilaChange message")
var errInvalidBLSPrefix = errors.New("withdrawal credential prefix is not a BLS prefix")
var errInvalidWithdrawalCredentials = errors.New("withdrawal credentials do not match")
var ErrInvalidSignature = errors.New("invalid signature")
var ErrInvalidProposerIndex = errors.New("invalid proposer index")
