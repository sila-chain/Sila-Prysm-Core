// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.
package deposit

import (
	"math/big"
	"strings"

	sila "github.com/sila-chain/Sila"
	"github.com/sila-chain/Sila/accounts/abi"
	"github.com/sila-chain/Sila/accounts/abi/bind"
	"github.com/sila-chain/Sila/common"
	"github.com/sila-chain/Sila/core/types"
	"github.com/sila-chain/Sila/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = sila.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// SilaDepositABI is the input ABI used to generate the binding from.
const SilaDepositABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"withdrawal_credentials\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"amount\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"index\",\"type\":\"bytes\"}],\"name\":\"DepositEvent\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"withdrawal_credentials\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"deposit_data_root\",\"type\":\"bytes32\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"get_deposit_count\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"get_deposit_root\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]"

// SilaDeposit is an auto generated Go binding around an Sila contract.
type SilaDeposit struct {
	SilaDepositCaller     // Read-only binding to the contract
	SilaDepositTransactor // Write-only binding to the contract
	SilaDepositFilterer   // Log filterer for contract events
}

// SilaDepositCaller is an auto generated read-only Go binding around an Sila contract.
type SilaDepositCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SilaDepositTransactor is an auto generated write-only Go binding around an Sila contract.
type SilaDepositTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SilaDepositFilterer is an auto generated log filtering Go binding around an Sila contract events.
type SilaDepositFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SilaDepositSession is an auto generated Go binding around an Sila contract,
// with pre-set call and transact options.
type SilaDepositSession struct {
	Contract     *SilaDeposit  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SilaDepositCallerSession is an auto generated read-only Go binding around an Sila contract,
// with pre-set call options.
type SilaDepositCallerSession struct {
	Contract *SilaDepositCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// SilaDepositTransactorSession is an auto generated write-only Go binding around an Sila contract,
// with pre-set transact options.
type SilaDepositTransactorSession struct {
	Contract     *SilaDepositTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// SilaDepositRaw is an auto generated low-level Go binding around an Sila contract.
type SilaDepositRaw struct {
	Contract *SilaDeposit // Generic contract binding to access the raw methods on
}

// SilaDepositCallerRaw is an auto generated low-level read-only Go binding around an Sila contract.
type SilaDepositCallerRaw struct {
	Contract *SilaDepositCaller // Generic read-only contract binding to access the raw methods on
}

// SilaDepositTransactorRaw is an auto generated low-level write-only Go binding around an Sila contract.
type SilaDepositTransactorRaw struct {
	Contract *SilaDepositTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSilaDeposit creates a new instance of SilaDeposit, bound to a specific deployed contract.
func NewSilaDeposit(address common.Address, backend bind.ContractBackend) (*SilaDeposit, error) {
	contract, err := bindSilaDeposit(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SilaDeposit{SilaDepositCaller: SilaDepositCaller{contract: contract}, SilaDepositTransactor: SilaDepositTransactor{contract: contract}, SilaDepositFilterer: SilaDepositFilterer{contract: contract}}, nil
}

// NewSilaDepositCaller creates a new read-only instance of SilaDeposit, bound to a specific deployed contract.
func NewSilaDepositCaller(address common.Address, caller bind.ContractCaller) (*SilaDepositCaller, error) {
	contract, err := bindSilaDeposit(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SilaDepositCaller{contract: contract}, nil
}

// NewSilaDepositTransactor creates a new write-only instance of SilaDeposit, bound to a specific deployed contract.
func NewSilaDepositTransactor(address common.Address, transactor bind.ContractTransactor) (*SilaDepositTransactor, error) {
	contract, err := bindSilaDeposit(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SilaDepositTransactor{contract: contract}, nil
}

// NewSilaDepositFilterer creates a new log filterer instance of SilaDeposit, bound to a specific deployed contract.
func NewSilaDepositFilterer(address common.Address, filterer bind.ContractFilterer) (*SilaDepositFilterer, error) {
	contract, err := bindSilaDeposit(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SilaDepositFilterer{contract: contract}, nil
}

// bindSilaDeposit binds a generic wrapper to an already deployed contract.
func bindSilaDeposit(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SilaDepositABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SilaDeposit *SilaDepositRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SilaDeposit.Contract.SilaDepositCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SilaDeposit *SilaDepositRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SilaDeposit.Contract.SilaDepositTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SilaDeposit *SilaDepositRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SilaDeposit.Contract.SilaDepositTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SilaDeposit *SilaDepositCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SilaDeposit.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SilaDeposit *SilaDepositTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SilaDeposit.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SilaDeposit *SilaDepositTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SilaDeposit.Contract.contract.Transact(opts, method, params...)
}

// GetDepositCount is a free data retrieval call binding the contract method 0x621fd130.
//
// Solidity: function get_deposit_count() view returns(bytes)
func (_SilaDeposit *SilaDepositCaller) GetDepositCount(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _SilaDeposit.contract.Call(opts, &out, "get_deposit_count")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetDepositCount is a free data retrieval call binding the contract method 0x621fd130.
//
// Solidity: function get_deposit_count() view returns(bytes)
func (_SilaDeposit *SilaDepositSession) GetDepositCount() ([]byte, error) {
	return _SilaDeposit.Contract.GetDepositCount(&_SilaDeposit.CallOpts)
}

// GetDepositCount is a free data retrieval call binding the contract method 0x621fd130.
//
// Solidity: function get_deposit_count() view returns(bytes)
func (_SilaDeposit *SilaDepositCallerSession) GetDepositCount() ([]byte, error) {
	return _SilaDeposit.Contract.GetDepositCount(&_SilaDeposit.CallOpts)
}

// GetDepositRoot is a free data retrieval call binding the contract method 0xc5f2892f.
//
// Solidity: function get_deposit_root() view returns(bytes32)
func (_SilaDeposit *SilaDepositCaller) GetDepositRoot(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SilaDeposit.contract.Call(opts, &out, "get_deposit_root")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetDepositRoot is a free data retrieval call binding the contract method 0xc5f2892f.
//
// Solidity: function get_deposit_root() view returns(bytes32)
func (_SilaDeposit *SilaDepositSession) GetDepositRoot() ([32]byte, error) {
	return _SilaDeposit.Contract.GetDepositRoot(&_SilaDeposit.CallOpts)
}

// GetDepositRoot is a free data retrieval call binding the contract method 0xc5f2892f.
//
// Solidity: function get_deposit_root() view returns(bytes32)
func (_SilaDeposit *SilaDepositCallerSession) GetDepositRoot() ([32]byte, error) {
	return _SilaDeposit.Contract.GetDepositRoot(&_SilaDeposit.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_SilaDeposit *SilaDepositCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _SilaDeposit.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_SilaDeposit *SilaDepositSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _SilaDeposit.Contract.SupportsInterface(&_SilaDeposit.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_SilaDeposit *SilaDepositCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _SilaDeposit.Contract.SupportsInterface(&_SilaDeposit.CallOpts, interfaceId)
}

// Deposit is a paid mutator transaction binding the contract method 0x22895118.
//
// Solidity: function deposit(bytes pubkey, bytes withdrawal_credentials, bytes signature, bytes32 deposit_data_root) payable returns()
func (_SilaDeposit *SilaDepositTransactor) Deposit(opts *bind.TransactOpts, pubkey []byte, withdrawal_credentials []byte, signature []byte, deposit_data_root [32]byte) (*types.Transaction, error) {
	return _SilaDeposit.contract.Transact(opts, "deposit", pubkey, withdrawal_credentials, signature, deposit_data_root)
}

// Deposit is a paid mutator transaction binding the contract method 0x22895118.
//
// Solidity: function deposit(bytes pubkey, bytes withdrawal_credentials, bytes signature, bytes32 deposit_data_root) payable returns()
func (_SilaDeposit *SilaDepositSession) Deposit(pubkey []byte, withdrawal_credentials []byte, signature []byte, deposit_data_root [32]byte) (*types.Transaction, error) {
	return _SilaDeposit.Contract.Deposit(&_SilaDeposit.TransactOpts, pubkey, withdrawal_credentials, signature, deposit_data_root)
}

// Deposit is a paid mutator transaction binding the contract method 0x22895118.
//
// Solidity: function deposit(bytes pubkey, bytes withdrawal_credentials, bytes signature, bytes32 deposit_data_root) payable returns()
func (_SilaDeposit *SilaDepositTransactorSession) Deposit(pubkey []byte, withdrawal_credentials []byte, signature []byte, deposit_data_root [32]byte) (*types.Transaction, error) {
	return _SilaDeposit.Contract.Deposit(&_SilaDeposit.TransactOpts, pubkey, withdrawal_credentials, signature, deposit_data_root)
}

// SilaDepositDepositEventIterator is returned from FilterDepositEvent and is used to iterate over the raw logs and unpacked data for DepositEvent events raised by the SilaSila deposit.
type SilaDepositDepositEventIterator struct {
	Event *SilaDepositDepositEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  sila.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SilaDepositDepositEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SilaDepositDepositEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SilaDepositDepositEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SilaDepositDepositEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SilaDepositDepositEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SilaDepositDepositEvent represents a DepositEvent event raised by the SilaSila deposit.
type SilaDepositDepositEvent struct {
	Pubkey                []byte
	WithdrawalCredentials []byte
	Amount                []byte
	Signature             []byte
	Index                 []byte
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterDepositEvent is a free log retrieval operation binding the contract event 0x649bbc62d0e31342afea4e5cd82d4049e7e1ee912fc0889aa790803be39038c5.
//
// Solidity: event DepositEvent(bytes pubkey, bytes withdrawal_credentials, bytes amount, bytes signature, bytes index)
func (_SilaDeposit *SilaDepositFilterer) FilterDepositEvent(opts *bind.FilterOpts) (*SilaDepositDepositEventIterator, error) {

	logs, sub, err := _SilaDeposit.contract.FilterLogs(opts, "DepositEvent")
	if err != nil {
		return nil, err
	}
	return &SilaDepositDepositEventIterator{contract: _SilaDeposit.contract, event: "DepositEvent", logs: logs, sub: sub}, nil
}

// WatchDepositEvent is a free log subscription operation binding the contract event 0x649bbc62d0e31342afea4e5cd82d4049e7e1ee912fc0889aa790803be39038c5.
//
// Solidity: event DepositEvent(bytes pubkey, bytes withdrawal_credentials, bytes amount, bytes signature, bytes index)
func (_SilaDeposit *SilaDepositFilterer) WatchDepositEvent(opts *bind.WatchOpts, sink chan<- *SilaDepositDepositEvent) (event.Subscription, error) {

	logs, sub, err := _SilaDeposit.contract.WatchLogs(opts, "DepositEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SilaDepositDepositEvent)
				if err := _SilaDeposit.contract.UnpackLog(event, "DepositEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDepositEvent is a log parse operation binding the contract event 0x649bbc62d0e31342afea4e5cd82d4049e7e1ee912fc0889aa790803be39038c5.
//
// Solidity: event DepositEvent(bytes pubkey, bytes withdrawal_credentials, bytes amount, bytes signature, bytes index)
func (_SilaDeposit *SilaDepositFilterer) ParseDepositEvent(log types.Log) (*SilaDepositDepositEvent, error) {
	event := new(SilaDepositDepositEvent)
	if err := _SilaDeposit.contract.UnpackLog(event, "DepositEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
