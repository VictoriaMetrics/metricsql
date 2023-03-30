package metricsql

import (
	"fmt"
	"strings"
)

const indentString = "  "

func pad(indent int, s string) string {
	p := strings.Repeat(indentString, indent)
	return p + s
}

func shouldWrap(expr Expr) bool {
	switch expr.(type) {
	case *MetricExpr, *NumberExpr, *StringExpr, *FuncExpr:
		return false
	default:
		return true
	}
}

func wrapWithBraces(expr Expr, b *strings.Builder, indent, maxLineLength int) {
	if shouldWrap(expr) {
		b.WriteString(pad(indent, "(\n"))
		b.WriteString(prettify(expr, indent+1, maxLineLength))
		b.WriteString("\n" + pad(indent, ")"))
	} else {
		b.WriteString(prettify(expr, indent, maxLineLength))
	}
}

func prettify(expr Expr, indent, maxLineLength int) string {
	var str strings.Builder
	var b []byte
	b = expr.AppendString(b)
	if len(b) <= maxLineLength {
		str.WriteString(pad(indent, ""))
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
		var b []byte
		str.WriteString(pad(indent, ""))
		str.Write(e.AppendString(b))
	case *FuncExpr:
		buildFuncExpr(&str, e, indent, maxLineLength)
	default:
		e.AppendString(b)
		str.WriteString(pad(indent, ""))
		str.Write(b)
	}
	return str.String()
}

func buildRollupFunc(rollupStr *strings.Builder, e *RollupExpr) {
	if e.Window != nil || e.InheritStep || e.Step != nil {
		rollupStr.WriteString("[")
		if e.Window != nil {
			var b []byte
			b = e.Window.AppendString(b)
			rollupStr.Write(b)
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
	binaryOpStr.WriteString(fmt.Sprintf("\n%s%s", pad(indent, ""), e.Op))
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
	aggrFuncStr.WriteString(pad(indent, e.Name))
	if e.Modifier.Op != "" {
		aggrFuncStr.WriteString(" ")
		aggrFuncStr.Write(e.Modifier.AppendString(nil))
	}
	aggrFuncStr.WriteString(" (\n")
	for i, a := range e.Args {
		aggrFuncStr.WriteString(prettify(a, indent+1, maxLineLength))
		if i < len(e.Args)-1 {
			aggrFuncStr.WriteString(",")
		}
		aggrFuncStr.WriteString("\n")
	}
	aggrFuncStr.WriteString(pad(indent, ")"))
}

func buildFuncExpr(funcStr *strings.Builder, e *FuncExpr, indent, maxLineLength int) {
	if e.Name == "time" {
		funcStr.WriteString(pad(indent, "time ()"))
	} else {
		var funcExprStr strings.Builder

		funcExprStr.WriteString(pad(indent, e.Name) + " (\n")
		for i, a := range e.Args {
			funcExprStr.WriteString(prettify(a, indent+1, maxLineLength))
			if i < len(e.Args)-1 {
				funcExprStr.WriteString(",")
			}
			funcExprStr.WriteString("\n")
		}
		funcExprStr.WriteString(pad(indent, ")"))

		funcStr.WriteString(funcExprStr.String())
	}
}

func Prettify(s string, maxLineLength int) (string, error) {
	expr, err := Parse(s)
	if err != nil {
		return "", err
	}

	return prettify(expr, 0, maxLineLength), nil
}
