package generator

import (
	"math"
	"strconv"
	"unsafe"

	"github.com/t14raptor/go-fast/ast"
)

// Options controls code generation behavior.
type Options struct {
	// Minified disables pretty printing: no newlines, no indentation, and
	// line (`//`) comments are omitted to keep the output on a single line.
	Minified bool
}

// Generate renders node as JavaScript source using the default (pretty) options.
func Generate(node ast.VisitableNode) string {
	return GenerateWithOptions(node, Options{})
}

// GenerateMinified renders node without pretty printing (no newlines or
// indentation). It is equivalent to GenerateWithOptions(node, Options{Minified: true}).
func GenerateMinified(node ast.VisitableNode) string {
	return GenerateWithOptions(node, Options{Minified: true})
}

// GenerateWithOptions renders node as JavaScript source using the supplied options.
func GenerateWithOptions(node ast.VisitableNode, opts Options) string {
	g := &GenVisitor{opts: opts}
	g.V = g
	g.gen(node)
	return unsafe.String(unsafe.SliceData(g.buf), len(g.buf))
}

type GenVisitor struct {
	ast.NoopVisitor

	buf  []byte
	opts Options

	indent int

	// Precedence/context state threaded through expression generation.
	// Set by genExpr before calling VisitWith; read by each expression visitor.
	prec ast.Precedence
	ctx  context
}

func (g *GenVisitor) writeByte(c byte) {
	g.buf = append(g.buf, c)
}

func (g *GenVisitor) writeString(s string) {
	g.buf = append(g.buf, s...)
}

func (g *GenVisitor) gen(node ast.VisitableNode) {
	node.VisitWith(g)
}

// genExpr sets the minimum precedence and context, then visits expr. Each
// expression visitor reads g.prec/g.ctx to decide whether to wrap in parens,
// and calls genExpr on children with the appropriate child precedence.
func (g *GenVisitor) genExpr(expr ast.Expr, prec ast.Precedence, ctx context) {
	savedPrec, savedCtx := g.prec, g.ctx
	g.prec, g.ctx = prec, ctx
	expr.VisitWith(g)
	g.prec, g.ctx = savedPrec, savedCtx
}

func (g *GenVisitor) line() {
	if g.opts.Minified {
		return
	}
	g.writeByte('\n')
}

func (g *GenVisitor) lineAndPad() {
	if g.opts.Minified {
		return
	}
	g.writeByte('\n')
	for range g.indent {
		g.writeByte('\t')
	}
}

func (g *GenVisitor) space() {
	if g.opts.Minified {
		return
	}
	g.writeByte(' ')
}

func (g *GenVisitor) VisitBinaryExpression(n *ast.BinaryExpression) {
	g.genBinaryExpr(n, g.prec, g.ctx)
}

func (g *GenVisitor) VisitLogicalExpression(n *ast.LogicalExpression) {
	g.genBinaryExpr(n, g.prec, g.ctx)
}

func (g *GenVisitor) VisitAssignExpression(n *ast.AssignExpression) {
	ctx := g.ctx
	wrap := g.prec > ast.PrecedenceAssign
	if wrap {
		g.writeByte('(')
	}

	g.genExpr(n.Left.Expr, ast.PrecedenceAssign, ctx)
	g.space()
	g.writeString(n.Operator.String())
	g.space()
	// Propagate ctxForbidIn into the RHS: `for (x = a in b; ...)` is
	// ambiguous with for-in and must render as `for (x = (a in b); ...)`.
	g.genExpr(n.Right.Expr, ast.PrecedenceAssign, ctx&ctxForbidIn)

	if wrap {
		g.writeByte(')')
	}
}

func (g *GenVisitor) VisitConditionalExpression(n *ast.ConditionalExpression) {
	ctx := g.ctx
	wrap := g.prec > ast.PrecedenceConditional
	if wrap {
		g.writeByte('(')
	}

	g.genExpr(n.Test.Expr, ast.PrecedenceConditional+1, ctx&ctxForbidIn)
	g.space()
	g.writeByte('?')
	g.space()
	g.genExpr(n.Consequent.Expr, ast.PrecedenceAssign, 0)
	g.space()
	g.writeByte(':')
	g.space()
	g.genExpr(n.Alternate.Expr, ast.PrecedenceAssign, ctx&ctxForbidIn)

	if wrap {
		g.writeByte(')')
	}
}

func (g *GenVisitor) VisitUnaryExpression(n *ast.UnaryExpression) {
	wrap := g.prec > ast.PrecedencePrefix
	if wrap {
		g.writeByte('(')
	}

	g.writeString(n.Operator.String())
	if n.Operator.IsKeyword() {
		g.writeByte(' ')
	}
	g.genExpr(n.Operand.Expr, ast.PrecedencePrefix, 0)

	if wrap {
		g.writeByte(')')
	}
}

func (g *GenVisitor) VisitUpdateExpression(n *ast.UpdateExpression) {
	if n.Postfix {
		wrap := g.prec > ast.PrecedencePostfix
		if wrap {
			g.writeByte('(')
		}

		g.genExpr(n.Operand.Expr, ast.PrecedencePostfix, 0)
		g.writeString(n.Operator.String())

		if wrap {
			g.writeByte(')')
		}
	} else {
		wrap := g.prec > ast.PrecedencePrefix
		if wrap {
			g.writeByte('(')
		}

		g.writeString(n.Operator.String())
		g.genExpr(n.Operand.Expr, ast.PrecedencePrefix, 0)

		if wrap {
			g.writeByte(')')
		}
	}
}

func (g *GenVisitor) VisitSequenceExpression(n *ast.SequenceExpression) {
	ctx := g.ctx
	wrap := g.prec > ast.PrecedenceComma
	if wrap {
		g.writeByte('(')
		// Wrapping this sequence in parens means any inner `in` is no
		// longer at for-init top-level, so the forbid-in guard doesn't
		// need to carry further.
		ctx &^= ctxForbidIn
	}

	for i, e := range n.Sequence {
		g.genExpr(e.Expr, ast.PrecedenceAssign, ctx&ctxForbidIn)
		if i < len(n.Sequence)-1 {
			g.writeByte(',')
			g.space()
		}
	}

	if wrap {
		g.writeByte(')')
	}
}

func (g *GenVisitor) VisitYieldExpression(n *ast.YieldExpression) {
	wrap := g.prec > ast.PrecedenceYield
	if wrap {
		g.writeByte('(')
	}

	g.writeString("yield")
	if n.Delegate {
		g.writeByte('*')
	}
	if n.Argument != nil {
		g.writeByte(' ')
		g.genExpr(n.Argument.Expr, ast.PrecedenceAssign, 0)
	}

	if wrap {
		g.writeByte(')')
	}
}

func (g *GenVisitor) VisitAwaitExpression(n *ast.AwaitExpression) {
	wrap := g.prec > ast.PrecedencePrefix
	if wrap {
		g.writeByte('(')
	}

	g.writeString("await ")
	g.genExpr(n.Argument.Expr, ast.PrecedencePrefix, 0)

	if wrap {
		g.writeByte(')')
	}
}

func (g *GenVisitor) VisitSpreadElement(n *ast.SpreadElement) {
	g.writeString("...")
	g.genExpr(n.Expression.Expr, ast.PrecedenceAssign, 0)
}

func (g *GenVisitor) VisitCallExpression(n *ast.CallExpression) {
	wrap := g.ctx&ctxForbidCall != 0
	if wrap {
		g.writeByte('(')
	}

	switch n.Callee.Expr.(type) {
	case *ast.FunctionLiteral, *ast.ArrowFunctionLiteral:
		g.writeByte('(')
		g.genExpr(n.Callee.Expr, ast.PrecedenceLowest, 0)
		g.writeByte(')')
	default:
		g.genExpr(n.Callee.Expr, ast.PrecedenceCall, 0)
	}
	g.writeByte('(')
	for i, a := range n.ArgumentList {
		g.genExpr(a.Expr, ast.PrecedenceAssign, 0)
		if i < len(n.ArgumentList)-1 {
			g.writeByte(',')
			g.space()
		}
	}
	g.writeByte(')')

	if wrap {
		g.writeByte(')')
	}
}

func (g *GenVisitor) VisitNewExpression(n *ast.NewExpression) {
	g.writeString("new ")
	g.genExpr(n.Callee.Expr, ast.PrecedenceNew, ctxForbidCall)
	g.writeByte('(')
	for i, a := range n.ArgumentList {
		g.genExpr(a.Expr, ast.PrecedenceAssign, 0)
		if i < len(n.ArgumentList)-1 {
			g.writeByte(',')
			g.space()
		}
	}
	g.writeByte(')')
}

func (g *GenVisitor) VisitMemberExpression(n *ast.MemberExpression) {
	switch n.Object.Expr.(type) {
	case *ast.NumberLiteral:
		g.writeByte('(')
		g.genExpr(n.Object.Expr, ast.PrecedenceLowest, 0)
		g.writeByte(')')
	default:
		g.genExpr(n.Object.Expr, ast.PrecedenceMember, 0)
	}
	g.gen(n.Property)
}

func (g *GenVisitor) VisitPrivateDotExpression(n *ast.PrivateDotExpression) {
	g.genExpr(n.Left.Expr, ast.PrecedenceMember, 0)
	g.writeString(".#")
	g.writeString(n.Identifier.Identifier.Name)
}

func (g *GenVisitor) VisitOptionalChain(n *ast.OptionalChain) {
	g.genExpr(n.Base.Expr, ast.PrecedenceCall, 0)
}

func (g *GenVisitor) VisitOptional(n *ast.Optional) {
	g.genExpr(n.Expr.Expr, ast.PrecedenceCall, 0)
}

func (g *GenVisitor) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {
	wrap := g.prec > ast.PrecedenceAssign
	if wrap {
		g.writeByte('(')
	}

	if n.Async {
		g.writeString("async ")
	}
	g.gen(n.ParameterList)
	g.space()
	g.writeString("=>")
	g.space()
	switch body := n.Body.Body.(type) {
	case *ast.BlockStatement:
		g.gen(body)
	case *ast.Expression:
		if _, ok := body.Expr.(*ast.ObjectLiteral); ok {
			g.writeByte('(')
			g.genExpr(body.Expr, ast.PrecedenceLowest, 0)
			g.writeByte(')')
		} else {
			g.genExpr(body.Expr, ast.PrecedenceAssign, 0)
		}
	}

	if wrap {
		g.writeByte(')')
	}
}

func (g *GenVisitor) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	if n.Async {
		g.writeString("async ")
	}
	if n.Name != nil {
		g.writeString("function ")
		g.gen(n.Name)
	} else {
		g.writeString("function")
	}
	g.gen(n.ParameterList)
	g.space()
	g.gen(n.Body)
}

func (g *GenVisitor) VisitClassLiteral(n *ast.ClassLiteral) {
	g.writeString("class")
	if n.Name != nil {
		g.writeByte(' ')
		g.gen(n.Name)
	}
	g.space()
	g.writeByte('{')

	g.indent++
	for _, element := range n.Body {
		g.lineAndPad()
		switch e := element.Element.(type) {
		case *ast.MethodDefinition:
			if e.Static {
				g.writeString("static ")
			}
			if e.Kind == ast.PropertyKindGet {
				g.writeString("get ")
			} else if e.Kind == ast.PropertyKindSet {
				g.writeString("set ")
			}
			if e.Computed {
				g.writeByte('[')
				g.gen(e.Key)
				g.writeByte(']')
			} else {
				g.gen(e.Key)
			}
			g.gen(e.Body.ParameterList)
			g.space()
			g.gen(e.Body.Body)
		}
	}
	g.indent--

	g.lineAndPad()
	g.writeByte('}')
}

func (g *GenVisitor) VisitIdentifier(n *ast.Identifier) {
	if n != nil {
		g.writeString(n.Name)
	}
}

func (g *GenVisitor) VisitPrivateIdentifier(n *ast.PrivateIdentifier) {
	g.writeByte('#')
	g.writeString(n.Identifier.Name)
}

func (g *GenVisitor) VisitThisExpression(n *ast.ThisExpression) {
	g.writeString("this")
}

func (g *GenVisitor) VisitNullLiteral(n *ast.NullLiteral) {
	g.writeString("null")
}

func (g *GenVisitor) VisitBooleanLiteral(n *ast.BooleanLiteral) {
	if n.Value {
		g.writeString("true")
	} else {
		g.writeString("false")
	}
}

func (g *GenVisitor) VisitNumberLiteral(n *ast.NumberLiteral) {
	if n.Raw != nil {
		g.writeString(*n.Raw)
	} else if math.IsInf(n.Value, 1) {
		g.writeString("Infinity")
	} else if math.IsInf(n.Value, -1) {
		wrap := g.prec > ast.PrecedencePrefix
		if wrap {
			g.writeByte('(')
		}
		g.writeString("-Infinity")
		if wrap {
			g.writeByte(')')
		}
	} else {
		g.writeString(strconv.FormatFloat(n.Value, 'f', -1, 64))
	}
}

func (g *GenVisitor) VisitBigIntLiteral(n *ast.BigIntLiteral) {
	if n.Raw != nil {
		g.writeString(*n.Raw)
		return
	}
	if n.Value != nil {
		g.writeString(n.Value.String())
	} else {
		g.writeByte('0')
	}
	g.writeByte('n')
}

func (g *GenVisitor) VisitStringLiteral(n *ast.StringLiteral) {
	if n.Raw != nil {
		g.writeString(*n.Raw)
	} else {
		g.writeString(strconv.Quote(n.Value))
	}
}

func (g *GenVisitor) VisitRegExpLiteral(n *ast.RegExpLiteral) {
	g.writeString(n.Literal)
}

func (g *GenVisitor) VisitTemplateLiteral(n *ast.TemplateLiteral) {
	g.writeByte('`')
	for i, e := range n.Elements {
		g.writeString(e.Parsed)
		if i < len(n.Expressions) {
			g.writeString("${")
			g.genExpr(n.Expressions[i].Expr, ast.PrecedenceLowest, 0)
			g.writeByte('}')
		}
	}
	g.writeByte('`')
}

func (g *GenVisitor) VisitArrayLiteral(n *ast.ArrayLiteral) {
	g.writeByte('[')
	for i, ex := range n.Value {
		if ex.Expr != nil {
			g.genExpr(ex.Expr, ast.PrecedenceAssign, 0)
		}
		if i < len(n.Value)-1 {
			g.writeByte(',')
			g.space()
		}
	}
	g.writeByte(']')
}

func (g *GenVisitor) VisitObjectLiteral(n *ast.ObjectLiteral) {
	g.writeByte('{')

	g.indent++
	for i, p := range n.Value {
		g.lineAndPad()
		g.gen(p.Prop)
		if i < len(n.Value)-1 {
			g.writeByte(',')
		}
	}
	g.indent--

	if len(n.Value) > 0 {
		g.lineAndPad()
	}
	g.writeByte('}')
}

func (g *GenVisitor) VisitArrayPattern(n *ast.ArrayPattern) {
	g.writeByte('[')
	for i, elem := range n.Elements {
		if elem.Expr != nil {
			g.genExpr(elem.Expr, ast.PrecedenceAssign, 0)
		}
		if i < len(n.Elements)-1 {
			g.writeByte(',')
			g.space()
		}
	}
	g.writeByte(']')
}

func (g *GenVisitor) VisitObjectPattern(n *ast.ObjectPattern) {
	g.writeByte('{')
	for i, prop := range n.Properties {
		g.gen(prop.Prop)
		if i < len(n.Properties)-1 {
			g.writeByte(',')
			g.space()
		}
	}
	if n.Rest != nil {
		if len(n.Properties) > 0 {
			g.writeByte(',')
			g.space()
		}
		g.writeString("...")
		g.gen(n.Rest)
	}
	g.writeByte('}')
}

func (g *GenVisitor) VisitMetaProperty(n *ast.MetaProperty) {
	g.gen(n.Meta)
	g.writeByte('.')
	g.gen(n.Property)
}

func (g *GenVisitor) VisitBindingTarget(n *ast.BindingTarget) {
	g.genExpr(n.Target, ast.PrecedenceLowest, 0)
}

func (g *GenVisitor) VisitExpression(n *ast.Expression) {
	if n.Expr != nil {
		g.genExpr(n.Expr, ast.PrecedenceLowest, 0)
	}
}

func (g *GenVisitor) VisitProgram(n *ast.Program) {
	for _, b := range n.Body {
		g.gen(b.Stmt)
		g.line()
	}
}

func (g *GenVisitor) VisitStatements(n *ast.Statements) {
	for _, st := range *n {
		g.lineAndPad()
		g.gen(st.Stmt)
	}
}

func (g *GenVisitor) VisitBlockStatement(n *ast.BlockStatement) {
	g.writeByte('{')

	g.indent++
	g.VisitStatements(&n.List)
	g.indent--

	if len(n.List) > 0 {
		g.lineAndPad()
	}
	g.writeByte('}')
}

func (g *GenVisitor) VisitExpressionStatement(n *ast.ExpressionStatement) {
	switch e := n.Expression.Expr.(type) {
	case *ast.ObjectLiteral, *ast.FunctionLiteral, *ast.ClassLiteral:
		g.writeByte('(')
		g.genExpr(n.Expression.Expr, ast.PrecedenceLowest, 0)
		g.writeByte(')')
	case *ast.AssignExpression:
		switch e.Left.Expr.(type) {
		case *ast.ObjectPattern, *ast.ArrayPattern:
			g.writeByte('(')
			g.genExpr(n.Expression.Expr, ast.PrecedenceLowest, 0)
			g.writeByte(')')
		default:
			g.genExpr(n.Expression.Expr, ast.PrecedenceLowest, 0)
		}
	default:
		g.genExpr(n.Expression.Expr, ast.PrecedenceLowest, 0)
	}
	g.writeByte(';')
	if len(n.Comment) > 0 && !g.opts.Minified {
		g.writeString(" // " + n.Comment)
	}
}

func (g *GenVisitor) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	g.writeString(n.Token.String())
	g.writeByte(' ')
	for i, b := range n.List {
		g.gen(&b)
		if i < len(n.List)-1 {
			g.writeByte(',')
			g.space()
		}
	}
	g.writeByte(';')
	if len(n.Comment) > 0 && !g.opts.Minified {
		g.writeString(" // " + n.Comment)
	}
}

func (g *GenVisitor) VisitVariableDeclarator(n *ast.VariableDeclarator) {
	g.gen(n.Target)
	if n.Initializer != nil {
		g.space()
		g.writeByte('=')
		g.space()
		g.genExpr(n.Initializer.Expr, ast.PrecedenceAssign, 0)
	}
}

func (g *GenVisitor) VisitReturnStatement(n *ast.ReturnStatement) {
	g.writeString("return")
	if n.Argument != nil {
		g.writeByte(' ')
		g.genExpr(n.Argument.Expr, ast.PrecedenceAssign, 0)
	}
	g.writeByte(';')
}

func (g *GenVisitor) VisitThrowStatement(n *ast.ThrowStatement) {
	g.writeString("throw ")
	g.genExpr(n.Argument.Expr, ast.PrecedenceAssign, 0)
	g.writeByte(';')
}

func (g *GenVisitor) VisitIfStatement(n *ast.IfStatement) {
	g.writeString("if")
	g.space()
	g.writeByte('(')
	g.genExpr(n.Test.Expr, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()

	switch n.Consequent.Stmt.(type) {
	case *ast.EmptyStatement, *ast.BlockStatement:
		g.gen(n.Consequent.Stmt)
	default:
		g.indent++
		g.gen(n.Consequent.Stmt)
		g.indent--
		g.lineAndPad()
	}

	if n.Alternate != nil {
		g.writeString(" else ")

		switch n.Alternate.Stmt.(type) {
		case *ast.EmptyStatement, *ast.BlockStatement, *ast.IfStatement:
			g.gen(n.Alternate.Stmt)
		default:
			g.indent++
			g.gen(n.Alternate.Stmt)
			g.indent--
			g.lineAndPad()
		}
	}
}

func (g *GenVisitor) VisitForStatement(n *ast.ForStatement) {
	g.writeString("for")
	g.space()
	g.writeByte('(')
	if n.Initializer != nil {
		g.gen(n.Initializer)
	} else {
		g.writeByte(';')
	}
	g.space()

	if n.Test.Expr != nil {
		g.genExpr(n.Test.Expr, ast.PrecedenceLowest, 0)
	}
	g.writeByte(';')
	g.space()
	if n.Update.Expr != nil {
		g.genExpr(n.Update.Expr, ast.PrecedenceLowest, 0)
	}
	g.writeByte(')')
	g.space()

	switch n.Body.Stmt.(type) {
	case *ast.EmptyStatement, *ast.BlockStatement:
		g.gen(n.Body.Stmt)
	default:
		g.indent++
		g.gen(n.Body.Stmt)
		g.indent--
		g.lineAndPad()
	}
}

func (g *GenVisitor) VisitForLoopInitializer(n *ast.ForLoopInitializer) {
	switch init := n.Initializer.(type) {
	case *ast.Expression:
		g.genExpr(init.Expr, ast.PrecedenceLowest, ctxForbidIn)
		g.writeByte(';')
	case *ast.VariableDeclaration:
		g.gen(init)
	}
}

func (g *GenVisitor) VisitForInStatement(n *ast.ForInStatement) {
	g.writeString("for")
	g.space()
	g.writeByte('(')
	g.gen(n.Into)
	g.writeString(" in ")
	g.genExpr(n.Source.Expr, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()
	g.gen(n.Body.Stmt)
}

func (g *GenVisitor) VisitForOfStatement(n *ast.ForOfStatement) {
	g.writeString("for")
	if n.Await {
		g.writeString(" await")
	}
	g.space()
	g.writeByte('(')
	g.gen(n.Into)
	g.writeString(" of ")
	g.genExpr(n.Source.Expr, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()
	g.gen(n.Body.Stmt)
}

func (g *GenVisitor) VisitForInto(n *ast.ForInto) {
	switch into := n.Into.(type) {
	case *ast.VariableDeclaration:
		g.writeString(into.Token.String())
		g.writeByte(' ')
		g.gen(&into.List)
	case *ast.Expression:
		g.gen(into)
	}
}

func (g *GenVisitor) VisitDoWhileStatement(n *ast.DoWhileStatement) {
	g.writeString("do ")
	g.gen(n.Body.Stmt)
	g.writeString(" while(")
	g.genExpr(n.Test.Expr, ast.PrecedenceLowest, 0)
	g.writeString(");")
}

func (g *GenVisitor) VisitWhileStatement(n *ast.WhileStatement) {
	g.writeString("while")
	g.space()
	g.writeByte('(')
	g.genExpr(n.Test.Expr, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()
	g.gen(n.Body.Stmt)
}

func (g *GenVisitor) VisitSwitchStatement(n *ast.SwitchStatement) {
	g.writeString("switch")
	g.space()
	g.writeByte('(')
	g.genExpr(n.Discriminant.Expr, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()
	g.writeByte('{')

	g.indent++
	for _, c := range n.Body {
		g.lineAndPad()
		g.gen(&c)
	}
	g.indent--

	if len(n.Body) > 0 {
		g.lineAndPad()
	}
	g.writeByte('}')
}

func (g *GenVisitor) VisitCaseStatement(n *ast.CaseStatement) {
	if n.Test != nil {
		g.writeString("case ")
		g.genExpr(n.Test.Expr, ast.PrecedenceLowest, 0)
		g.writeByte(':')
	} else {
		g.writeString("default:")
	}
	g.indent++
	for i := range n.Consequent {
		g.lineAndPad()
		g.gen(n.Consequent[i].Stmt)
	}
	g.indent--
}

func (g *GenVisitor) VisitTryStatement(n *ast.TryStatement) {
	g.writeString("try")
	g.space()

	g.gen(n.Body)

	if n.Catch != nil {
		g.space()
		g.writeString("catch")
		g.space()
		if n.Catch.Parameter != nil {
			g.writeByte('(')
			g.gen(n.Catch.Parameter)
			g.writeByte(')')
			g.space()
		}
		g.gen(n.Catch.Body)
	}
	if n.Finally != nil {
		g.space()
		g.writeString("finally")
		g.space()
		g.gen(n.Finally)
	}
}

func (g *GenVisitor) VisitCatchStatement(n *ast.CatchStatement) {
	if n.Parameter != nil {
		g.gen(n.Parameter)
	}
	g.gen(n.Body)
}

func (g *GenVisitor) VisitBreakStatement(n *ast.BreakStatement) {
	g.writeString("break")
	if n.Label != nil {
		g.writeByte(' ')
		g.gen(n.Label)
	}
	g.writeByte(';')
}

func (g *GenVisitor) VisitContinueStatement(n *ast.ContinueStatement) {
	g.writeString("continue")
	if n.Label != nil {
		g.writeByte(' ')
		g.gen(n.Label)
	}
	g.writeByte(';')
}

func (g *GenVisitor) VisitLabelledStatement(n *ast.LabelledStatement) {
	g.gen(n.Label)
	g.writeByte(':')
	g.space()
	g.gen(n.Statement.Stmt)
}

func (g *GenVisitor) VisitWithStatement(n *ast.WithStatement) {
	g.writeString("with")
	g.space()
	g.writeByte('(')
	g.genExpr(n.Object.Expr, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()
	g.gen(n.Body.Stmt)
}

func (g *GenVisitor) VisitDebuggerStatement(n *ast.DebuggerStatement) {
	g.writeString("debugger;")
}

func (g *GenVisitor) VisitEmptyStatement(n *ast.EmptyStatement) {
	g.writeByte(';')
}

func (g *GenVisitor) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	g.lineAndPad()
	g.VisitFunctionLiteral(n.Function)
}

func (g *GenVisitor) VisitClassDeclaration(n *ast.ClassDeclaration) {
	g.VisitClassLiteral(n.Class)
}

func (g *GenVisitor) VisitParameterList(n *ast.ParameterList) {
	g.writeByte('(')
	for i, p := range n.List {
		g.gen(&p)
		if i < len(n.List)-1 {
			g.writeByte(',')
			g.space()
		}
	}

	if n.Rest != nil {
		g.writeString("...")
		g.gen(n.Rest)
	}
	g.writeByte(')')
}

func (g *GenVisitor) VisitMemberProperty(n *ast.MemberProperty) {
	switch prop := n.Prop.(type) {
	case *ast.Identifier:
		g.writeByte('.')
		g.gen(prop)
	case *ast.ComputedProperty:
		g.writeByte('[')
		g.genExpr(prop.Expr.Expr, ast.PrecedenceLowest, 0)
		g.writeByte(']')
	}
}

func (g *GenVisitor) VisitPropertyKeyed(n *ast.PropertyKeyed) {
	if n.Kind == ast.PropertyKindGet || n.Kind == ast.PropertyKindSet {
		g.writeString(string(n.Kind))
		g.writeByte(' ')
		if n.Computed {
			g.writeByte('[')
			g.genExpr(n.Key.Expr, ast.PrecedenceLowest, 0)
			g.writeByte(']')
		} else {
			g.genExpr(n.Key.Expr, ast.PrecedenceLowest, 0)
		}
		f := n.Value.Expr.(*ast.FunctionLiteral)
		g.gen(f.ParameterList)
		g.space()
		g.gen(f.Body)
		return
	}
	if n.Computed {
		g.writeByte('[')
		g.genExpr(n.Key.Expr, ast.PrecedenceLowest, 0)
		g.writeByte(']')
	} else {
		g.genExpr(n.Key.Expr, ast.PrecedenceLowest, 0)
	}
	g.writeByte(':')
	g.space()
	g.genExpr(n.Value.Expr, ast.PrecedenceAssign, 0)
}
