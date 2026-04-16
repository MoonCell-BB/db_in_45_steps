package parser

import (
	"errors"
	"slices"

	"github.com/mooncell-bb/db_in_45_steps/database"
)

func evalExpr(schema *database.Schema, row database.Row, expr any) (*database.Cell, error) {
	switch e := expr.(type) {
	case string:
		idx := slices.IndexFunc(schema.Cols, func(col database.Column) bool {
			return col.Name == e
		})

		if idx < 0 {
			return nil, errors.New("unknown colnum")
		}
		return &row[idx], nil
	case *database.Cell:
		return e, nil
	case *ExprBinOp:
		left, err := evalExpr(schema, row, e.left)
		if err != nil {
			return nil, err
		}

		right, err := evalExpr(schema, row, e.right)
		if err != nil {
			return nil, err
		}

		out := &database.Cell{Type: left.Type}
		switch {
		case e.op == database.OP_ADD, out.Type == database.TypeStr:
			out.Str = slices.Concat(left.Str, right.Str)
		case e.op == database.OP_ADD && out.Type == database.TypeI64:
			out.I64 = left.I64 + right.I64
		case e.op == database.OP_SUB && out.Type == database.TypeI64:
			out.I64 = left.I64 - right.I64
		default:
			return nil, errors.New("bad binary op")
		}

		return out, nil
	default:
		panic("unknown expr type")
	}
}
