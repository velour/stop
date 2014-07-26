package ast

func (n *StructType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *InterfaceType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Signature) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *ChannelType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *MapType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *ArrayType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *SliceType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Star) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *TypeName) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *CompositeLiteral) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Index) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Slice) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *TypeAssertion) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Selector) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Call) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *BinaryOp) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *UnaryOp) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Identifier) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *IntegerLiteral) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *FloatLiteral) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *ComplexLiteral) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *RuneLiteral) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *StringLiteral) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *NilLiteral) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n Untyped) Check(*symtab, int) (Expression, error) { return n, nil }
