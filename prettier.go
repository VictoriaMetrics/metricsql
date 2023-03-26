package metricsql

import (
	"fmt"
	"log"
	"strings"
)

const indentString = "  "

func paddings(indent int) string {
	var str strings.Builder
	for i := 0; i < indent; i++ {
		str.WriteString(indentString)
	}
	return str.String()
}

func shouldWrap(expr Expr) bool {
	switch expr.(type) {
	case *MetricExpr, *NumberExpr, *StringExpr, *FuncExpr:
		return false
	// case *BinaryOpExpr:
	// 	var buf []byte
	// 	b := e.AppendString(buf)
	// 	log.Printf("B => %s", b)
	// 	return true
	default:
		return true
	}
}

func wrapWithBraces(expr Expr, b *strings.Builder, indent, maxLineLength int) {
	paddings := paddings(indent)
	if shouldWrap(expr) {
		b.WriteString(paddings + "(\n")
		b.WriteString(prettier(expr, indent+1, maxLineLength))
		b.WriteString("\n" + paddings + ")")
	} else {
		b.WriteString(prettier(expr, indent, maxLineLength))
	}
}

func prettier(expr Expr, indent, maxLineLength int) string {
	var str strings.Builder
	var b []byte
	b = expr.AppendString(b)
	log.Printf("LEN => %d", len(b))
	if len(b) <= maxLineLength {
		paddings := paddings(indent)
		str.WriteString(paddings)
		str.Write(b)
		return str.String()
	}
	switch e := expr.(type) {
	case *BinaryOpExpr:
		var binaryOpStr strings.Builder

		wrapWithBraces(e.Left, &binaryOpStr, indent+1, maxLineLength)
		buildBinaryOp(&binaryOpStr, e, indent)
		wrapWithBraces(e.Right, &binaryOpStr, indent+1, maxLineLength)

		str.WriteString(binaryOpStr.String())
	case *RollupExpr:
		var rollupStr strings.Builder

		wrapWithBraces(e.Expr, &rollupStr, indent, maxLineLength)
		buildRollupFunc(&rollupStr, e)

		str.WriteString(rollupStr.String())
	case *AggrFuncExpr:
		var aggrFuncStr strings.Builder

		buildAggrFuncString(&aggrFuncStr, e, indent, maxLineLength)

		str.WriteString(aggrFuncStr.String())
	case *MetricExpr, *NumberExpr, *StringExpr:
		paddings := paddings(indent)
		var b []byte
		str.WriteString(paddings)
		str.Write(e.AppendString(b))
	case *FuncExpr:
		buildFuncExpr(&str, e, indent, maxLineLength)
	}

	return str.String()
}

func buildRollupFunc(rollupStr *strings.Builder, e *RollupExpr) {
	if e.Window != nil || e.InheritStep || e.Step != nil {
		rollupStr.WriteString("[")
		if e.Window != nil {
			var b []byte
			b = e.Window.AppendString(b)
			rollupStr.WriteString(string(b))
		}
		if e.Step != nil {
			rollupStr.WriteString(":")
			rollupStr.WriteString(e.Step.s)
		} else if e.InheritStep {
			rollupStr.WriteString(":")
		}
		rollupStr.WriteString("]")
	}
	if e.Offset != nil {
		rollupStr.WriteString(fmt.Sprintf(" offset %s", e.Offset.s))
	}
	if e.At != nil {
		var b []byte
		rollupStr.WriteString(fmt.Sprintf(" @ (%s)", e.At.AppendString(b)))
	}
}

func buildBinaryOp(binaryOpStr *strings.Builder, e *BinaryOpExpr, indent int) {
	paddings := paddings(indent)
	binaryOpStr.WriteString(fmt.Sprintf("\n%s%s", paddings, e.Op))
	if e.Bool {
		binaryOpStr.WriteString(" bool")
	}
	if e.GroupModifier.Op != "" {
		binaryOpStr.WriteString(" ")
		binaryOpStr.Write(e.GroupModifier.AppendString(nil))
	}
	if e.JoinModifier.Op != "" {
		binaryOpStr.WriteString(" ")
		binaryOpStr.Write(e.JoinModifier.AppendString(nil))
	}
	binaryOpStr.WriteString("\n")
}

func buildAggrFuncString(aggrFuncStr *strings.Builder, e *AggrFuncExpr, indent, maxLineLength int) {
	paddings := paddings(indent)
	aggrFuncStr.WriteString(paddings + e.Name)
	if e.Modifier.Op != "" {
		aggrFuncStr.WriteString(" ")
		aggrFuncStr.Write(e.Modifier.AppendString(nil))
	}
	aggrFuncStr.WriteString(" (\n")
	for i, a := range e.Args {
		aggrFuncStr.WriteString(prettier(a, indent+1, maxLineLength))
		if i < len(e.Args)-1 {
			aggrFuncStr.WriteString(",")
		}
		aggrFuncStr.WriteString("\n")
	}
	aggrFuncStr.WriteString(paddings + ")")
}

func buildFuncExpr(funcStr *strings.Builder, e *FuncExpr, indent, maxLineLength int) {
	paddings := paddings(indent)
	if e.Name == "time" {
		funcStr.WriteString(paddings + "time ()")
	} else {
		var funcExprStr strings.Builder

		funcExprStr.WriteString(paddings + e.Name + " (\n")
		for i, a := range e.Args {
			funcExprStr.WriteString(prettier(a, indent+1, maxLineLength))
			if i < len(e.Args)-1 {
				funcExprStr.WriteString(",")
			}
			funcExprStr.WriteString("\n")
		}
		funcExprStr.WriteString(paddings + ")")

		funcStr.WriteString(funcExprStr.String())
	}
}

func Prettier(s string, maxLineLength int) (string, error) {
	expr, err := Parse(s)
	if err != nil {
		return "", err
	}

	// var b []byte
	// b = expr.AppendString(b)
	// if len(b) <= 80 {
	// 	return string(b), nil
	// }
	return prettier(expr, 0, maxLineLength), nil
}
