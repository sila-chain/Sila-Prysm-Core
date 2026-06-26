package deposit

import (
	"github.com/sila-chain/Sila/accounts/abi/bind"
)

// NewSilaDepositCallerFromBoundContract creates a new instance of SilaDepositCaller, bound to
// a specific deployed contract.
func NewSilaDepositCallerFromBoundContract(contract *bind.BoundContract) SilaDepositCaller {
	return SilaDepositCaller{contract: contract}
}

// NewSilaDepositTransactorFromBoundContract creates a new instance of
// SilaDepositTransactor, bound to a specific deployed contract.
func NewSilaDepositTransactorFromBoundContract(contract *bind.BoundContract) SilaDepositTransactor {
	return SilaDepositTransactor{contract: contract}
}

// NewSilaDepositFiltererFromBoundContract creates a new instance of
// SilaDepositFilterer, bound to a specific deployed contract.
func NewSilaDepositFiltererFromBoundContract(contract *bind.BoundContract) SilaDepositFilterer {
	return SilaDepositFilterer{contract: contract}
}
