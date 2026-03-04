package generator

import (
	"math"
	"strconv"
	"strings"

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
	return g.out.String()
}

type GenVisitor struct {
	ast.NoopVisitor

	out  strings.Builder
	opts Options

	indent int

	p ast.VisitableNode
	s ast.VisitableNode
}

func (g *GenVisitor) gen(node ast.VisitableNode) {
	old := g.p
	g.p, g.s = g.s, node
	node.VisitWith(g)
	g.s, g.p = g.p, old
}

func (g *GenVisitor) line() {
	if g.opts.Minified {
		return
	}
	g.out.WriteString("\n")
}

func (g *GenVisitor) lineAndPad() {
	if g.opts.Minified {
		return
	}
	g.out.WriteString("\n")
	for i := 0; i < g.indent; i++ {
		g.out.WriteString("    ")
	}
}

func (g *GenVisitor) space() {
	if g.opts.Minified {
		return
	}
	g.out.WriteString(" ")
}

func (g *GenVisitor) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {
	if n.Async {
		g.out.WriteString("async ")
	}
	g.gen(n.ParameterList)
	g.space()
	g.out.WriteString("=>")
	g.space()
	g.gen(n.Body.Unwrap())
}

func (g *GenVisitor) VisitAwaitExpression(n *ast.AwaitExpression) {
	g.out.WriteString("await ")
	g.gen(n.Argument.Unwrap())
}

func (g *GenVisitor) VisitArrayLiteral(n *ast.ArrayLiteral) {
	g.out.WriteString("[")
	for i, ex := range n.Value {
		if !ex.IsNone() {
			g.gen(ex.Unwrap())
		}
		if i < len(n.Value)-1 {
			g.out.WriteString(",")
			g.space()
		}
	}
	g.out.WriteString("]")
}

func (g *GenVisitor) VisitAssignExpression(n *ast.AssignExpression) {
	needsParens := false

	switch n.Left.Kind() {
	case ast.ExprObjPat, ast.ExprArrPat:
		if _, ok := g.p.(*ast.ExpressionStatement); ok {
			needsParens = true
		}
	}
	// we also need parentheses if parent is binary expression
	switch g.p.(type) {
	case *ast.BinaryExpression, *ast.LogicalExpression:
		needsParens = true
	}

	if needsParens {
		g.out.WriteString("(")
		defer g.out.WriteString(")")
	}

	g.gen(n.Left.Unwrap())

	g.space()
	g.out.WriteString(n.Operator.String())
	g.space()

	g.gen(n.Right.Unwrap())
}

func (g *GenVisitor) VisitArrayPattern(n *ast.ArrayPattern) {
	g.out.WriteString("[")
	for i, elem := range n.Elements {
		if !elem.IsNone() {
			g.gen(elem.Unwrap())
		}
		if i < len(n.Elements)-1 {
			g.out.WriteString(",")
			g.space()
		}
	}
	g.out.WriteString("]")
}

func (g *GenVisitor) VisitObjectPattern(n *ast.ObjectPattern) {
	g.out.WriteString("{")
	for i := range n.Properties {
		g.gen(n.Properties[i].Unwrap())
		if i < len(n.Properties)-1 {
			g.out.WriteString(",")
			g.space()
		}
	}
	if n.Rest != nil {
		if len(n.Properties) > 0 {
			g.out.WriteString(",")
			g.space()
		}
		g.out.WriteString("...")
		g.gen(n.Rest.Unwrap())
	}
	g.out.WriteString("}")
}

func (g *GenVisitor) VisitBinaryExpression(n *ast.BinaryExpression) {
	var (
		parentPrec  ast.Precedence
		parentRight *ast.Expression
	)
	switch g.p {
	case *ast.BinaryExpression:
		parentPrec = pn.Operator.Precedence()
		parentRight = pn.Right
	case *ast.LogicalExpression:
		parentPrec = pn.Operator.Precedence()
		parentRight = pn.Right
	}
	if parentPrec > ast.PrecedenceLowest {
		prec := n.Operator.Precedence()
		if prec < parentPrec || prec == parentPrec && parentRight.Expr == n {
			g.out.WriteString("(")
			defer g.out.WriteString(")")
		}
	}

	g.gen(n.Left.Unwrap())

	if op := n.Operator.String(); len(op) > 2 {
		g.out.WriteString(" " + op + " ")
	} else {
		g.space()
		g.out.WriteString(op)
		g.space()
	}

	g.gen(n.Right.Unwrap())
}

func (g *GenVisitor) VisitLogicalExpression(n *ast.LogicalExpression) {
	var (
		parentPrec  ast.Precedence
		parentRight *ast.Expression
	)
	switch pn := g.p.(type) {
	case *ast.BinaryExpression:
		parentPrec = pn.Operator.Precedence()
		parentRight = pn.Right
	case *ast.LogicalExpression:
		parentPrec = pn.Operator.Precedence()
		parentRight = pn.Right
	}
	if parentPrec > ast.PrecedenceLowest {
		prec := n.Operator.Precedence()
		if prec < parentPrec || prec == parentPrec && parentRight.Expr == n {
			g.out.WriteString("(")
			defer g.out.WriteString(")")
		}
	}

	g.gen(n.Left.Unwrap())

	g.space()
	g.out.WriteString(n.Operator.String())
	g.space()

	g.gen(n.Right.Unwrap())
}

func (g *GenVisitor) VisitBlockStatement(n *ast.BlockStatement) {
	g.out.WriteString("{")

	g.indent++
	g.VisitStatements(&n.List)
	g.indent--

	if len(n.List) > 0 {
		g.lineAndPad()
	}
	g.out.WriteString("}")
}

func (g *GenVisitor) VisitStatements(n *ast.Statements) {
	for _, st := range *n {
		g.lineAndPad()
		g.gen(st.Unwrap())
	}
}

func (g *GenVisitor) VisitBooleanLiteral(n *ast.BooleanLiteral) {
	if n.Value {
		g.out.WriteString("true")
	} else {
		g.out.WriteString("false")
	}
}

func (g *GenVisitor) VisitBreakStatement(n *ast.BreakStatement) {
	g.out.WriteString("break")
	if n.Label != nil {
		g.out.WriteString(" ")
		g.gen(n.Label)
	}
	g.out.WriteString(";")
}

func (g *GenVisitor) VisitContinueStatement(n *ast.ContinueStatement) {
	g.out.WriteString("continue")
	if n.Label != nil {
		g.out.WriteString(" ")
		g.gen(n.Label)
	}
	g.out.WriteString(";")
}

func (g *GenVisitor) VisitCallExpression(n *ast.CallExpression) {
	switch n.Callee.Kind() {
	case ast.ExprFuncLit, ast.ExprArrowFuncLit, ast.ExprAssign:
		g.out.WriteString("(")
		g.gen(n.Callee.Unwrap())
		g.out.WriteString(")")
	default:
		g.gen(n.Callee.Unwrap())
	}
	g.out.WriteString("(")
	for i, a := range n.ArgumentList {
		g.gen(a.Unwrap())
		if i < len(n.ArgumentList)-1 {
			g.out.WriteString(",")
			g.space()
		}
	}
	g.out.WriteString(")")
}

func (g *GenVisitor) VisitCaseStatement(n *ast.CaseStatement) {
	if n.Test != nil {
		g.out.WriteString("case ")
		g.gen(n.Test.Unwrap())
		g.out.WriteString(":")
	} else {
		g.out.WriteString("default:")
	}
	g.indent++
	for i := range n.Consequent {
		g.lineAndPad()
		g.gen(n.Consequent[i].Unwrap())
	}
	g.indent--
}

func (g *GenVisitor) VisitCatchStatement(n *ast.CatchStatement) {
	if n.Parameter != nil && !n.Parameter.IsNone() {
		g.gen(n.Parameter.Unwrap())
	}
	g.gen(n.Body)
}

func (g *GenVisitor) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	g.lineAndPad()
	g.gen(n.Function)
}

func (g *GenVisitor) VisitConditionalExpression(n *ast.ConditionalExpression) {
	switch g.p.(type) {
	case *ast.BinaryExpression, *ast.LogicalExpression, *ast.NewExpression:
		g.out.WriteString("(")
		defer g.out.WriteString(")")
	}
	switch n.Test.Kind() {
	case ast.ExprAssign, ast.ExprConditional:
		g.out.WriteString("(")
		g.gen(n.Test.Unwrap())
		g.out.WriteString(")")
	default:
		g.gen(n.Test.Unwrap())
	}
	g.space()
	g.out.WriteString("?")
	g.space()
	g.gen(n.Consequent.Unwrap())
	g.space()
	g.out.WriteString(":")
	g.space()
	g.gen(n.Alternate.Unwrap())
}

func (g *GenVisitor) VisitDebuggerStatement(n *ast.DebuggerStatement) {
	g.out.WriteString("debugger;")
}

func (g *GenVisitor) VisitDoWhileStatement(n *ast.DoWhileStatement) {
	g.out.WriteString("do ")
	g.gen(n.Body.Unwrap())
	g.out.WriteString(" while(")
	g.gen(n.Test.Unwrap())
	g.out.WriteString(");")
}

func (g *GenVisitor) VisitMemberExpression(n *ast.MemberExpression) {
	switch n.Object.Kind() {
	case ast.ExprAssign, ast.ExprBinary, ast.ExprLogical, ast.ExprUnary, ast.ExprSequence, ast.ExprConditional, ast.ExprNumLit,
		ast.ExprFuncLit, ast.ExprArrowFuncLit, ast.ExprUpdate:
		g.out.WriteString("(")
		g.gen(n.Object.Unwrap())
		g.out.WriteString(")")
	default:
		g.gen(n.Object.Unwrap())
	}

	g.gen(n.Property)
}

func (g *GenVisitor) VisitMemberProperty(n *ast.MemberProperty) {
	switch {
	case n.IsComputed():
		g.out.WriteString("[")
		g.gen(n.MustComputed().Expr.Unwrap())
		g.out.WriteString("]")
	case n.IsIdent():
		g.out.WriteString(".")
		g.gen(n.MustIdent())
	}
}

func (g *GenVisitor) VisitComputedProperty(n *ast.ComputedProperty) {
	g.out.WriteString("[")
	if n.Expr != nil {
		g.gen(n.Expr.Unwrap())
	}
	g.out.WriteString("]")
}

func (g *GenVisitor) VisitEmptyStatement(n *ast.EmptyStatement) {
	g.out.WriteString(";")
}

func (g *GenVisitor) VisitExpressionStatement(n *ast.ExpressionStatement) {
	g.gen(n.Expression.Unwrap())
	g.out.WriteString(";")
	if len(n.Comment) > 0 && !g.opts.Minified {
		g.out.WriteString(" // " + n.Comment)
	}
}

func (g *GenVisitor) VisitForInStatement(n *ast.ForInStatement) {
	g.out.WriteString("for")
	g.space()
	g.out.WriteString("(")
	g.gen(n.Into)
	g.out.WriteString(" in ")
	g.gen(n.Source.Unwrap())
	g.out.WriteString(")")
	g.space()
	g.gen(n.Body.Unwrap())
}

func (g *GenVisitor) VisitForOfStatement(n *ast.ForOfStatement) {
	g.out.WriteString("for")
	if n.Await {
		g.out.WriteString(" await")
	}
	g.space()
	g.out.WriteString("(")
	g.gen(n.Into)
	g.out.WriteString(" of ")
	g.gen(n.Source.Unwrap())
	g.out.WriteString(")")
	g.space()
	g.gen(n.Body.Unwrap())
}

func (g *GenVisitor) VisitForStatement(n *ast.ForStatement) {
	g.out.WriteString("for")
	g.space()
	g.out.WriteString("(")
	if n.Initializer != nil {
		g.gen(n.Initializer)
	} else {
		g.out.WriteString(";")
	}
	g.space()

	if n.Test != nil && !n.Test.IsNone() {
		g.gen(n.Test.Unwrap())
	}
	g.out.WriteString(";")
	g.space()
	if n.Update != nil && !n.Update.IsNone() {
		g.gen(n.Update.Unwrap())
	}
	g.out.WriteString(")")
	g.space()

	switch n.Body.Kind() {
	case ast.StmtEmpty, ast.StmtBlock:
		g.gen(n.Body.Unwrap())
	default:
		g.indent++
		g.gen(n.Body.Unwrap())
		g.indent--
		g.lineAndPad()
	}
}

func (g *GenVisitor) VisitForLoopInitializer(n *ast.ForLoopInitializer) {
	switch n.Kind() {
	case ast.ForInitExpr:
		g.gen(n.MustExpr().Unwrap())
		g.out.WriteString(";")
	case ast.ForInitVarDecl:
		g.gen(n.MustVarDecl())
	}
}

func (g *GenVisitor) VisitForInto(n *ast.ForInto) {
	switch n.Kind() {
	case ast.ForIntoVarDecl:
		vd := n.MustVarDecl()
		g.out.WriteString(vd.Token.String())
		g.out.WriteString(" ")
		g.gen(&vd.List)
	case ast.ForIntoExpr:
		g.gen(n.MustExpr().Unwrap())
	}
}

func (g *GenVisitor) VisitParameterList(n *ast.ParameterList) {
	g.out.WriteString("(")
	for i, p := range n.List {
		g.gen(&p)
		if i < len(n.List)-1 {
			g.out.WriteString(",")
			g.space()
		}
	}

	if n.Rest != nil {
		g.out.WriteString("...")
		g.gen(n.Rest.Unwrap())
	}
	g.out.WriteString(")")
}

func (g *GenVisitor) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	if n.Async {
		g.out.WriteString("async ")
	}

	if n.Name != nil {
		g.out.WriteString("function ")
		g.gen(n.Name)
	} else {
		g.out.WriteString("function")
	}
	g.gen(n.ParameterList)
	g.space()
	g.gen(n.Body)
}

func (g *GenVisitor) VisitIdentifier(n *ast.Identifier) {
	if n != nil {
		g.out.WriteString(n.Name)
	}
}

func (g *GenVisitor) VisitIfStatement(n *ast.IfStatement) {
	g.out.WriteString("if")
	g.space()
	g.out.WriteString("(")
	g.gen(n.Test.Unwrap())
	g.out.WriteString(")")
	g.space()

	switch n.Consequent.Kind() {
	case ast.StmtEmpty, ast.StmtBlock:
		g.gen(n.Consequent.Unwrap())
	default:
		g.indent++
		g.gen(n.Consequent.Unwrap())
		g.indent--
		g.lineAndPad()
	}

	if n.Alternate != nil {
		g.out.WriteString(" else ")

		switch n.Alternate.Kind() {
		case ast.StmtEmpty, ast.StmtBlock, ast.StmtIf:
			g.gen(n.Alternate.Unwrap())
		default:
			g.indent++
			g.gen(n.Alternate.Unwrap())
			g.indent--
			g.lineAndPad()
		}
	}
}

func (g *GenVisitor) VisitLabelledStatement(n *ast.LabelledStatement) {
	g.gen(n.Label)
	g.out.WriteString(":")
	g.space()
	g.gen(n.Statement.Unwrap())
}

func (g *GenVisitor) VisitNewExpression(n *ast.NewExpression) {
	g.out.WriteString("new ")
	switch n.Callee.Kind() {
	case ast.ExprBinary, ast.ExprLogical, ast.ExprCall, ast.ExprConditional, ast.ExprAssign, ast.ExprUnary, ast.ExprSequence:
		g.out.WriteString("(")
		g.gen(n.Callee.Unwrap())
		g.out.WriteString(")")
	default:
		g.gen(n.Callee.Unwrap())
	}
	g.out.WriteString("(")
	for i, a := range n.ArgumentList {
		g.gen(a.Unwrap())
		if i < len(n.ArgumentList)-1 {
			g.out.WriteString(",")
			g.space()
		}
	}
	g.out.WriteString(")")
}

func (g *GenVisitor) VisitNullLiteral(n *ast.NullLiteral) {
	g.out.WriteString("null")
}

func (g *GenVisitor) VisitNumberLiteral(n *ast.NumberLiteral) {
	if n.Raw != nil {
		g.out.WriteString(*n.Raw)
	} else if math.IsInf(n.Value, 1) {
		g.out.WriteString("Infinity")
	} else if math.IsInf(n.Value, -1) {
		g.out.WriteString("-Infinity")
	} else {
		g.out.WriteString(strconv.FormatFloat(n.Value, 'f', -1, 64))
	}
}

func (g *GenVisitor) VisitBigIntLiteral(n *ast.BigIntLiteral) {
	if n.Raw != nil {
		g.out.WriteString(*n.Raw)
		return
	}
	if n.Value != nil {
		g.out.WriteString(n.Value.String())
	} else {
		g.out.WriteString("0")
	}
	g.out.WriteString("n")
}

func (g *GenVisitor) VisitObjectLiteral(n *ast.ObjectLiteral) {
	switch g.p.(type) {
	case *ast.BinaryExpression, *ast.LogicalExpression, *ast.ArrowFunctionLiteral:
		g.out.WriteString("(")
		defer g.out.WriteString(")")
	}

	g.out.WriteString("{")

	g.indent++
	for i, p := range n.Value {
		g.lineAndPad()
		g.gen(p.Unwrap())
		if i < len(n.Value)-1 {
			g.out.WriteString(",")
		}
	}
	g.indent--

	if len(n.Value) > 0 {
		g.lineAndPad()
	}
	g.out.WriteString("}")
}

func (g *GenVisitor) VisitPropertyKeyed(n *ast.PropertyKeyed) {
	switch n.Kind {
	case ast.PropertyKindGet:
		g.out.WriteString("get ")
	case ast.PropertyKindSet:
		g.out.WriteString("set ")
	}
	if n.Kind == ast.PropertyKindGet || n.Kind == ast.PropertyKindSet {
		if n.Computed {
			g.out.WriteString("[")
			g.gen(n.Key.Unwrap())
			g.out.WriteString("]")
		} else {
			g.gen(n.Key.Unwrap())
		}
		f := n.Value.MustFuncLit()
		g.gen(f.ParameterList)
		g.space()
		g.gen(f.Body)
		return
	}
	if n.Computed {
		g.out.WriteString("[")
		g.gen(n.Key.Unwrap())
		g.out.WriteString("]")
	} else {
		g.gen(n.Key.Unwrap())
	}
	g.out.WriteString(":")
	g.space()
	g.gen(n.Value.Unwrap())
}

func (g *GenVisitor) VisitProgram(n *ast.Program) {
	for _, b := range n.Body {
		g.gen(b.Unwrap())
		g.line()
	}
}

func (g *GenVisitor) VisitRegExpLiteral(n *ast.RegExpLiteral) {
	g.out.WriteString(n.Literal)
}

func (g *GenVisitor) VisitReturnStatement(n *ast.ReturnStatement) {
	g.out.WriteString("return")
	if n.Argument != nil {
		g.out.WriteString(" ")
		g.gen(n.Argument.Unwrap())
	}
	g.out.WriteString(";")
}

func (g *GenVisitor) VisitSequenceExpression(n *ast.SequenceExpression) {
	switch g.p.(type) {
	case *ast.VariableDeclarator, *ast.PropertyKeyed,
		*ast.UnaryExpression, *ast.UpdateExpression, *ast.BinaryExpression, *ast.LogicalExpression,
		*ast.ConditionalExpression, *ast.AssignExpression, *ast.CallExpression,
		*ast.ArrayLiteral:
		g.out.WriteString("(")
		defer g.out.WriteString(")")
	}
	for i, e := range n.Sequence {
		g.gen(e.Unwrap())
		if i < len(n.Sequence)-1 {
			g.out.WriteString(",")
			g.space()
		}
	}
}

func (g *GenVisitor) VisitStringLiteral(n *ast.StringLiteral) {
	if n.Raw != nil {
		g.out.WriteString(*n.Raw)
		return
	}
	g.out.WriteString(strconv.Quote(n.Value))
}

func (g *GenVisitor) VisitSwitchStatement(n *ast.SwitchStatement) {
	g.out.WriteString("switch")
	g.space()
	g.out.WriteString("(")
	g.gen(n.Discriminant.Unwrap())
	g.out.WriteString(")")
	g.space()
	g.out.WriteString("{")

	g.indent++
	for _, c := range n.Body {
		g.lineAndPad()
		g.gen(&c)
	}
	g.indent--

	if len(n.Body) > 0 {
		g.lineAndPad()
	}
	g.out.WriteString("}")
}

func (g *GenVisitor) VisitThisExpression(n *ast.ThisExpression) {
	g.out.WriteString("this")
}

func (g *GenVisitor) VisitThrowStatement(n *ast.ThrowStatement) {
	g.out.WriteString("throw ")
	g.gen(n.Argument.Unwrap())
	g.out.WriteString(";")
}

func (g *GenVisitor) VisitTryStatement(n *ast.TryStatement) {
	g.out.WriteString("try")
	g.space()

	g.gen(n.Body)

	if n.Catch != nil {
		g.space()
		g.out.WriteString("catch")
		g.space()
		if n.Catch.Parameter != nil && !n.Catch.Parameter.IsNone() {
			g.out.WriteString("(")
			g.gen(n.Catch.Parameter.Unwrap())
			g.out.WriteString(")")
			g.space()
		}
		g.gen(n.Catch.Body)
	}
	if n.Finally != nil {
		g.space()
		g.out.WriteString("finally")
		g.space()
		g.gen(n.Finally)
	}
}

func (g *GenVisitor) VisitUnaryExpression(n *ast.UnaryExpression) {
	g.out.WriteString(n.Operator.String())
	if len(n.Operator.String()) > 2 {
		g.out.WriteString(" ")
	}

	wrap := false
	switch n.Operand.Kind() {
	case ast.ExprBinary, ast.ExprLogical, ast.ExprConditional, ast.ExprAssign, ast.ExprUnary, ast.ExprUpdate:
		wrap = true
	}

	if wrap {
		g.out.WriteString("(")
	}
	g.gen(n.Operand.Unwrap())
	if wrap {
		g.out.WriteString(")")
	}
}

func (g *GenVisitor) VisitUpdateExpression(n *ast.UpdateExpression) {
	if !n.Postfix {
		g.out.WriteString(n.Operator.String())
		if len(n.Operator.String()) > 2 {
			g.out.WriteString(" ")
		}
	}

	wrap := false
	switch n.Operand.Kind() {
	case ast.ExprBinary, ast.ExprLogical, ast.ExprConditional, ast.ExprAssign, ast.ExprUnary, ast.ExprUpdate:
		wrap = true
	}

	if wrap {
		g.out.WriteString("(")
	}
	g.gen(n.Operand.Unwrap())
	if wrap {
		g.out.WriteString(")")
	}

	if n.Postfix {
		g.out.WriteString(n.Operator.String())
	}
}

func (g *GenVisitor) VisitWhileStatement(n *ast.WhileStatement) {
	g.out.WriteString("while")
	g.space()
	g.out.WriteString("(")
	g.gen(n.Test.Unwrap())
	g.out.WriteString(")")
	g.space()
	g.gen(n.Body.Unwrap())
}

func (g *GenVisitor) VisitWithStatement(n *ast.WithStatement) {
	g.out.WriteString("with")
	g.space()
	g.out.WriteString("(")
	g.gen(n.Object.Unwrap())
	g.out.WriteString(")")
	g.space()
	g.gen(n.Body.Unwrap())
}

func (g *GenVisitor) VisitVariableDeclarator(n *ast.VariableDeclarator) {
	g.gen(n.Target.Unwrap())
	if n.Initializer != nil {
		g.space()
		g.out.WriteString("=")
		g.space()
		g.gen(n.Initializer.Unwrap())
	}
}

func (g *GenVisitor) VisitTemplateLiteral(n *ast.TemplateLiteral) {
	g.out.WriteString("`")
	for i, e := range n.Elements {
		g.out.WriteString(e.Parsed)
		if i < len(n.Expressions) {
			g.out.WriteString("${")
			g.gen(n.Expressions[i].Unwrap())
			g.out.WriteString("}")
		}
	}
	g.out.WriteString("`")
}

func (g *GenVisitor) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	g.out.WriteString(n.Token.String())
	g.out.WriteString(" ")
	for i, b := range n.List {
		g.gen(&b)
		if i < len(n.List)-1 {
			g.out.WriteString(",")
			g.space()
		}
	}

	g.out.WriteString(";")
	if len(n.Comment) > 0 && !g.opts.Minified {
		g.out.WriteString(" // " + n.Comment)
	}
}

func (g *GenVisitor) VisitClassLiteral(n *ast.ClassLiteral) {
	g.out.WriteString("class")
	if n.Name != nil {
		g.out.WriteString(" ")
		g.gen(n.Name)
	}
	g.space()
	g.out.WriteString("{")

	g.indent++
	for _, element := range n.Body {
		g.lineAndPad()
		switch element.Kind() {
		case ast.ClassElemMethod:
			e := element.MustMethod()
			if e.Static {
				g.out.WriteString("static ")
			}
			switch e.Kind {
			case ast.PropertyKindGet:
				g.out.WriteString("get ")
			case ast.PropertyKindSet:
				g.out.WriteString("set ")
			}
			if e.Computed {
				g.out.WriteString("[")
				g.gen(e.Key.Unwrap())
				g.out.WriteString("]")
			} else {
				g.gen(e.Key.Unwrap())
			}
			g.gen(e.Body.ParameterList)
			g.space()
			g.gen(e.Body.Body)
		}
	}
	g.indent--

	g.lineAndPad()
	g.out.WriteString("}")
}

func (g *GenVisitor) VisitMetaProperty(n *ast.MetaProperty) {
	g.gen(n.Meta)
	g.out.WriteString(".")
	g.gen(n.Property)
}

func (g *GenVisitor) VisitSpreadElement(n *ast.SpreadElement) {
	g.out.WriteString("...")
	g.gen(n.Expression.Unwrap())
}
