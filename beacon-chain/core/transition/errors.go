package transition

import "errors"

var (
	ErrAttestationsSignatureInvalid          = errors.New("attestations signature invalid")
	ErrRandaoSignatureInvalid                = errors.New("randao signature invalid")
	ErrBLSToSilaChangesSignatureInvalid = errors.New("BLS to Sila changes signature invalid")
	ErrProcessWithdrawalsFailed              = errors.New("process withdrawals failed")
	ErrProcessRandaoFailed                   = errors.New("process randao failed")
	ErrProcessSilaExecutionDataFailed                 = errors.New("process silaexec data failed")
	ErrProcessProposerSlashingsFailed        = errors.New("process proposer slashings failed")
	ErrProcessAttesterSlashingsFailed        = errors.New("process attester slashings failed")
	ErrProcessAttestationsFailed             = errors.New("process attestations failed")
	ErrProcessDepositsFailed                 = errors.New("process deposits failed")
	ErrProcessVoluntaryExitsFailed           = errors.New("process voluntary exits failed")
	ErrProcessBLSChangesFailed               = errors.New("process BLS to Sila changes failed")
	ErrProcessPayloadAttestationsFailed      = errors.New("process payload attestations failed")
	ErrProcessSyncAggregateFailed            = errors.New("process sync aggregate failed")
)
