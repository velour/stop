package ast

import (
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"testing"

	"github.com/eaburns/eq"
	"github.com/eaburns/pretty"
	"github.com/velour/stop/token"
)

var (
	boolType       = typ("bool")
	runeType       = typ("rune")
	intType        = typ("int")
	int8Type       = typ("int8")
	int16Type      = typ("int16")
	int32Type      = typ("int32")
	int64Type      = typ("int64")
	uintType       = typ("uint")
	byteType       = typ("byte")
	uint8Type      = typ("uint8")
	uint16Type     = typ("uint16")
	uint32Type     = typ("uint32")
	uint64Type     = typ("uint64")
	complex64Type  = typ("complex64")
	complex128Type = typ("complex128")
	float32Type    = typ("float32")
	float64Type    = typ("float64")
	stringType     = typ("string")

	t1Ident = typ("T1")
	t1Diff  = typ("T1")
)

func init() {
	boolType.Identifier.decl = univScope.Decls["bool"]
	runeType.Identifier.decl = univScope.Decls["rune"]
	intType.Identifier.decl = univScope.Decls["int"]
	int8Type.Identifier.decl = univScope.Decls["int8"]
	int16Type.Identifier.decl = univScope.Decls["int16"]
	int32Type.Identifier.decl = univScope.Decls["int32"]
	int64Type.Identifier.decl = univScope.Decls["int64"]
	uintType.Identifier.decl = univScope.Decls["uint"]
	byteType.Identifier.decl = univScope.Decls["byte"]
	uint8Type.Identifier.decl = univScope.Decls["uint8"]
	uint16Type.Identifier.decl = univScope.Decls["uint16"]
	uint32Type.Identifier.decl = univScope.Decls["uint32"]
	uint64Type.Identifier.decl = univScope.Decls["uint64"]
	int32Type.Identifier.decl = univScope.Decls["int32"]
	complex64Type.Identifier.decl = univScope.Decls["complex64"]
	complex128Type.Identifier.decl = univScope.Decls["complex128"]
	float32Type.Identifier.decl = univScope.Decls["float32"]
	float64Type.Identifier.decl = univScope.Decls["float64"]
	stringType.Identifier.decl = univScope.Decls["string"]

	t0.Identifier.decl = &TypeSpec{Identifier: *id("T0"), Type: intType}
	t1.Identifier.decl = &TypeSpec{Identifier: *id("T1"), Type: t0}
	// t1Ident is just like t1 and from the same declaration as t1.
	t1Ident.Identifier.decl = t1.Identifier.decl
	// t1Diff is just like t1 but from a different declaration.
	t1Diff.Identifier.decl = &TypeSpec{Identifier: *id("T1"), Type: intType}
}

func TestCheckErrors(t *testing.T) {
	tests := []struct {
		src  []string
		errs []reflect.Type
	}{
		{[]string{`package a`, `package a`}, []reflect.Type{}},
		{
			[]string{
				`package a
				const (
					complexConst = 5i
					floatConst = 3.1415926535
					intConst = 6
					runeConst = 'α'
					stringConst = "Hello, World!"
					trueConst, falseConst = true, false
				)`,
			},
			[]reflect.Type{},
		},
		{
			[]string{
				`package a; const pi = 3.1415926535`,
				`package a; const π = pi`,
			},
			[]reflect.Type{},
		},
		{
			[]string{
				`package a
				const (
					zero = iota
					one
					two
				)`,
			},
			[]reflect.Type{},
		},
		{
			[]string{`package a; const a = notDeclared`},
			[]reflect.Type{reflect.TypeOf(Undeclared{})},
		},
		{
			[]string{`package a; const a, b = 1`},
			[]reflect.Type{reflect.TypeOf(AssignCountMismatch{})},
		},
		{
			[]string{`package a; const nilConst = nil`},
			[]reflect.Type{reflect.TypeOf(NotConstant{})},
		},
		{
			[]string{`package a; const a = a`},
			[]reflect.Type{reflect.TypeOf(ConstantLoop{})},
		},
		{
			[]string{
				`package a; const a = b`,
				`package a; const b = c`,
				`package a; const c = a`,
			},
			[]reflect.Type{reflect.TypeOf(ConstantLoop{})},
		},
	}
	for _, test := range tests {
		want := make(map[reflect.Type]int)
		for _, e := range test.errs {
			want[e]++
		}

		var files []*File
		for _, src := range test.src {
			l := token.NewLexer("", src)
			p := NewParser(l)
			files = append(files, parseFile(p))
		}

		var got []reflect.Type
		if err := Check(files); err != nil {
			for _, e := range err.(errors).All() {
				got = append(got, reflect.TypeOf(e))
			}
		}
		for _, t := range got {
			want[t]--
		}

		diff := false
		for _, v := range want {
			if v != 0 {
				diff = true
				break
			}
		}
		if diff {
			t.Errorf("Check(%v)=%v, want %v", test.src, got, test.errs)
		}
	}
}

func TestIsAssignable(t *testing.T) {
	anInt32 := intLit("42")
	anInt32.typ = int32Type
	aUint8 := intLit("42")
	aUint8.typ = uint8Type
	t0Ident := id("t0")
	t0Ident.decl = &VarSpec{Type: t0}
	nilLit := &NilLiteral{typ: Untyped(NilConst)}

	namedT0Slice0 := typ("T0Slice")
	namedT0Slice0.Identifier.decl = &TypeSpec{
		Identifier: *id("T0Slice"),
		Type:       &SliceType{Element: t0},
	}
	// Just like namedT0Slice0, but via a different declaration.
	namedT0Slice1 := typ("T0Slice")
	namedT0Slice1.Identifier.decl = &TypeSpec{
		Identifier: *id("T0Slice"),
		Type:       &SliceType{Element: t0},
	}
	t0SliceIdent := id("t0slice")
	t0SliceIdent.decl = &VarSpec{Type: namedT0Slice0}

	bidirIntChan := id("ch")
	bidirIntChan.decl = &VarSpec{Type: &ChannelType{Send: true, Receive: true, Element: intType}}
	sendIntChan := id("ch")
	sendIntChan.decl = &VarSpec{Type: &ChannelType{Send: true, Element: intType}}
	recvIntChan := id("ch")
	recvIntChan.decl = &VarSpec{Type: &ChannelType{Receive: true, Element: intType}}

	tests := []struct {
		x  Expression
		t  Type
		ok bool
	}{
		{intLit("42"), intType, true},
		{intLit("42"), int8Type, true},
		{intLit("42"), int16Type, true},
		{intLit("42"), int32Type, true},
		{intLit("42"), int64Type, true},
		{intLit("42"), uintType, true},
		{intLit("42"), uint8Type, true},
		{intLit("42"), uint16Type, true},
		{intLit("42"), uint32Type, true},
		{intLit("42"), uint64Type, true},
		{intLit("42"), float32Type, true},
		{intLit("42"), float64Type, true},
		{intLit("42"), complex64Type, true},
		{intLit("42"), complex128Type, true},
		{strLit("Hello, World!"), intType, false},
		{anInt32, int32Type, true},
		{anInt32, runeType, true},
		{anInt32, intType, false},
		{aUint8, uint8Type, true},
		{aUint8, byteType, true},
		{aUint8, intType, false},
		{strLit("Hello, World!"), stringType, true},
		{intLit("42"), stringType, false},
		{t0Ident, t0, true},
		{t0Ident, intType, false},
		{t0Ident, &SliceType{Element: t0}, false},
		{t0SliceIdent, &SliceType{Element: t0}, true},
		{t0SliceIdent, namedT0Slice0, true},
		{t0SliceIdent, namedT0Slice1, false},
		{nilLit, intType, false},
		{nilLit, &StructType{}, false},
		{nilLit, &Star{Target: t0}, true},
		{nilLit, &SliceType{Element: t0}, true},
		{nilLit, &MapType{Key: t0, Value: intType}, true},
		{nilLit, &ChannelType{Element: t0}, true},
		{nilLit, &InterfaceType{}, true},
		{bidirIntChan, &ChannelType{Send: true, Element: intType}, true},
		{bidirIntChan, &ChannelType{Receive: true, Element: intType}, true},
		{bidirIntChan, &ChannelType{Send: true, Receive: true, Element: intType}, true},
		{bidirIntChan, &ChannelType{Send: true, Element: t0}, false},
		{sendIntChan, &ChannelType{Send: true, Element: intType}, true},
		{sendIntChan, &ChannelType{Receive: true, Element: intType}, false},
		{sendIntChan, &ChannelType{Send: true, Receive: true, Element: intType}, false},
		{sendIntChan, &ChannelType{Send: true, Element: t0}, false},
		{recvIntChan, &ChannelType{Send: true, Element: intType}, false},
		{recvIntChan, &ChannelType{Receive: true, Element: intType}, true},
		{recvIntChan, &ChannelType{Send: true, Receive: true, Element: intType}, false},
		{recvIntChan, &ChannelType{Receive: true, Element: t0}, false},
	}
	for _, test := range tests {
		if ok := IsAssignable(test.x, test.t); ok != test.ok {
			t.Errorf("IsAssignable(%s, %s)=%t, want %t", pretty.String(test.x), pretty.String(test.t), ok, test.ok)
		}
	}
}

func TestIsRepresentable(t *testing.T) {
	zero, neg1 := intLit("0"), intLit("-1")
	hello := strLit("Hello, World!")
	tests := []struct {
		x  Expression
		t  Type
		ok bool
	}{
		{&BoolLiteral{}, boolType, true},
		{zero, boolType, false},

		{imgLit("5i"), complex64Type, true},
		{floatLit("5"), complex64Type, true},
		{intLit("5"), complex64Type, true},
		{hello, complex64Type, false},
		{imgLit("5i"), complex128Type, true},
		{floatLit("5"), complex128Type, true},
		{intLit("5"), complex128Type, true},
		{hello, complex128Type, false},

		{floatLit("5"), float32Type, true},
		{intLit("5"), float32Type, true},
		{hello, float32Type, false},
		{floatLit("5"), float64Type, true},
		{intLit("5"), float64Type, true},
		{hello, float64Type, false},

		{hello, stringType, true},
		{zero, stringType, false},

		{intLit(strconv.Itoa(minInt)), intType, true},
		{intLit(strconv.Itoa(maxInt)), intType, true},
		{intLit(strconv.Itoa(math.MinInt8 - 1)), int8Type, false},
		{intLit(strconv.Itoa(math.MinInt8)), int8Type, true},
		{intLit(strconv.Itoa(math.MinInt8)), int8Type, true},
		{intLit(strconv.Itoa(math.MaxInt8 + 1)), int8Type, false},
		{intLit(strconv.Itoa(math.MinInt16 - 1)), int16Type, false},
		{intLit(strconv.Itoa(math.MinInt16)), int16Type, true},
		{intLit(strconv.Itoa(math.MaxInt16)), int16Type, true},
		{intLit(strconv.Itoa(math.MaxInt16 + 1)), int16Type, false},
		{intLit(strconv.Itoa(math.MinInt32 - 1)), int32Type, false},
		{intLit(strconv.Itoa(math.MinInt32)), int32Type, true},
		{intLit(strconv.Itoa(math.MaxInt32)), int32Type, true},
		{intLit(strconv.Itoa(math.MaxInt32 + 1)), int32Type, false},
		{intLit("-9223372036854775809"), int32Type, false},
		{intLit(strconv.Itoa(math.MinInt64)), int64Type, true},
		{intLit(strconv.Itoa(math.MaxInt64)), int64Type, true},
		{intLit("9223372036854775808"), int32Type, false},
		{neg1, uintType, false},
		{zero, uintType, true},
		{intLit(strconv.FormatUint(maxUint, 10)), uintType, true},
		{neg1, byteType, false},
		{zero, byteType, true},
		{intLit(strconv.Itoa(math.MaxUint8)), byteType, true},
		{intLit(strconv.Itoa(math.MaxUint8 + 1)), byteType, false},
		{neg1, uint8Type, false},
		{zero, uint8Type, true},
		{intLit(strconv.Itoa(math.MaxUint8)), uint8Type, true},
		{intLit(strconv.Itoa(math.MaxUint8 + 1)), uint8Type, false},
		{neg1, uint16Type, false},
		{zero, uint16Type, true},
		{intLit(strconv.Itoa(math.MaxUint16)), uint16Type, true},
		{intLit(strconv.Itoa(math.MaxUint16 + 1)), uint16Type, false},
		{neg1, uint32Type, false},
		{zero, uint32Type, true},
		{intLit(strconv.Itoa(math.MaxUint32)), uint32Type, true},
		{intLit(strconv.Itoa(math.MaxUint32 + 1)), uint32Type, false},
		{neg1, uint64Type, false},
		{zero, uint64Type, true},
		{intLit(strconv.FormatUint(math.MaxUint64, 10)), uint64Type, true},
		{intLit("18446744073709551616"), uint64Type, false},
		{&BoolLiteral{}, intType, false},

		{&BoolLiteral{}, &SliceType{Element: intType}, false},
	}
	for _, test := range tests {
		if ok := IsRepresentable(test.x, test.t); ok != test.ok {
			t.Errorf("IsRepresentable(%s, %s)=%t, want %t", pretty.String(test.x), pretty.String(test.t), ok, test.ok)
		}
	}
}

func TestTypeIdentical(t *testing.T) {
	// We don't care about the contents of these, just that &pkgA != &pkgB.
	var pkgA, pkgB packageDecl

	tests := []struct {
		u, v  Type
		ident bool
	}{
		{intType, intType, true},
		{byteType, byteType, true},
		{byteType, uint8Type, true},
		{runeType, int32Type, true},
		{runeType, intType, false},
		{intType, t0, false},
		{t0, t0, true},
		{t0, t1, false},
		{t1, t1Ident, true},
		{t1, t1Diff, false},
		{t0, &Star{Target: t0}, false},

		{&Star{Target: intType}, &Star{Target: intType}, true},
		{&Star{Target: int32Type}, &Star{Target: runeType}, true},
		{&Star{Target: int32Type}, &Star{Target: intType}, false},
		{
			&Star{Target: &Star{Target: intType}},
			&Star{Target: &Star{Target: intType}},
			true,
		},
		{
			&Star{Target: &Star{Target: intType}},
			&Star{Target: intType},
			false,
		},
		{
			&SliceType{Element: intType},
			&SliceType{Element: intType},
			true,
		},
		{
			&SliceType{Element: int32Type},
			&SliceType{Element: runeType},
			true,
		},
		{
			&SliceType{Element: int32Type},
			&SliceType{Element: intType},
			false,
		},
		{
			&SliceType{Element: &SliceType{Element: intType}},
			&SliceType{Element: &SliceType{Element: intType}},
			true,
		},
		{
			&SliceType{Element: &SliceType{Element: intType}},
			&SliceType{Element: intType},
			false,
		},
		{&SliceType{Element: int32Type}, t0, false},
		{
			&ArrayType{Size: intLit("1"), Element: intType},
			&ArrayType{Size: intLit("1"), Element: intType},
			true,
		},
		{
			&ArrayType{Size: intLit("1"), Element: intType},
			&ArrayType{Size: intLit("2"), Element: intType},
			false,
		},
		{
			&ArrayType{Size: intLit("1"), Element: intType},
			&ArrayType{Size: intLit("1"), Element: t0},
			false,
		},
		{
			&ArrayType{
				Size: intLit("1"),
				Element: &ArrayType{
					Size:    intLit("1"),
					Element: t0,
				},
			},
			&ArrayType{
				Size: intLit("1"),
				Element: &ArrayType{
					Size:    intLit("1"),
					Element: t0,
				},
			},
			true,
		},
		{
			&ArrayType{
				Size: intLit("2"),
				Element: &ArrayType{
					Size:    intLit("1"),
					Element: t0,
				},
			},
			&ArrayType{
				Size: intLit("1"),
				Element: &ArrayType{
					Size:    intLit("2"),
					Element: t0,
				},
			},
			false,
		},
		{
			&ArrayType{Size: intLit("1"), Element: intType},
			&SliceType{Element: intType},
			false,
		},
		{
			&MapType{Key: intType, Value: t1},
			&MapType{Key: intType, Value: t1Ident},
			true,
		},
		{
			&MapType{Key: intType, Value: t1},
			&MapType{Key: intType, Value: t1Diff},
			false,
		},
		{
			&MapType{Key: t0, Value: t1},
			&MapType{Key: intType, Value: t1},
			false,
		},
		{
			&ChannelType{Element: intType},
			&ChannelType{Element: intType},
			true,
		},
		{
			&ChannelType{Receive: true, Element: intType},
			&ChannelType{Receive: true, Element: intType},
			true,
		},
		{
			&ChannelType{Receive: true, Send: true, Element: intType},
			&ChannelType{Receive: true, Send: true, Element: intType},
			true,
		},
		{
			&ChannelType{Element: intType},
			&ChannelType{Send: true, Element: intType},
			false,
		},
		{
			&ChannelType{Receive: true, Element: intType},
			&ChannelType{Element: intType},
			false,
		},
		{
			&ChannelType{Element: t1},
			&ChannelType{Element: t1Diff},
			false,
		},
		{
			&ChannelType{Element: &ChannelType{Element: runeType}},
			&ChannelType{Element: &ChannelType{Element: int32Type}},
			true,
		},
		{
			&ChannelType{Receive: true, Element: &ChannelType{Element: t0}},
			&ChannelType{Element: &ChannelType{Receive: true, Element: t0}},
			false,
		},
		{&ChannelType{Element: t1}, t0, false},
		{&FunctionType{}, &FunctionType{}, true},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{
					{Type: intType, Identifier: id("a")},
				},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{
					{Type: intType, Identifier: id("a")},
				},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{
					{Type: intType, Identifier: id("a")},
				},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{
					{Type: intType, Identifier: id("notA")},
				},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t1Ident}},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t1Diff}},
			}},
			false,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}},
			}},
			&FunctionType{Signature: Signature{
				Results: []ParameterDecl{{Type: t0}},
			}},
			false,
		},
		{
			&FunctionType{Signature: Signature{
				Results: []ParameterDecl{{Type: t0}},
			}},
			&FunctionType{Signature: Signature{
				Results: []ParameterDecl{{Type: t0}},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t1}, {Type: t0}},
			}},
			false,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1, DotDotDot: true}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			false,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1, DotDotDot: true}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1, DotDotDot: true}},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
				Results:    []ParameterDecl{{Type: intType}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
				Results:    []ParameterDecl{{Type: intType}},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: intType}},
				Results:    []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
				Results:    []ParameterDecl{{Type: intType}},
			}},
			false,
		},
		{
			&FunctionType{Signature: Signature{
				Results: []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Results: []ParameterDecl{{Type: t1}, {Type: t0}},
			}},
			false,
		},
		{&FunctionType{}, t0, false},

		// The Methods field is not used when checking for interface identity, but
		// the unexported methodSet field won't print on test errors, so we kindly
		// populate both.
		{&InterfaceType{}, &InterfaceType{}, true},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			true,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgB},
				},
			},
			false,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
				},
				methodSet: []*Method{
					{Identifier: *id("A"), pkg: &pkgA},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
				},
				methodSet: []*Method{
					{Identifier: *id("A"), pkg: &pkgB},
				},
			},
			true,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			&InterfaceType{},
			false,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("b")},
				},
				methodSet: []*Method{
					{Identifier: *id("b"), pkg: &pkgA},
				},
			},
			false,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
					},
				},
				methodSet: []*Method{
					{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
						pkg: &pkgA,
					},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
					},
				},
				methodSet: []*Method{
					{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
						pkg: &pkgA,
					},
				},
			},
			true,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{
						Identifier: *id("a"),
						Signature: Signature{
							Results: []ParameterDecl{{Type: intType}},
						},
					},
				},
				methodSet: []*Method{
					{
						Identifier: *id("a"),
						Signature: Signature{
							Results: []ParameterDecl{{Type: intType}},
						},
						pkg: &pkgA,
					},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{
						Identifier: *id("a"),
						Signature: Signature{
							Results: []ParameterDecl{{Type: t0}},
						},
					},
				},
				methodSet: []*Method{
					{
						Identifier: *id("a"),
						Signature: Signature{
							Results: []ParameterDecl{{Type: t0}},
						},
						pkg: &pkgA,
					},
				},
			},
			false,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
					},
				},
				methodSet: []*Method{
					{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
						pkg: &pkgA,
					},
				},
			},
			false,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
					&Method{Identifier: *id("B")},
				},
				methodSet: []*Method{
					{Identifier: *id("A")},
					{Identifier: *id("B")},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
					&Method{Identifier: *id("B")},
				},
				methodSet: []*Method{
					{Identifier: *id("A")},
					{Identifier: *id("B")},
				},
			},
			true,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
					&Method{Identifier: *id("B")},
				},
				methodSet: []*Method{
					{Identifier: *id("A")},
					{Identifier: *id("B")},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
					&Method{Identifier: *id("C")},
				},
				methodSet: []*Method{
					{Identifier: *id("A")},
					{Identifier: *id("C")},
				},
			},
			false,
		},
		{&InterfaceType{}, t0, false},
		{&StructType{}, &StructType{}, true},
		{
			&StructType{Fields: []FieldDecl{{Type: intType}}},
			&StructType{Fields: []FieldDecl{{Type: intType}}},
			true,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("a"), Type: intType},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("a"), Type: intType},
			}},
			true,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("y"), Type: &SliceType{Element: t0}},
				{Identifier: id("z"), Type: t1},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("y"), Type: &SliceType{Element: t0}},
				{Identifier: id("z"), Type: t1},
			}},
			true,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("y"), Type: &SliceType{Element: t0}},
				{Identifier: id("z"), Type: t1},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("z"), Type: t1},
			}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("y"), Type: &SliceType{Element: t0}},
				{Identifier: id("z"), Type: t1},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("z"), Type: t1},
			}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t1},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
			}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t1, Tag: strLit("tag")},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t1, Tag: strLit("tag")},
			}},
			true,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t1, Tag: strLit("tag")},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t1, Tag: strLit("different tag")},
			}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{{Identifier: id("x")}}},
			&StructType{Fields: []FieldDecl{{Identifier: id("y")}}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{{Identifier: id("x"), Type: t0}}},
			&StructType{Fields: []FieldDecl{{Type: t0}}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t0, pkg: &pkgA},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t0, pkg: &pkgA},
			}},
			true,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), pkg: &pkgA},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), pkg: &pkgB},
			}},
			false,
		},
		{&StructType{}, t0, false},
	}
	for _, test := range tests {
		ident := test.u.Identical(test.v)
		if ident != test.ident {
			t.Errorf("(%s).Identical(%s)=%t, want %t", pretty.String(test.u), pretty.String(test.v), ident, test.ident)
		}
		identReflect := test.v.Identical(test.u)
		if ident == identReflect {
			continue
		}
		u := pretty.String(test.u)
		v := pretty.String(test.v)
		t.Errorf("(%s).Identical(%s)=%t, but (%s).Identical(%s)=%t", u, v, ident, v, u, identReflect)
	}
}

func TestTypeUnderlying(t *testing.T) {
	tests := []struct {
		t Type
		u Type
	}{
		{t: intType, u: intType},
		{t: byteType, u: byteType},
		{t: t0, u: intType},
		{t: t1, u: intType},
		{t: &Star{Target: t0}, u: &Star{Target: t0}},
		{
			t: &SliceType{Element: intType},
			u: &SliceType{Element: intType},
		},
		{
			t: &ArrayType{Size: intLit("5"), Element: intType},
			u: &ArrayType{Size: intLit("5"), Element: intType},
		},
		{
			t: &MapType{Key: t0, Value: intType},
			u: &MapType{Key: t0, Value: intType},
		},
		{
			t: &ChannelType{Send: true, Element: intType},
			u: &ChannelType{Send: true, Element: intType},
		},
		{t: &FunctionType{}, u: &FunctionType{}},
		{t: &InterfaceType{}, u: &InterfaceType{}},
		{t: &StructType{}, u: &StructType{}},
	}
	for _, test := range tests {
		u := test.t.Underlying()
		if eq.Deep(u, test.u) {
			continue
		}
		t.Errorf("(%s).Underlying()=%s, want %s", pretty.String(test.t), pretty.String(u), pretty.String(test.u))
	}
}

func TestPkgDecls(t *testing.T) {
	tests := []struct {
		src []string
		// The top-level identifiers.
		ids []string
		// An expected error, if non-empty.
		err string
	}{
		{
			src: []string{
				`package a; const a = 1`,
				`package a; const B = 1`,
			},
			ids: []string{"a", "B"},
		},
		{
			src: []string{`package a; const a, B = B, 1`},
			ids: []string{"a", "B"},
		},
		{
			src: []string{
				`package a
			const (
				a, B = B, 1
				c, d
			)`,
			},
			ids: []string{"a", "B", "c", "d"},
		},
		{
			src: []string{
				`package a; var a = 1`,
				`package a; var B = 1`,
			},
			ids: []string{"a", "B"},
		},
		{
			src: []string{`package a; var a, B = 1, 2`},
			ids: []string{"a", "B"},
		},
		{
			src: []string{
				`package a
			var (
				a, B = 0, 1
				c, d = 3.0, 4.0
			)`,
			},
			ids: []string{"a", "B", "c", "d"},
		},
		{
			src: []string{
				`package a
			type (
				a struct{}
				B struct { C }
			)`,
			},
			ids: []string{"a", "B"},
		},
		{
			src: []string{
				`package a
			func a(){}
			func B() int { return 0 }`,
			},
			ids: []string{"a", "B"},
		},
		{
			src: []string{
				`package a
			func a(){}
			func B() int { return 0 }`,
				`package a
			func c(e int){}
			func d() (f int) { return 0 }`,
			},
			ids: []string{"a", "B", "c", "d"},
		},
		{
			src: []string{
				`package a
			func (z int) a(){}
			func (z float64) B() int { return 0 }`,
			},
			ids: []string{"a", "B"},
		},
		{
			src: []string{
				`package a
			func (z T0) a(){}
			func (z T1) B() int { return 0 }`,
				`package a
			func (z T2) c(e int){}
			func (z T3) d() (f int) { return 0 }`,
			},
			ids: []string{"a", "B", "c", "d"},
		},

		// Redeclaration errors.
		{
			src: []string{`package a; const a = 1; const a = 2`},
			err: "a redeclared",
		},
		{
			src: []string{`package a; const a = 1; var a = 2`},
			err: "a redeclared",
		},
		{
			src: []string{`package a; const a = 1; func a(){}`},
			err: "a redeclared",
		},
		{
			src: []string{
				`package a; const a = 1`,
				`package a; const a = 2`,
			},
			err: "a redeclared",
		},
		{
			src: []string{`
				package a
				import (
					"fmt"
					"fmt"
				)
			`},
			err: "fmt redeclared",
		},
		{
			src: []string{`
				package a
				import (
					"fmt"
					"foo/bar/fmt"
				)
			`},
			err: "fmt redeclared",
		},
		{
			src: []string{`
				package a
				import fmt "foo"
				import "fmt"
			`},
			err: "fmt redeclared",
		},
		{
			src: []string{`
				package a
				import (
					"fmt"
					fmt "foo/bar"
				)
			`},
			err: "fmt redeclared",
		},
		{
			// Not an error to redeclare the same import in different files.
			src: []string{
				`package a; import "fmt"`,
				`package a; import "fmt"`,
			},
		},

		// The blank identifier is not redeclared.
		{
			src: []string{`package a; const _ = 1; const _ = 2`},
		},
	}

	for _, test := range tests {
		files := parseSrcFiles(t, test.src)
		if files == nil {
			continue
		}
		var ids []string
		seen := make(map[Declaration]bool)
		s, err := pkgDecls(files)
		if test.err != "" {
			if err == nil {
				t.Errorf("pkgDecls(%v), err=nil, want matching %s", test.src, test.err)
			} else if !regexp.MustCompile(test.err).MatchString(err.Error()) {
				t.Errorf("pkgDecls(%v), err=[%v], want matching [%s]", test.src, err, test.err)
			}
			continue
		}
		if err != nil {
			t.Errorf("pkgDecls(%v), unexpected err=[%v]", test.src, err)
			continue
		}
		for id, d := range s.Decls {
			ids = append(ids, id)
			if seen[d] {
				t.Errorf("pkgDecls(%v), multiple idents map to declaration %s", test.src, pretty.String(d))
			}
			seen[d] = true
		}
		sort.Strings(ids)
		sort.Strings(test.ids)
		if reflect.DeepEqual(ids, test.ids) {
			continue
		}
		t.Errorf("pkgDecls(%v)=%v, want %v", test.src, ids, test.ids)
	}
}

func parseSrcFiles(t *testing.T, srcFiles []string) []*File {
	var files []*File
	for _, src := range srcFiles {
		p := NewParser(token.NewLexer("", src))
		file, err := Parse(p)
		if err == nil {
			files = append(files, file)
			continue
		}
		t.Fatalf("Parse(%s), got unexpected error %v", src, err)
		return nil
	}
	return files
}

func TestScopeFind(t *testing.T) {
	src := `
		package testpkg
		const (
			π = 3.1415926535
			name = "eaburns"
		)
		var mayChange int = f()
		var len = 1	// Shadows the predeclared len function.
		func f() int { return 0 }
	`
	p := NewParser(token.NewLexer("", src))
	s, err := pkgDecls([]*File{parseFile(p)})
	if err != nil {
		panic(err)
	}

	tests := []struct {
		id       string
		declType reflect.Type
	}{
		{"π", reflect.TypeOf(&constSpecView{})},
		{"name", reflect.TypeOf(&constSpecView{})},
		{"mayChange", reflect.TypeOf(&varSpecView{})},
		{"len", reflect.TypeOf(&varSpecView{})},
		{"f", reflect.TypeOf(&FunctionDecl{})},
		{"notDefined", nil},
	}
	for _, test := range tests {
		d := s.Find(test.id)
		typ := reflect.TypeOf(d)
		if (d == nil && test.declType != nil) || (d != nil && typ != test.declType) {
			t.Errorf("Find(%s)=%s, want %s", test.id, typ, test.declType)
		}
	}

	// Our testing source file shadowed len, but let's check that it's
	// still findable in the universal scope.
	d := univScope.Find("len")
	typ := reflect.TypeOf(d)
	lenType := reflect.TypeOf(&predeclaredFunc{})
	if d == nil || typ != lenType {
		t.Errorf("Find(len)=%s, want %s", typ, lenType)
	}

	// Rune and byte are aliases for int32 and uint8 respectively;
	// the have the same declaration.
	if univScope.Find("rune") != univScope.Find("int32") {
		t.Errorf("rune is not an alias for int32")
	}
	if univScope.Find("byte") != univScope.Find("uint8") {
		t.Errorf("byte is not an alias for uint8")
	}
}
