// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License included
// in the file licenses/BSL.txt and at www.mariadb.com/bsl11.
//
// Change Date: 2022-10-01
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt and at
// https://www.apache.org/licenses/LICENSE-2.0

// {{/*
// +build execgen_template
//
// This file is the execgen template for sum_agg.eg.go. It's formatted in a
// special way, so it's both valid Go and a valid text/template input. This
// permits editing this file with editor support.
//
// */}}

package exec

import (
	"github.com/cockroachdb/apd"
	"github.com/cockroachdb/cockroach/pkg/sql/exec/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/exec/types"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/pkg/errors"
)

// {{/*
// Declarations to make the template compile properly

// Dummy import to pull in "apd" package.
var _ apd.Decimal

// Dummy import to pull in "tree" package.
var _ tree.Datum

// _ASSIGN_DIV_INT64 is the template division function for assigning the first
// input to the result of the second input / the third input, where the third
// input is an int64.
func _ASSIGN_DIV_INT64(_, _, _ string) {
	panic("")
}

// */}}

func newAvgAgg(t types.T) (aggregateFunc, error) {
	switch t {
	// {{range .}}
	case _TYPES_T:
		return &avg_TYPEAgg{}, nil
	// {{end}}
	default:
		return nil, errors.Errorf("unsupported avg agg type %s", t)
	}
}

// {{range .}}

type avg_TYPEAgg struct {
	done bool

	groups  []bool
	scratch struct {
		curIdx int
		// groupSums[i] keeps track of the sum of elements belonging to the ith
		// group.
		groupSums []_GOTYPE
		// groupCounts[i] keeps track of the number of elements that we've seen
		// belonging to the ith group.
		groupCounts []int64
		// vec points to the output vector.
		vec []_GOTYPE
	}
}

var _ aggregateFunc = &avg_TYPEAgg{}

func (a *avg_TYPEAgg) Init(groups []bool, v coldata.Vec) {
	a.groups = groups
	a.scratch.vec = v._TemplateType()
	a.scratch.groupSums = make([]_GOTYPE, len(a.scratch.vec))
	a.scratch.groupCounts = make([]int64, len(a.scratch.vec))
	a.Reset()
}

func (a *avg_TYPEAgg) Reset() {
	copy(a.scratch.groupSums, zero_TYPEColumn)
	copy(a.scratch.groupCounts, zeroInt64Column)
	copy(a.scratch.vec, zero_TYPEColumn)
	a.scratch.curIdx = -1
	a.done = false
}

func (a *avg_TYPEAgg) CurrentOutputIndex() int {
	return a.scratch.curIdx
}

func (a *avg_TYPEAgg) SetOutputIndex(idx int) {
	if a.scratch.curIdx != -1 {
		a.scratch.curIdx = idx
		copy(a.scratch.groupSums[idx+1:], zero_TYPEColumn)
		copy(a.scratch.groupCounts[idx+1:], zeroInt64Column)
		// TODO(asubiotto): We might not have to zero a.scratch.vec since we
		// overwrite with an independent value.
		copy(a.scratch.vec[idx+1:], zero_TYPEColumn)
	}
}

func (a *avg_TYPEAgg) Compute(b coldata.Batch, inputIdxs []uint32) {
	if a.done {
		return
	}
	inputLen := b.Length()
	if inputLen == 0 {
		// The aggregation is finished. Flush the last value.
		if a.scratch.curIdx >= 0 {
			_ASSIGN_DIV_INT64("a.scratch.vec[a.scratch.curIdx]", "a.scratch.groupSums[a.scratch.curIdx]", "a.scratch.groupCounts[a.scratch.curIdx]")
		}
		a.scratch.curIdx++
		a.done = true
		return
	}
	col, sel := b.ColVec(int(inputIdxs[0]))._TemplateType(), b.Selection()
	if sel != nil {
		sel = sel[:inputLen]
		for _, i := range sel {
			x := 0
			if a.groups[i] {
				x = 1
			}
			a.scratch.curIdx += x
			_ASSIGN_ADD("a.scratch.groupSums[a.scratch.curIdx]", "a.scratch.groupSums[a.scratch.curIdx]", "col[i]")
			a.scratch.groupCounts[a.scratch.curIdx]++
		}
	} else {
		col = col[:inputLen]
		for i := range col {
			x := 0
			if a.groups[i] {
				x = 1
			}
			a.scratch.curIdx += x
			_ASSIGN_ADD("a.scratch.groupSums[a.scratch.curIdx]", "a.scratch.groupSums[a.scratch.curIdx]", "col[i]")
			a.scratch.groupCounts[a.scratch.curIdx]++
		}
	}

	for i := 0; i < a.scratch.curIdx; i++ {
		_ASSIGN_DIV_INT64("a.scratch.vec[i]", "a.scratch.groupSums[i]", "a.scratch.groupCounts[i]")
	}
}

// {{end}}
