// Copyright 2014 The go-ethereum Authors
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
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

type instrFn func(instr instruction, env Environment, context *Context, memory *Memory, stack *stack)
type instrExFn func(instr instruction, ret *big.Int, env Environment, context *Context, memory *Memory, stack *stack)

type instruction struct {
	op     OpCode
	pc     uint64
	fn     instrFn
	specFn instrExFn
	data   *big.Int

	gas   *big.Int
	spop  int
	spush int
}

func opStaticJump(instr instruction, ret *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	ret.Set(instr.data)
}

func opAdd(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	stack.push(U256(x.Add(x, y)))
}

func opSub(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	stack.push(U256(x.Sub(x, y)))
}

func opMul(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	stack.push(U256(x.Mul(x, y)))
}

func opDiv(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	if y.Cmp(common.Big0) != 0 {
		stack.push(U256(x.Div(x, y)))
	} else {
		stack.push(new(big.Int))
	}
}

func opSdiv(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := S256(stack.pop()), S256(stack.pop())
	if y.Cmp(common.Big0) == 0 {
		stack.push(new(big.Int))
		return
	} else {
		n := new(big.Int)
		if new(big.Int).Mul(x, y).Cmp(common.Big0) < 0 {
			n.SetInt64(-1)
		} else {
			n.SetInt64(1)
		}

		res := x.Div(x.Abs(x), y.Abs(y))
		res.Mul(res, n)

		stack.push(U256(res))
	}
}

func opMod(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	if y.Cmp(common.Big0) == 0 {
		stack.push(new(big.Int))
	} else {
		stack.push(U256(x.Mod(x, y)))
	}
}

func opSmod(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := S256(stack.pop()), S256(stack.pop())

	if y.Cmp(common.Big0) == 0 {
		stack.push(new(big.Int))
	} else {
		n := new(big.Int)
		if x.Cmp(common.Big0) < 0 {
			n.SetInt64(-1)
		} else {
			n.SetInt64(1)
		}

		res := x.Mod(x.Abs(x), y.Abs(y))
		res.Mul(res, n)

		stack.push(U256(res))
	}
}

func opExp(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	stack.push(U256(x.Exp(x, y, Pow256)))
}

func opSignExtend(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	back := stack.pop()
	if back.Cmp(big.NewInt(31)) < 0 {
		bit := uint(back.Uint64()*8 + 7)
		num := stack.pop()
		mask := back.Lsh(common.Big1, bit)
		mask.Sub(mask, common.Big1)
		if common.BitTest(num, int(bit)) {
			num.Or(num, mask.Not(mask))
		} else {
			num.And(num, mask)
		}

		stack.push(U256(num))
	}
}

func opNot(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x := stack.pop()
	stack.push(U256(x.Not(x)))
}

func opLt(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	if x.Cmp(y) < 0 {
		stack.push(big.NewInt(1))
	} else {
		stack.push(new(big.Int))
	}
}

func opGt(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	if x.Cmp(y) > 0 {
		stack.push(big.NewInt(1))
	} else {
		stack.push(new(big.Int))
	}
}

func opSlt(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := S256(stack.pop()), S256(stack.pop())
	if x.Cmp(S256(y)) < 0 {
		stack.push(big.NewInt(1))
	} else {
		stack.push(new(big.Int))
	}
}

func opSgt(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := S256(stack.pop()), S256(stack.pop())
	if x.Cmp(y) > 0 {
		stack.push(big.NewInt(1))
	} else {
		stack.push(new(big.Int))
	}
}

func opEq(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	if x.Cmp(y) == 0 {
		stack.push(big.NewInt(1))
	} else {
		stack.push(new(big.Int))
	}
}

func opIszero(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x := stack.pop()
	if x.Cmp(common.Big0) > 0 {
		stack.push(new(big.Int))
	} else {
		stack.push(big.NewInt(1))
	}
}

func opAnd(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	stack.push(x.And(x, y))
}
func opOr(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	stack.push(x.Or(x, y))
}
func opXor(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()
	stack.push(x.Xor(x, y))
}
func opByte(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	th, val := stack.pop(), stack.pop()
	if th.Cmp(big.NewInt(32)) < 0 {
		byte := big.NewInt(int64(common.LeftPadBytes(val.Bytes(), 32)[th.Int64()]))
		stack.push(byte)
	} else {
		stack.push(new(big.Int))
	}
}
func opAddmod(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	if z.Cmp(Zero) > 0 {
		add := x.Add(x, y)
		add.Mod(add, z)
		stack.push(U256(add))
	} else {
		stack.push(new(big.Int))
	}
}
func opMulmod(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	if z.Cmp(Zero) > 0 {
		mul := x.Mul(x, y)
		mul.Mod(mul, z)
		stack.push(U256(mul))
	} else {
		stack.push(new(big.Int))
	}
}

func opSha3(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	offset, size := stack.pop(), stack.pop()
	hash := crypto.Sha3(memory.Get(offset.Int64(), size.Int64()))

	stack.push(common.BytesToBig(hash))
}

func opAddress(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(common.Bytes2Big(context.Address().Bytes()))
}

func opBalance(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	addr := common.BigToAddress(stack.pop())
	balance := env.State().GetBalance(addr)

	stack.push(new(big.Int).Set(balance))
}

func opOrigin(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(env.Origin().Big())
}

func opCaller(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(common.Bytes2Big(context.caller.Address().Bytes()))
}

func opCallValue(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).Set(context.value))
}

func opCalldataLoad(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(common.Bytes2Big(getData(context.Input, stack.pop(), common.Big32)))
}

func opCalldataSize(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(big.NewInt(int64(len(context.Input))))
}

func opCalldataCopy(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	var (
		mOff = stack.pop()
		cOff = stack.pop()
		l    = stack.pop()
	)
	memory.Set(mOff.Uint64(), l.Uint64(), getData(context.Input, cOff, l))
}

func opExtCodeSize(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	addr := common.BigToAddress(stack.pop())
	l := big.NewInt(int64(len(env.State().GetCode(addr))))
	stack.push(l)
}

func opCodeSize(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	l := big.NewInt(int64(len(context.Code)))
	stack.push(l)
}

func opCodeCopy(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	var (
		mOff = stack.pop()
		cOff = stack.pop()
		l    = stack.pop()
	)
	codeCopy := getData(context.Code, cOff, l)

	memory.Set(mOff.Uint64(), l.Uint64(), codeCopy)
}

func opExtCodeCopy(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	var (
		addr = common.BigToAddress(stack.pop())
		mOff = stack.pop()
		cOff = stack.pop()
		l    = stack.pop()
	)
	codeCopy := getData(env.State().GetCode(addr), cOff, l)

	memory.Set(mOff.Uint64(), l.Uint64(), codeCopy)
}

func opGasprice(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).Set(context.Price))
}

func opBlockhash(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	num := stack.pop()

	n := new(big.Int).Sub(env.BlockNumber(), common.Big257)
	if num.Cmp(n) > 0 && num.Cmp(env.BlockNumber()) < 0 {
		stack.push(env.GetHash(num.Uint64()).Big())
	} else {
		stack.push(new(big.Int))
	}
}

func opCoinbase(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(env.Coinbase().Big())
}

func opTimestamp(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(U256(new(big.Int).Set(env.Time())))
}

func opNumber(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(U256(new(big.Int).Set(env.BlockNumber())))
}

func opDifficulty(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(U256(new(big.Int).Set(env.Difficulty())))
}

func opGasLimit(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(U256(new(big.Int).Set(env.GasLimit())))
}

func opPop(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.pop()
}

func opPush(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).Set(instr.data))
}

func opDup(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.dup(int(instr.data.Int64()))
}

func opSwap(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.swap(int(instr.data.Int64()))
}

func opLog(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	n := int(instr.data.Int64())
	topics := make([]common.Hash, n)
	mStart, mSize := stack.pop(), stack.pop()
	for i := 0; i < n; i++ {
		topics[i] = common.BigToHash(stack.pop())
	}

	d := memory.Get(mStart.Int64(), mSize.Int64())
	log := state.NewLog(context.Address(), topics, d, env.BlockNumber().Uint64())
	env.AddLog(log)
}

func opMload(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	offset := stack.pop()
	val := common.BigD(memory.Get(offset.Int64(), 32))
	stack.push(val)
}

func opMstore(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	// pop value of the stack
	mStart, val := stack.pop(), stack.pop()
	memory.Set(mStart.Uint64(), 32, common.BigToBytes(val, 256))
}

func opMstore8(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	off, val := stack.pop().Int64(), stack.pop().Int64()
	memory.store[off] = byte(val & 0xff)
}

func opSload(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	loc := common.BigToHash(stack.pop())
	val := env.State().GetState(context.Address(), loc).Big()
	stack.push(val)
}

func opSstore(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	loc := common.BigToHash(stack.pop())
	val := stack.pop()

	env.State().SetState(context.Address(), loc, common.BigToHash(val))
}

func opJump(instr instruction, env Environment, context *Context, memory *Memory, stack *stack)     {}
func opJumpi(instr instruction, env Environment, context *Context, memory *Memory, stack *stack)    {}
func opJumpdest(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {}

func opPc(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).Set(instr.data))
}

func opMsize(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(big.NewInt(int64(memory.Len())))
}

func opGas(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).Set(context.Gas))
}

func opCreate(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	var (
		value        = stack.pop()
		offset, size = stack.pop(), stack.pop()
		input        = memory.Get(offset.Int64(), size.Int64())
		gas          = new(big.Int).Set(context.Gas)
		addr         common.Address
	)

	context.UseGas(context.Gas)
	ret, suberr, ref := env.Create(context, input, gas, context.Price, value)
	if suberr != nil {
		stack.push(new(big.Int))

	} else {
		// gas < len(ret) * Createinstr.dataGas == NO_CODE
		dataGas := big.NewInt(int64(len(ret)))
		dataGas.Mul(dataGas, params.CreateDataGas)
		if context.UseGas(dataGas) {
			ref.SetCode(ret)
		}
		addr = ref.Address()

		stack.push(addr.Big())

	}
}

func opCall(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	gas := stack.pop()
	// pop gas and value of the stack.
	addr, value := stack.pop(), stack.pop()
	value = U256(value)
	// pop input size and offset
	inOffset, inSize := stack.pop(), stack.pop()
	// pop return size and offset
	retOffset, retSize := stack.pop(), stack.pop()

	address := common.BigToAddress(addr)

	// Get the arguments from the memory
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if len(value.Bytes()) > 0 {
		gas.Add(gas, params.CallStipend)
	}

	ret, err := env.Call(context, address, args, gas, context.Price, value)

	if err != nil {
		stack.push(new(big.Int))

	} else {
		stack.push(big.NewInt(1))

		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
}

func opCallCode(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	gas := stack.pop()
	// pop gas and value of the stack.
	addr, value := stack.pop(), stack.pop()
	value = U256(value)
	// pop input size and offset
	inOffset, inSize := stack.pop(), stack.pop()
	// pop return size and offset
	retOffset, retSize := stack.pop(), stack.pop()

	address := common.BigToAddress(addr)

	// Get the arguments from the memory
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if len(value.Bytes()) > 0 {
		gas.Add(gas, params.CallStipend)
	}

	ret, err := env.CallCode(context, address, args, gas, context.Price, value)

	if err != nil {
		stack.push(new(big.Int))

	} else {
		stack.push(big.NewInt(1))

		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
}

func opReturn(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {}
func opStop(instr instruction, env Environment, context *Context, memory *Memory, stack *stack)   {}

func opSuicide(instr instruction, env Environment, context *Context, memory *Memory, stack *stack) {
	receiver := env.State().GetOrNewStateObject(common.BigToAddress(stack.pop()))
	balance := env.State().GetBalance(context.Address())

	receiver.AddBalance(balance)

	env.State().Delete(context.Address())
}
