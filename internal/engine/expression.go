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
// It supports simple binary expressions and AND (&&) / OR (||) operators.
func NewPredicate(expression string) (Predicate, error) {
	expr := strings.TrimSpace(expression)
	if strings.Contains(expr, "&&") {
		parts := strings.Split(expr, "&&")
		left, err := NewPredicate(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, err
		}
		right, err := NewPredicate(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		return &andPredicate{left: left, right: right}, nil
	}
	if strings.Contains(expr, "||") {
		parts := strings.Split(expr, "||")
		left, err := NewPredicate(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, err
		}
		right, err := NewPredicate(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		return &orPredicate{left: left, right: right}, nil
	}
	return &binaryExprPredicate{expression: expr}, nil
}

type andPredicate struct {
	left  Predicate
	right Predicate
}

func (p *andPredicate) Eval(rec arrow.Record) (arrow.Array, error) {
	leftArr, err := p.left.Eval(rec)
	if err != nil {
		return nil, err
	}
	defer leftArr.Release()

	rightArr, err := p.right.Eval(rec)
	if err != nil {
		return nil, err
	}
	defer rightArr.Release()

	left := leftArr.(*array.Boolean)
	right := rightArr.(*array.Boolean)

	mem := memory.NewGoAllocator()
	bldr := array.NewBooleanBuilder(mem)
	defer bldr.Release()

	for i := 0; i < left.Len(); i++ {
		bldr.Append(left.Value(i) && right.Value(i))
	}

	return bldr.NewArray(), nil
}

type orPredicate struct {
	left  Predicate
	right Predicate
}

func (p *orPredicate) Eval(rec arrow.Record) (arrow.Array, error) {
	leftArr, err := p.left.Eval(rec)
	if err != nil {
		return nil, err
	}
	defer leftArr.Release()

	rightArr, err := p.right.Eval(rec)
	if err != nil {
		return nil, err
	}
	defer rightArr.Release()

	left := leftArr.(*array.Boolean)
	right := rightArr.(*array.Boolean)

	mem := memory.NewGoAllocator()
	bldr := array.NewBooleanBuilder(mem)
	defer bldr.Release()

	for i := 0; i < left.Len(); i++ {
		bldr.Append(left.Value(i) || right.Value(i))
	}

	return bldr.NewArray(), nil
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

	literalStr = strings.Trim(literalStr, "'\"")
	isStringLiteral := strings.HasPrefix(parts[2], "'") || strings.HasPrefix(parts[2], "\"")

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
		if isStringLiteral {
			return nil, fmt.Errorf("cannot compare int64 column to string literal")
		}
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
		if isStringLiteral {
			return nil, fmt.Errorf("cannot compare float64 column to string literal")
		}
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
	case *array.String:
		if !isStringLiteral {
			return nil, fmt.Errorf("string column requires quoted string literal")
		}
		for i := 0; i < typedCol.Len(); i++ {
			val := typedCol.Value(i)
			switch op {
			case "==":
				bldr.Append(val == literalStr)
			case "!=":
				bldr.Append(val != literalStr)
			default:
				return nil, fmt.Errorf("unsupported operator for strings: %s", op)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported column type for filtering: %s", col.DataType())
	}

	return bldr.NewArray(), nil
}
