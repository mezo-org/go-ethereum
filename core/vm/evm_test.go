// Copyright 2026 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

// TestEVMCallCanReceiveTransferGuard verifies that a non-nil
// CanReceiveTransfer returning false aborts a value-bearing CALL with
// ErrTransferNotAllowed before any state mutation: gas is left intact,
// the recipient is not created, and Transfer is not invoked.
func TestEVMCallCanReceiveTransferGuard(t *testing.T) {
	var (
		caller    = common.HexToAddress("0x01")
		recipient = common.HexToAddress("0x02")
		gas       = uint64(100_000)
		value     = uint256.NewInt(1)
	)

	statedb, err := state.New(types.EmptyRootHash, state.NewDatabaseForTesting())
	require.NoError(t, err)

	var transferCalled bool
	evm := NewEVM(BlockContext{
		CanTransfer: func(StateDB, common.Address, *uint256.Int) bool {
			return true
		},
		CanReceiveTransfer: func(StateDB, common.Address, *uint256.Int) bool {
			return false
		},
		Transfer: func(StateDB, common.Address, common.Address, *uint256.Int) {
			transferCalled = true
		},
	}, statedb, params.AllEthashProtocolChanges, Config{})

	_, leftOverGas, err := evm.Call(caller, recipient, nil, gas, value)

	require.True(t, errors.Is(err, ErrTransferNotAllowed))
	require.Equal(t, gas, leftOverGas)
	require.False(t, statedb.Exist(recipient))
	require.False(t, transferCalled)
}

// TestEVMCallCanReceiveTransferGuardIgnoresZeroValue verifies that the
// recipient-side guard is short-circuited by the value.IsZero() check
// and is not invoked for zero-value CALLs.
func TestEVMCallCanReceiveTransferGuardIgnoresZeroValue(t *testing.T) {
	var (
		caller    = common.HexToAddress("0x01")
		recipient = common.HexToAddress("0x02")
		gas       = uint64(100_000)
		value     = uint256.NewInt(0)
	)

	statedb, err := state.New(types.EmptyRootHash, state.NewDatabaseForTesting())
	require.NoError(t, err)

	var guardCalled bool
	evm := NewEVM(BlockContext{
		CanTransfer: func(StateDB, common.Address, *uint256.Int) bool {
			return true
		},
		CanReceiveTransfer: func(StateDB, common.Address, *uint256.Int) bool {
			guardCalled = true
			return false
		},
		Transfer: func(StateDB, common.Address, common.Address, *uint256.Int) {},
	}, statedb, params.AllEthashProtocolChanges, Config{})

	_, _, err = evm.Call(caller, recipient, nil, gas, value)

	require.NoError(t, err)
	require.False(t, guardCalled)
}
