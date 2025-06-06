// Copyright (c) 2025 Fantom Foundation
//
// Use of this software is governed by the Business Source License included
// in the LICENSE file and at fantom.foundation/bsl11.
//
// Change Date: 2028-4-16
//
// On the date above, in accordance with the Business Source License, use of
// this software will be governed by the GNU Lesser General Public License v3.

package cachescale

// Ratio alters the cache sizes proportionally to a ratio
type Ratio struct {
	Base   uint64
	Target uint64
}

var _ Func = (*Ratio)(nil)

// Identity doesn't alter the cache sizes
var Identity = Ratio{1, 1}

func (r Ratio) U64(v uint64) uint64 {
	multipliedValue := v * r.Target
	if multipliedValue%r.Base == 0 {
		return multipliedValue / r.Base
	}
	return multipliedValue/r.Base + 1
}

func (r Ratio) F32(v float32) float32 {
	return v * (float32(r.Target) / float32(r.Base))
}

func (r Ratio) F64(v float64) float64 {
	return v * (float64(r.Target) / float64(r.Base))
}

func (r Ratio) U(v uint) uint {
	return uint(r.U64(uint64(v)))
}

func (r Ratio) U32(v uint32) uint32 {
	return uint32(r.U64(uint64(v)))
}

func (r Ratio) I(v int) int {
	return int(r.U64(uint64(v)))
}

func (r Ratio) I32(v int32) int32 {
	return int32(r.U64(uint64(v)))
}

func (r Ratio) I64(v int64) int64 {
	return int64(r.U64(uint64(v)))
}
