// This file contains various things that are not portable across different target
// architectures. Eventually we will need a solution so that we can generate
// code for multiple targets.

package ast

import "math"

const (
	intBits = 64
	minInt  = math.MinInt64
	maxInt  = math.MaxInt64

	uintBits = 64
	maxUint  = math.MaxUint64

	uintptrBits = 64
	maxUintptr  = math.MaxUint64
)
