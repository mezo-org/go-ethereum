// Copyright 2025 The go-ethereum Authors
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
	"math"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

// TestEIP90000DisablesSelfdestruct checks that activating EIP 90000 turns the
// SELFDESTRUCT opcode (0xff) into an invalid opcode, while leaving it working
// when the EIP is not active.
func TestEIP90000DisablesSelfdestruct(t *testing.T) {
	// PUSH1 0x00; SELFDESTRUCT - self-destruct, sending to the zero address.
	code := hexutil.MustDecode("0x6000ff")

	run := func(t *testing.T, extraEips []int) error {
		address := common.BytesToAddress([]byte("contract"))

		statedb, err := state.New(types.EmptyRootHash, state.NewDatabaseForTesting())
		require.NoError(t, err)
		statedb.CreateAccount(address)
		statedb.SetCode(address, code, tracing.CodeChangeUnspecified)
		statedb.Finalise(true)

		vmctx := BlockContext{
			CanTransfer: func(StateDB, common.Address, *uint256.Int) bool { return true },
			Transfer:    func(StateDB, common.Address, common.Address, *uint256.Int) {},
		}
		evm := NewEVM(vmctx, statedb, params.AllEthashProtocolChanges, Config{ExtraEips: extraEips})

		_, _, err = evm.Call(common.Address{}, address, nil, math.MaxUint64, new(uint256.Int))
		return err
	}

	t.Run("without the EIP: SELFDESTRUCT is valid", func(t *testing.T) {
		require.NoError(t, run(t, nil))
	})

	t.Run("with EIP 90000: SELFDESTRUCT is invalid", func(t *testing.T) {
		err := run(t, []int{90000})

		var invalidOp *ErrInvalidOpCode
		require.ErrorAs(t, err, &invalidOp)
		require.Equal(t, SELFDESTRUCT, invalidOp.opcode)
	})
}
