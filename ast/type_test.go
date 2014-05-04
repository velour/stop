package ast

import (
	"testing"

	"github.com/velour/stop/token"
)

func TestType(t *testing.T) {
	tests := parserTests{
		{`a`, ident("a")},
		{`(a)`, ident("a")},
		{`((a))`, ident("a")},
		{`*(a)`, pointer(ident("a"))},
		{`[](a)`, sliceType(ident("a"))},
		{`[]*(a)`, sliceType(pointer(ident("a")))},
		{`*[](a)`, pointer(sliceType(ident("a")))},
		{`map[a]b`, mapType(ident("a"), ident("b"))},
		//{`[](map[([4]a)*b)`, sliceType(mapType(arrayType(intLit("4"), ident("a")), pointer(ident("b"))))},

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
	ta, tb, tc := ident("a"), ident("b"), ident("c")
	bi := sel(ident("big"), ident("Int"))
	tests := parserTests{
		{`struct{}`, structType()},
		{`struct{a}`, structType(fieldDecl(ta))},
		{`struct{a; b; c}`, structType(fieldDecl(ta), fieldDecl(tb), fieldDecl(tc))},
		{`struct{a, b c}`, structType(fieldDecl(tc, a, b))},
		{`struct{a b; c}`, structType(fieldDecl(tb, a), fieldDecl(tc))},
		{`struct{a; b c}`, structType(fieldDecl(ta), fieldDecl(tc, b))},
		{`struct{big.Int}`, structType(fieldDecl(bi))},
		{`struct{big.Int; a}`, structType(fieldDecl(bi), fieldDecl(ta))},
		{`struct{big.Int; a b}`, structType(fieldDecl(bi), fieldDecl(tb, a))},
		{`struct{big.Int; a big.Int}`, structType(fieldDecl(bi), fieldDecl(bi, a))},
		{`struct{*big.Int}`, structType(fieldDecl(pointer(bi)))},
		{`struct{*big.Int; a}`, structType(fieldDecl(pointer(bi)), fieldDecl(ta))},
		{`struct{a; *big.Int}`, structType(fieldDecl(ta), fieldDecl(pointer(bi)))},

		// Tagged.
		{`struct{a "your it"}`, structType(tagFieldDecl(ta, strLit("your it")))},
		{`struct{*a "your it"}`, structType(tagFieldDecl(pointer(ta), strLit("your it")))},
		{`struct{big.Int "your it"}`, structType(tagFieldDecl(bi, strLit("your it")))},
		{`struct{*big.Int "your it"}`, structType(tagFieldDecl(pointer(bi), strLit("your it")))},
		{`struct{a "your it"; b}`, structType(tagFieldDecl(ta, strLit("your it")), fieldDecl(tb))},
		{`struct{a b "your it"}`, structType(tagFieldDecl(tb, strLit("your it"), a))},

		// Trailing ;
		{`struct{a;}`, structType(fieldDecl(ta))},
		{`struct{a; b; c;}`, structType(fieldDecl(ta), fieldDecl(tb), fieldDecl(tc))},

		// Embedded pointers must be type names.
		{`struct{**big.Int}`, parseErr("expected.*got \\*")},
		{`struct{*[]big.Int}`, parseErr("expected.*got \\[")},
	}
	tests.runType(t)
}

func TestInterfaceType(t *testing.T) {
	ta, tb, tc := ident("a"), ident("b"), ident("c")
	tests := parserTests{
		{`interface{}`, ifaceType()},
		{`interface{a; b; c}`, ifaceType(ta, tb, tc)},
		{`interface{a; b; big.Int}`, ifaceType(ta, tb, sel(ident("big"), ident("Int")))},
		{`interface{a; big.Int; b}`, ifaceType(ta, sel(ident("big"), ident("Int")), tb)},
		{`interface{big.Int; a; b}`, ifaceType(sel(ident("big"), ident("Int")), ta, tb)},

		{`interface{a; b; c()}`, ifaceType(ta, tb, method(c, parmList(), parmList()))},
		{`interface{a; b(); c}`, ifaceType(ta, method(b, parmList(), parmList()), tc)},
		{`interface{a(); b; c}`, ifaceType(method(a, parmList(), parmList()), tb, tc)},
		{`interface{a; big.Int; c()}`, ifaceType(ta, sel(ident("big"), ident("Int")), method(c, parmList(), parmList()))},

		// Trailing ;
		{`interface{a; b; c;}`, ifaceType(ta, tb, tc)},
	}
	tests.runType(t)
}

func TestFunctionType(t *testing.T) {
	ta, tb, tc := ident("a"), ident("b"), ident("c")
	tests := parserTests{
		{`func()`, funcType(parmList(), parmList())},
		{`func(a) b`, funcType(
			parmList(parmDecl(ta, false)),
			parmList(parmDecl(tb, false)))},
		{`func(...a) b`, funcType(
			parmList(parmDecl(ta, true)),
			parmList(parmDecl(tb, false)))},
		{`func(a, b) c`, funcType(
			parmList(parmDecl(ta, false), parmDecl(tb, false)),
			parmList(parmDecl(tc, false)))},
		{`func(a) (b, c)`, funcType(
			parmList(parmDecl(ta, false)),
			parmList(parmDecl(tb, false), parmDecl(tc, false)))},
		{`func(...a) (b, c)`, funcType(
			parmList(parmDecl(ta, true)),
			parmList(parmDecl(tb, false), parmDecl(tc, false)))},

		// Invalid, but will have to be caught during type checking.
		{`func(a) (b, ...c)`, funcType(
			parmList(parmDecl(ta, false)),
			parmList(parmDecl(tb, false), parmDecl(tc, true)))},
	}
	tests.runType(t)
}

func TestParameterList(t *testing.T) {
	ta, tb, tc := ident("a"), ident("b"), ident("c")
	tests := parserTests{
		{`()`, parmList()},

		// Parameter declarations without any identifiers.
		{`(a, b, c)`, parmList(
			parmDecl(ta, false),
			parmDecl(tb, false),
			parmDecl(tc, false))},
		{`(a, b, ...c)`, parmList(
			parmDecl(ta, false),
			parmDecl(tb, false),
			parmDecl(tc, true))},
		{`(a, b, big.Int)`, parmList(
			parmDecl(ta, false),
			parmDecl(tb, false),
			parmDecl(sel(ident("big"), ident("Int")), false))},
		{`(a, b, ...big.Int)`, parmList(
			parmDecl(ta, false),
			parmDecl(tb, false),
			parmDecl(sel(ident("big"), ident("Int")), true))},
		{`(a, b, []c)`, parmList(
			parmDecl(ta, false),
			parmDecl(tb, false),
			parmDecl(sliceType(tc), false))},
		{`(a, b, ...[]c)`, parmList(
			parmDecl(ta, false),
			parmDecl(tb, false),
			parmDecl(sliceType(tc), true))},
		{`([]a, b, c)`, parmList(
			parmDecl(sliceType(ta), false),
			parmDecl(tb, false),
			parmDecl(tc, false))},
		{`([]a, b, ...c)`, parmList(
			parmDecl(sliceType(ta), false),
			parmDecl(tb, false),
			parmDecl(tc, true))},
		{`(...a)`, parmList(parmDecl(ta, true))},

		// Parameter declarations with identifiers
		{`(a, b c)`, parmList(parmDecl(tc, false, a, b))},
		{`(a, b ...c)`, parmList(parmDecl(tc, true, a, b))},
		{`(a, b big.Int)`, parmList(parmDecl(sel(ident("big"), ident("Int")), false, a, b))},
		{`(a, b []int)`, parmList(parmDecl(sliceType(ident("int")), false, a, b))},
		{`(a, b ...[]int)`, parmList(parmDecl(sliceType(ident("int")), true, a, b))},
		{`(a, b []int, c, d ...[]int)`, parmList(
			parmDecl(sliceType(ident("int")), false, a, b),
			parmDecl(sliceType(ident("int")), true, c, d))},

		// Trailing comma is OK.
		{`(a, b, c,)`, parmList(
			parmDecl(ta, false),
			parmDecl(tb, false),
			parmDecl(tc, false))},
		{`(a, []b, c,)`, parmList(
			parmDecl(ta, false),
			parmDecl(sliceType(tb), false),
			parmDecl(tc, false))},
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
		{`chan a`, chanType(ident("a"))},
		{`<-chan a`, recvChan(ident("a"))},
		{`chan<- a`, sendChan(ident("a"))},
		{`chan<- <- chan a`, sendChan(recvChan(ident("a")))},
		{`chan<- chan a`, sendChan(chanType(ident("a")))},
		{`<- chan <-chan a`, recvChan(recvChan(ident("a")))},

		{`<- chan<- a`, parseErr("expected")},
		{`chan<- <- a`, parseErr("expected")},
		{`chan<- <- <- chan a`, parseErr("expected")},
	}
	tests.runType(t)
}

func TestMapType(t *testing.T) {
	tests := parserTests{
		{`map[int]a`, mapType(ident("int"), ident("a"))},
		{`map[*int]a.b`, mapType(pointer(ident("int")), sel(ident("a"), ident("b")))},
		{`map[*int]map[string]int]`, mapType(pointer(ident("int")), mapType(ident("string"), ident("int")))},
	}
	tests.runType(t)
}

func TestArrayType(t *testing.T) {
	tests := parserTests{
		{`[4]a`, arrayType(intLit("4"), ident("a"))},
		{`[4]a.b`, arrayType(intLit("4"), sel(ident("a"), ident("b")))},
		{`[4](a)`, arrayType(intLit("4"), ident("a"))},
		{`[4][]a`, arrayType(intLit("4"), sliceType(ident("a")))},
		{`[4][42*z]a`, arrayType(intLit("4"), arrayType(binOp(token.Star, intLit("42"), ident("z")), ident("a")))},

		// [...]Type notation is only allowed in composite literals,
		// not types in general
		{`[...]int`, parseErr("expected.*got \\.\\.\\.")},
	}
	tests.runType(t)
}

func TestSliceType(t *testing.T) {
	tests := parserTests{
		{`[]a`, sliceType(ident("a"))},
		{`[]a.b`, sliceType(sel(ident("a"), ident("b")))},
		{`[](a)`, sliceType(ident("a"))},
		{`[][]a)`, sliceType(sliceType(ident("a")))},
	}
	tests.runType(t)
}

func TestPointerType(t *testing.T) {
	tests := parserTests{
		{`*a`, pointer(ident("a"))},
		{`*a.b`, pointer(sel(ident("a"), ident("b")))},
		{`*(a)`, pointer(ident("a"))},
		{`**(a)`, pointer(pointer(ident("a")))},

		{`α.`, parseErr("expected.*Identifier.*got EOF")},
	}
	tests.runType(t)
}

func TestTypeName(t *testing.T) {
	tests := parserTests{
		{`a`, ident("a")},
		{`a.b`, sel(ident("a"), ident("b"))},
		{`α.b`, sel(ident("α"), ident("b"))},
	}
	tests.runType(t)
}
