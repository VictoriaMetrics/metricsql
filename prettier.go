package metricsql

const (
	maxLineLength = 130
	defaultIndent = 0
)

func Prettier(s string, withDefaultMaxLine bool) (string, error) {
	expr, err := Parse(s)
	if err != nil {
		return "", err
	}

	if len(s) <= maxLineLength && withDefaultMaxLine {
		return string(expr.AppendString(nil)), nil
	}

	pr := prettier(expr, defaultIndent)
	return string(pr), nil
}

func prettier(expr Expr, ident int) []byte {
	paddings := padding(ident)
	var buf []byte

	switch e := expr.(type) {
	case *MetricExpr, *NumberExpr, *StringExpr:
		buf = append(buf, paddings...)
		buf = e.AppendString(buf)
	case *RollupExpr:
		var rollupDst []byte

		rollupDst = append(rollupDst, wrapParens(e.Expr, rollupDst, ident)...)

		if e.Window != nil || e.InheritStep || e.Step != nil {
			rollupDst = append(rollupDst, '[')
			rollupDst = append(rollupDst, e.Window.s...)

			if e.Step != nil {
				rollupDst = append(rollupDst, ':')
				rollupDst = append(rollupDst, e.Step.s...)
			} else if e.InheritStep {
				rollupDst = append(rollupDst, ':')
			}
			rollupDst = append(rollupDst, ']')
		}
		buf = append(buf, rollupDst...)
	case *BinaryOpExpr:
		var binaryDst []byte

		binaryDst = append(binaryDst, wrapParens(e.Left, binaryDst, ident+1)...)
		binaryDst = append(binaryDst, '\n')
		binaryDst = append(binaryDst, paddings...)
		binaryDst = append(binaryDst, e.Op...)
		if e.Bool {
			binaryDst = append(binaryDst, ' ')
			binaryDst = append(binaryDst, "bool"...)
		}
		if e.GroupModifier.Op != "" {
			binaryDst = append(binaryDst, ' ')
			binaryDst = append(binaryDst, e.GroupModifier.AppendString(nil)...)
		}
		if e.JoinModifier.Op != "" {
			binaryDst = append(binaryDst, ' ')
			binaryDst = append(binaryDst, e.JoinModifier.AppendString(nil)...)
		}
		binaryDst = append(binaryDst, '\n')
		binaryDst = append(binaryDst, wrapParens(e.Right, binaryDst, ident+1)...)

		buf = append(buf, binaryDst...)
	case *AggrFuncExpr:
		var b []byte

		b = append(b, paddings...)
		b = append(b, e.Name...)
		if e.Modifier.Op != "" {
			b = append(b, ' ')
			mod := e.Modifier.AppendString(nil)
			b = append(b, mod...)
		}
		// b = append(b, ' ')
		// b = append(b, '(')
		// b = append(b, '\n')
		for i, a := range e.Args {
			pt := prettier(a, ident+1)
			b = append(b, pt...)
			if i < len(e.Args)-1 {
				b = append(b, ',')
			}
			b = append(b, '\n')
		}
		b = append(b, paddings...)
		b = append(b, ')')

		buf = append(buf, b...)
	case *FuncExpr:
		if e.Name == "time" {
			buf = append(buf, paddings...)
			buf = append(buf, "time ()"...)
		} else {
			var b []byte

			b = append(b, paddings...)
			b = append(b, e.Name...)
			// b = append(b, ' ')
			// b = append(b, '(')
			// b = append(b, '\n')
			for i, a := range e.Args {
				pt := prettier(a, ident+1)
				b = append(b, pt...)
				if i < len(e.Args)-1 {
					b = append(b, ',')
				}
				b = append(b, '\n')
			}
			b = append(b, paddings...)
			b = append(b, ')')

			buf = append(buf, b...)
		}
	}

	return buf
}

func padding(indent int) []byte {
	var b []byte
	for i := 0; i < indent; i++ {
		b = append(b, "  "...)
	}
	return b
}

func needParens(expr Expr) bool {
	if _, ok := expr.(*RollupExpr); ok {
		return true
	}
	if _, ok := expr.(*BinaryOpExpr); ok {
		return true
	}
	if ae, ok := expr.(*AggrFuncExpr); ok && ae.Modifier.Op != "" {
		return true
	}
	return false
}

func wrapParens(expr Expr, by []byte, ident int) []byte {
	var b []byte
	paddings := padding(ident)
	if needParens(expr) {
		b = append(b, paddings...)
		b = append(b, '(')
		b = append(b, '\n')
		b = append(b, prettier(expr, ident+1)...)
		b = append(b, '\n')
		b = append(b, paddings...)
		b = append(b, ')')
		return append(by, b...)
	}

	by = append(by, prettier(expr, ident)...)
	return by
}

// func genPadding(ident int) string {
// 	return strings.Repeat("  ", ident)
// }
//
// func needParens(expr Expr) bool {
// 	switch e := expr.(type) {
// 	case *MetricExpr, *NumberExpr, *StringExpr:
// 		return false
// 	case *FuncExpr:
// 		return e.Name != "time"
// 	default:
// 		return true
// 	}
// }
//
// func wrapParensWhenNecesary(expr Expr, b *bytes.Buffer, ident int) {
// 	paddings := genPadding(ident)
// 	if needParens(expr) {
// 		b.WriteString(paddings + "(\n")
// 		b.Write(prettier(expr, ident+1))
// 		b.WriteString("\n" + paddings + ")")
// 	} else {
// 		b.Write(prettier(expr, ident))
// 	}
// }
//
// func prettier(expr Expr, ident int) []byte {
// 	paddings := genPadding(ident)
// 	var buf []byte
//
// 	switch e := expr.(type) {
// 	case *MetricExpr, *NumberExpr, *StringExpr:
// 		buf = append(buf, paddings...)
// 		buf = e.AppendString(buf)
// 	case *RollupExpr:
// 		var b bytes.Buffer
//
// 		wrapParensWhenNecesary(e.Expr, &b, ident)
// 		if e.Window != nil || e.InheritStep || e.Step != nil {
// 			b.WriteString("[")
// 			b.WriteString(e.Window.s)
// 			b.WriteString(":")
//
// 			if e.Step != nil {
// 				b.WriteString(e.Step.s)
// 			} else if e.InheritStep {
//
// 			}
// 			b.WriteString("]")
// 		}
//
// 		buf = append(buf, b.Bytes()...)
// 	case *BinaryOpExpr:
// 		var b bytes.Buffer
//
// 		wrapParensWhenNecesary(e.Left, &b, ident+1)
// 		b.WriteString(fmt.Sprintf("\n%s%s", paddings, e.Op))
// 		if e.Bool {
// 			b.WriteString(" bool")
// 		}
// 		if e.GroupModifier.Op != "" {
// 			b.WriteString(" ")
// 			b.Write(e.GroupModifier.AppendString(nil))
// 		}
// 		if e.JoinModifier.Op != "" {
// 			b.WriteString(" ")
// 			b.Write(e.JoinModifier.AppendString(nil))
// 		}
// 		b.WriteString("\n")
// 		wrapParensWhenNecesary(e.Right, &b, ident+1)
//
// 		buf = append(buf, b.Bytes()...)
// 	case *AggrFuncExpr:
// 		var b bytes.Buffer
//
// 		b.WriteString(paddings + e.Name)
// 		if e.Modifier.Op != "" {
// 			b.WriteString(" ")
// 			b.Write(e.Modifier.AppendString(nil))
// 		}
// 		b.WriteString(" (\n")
// 		for i, a := range e.Args {
// 			b.Write(prettier(a, ident+1))
// 			if i < len(e.Args)-1 {
// 				b.WriteString(",")
// 			}
// 			b.WriteString("\n")
// 		}
// 		b.WriteString(paddings + ")")
//
// 		buf = append(buf, b.Bytes()...)
// 	case *FuncExpr:
// 		if e.Name == "time" {
// 			buf = append(buf, []byte(paddings+"time ()")...)
// 		} else {
// 			var b bytes.Buffer
//
// 			b.WriteString(paddings + e.Name + " (\n")
// 			for i, a := range e.Args {
// 				b.Write(prettier(a, ident+1))
// 				if i < len(e.Args)-1 {
// 					b.WriteString(",")
// 				}
// 				b.WriteString("\n")
// 			}
// 			b.WriteString(paddings + ")")
//
// 			buf = append(buf, b.Bytes()...)
// 		}
// 	}
//
// 	return buf
// }
//
// func Prettier(s string) (string, error) {
// 	expr, err := Parse(s)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return string(prettier(expr, 0)), nil
// }
