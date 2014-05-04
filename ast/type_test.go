package ast

import (
	"testing"

	"github.com/velour/stop/token"
)

func TestType(t *testing.T) {
	tests := parserTests{
		{`a`, a},
		{`(a)`, a},
		{`((a))`, a},
		{`*(a)`, pointer(a)},
		{`[](a)`, sliceType(a)},
		{`[]*(a)`, sliceType(pointer(a))},
		{`*[](a)`, pointer(sliceType(a))},
		{`map[a]b`, mapType(a, b)},

		{`[]func()`, sliceType(funcType(parmList(), parmList()))},
		{`[]func()<-chan int`,
			sliceType(funcType(
				parmList(),
				parmList(parmDecl(recvChan(ident("int")), false))))},
		{`[]interface{ c()<-chan[5]sort.Interface }`, sliceType(ifaceType(method(c, parmList(), parmList(parmDecl(recvChan(arrayType(intLit("5"), sel(ident("sort"), ident("Interface")))), false)))))},

		{`1`, parseErr("expected")},
	}
	tests.runType(t)
}

func TestStructType(t *testing.T) {
	bi := sel(ident("big"), ident("Int"))
	tests := parserTests{
		{`struct{}`, structType()},
		{`struct{a}`, structType(fieldDecl(a))},
		{`struct{a; b; c}`, structType(fieldDecl(a), fieldDecl(b), fieldDecl(c))},
		{`struct{a, b c}`, structType(fieldDecl(c, a, b))},
		{`struct{a b; c}`, structType(fieldDecl(b, a), fieldDecl(c))},
		{`struct{a; b c}`, structType(fieldDecl(a), fieldDecl(c, b))},
		{`struct{big.Int}`, structType(fieldDecl(bi))},
		{`struct{big.Int; a}`, structType(fieldDecl(bi), fieldDecl(a))},
		{`struct{big.Int; a b}`, structType(fieldDecl(bi), fieldDecl(b, a))},
		{`struct{big.Int; a big.Int}`, structType(fieldDecl(bi), fieldDecl(bi, a))},
		{`struct{*big.Int}`, structType(fieldDecl(pointer(bi)))},
		{`struct{*big.Int; a}`, structType(fieldDecl(pointer(bi)), fieldDecl(a))},
		{`struct{a; *big.Int}`, structType(fieldDecl(a), fieldDecl(pointer(bi)))},

		// Tagged.
		{`struct{a "your it"}`, structType(tagFieldDecl(a, strLit("your it")))},
		{`struct{*a "your it"}`, structType(tagFieldDecl(pointer(a), strLit("your it")))},
		{`struct{big.Int "your it"}`, structType(tagFieldDecl(bi, strLit("your it")))},
		{`struct{*big.Int "your it"}`, structType(tagFieldDecl(pointer(bi), strLit("your it")))},
		{`struct{a "your it"; b}`, structType(tagFieldDecl(a, strLit("your it")), fieldDecl(b))},
		{`struct{a b "your it"}`, structType(tagFieldDecl(b, strLit("your it"), a))},

		// Trailing ;
		{`struct{a;}`, structType(fieldDecl(a))},
		{`struct{a; b; c;}`, structType(fieldDecl(a), fieldDecl(b), fieldDecl(c))},

		// Embedded pointers must be type names.
		{`struct{**big.Int}`, parseErr("expected.*got \\*")},
		{`struct{*[]big.Int}`, parseErr("expected.*got \\[")},
	}
	tests.runType(t)
}

func TestInterfaceType(t *testing.T) {
	tests := parserTests{
		{`interface{}`, ifaceType()},
		{`interface{a; b; c}`, ifaceType(a, b, c)},
		{`interface{a; b; big.Int}`, ifaceType(a, b, sel(ident("big"), ident("Int")))},
		{`interface{a; big.Int; b}`, ifaceType(a, sel(ident("big"), ident("Int")), b)},
		{`interface{big.Int; a; b}`, ifaceType(sel(ident("big"), ident("Int")), a, b)},

		{`interface{a; b; c()}`, ifaceType(a, b, method(c, parmList(), parmList()))},
		{`interface{a; b(); c}`, ifaceType(a, method(b, parmList(), parmList()), c)},
		{`interface{a(); b; c}`, ifaceType(method(a, parmList(), parmList()), b, c)},
		{`interface{a; big.Int; c()}`, ifaceType(a, sel(ident("big"), ident("Int")), method(c, parmList(), parmList()))},

		// Trailing ;
		{`interface{a; b; c;}`, ifaceType(a, b, c)},
	}
	tests.runType(t)
}

func TestFunctionType(t *testing.T) {
	tests := parserTests{
		{`func()`, funcType(parmList(), parmList())},
		{`func(a) b`, funcType(
			parmList(parmDecl(a, false)),
			parmList(parmDecl(b, false)))},
		{`func(...a) b`, funcType(
			parmList(parmDecl(a, true)),
			parmList(parmDecl(b, false)))},
		{`func(a, b) c`, funcType(
			parmList(parmDecl(a, false), parmDecl(b, false)),
			parmList(parmDecl(c, false)))},
		{`func(a) (b, c)`, funcType(
			parmList(parmDecl(a, false)),
			parmList(parmDecl(b, false), parmDecl(c, false)))},
		{`func(...a) (b, c)`, funcType(
			parmList(parmDecl(a, true)),
			parmList(parmDecl(b, false), parmDecl(c, false)))},

		// Invalid, but will have to be caught during type checking.
		{`func(a) (b, ...c)`, funcType(
			parmList(parmDecl(a, false)),
			parmList(parmDecl(b, false), parmDecl(c, true)))},
	}
	tests.runType(t)
}

func TestParameterList(t *testing.T) {
	tests := parserTests{
		{`()`, parmList()},

		// Parameter declarations without any identifiers.
		{`(a, b, c)`, parmList(
			parmDecl(a, false),
			parmDecl(b, false),
			parmDecl(c, false))},
		{`(a, b, ...c)`, parmList(
			parmDecl(a, false),
			parmDecl(b, false),
			parmDecl(c, true))},
		{`(a, b, big.Int)`, parmList(
			parmDecl(a, false),
			parmDecl(b, false),
			parmDecl(sel(ident("big"), ident("Int")), false))},
		{`(a, b, ...big.Int)`, parmList(
			parmDecl(a, false),
			parmDecl(b, false),
			parmDecl(sel(ident("big"), ident("Int")), true))},
		{`(a, b, []c)`, parmList(
			parmDecl(a, false),
			parmDecl(b, false),
			parmDecl(sliceType(c), false))},
		{`(a, b, ...[]c)`, parmList(
			parmDecl(a, false),
			parmDecl(b, false),
			parmDecl(sliceType(c), true))},
		{`([]a, b, c)`, parmList(
			parmDecl(sliceType(a), false),
			parmDecl(b, false),
			parmDecl(c, false))},
		{`([]a, b, ...c)`, parmList(
			parmDecl(sliceType(a), false),
			parmDecl(b, false),
			parmDecl(c, true))},
		{`(...a)`, parmList(parmDecl(a, true))},

		// Parameter declarations with identifiers
		{`(a, b c)`, parmList(parmDecl(c, false, a, b))},
		{`(a, b ...c)`, parmList(parmDecl(c, true, a, b))},
		{`(a, b big.Int)`, parmList(parmDecl(sel(ident("big"), ident("Int")), false, a, b))},
		{`(a, b []int)`, parmList(parmDecl(sliceType(ident("int")), false, a, b))},
		{`(a, b ...[]int)`, parmList(parmDecl(sliceType(ident("int")), true, a, b))},
		{`(a, b []int, c, d ...[]int)`, parmList(
			parmDecl(sliceType(ident("int")), false, a, b),
			parmDecl(sliceType(ident("int")), true, c, d))},

		// Trailing comma is OK.
		{`(a, b, c,)`, parmList(
			parmDecl(a, false),
			parmDecl(b, false),
			parmDecl(c, false))},
		{`(a, []b, c,)`, parmList(
			parmDecl(a, false),
			parmDecl(sliceType(b), false),
			parmDecl(c, false))},
		{`(a, b []int,)`, parmList(parmDecl(sliceType(ident("int")), false, a, b))},
		{`(a, b []int, c, d ...[]int,)`, parmList(
			parmDecl(sliceType(ident("int")), false, a, b),
			parmDecl(sliceType(ident("int")), true, c, d))},

		// Strange, but OK.
		{`(int float64, float64 int)`, parmList(
			parmDecl(ident("float64"), false, ident("int")),
			parmDecl(ident("int"), false, ident("float64")))},

		// ... types must be the last in the list.
		{`(...a, b)`, parseErr("")},
		{`([]a, ...b, c)`, parseErr("")},
		{`(a, ...b, c)`, parseErr("")},
		{`(a ...b, c int)`, parseErr("")},

		// Can't mix declarations with identifiers with those without.
		{`([]a, b c)`, parseErr("")},
		{`(a b, c, d)`, parseErr("")},
	}
	tests.runParamList(t)
}

func TestChannelType(t *testing.T) {
	tests := parserTests{
		{`chan a`, chanType(a)},
		{`<-chan a`, recvChan(a)},
		{`chan<- a`, sendChan(a)},
		{`chan<- <- chan a`, sendChan(recvChan(a))},
		{`chan<- chan a`, sendChan(chanType(a))},
		{`<- chan <-chan a`, recvChan(recvChan(a))},

		{`<- chan<- a`, parseErr("expected")},
		{`chan<- <- a`, parseErr("expected")},
		{`chan<- <- <- chan a`, parseErr("expected")},
	}
	tests.runType(t)
}

func TestMapType(t *testing.T) {
	tests := parserTests{
		{`map[int]a`, mapType(ident("int"), a)},
		{`map[*int]a.b`, mapType(pointer(ident("int")), sel(a, b))},
		{`map[*int]map[string]int]`, mapType(pointer(ident("int")), mapType(ident("string"), ident("int")))},
	}
	tests.runType(t)
}

func TestArrayType(t *testing.T) {
	tests := parserTests{
		{`[4]a`, arrayType(intLit("4"), a)},
		{`[4]a.b`, arrayType(intLit("4"), sel(a, b))},
		{`[4](a)`, arrayType(intLit("4"), a)},
		{`[4][]a`, arrayType(intLit("4"), sliceType(a))},
		{`[4][42*z]a`, arrayType(intLit("4"), arrayType(binOp(token.Star, intLit("42"), ident("z")), a))},

		// [...]Type notation is only allowed in composite literals,
		// not types in general
		{`[...]int`, parseErr("expected.*got \\.\\.\\.")},
	}
	tests.runType(t)
}

func TestSliceType(t *testing.T) {
	tests := parserTests{
		{`[]a`, sliceType(a)},
		{`[]a.b`, sliceType(sel(a, b))},
		{`[](a)`, sliceType(a)},
		{`[][]a)`, sliceType(sliceType(a))},
	}
	tests.runType(t)
}

func TestPointerType(t *testing.T) {
	tests := parserTests{
		{`*a`, pointer(a)},
		{`*a.b`, pointer(sel(a, b))},
		{`*(a)`, pointer(a)},
		{`**(a)`, pointer(pointer(a))},

		{`α.`, parseErr("expected.*Identifier.*got EOF")},
	}
	tests.runType(t)
}

func TestTypeName(t *testing.T) {
	tests := parserTests{
		{`a`, a},
		{`a.b`, sel(a, b)},
		{`α.b`, sel(ident("α"), b)},
	}
	tests.runType(t)
}
