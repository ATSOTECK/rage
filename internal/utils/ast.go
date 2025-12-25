package utils

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/model"
)

func PrintAST(node model.Node, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	switch n := node.(type) {
	case *model.Module:
		fmt.Printf("%sModule\n", prefix)
		for _, stmt := range n.Body {
			PrintAST(stmt, indent+1)
		}
	case *model.FunctionDef:
		fmt.Printf("%sFunctionDef: %s\n", prefix, n.Name.Name)
		fmt.Printf("%s  Args: %d\n", prefix, len(n.Args.Args))
		fmt.Printf("%s  Body:\n", prefix)
		for _, stmt := range n.Body {
			PrintAST(stmt, indent+2)
		}
	case *model.If:
		fmt.Printf("%sIf\n", prefix)
		fmt.Printf("%s  Test: %T\n", prefix, n.Test)
		fmt.Printf("%s  Body:\n", prefix)
		for _, stmt := range n.Body {
			PrintAST(stmt, indent+2)
		}
		if len(n.OrElse) > 0 {
			fmt.Printf("%s  OrElse:\n", prefix)
			for _, stmt := range n.OrElse {
				PrintAST(stmt, indent+2)
			}
		}
	case *model.Return:
		fmt.Printf("%sReturn\n", prefix)
		if n.Value != nil {
			fmt.Printf("%s  Value: %T\n", prefix, n.Value)
		}
	case *model.Assign:
		fmt.Printf("%sAssign\n", prefix)
		fmt.Printf("%s  Targets: %d\n", prefix, len(n.Targets))
		fmt.Printf("%s  Value: %T\n", prefix, n.Value)
	case *model.ExprStmt:
		fmt.Printf("%sExprStmt\n", prefix)
		fmt.Printf("%s  Value: %T\n", prefix, n.Value)
	case *model.Call:
		fmt.Printf("%sCall\n", prefix)
		if id, ok := n.Func.(*model.Identifier); ok {
			fmt.Printf("%s  Func: %s\n", prefix, id.Name)
		}
		fmt.Printf("%s  Args: %d\n", prefix, len(n.Args))
	default:
		fmt.Printf("%s%T\n", prefix, node)
	}
}
