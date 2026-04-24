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

	binaryStack []binaryExprEntry
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
func (g *GenVisitor) genExpr(expr *ast.Expression, prec ast.Precedence, ctx context) {
	savedPrec, savedCtx := g.prec, g.ctx
	g.prec, g.ctx = prec, ctx
	expr.VisitChildrenWith(g)
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

func (g *GenVisitor) VisitAssignExpression(n *ast.AssignExpression) {
	ctx := g.ctx
	wrap := g.prec > ast.PrecedenceAssign
	if wrap {
		g.writeByte('(')
		ctx &^= ctxForbidIn
	}

	g.genExpr(n.Left, ast.PrecedenceAssign, ctx)
	g.space()
	g.writeString(n.Operator.String())
	g.space()
	g.genExpr(n.Right, ast.PrecedenceAssign, ctx&ctxForbidIn)

	if wrap {
		g.writeByte(')')
	}
}

func (g *GenVisitor) VisitConditionalExpression(n *ast.ConditionalExpression) {
	ctx := g.ctx
	wrap := g.prec > ast.PrecedenceConditional
	if wrap {
		g.writeByte('(')
		ctx &^= ctxForbidIn
	}

	g.genExpr(n.Test, ast.PrecedenceConditional+1, ctx&ctxForbidIn)
	g.space()
	g.writeByte('?')
	g.space()
	g.genExpr(n.Consequent, ast.PrecedenceAssign, 0)
	g.space()
	g.writeByte(':')
	g.space()
	g.genExpr(n.Alternate, ast.PrecedenceAssign, ctx&ctxForbidIn)

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
	g.genExpr(n.Operand, ast.PrecedencePrefix, 0)

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

		g.genExpr(n.Operand, ast.PrecedencePostfix, 0)
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
		g.genExpr(n.Operand, ast.PrecedencePrefix, 0)

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
		ctx &^= ctxForbidIn
	}

	for i := range n.Sequence {
		g.genExpr(&n.Sequence[i], ast.PrecedenceAssign, ctx&ctxForbidIn)
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
		g.genExpr(n.Argument, ast.PrecedenceAssign, 0)
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
	g.genExpr(n.Argument, ast.PrecedencePrefix, 0)

	if wrap {
		g.writeByte(')')
	}
}

func (g *GenVisitor) VisitSpreadElement(n *ast.SpreadElement) {
	g.writeString("...")
	g.genExpr(n.Expression, ast.PrecedenceAssign, 0)
}

func (g *GenVisitor) VisitCallExpression(n *ast.CallExpression) {
	wrap := g.ctx&ctxForbidCall != 0
	if wrap {
		g.writeByte('(')
	}

	switch n.Callee.Kind() {
	case ast.ExprFuncLit, ast.ExprArrowFuncLit:
		g.writeByte('(')
		g.genExpr(n.Callee, ast.PrecedenceLowest, 0)
		g.writeByte(')')
	default:
		g.genExpr(n.Callee, ast.PrecedenceCall, 0)
	}
	g.writeByte('(')
	for i := range n.ArgumentList {
		g.genExpr(&n.ArgumentList[i], ast.PrecedenceAssign, 0)
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
	g.genExpr(n.Callee, ast.PrecedenceNew, ctxForbidCall)
	g.writeByte('(')
	for i := range n.ArgumentList {
		g.genExpr(&n.ArgumentList[i], ast.PrecedenceAssign, 0)
		if i < len(n.ArgumentList)-1 {
			g.writeByte(',')
			g.space()
		}
	}
	g.writeByte(')')
}

func (g *GenVisitor) VisitMemberExpression(n *ast.MemberExpression) {
	switch n.Object.Kind() {
	case ast.ExprNumLit:
		g.writeByte('(')
		g.genExpr(n.Object, ast.PrecedenceLowest, 0)
		g.writeByte(')')
	default:
		g.genExpr(n.Object, ast.PrecedenceMember, 0)
	}
	g.gen(n.Property)
}

func (g *GenVisitor) VisitPrivateDotExpression(n *ast.PrivateDotExpression) {
	g.genExpr(n.Left, ast.PrecedenceMember, 0)
	g.writeString(".#")
	g.writeString(n.Identifier.Identifier.Name)
}

func (g *GenVisitor) VisitOptionalChain(n *ast.OptionalChain) {
	g.genExpr(n.Base, ast.PrecedenceCall, 0)
}

func (g *GenVisitor) VisitOptional(n *ast.Optional) {
	g.genExpr(n.Expr, ast.PrecedenceCall, 0)
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
	switch n.Body.Kind() {
	case ast.ConciseBodyBlock:
		g.gen(n.Body)
	case ast.ConciseBodyExpr:
		body := n.Body.MustExpr()
		if body.IsObjLit() {
			g.writeByte('(')
			g.genExpr(body, ast.PrecedenceLowest, 0)
			g.writeByte(')')
		} else {
			g.genExpr(body, ast.PrecedenceAssign, 0)
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
		switch element.Kind() {
		case ast.ClassElemMethod:
			e := element.MustMethod()
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
			g.genExpr(&n.Expressions[i], ast.PrecedenceLowest, 0)
			g.writeByte('}')
		}
	}
	g.writeByte('`')
}

func (g *GenVisitor) VisitArrayLiteral(n *ast.ArrayLiteral) {
	g.writeByte('[')
	for i, ex := range n.Value {
		if !ex.IsNone() {
			g.genExpr(&n.Value[i], ast.PrecedenceAssign, 0)
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
	for i := range n.Value {
		g.lineAndPad()
		g.gen(&n.Value[i])
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
	for i := range n.Elements {
		elem := &n.Elements[i]
		if !elem.IsNone() {
			g.genExpr(elem, ast.PrecedenceAssign, 0)
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
	for i := range n.Properties {
		g.gen(&n.Properties[i])
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
	expr := ast.ExpressionFromBindingTarget(n)
	g.genExpr(&expr, ast.PrecedenceLowest, 0)
}

func (g *GenVisitor) VisitExpression(n *ast.Expression) {
	switch n.Kind() {
	case ast.ExprBinary:
		g.genBinaryExpr(n, g.prec, g.ctx)
	case ast.ExprLogical:
		g.genBinaryExpr(n, g.prec, g.ctx)
	default:
		g.genExpr(n, ast.PrecedenceLowest, 0)
	}
}

func (g *GenVisitor) VisitProgram(n *ast.Program) {
	for i := range n.Body {
		g.gen(&n.Body[i])
		g.line()
	}
}

func (g *GenVisitor) VisitStatements(n *ast.Statements) {
	for i := range *n {
		g.lineAndPad()
		g.gen(&(*n)[i])
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
	switch n.Expression.Kind() {
	case ast.ExprObjLit, ast.ExprFuncLit, ast.ExprClassLit:
		g.writeByte('(')
		g.genExpr(n.Expression, ast.PrecedenceLowest, 0)
		g.writeByte(')')
	case ast.ExprAssign:
		switch n.Expression.MustAssign().Left.Kind() {
		case ast.ExprObjPat, ast.ExprArrPat:
			g.writeByte('(')
			g.genExpr(n.Expression, ast.PrecedenceLowest, 0)
			g.writeByte(')')
		default:
			g.genExpr(n.Expression, ast.PrecedenceLowest, 0)
		}
	default:
		g.genExpr(n.Expression, ast.PrecedenceLowest, 0)
	}
	g.writeByte(';')
	if len(n.Comment) > 0 && !g.opts.Minified {
		g.writeString(" // " + n.Comment)
	}
}

func (g *GenVisitor) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	g.writeString(n.Token.String())
	g.writeByte(' ')
	for i := range n.List {
		g.gen(&n.List[i])
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
		g.genExpr(n.Initializer, ast.PrecedenceAssign, 0)
	}
}

func (g *GenVisitor) VisitReturnStatement(n *ast.ReturnStatement) {
	g.writeString("return")
	if n.Argument != nil {
		g.writeByte(' ')
		g.genExpr(n.Argument, ast.PrecedenceAssign, 0)
	}
	g.writeByte(';')
}

func (g *GenVisitor) VisitThrowStatement(n *ast.ThrowStatement) {
	g.writeString("throw ")
	g.genExpr(n.Argument, ast.PrecedenceAssign, 0)
	g.writeByte(';')
}

func (g *GenVisitor) VisitIfStatement(n *ast.IfStatement) {
	g.writeString("if")
	g.space()
	g.writeByte('(')
	g.genExpr(n.Test, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()

	switch n.Consequent.Kind() {
	case ast.StmtEmpty, ast.StmtBlock:
		g.gen(n.Consequent)
	default:
		g.indent++
		g.gen(n.Consequent)
		g.indent--
		g.lineAndPad()
	}

	if n.Alternate != nil {
		g.writeString(" else ")

		switch n.Alternate.Kind() {
		case ast.StmtEmpty, ast.StmtBlock, ast.StmtIf:
			g.gen(n.Alternate)
		default:
			g.indent++
			g.gen(n.Alternate)
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

	if n.Test != nil {
		g.genExpr(n.Test, ast.PrecedenceLowest, 0)
	}
	g.writeByte(';')
	g.space()
	if n.Update != nil {
		g.genExpr(n.Update, ast.PrecedenceLowest, 0)
	}
	g.writeByte(')')
	g.space()

	switch n.Body.Kind() {
	case ast.StmtEmpty, ast.StmtBlock:
		g.gen(n.Body)
	default:
		g.indent++
		g.gen(n.Body)
		g.indent--
		g.lineAndPad()
	}
}

func (g *GenVisitor) VisitForLoopInitializer(n *ast.ForLoopInitializer) {
	switch n.Kind() {
	case ast.ForInitExpr:
		g.genExpr(n.MustExpr(), ast.PrecedenceLowest, ctxForbidIn)
		g.writeByte(';')
	case ast.ForInitVarDecl:
		g.gen(n.MustVarDecl())
	}
}

func (g *GenVisitor) VisitForInStatement(n *ast.ForInStatement) {
	g.writeString("for")
	g.space()
	g.writeByte('(')
	g.gen(n.Into)
	g.writeString(" in ")
	g.genExpr(n.Source, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()
	g.gen(n.Body)
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
	g.genExpr(n.Source, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()
	g.gen(n.Body)
}

func (g *GenVisitor) VisitForInto(n *ast.ForInto) {
	switch n.Kind() {
	case ast.ForIntoVarDecl:
		into := n.MustVarDecl()

		g.writeString(into.Token.String())
		g.writeByte(' ')
		g.gen(&into.List)
	case ast.ForIntoExpr:
		g.gen(n.MustExpr())
	}
}

func (g *GenVisitor) VisitDoWhileStatement(n *ast.DoWhileStatement) {
	g.writeString("do ")
	g.gen(n.Body)
	g.writeString(" while(")
	g.genExpr(n.Test, ast.PrecedenceLowest, 0)
	g.writeString(");")
}

func (g *GenVisitor) VisitWhileStatement(n *ast.WhileStatement) {
	g.writeString("while")
	g.space()
	g.writeByte('(')
	g.genExpr(n.Test, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()
	g.gen(n.Body)
}

func (g *GenVisitor) VisitSwitchStatement(n *ast.SwitchStatement) {
	g.writeString("switch")
	g.space()
	g.writeByte('(')
	g.genExpr(n.Discriminant, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()
	g.writeByte('{')

	g.indent++
	for i := range n.Body {
		g.lineAndPad()
		g.gen(&n.Body[i])
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
		g.genExpr(n.Test, ast.PrecedenceLowest, 0)
		g.writeByte(':')
	} else {
		g.writeString("default:")
	}
	g.indent++
	for i := range n.Consequent {
		g.lineAndPad()
		g.gen(&n.Consequent[i])
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
	g.gen(n.Statement)
}

func (g *GenVisitor) VisitWithStatement(n *ast.WithStatement) {
	g.writeString("with")
	g.space()
	g.writeByte('(')
	g.genExpr(n.Object, ast.PrecedenceLowest, 0)
	g.writeByte(')')
	g.space()
	g.gen(n.Body)
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
	for i := range n.List {
		g.gen(&n.List[i])
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
	switch n.Kind() {
	case ast.MemPropIdent:
		g.writeByte('.')
		g.gen(n.MustIdent())
	case ast.MemPropComputed:
		g.writeByte('[')
		g.genExpr(n.MustComputed().Expr, ast.PrecedenceLowest, 0)
		g.writeByte(']')
	}
}

func (g *GenVisitor) VisitPropertyKeyed(n *ast.PropertyKeyed) {
	if n.Kind == ast.PropertyKindGet || n.Kind == ast.PropertyKindSet {
		g.writeString(string(n.Kind))
		g.writeByte(' ')
		if n.Computed {
			g.writeByte('[')
			g.genExpr(n.Key, ast.PrecedenceLowest, 0)
			g.writeByte(']')
		} else {
			g.genExpr(n.Key, ast.PrecedenceLowest, 0)
		}
		f := n.Value.MustFuncLit()
		g.gen(f.ParameterList)
		g.space()
		g.gen(f.Body)
		return
	}
	if n.Computed {
		g.writeByte('[')
		g.genExpr(n.Key, ast.PrecedenceLowest, 0)
		g.writeByte(']')
	} else {
		g.genExpr(n.Key, ast.PrecedenceLowest, 0)
	}
	g.writeByte(':')
	g.space()
	g.genExpr(n.Value, ast.PrecedenceAssign, 0)
}
