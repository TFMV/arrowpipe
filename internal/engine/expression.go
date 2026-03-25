package engine

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

// Predicate is an interface for evaluating a filter condition on a record.
type Predicate interface {
	// Eval evaluates the predicate on the given record and returns a boolean array
	// indicating which rows match the condition.
	Eval(rec arrow.Record) (arrow.Array, error)
}

// NewPredicate creates a new predicate from a filter expression string.
// For now, it only supports simple binary expressions (e.g., "col > 10").
func NewPredicate(expression string) (Predicate, error) {
	return &binaryExprPredicate{expression: expression}, nil
}

type binaryExprPredicate struct {
	expression string
}

func (p *binaryExprPredicate) Eval(rec arrow.Record) (arrow.Array, error) {
	parts := strings.Fields(p.expression)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid expression: %s", p.expression)
	}

	colName := parts[0]
	op := parts[1]
	literalStr := parts[2]

	idx := rec.Schema().FieldIndices(colName)
	if len(idx) == 0 {
		return nil, fmt.Errorf("column not found: %s", colName)
	}
	col := rec.Column(idx[0])

	mem := memory.NewGoAllocator()
	bldr := array.NewBooleanBuilder(mem)
	defer bldr.Release()

	switch typedCol := col.(type) {
	case *array.Int64:
		literal, err := strconv.ParseInt(literalStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid literal: %s", literalStr)
		}
		for i := 0; i < typedCol.Len(); i++ {
			val := typedCol.Value(i)
			switch op {
			case ">":
				bldr.Append(val > literal)
			case "<":
				bldr.Append(val < literal)
			case ">=":
				bldr.Append(val >= literal)
			case "<=":
				bldr.Append(val <= literal)
			case "==":
				bldr.Append(val == literal)
			case "!=":
				bldr.Append(val != literal)
			default:
				return nil, fmt.Errorf("unsupported operator: %s", op)
			}
		}
	case *array.Float64:
		literal, err := strconv.ParseFloat(literalStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid literal: %s", literalStr)
		}
		for i := 0; i < typedCol.Len(); i++ {
			val := typedCol.Value(i)
			switch op {
			case ">":
				bldr.Append(val > literal)
			case "<":
				bldr.Append(val < literal)
			case ">=":
				bldr.Append(val >= literal)
			case "<=":
				bldr.Append(val <= literal)
			case "==":
				bldr.Append(val == literal)
			case "!=":
				bldr.Append(val != literal)
			default:
				return nil, fmt.Errorf("unsupported operator: %s", op)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported column type for filtering")
	}

	return bldr.NewArray(), nil
}
