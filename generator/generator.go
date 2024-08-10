package generator

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
	"strings"
	"unicode"
)

func Generate(node ast.VisitableNode) string {
	g := &GenVisitor{}
	g.gen(node)
	return g.out.String()
}

type GenVisitor struct {
	ast.NoopVisitor

	out strings.Builder

	p state
	s state
}

type state struct {
	node   ast.VisitableNode
	indent int
}

func (v *GenVisitor) gen(node ast.VisitableNode) {
	old := v.p
	v.p = v.s
	v.s = state{
		node:   node,
		indent: old.indent,
	}
	node.VisitWith(v)
	v.s = v.p
	v.p = old
}

func (v *GenVisitor) line() {
	v.out.WriteString("\n")
}

func (v *GenVisitor) lineAndPad() {
	v.line()
	v.out.WriteString(strings.Repeat("    ", v.s.indent))
}

func (g *GenVisitor) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {
	g.out.WriteString(n.Source)
}

func (g *GenVisitor) VisitArrayLiteral(n *ast.ArrayLiteral) {
	g.out.WriteString("[")
	for i, ex := range n.Value {
		if ex.Expr != nil {
			g.gen(ex.Expr)
			if i < len(n.Value)-1 {
				g.out.WriteString(", ")
			}
		}
	}
	g.out.WriteString("]")
}

func (g *GenVisitor) VisitAssignExpression(n *ast.AssignExpression) {
	if _, ok := g.p.node.(*ast.BinaryExpression); ok {
		g.out.WriteString("(")
		defer g.out.WriteString(")")
	}
	g.gen(n.Left.Expr)

	g.out.WriteString(" ")
	g.out.WriteString(n.Operator.String())
	if n.Operator != token.Assign {
		g.out.WriteString("=")
	}
	g.out.WriteString(" ")

	g.gen(n.Right.Expr)
}

func (g *GenVisitor) VisitBinaryExpression(n *ast.BinaryExpression) {
	if pn, ok := g.p.node.(*ast.BinaryExpression); ok {
		operatorPrecedence := n.Operator.Precedence(true)
		parentOperatorPrecedence := pn.Operator.Precedence(true)
		if operatorPrecedence < parentOperatorPrecedence || operatorPrecedence == parentOperatorPrecedence && pn.Right.Expr == n {
			g.out.WriteString("(")
			defer g.out.WriteString(")")
		}
	}
	g.gen(n.Left.Expr)
	g.out.WriteString(" " + n.Operator.String() + " ")
	g.gen(n.Right.Expr)
}

func (g *GenVisitor) VisitBlockStatement(n *ast.BlockStatement) {
	g.out.WriteString("{")

	g.s.indent++
	for _, st := range n.List {
		g.lineAndPad()
		g.gen(st.Stmt)
	}
	g.s.indent--

	g.lineAndPad()
	g.out.WriteString("}")
}

func (g *GenVisitor) VisitBooleanLiteral(n *ast.BooleanLiteral) {
	g.out.WriteString(n.Literal)
}

func (g *GenVisitor) VisitBranchStatement(n *ast.BranchStatement) {
	g.out.WriteString(n.Token.String())
	if n.Label != nil {
		g.out.WriteString(" ")
		g.gen(n.Label)
	}
	g.out.WriteString(";")
}

func (g *GenVisitor) VisitCallExpression(n *ast.CallExpression) {
	if _, ok := n.Callee.Expr.(*ast.FunctionLiteral); ok {
		g.out.WriteString("(")
		g.gen(n.Callee.Expr)
		g.out.WriteString(")")
	} else {
		g.gen(n.Callee.Expr)
	}
	g.out.WriteString("(")
	for i, a := range n.ArgumentList {
		g.gen(a.Expr)
		if i < len(n.ArgumentList)-1 {
			g.out.WriteString(", ")
		}
	}
	g.out.WriteString(")")
}

func (g *GenVisitor) VisitCaseStatement(n *ast.CaseStatement) {
	if n.Test != nil {
		g.out.WriteString("case ")
		g.gen(n.Test.Expr)
		g.out.WriteString(": ")
	} else {
		g.out.WriteString("default: ")
	}
	g.gen(&ast.BlockStatement{List: n.Consequent})
}

func (g *GenVisitor) VisitCatchStatement(n *ast.CatchStatement) {
	g.gen(*n.Parameter)
	g.gen(n.Body)
}

func (g *GenVisitor) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	g.lineAndPad()
	g.gen(n.Function)
}

func (g *GenVisitor) VisitConditionalExpression(n *ast.ConditionalExpression) {
	if _, ok := g.p.node.(*ast.BinaryExpression); ok {
		g.out.WriteString("(")
		defer g.out.WriteString(")")
	}
	g.gen(n.Test.Expr)
	g.out.WriteString(" ? ")
	g.gen(n.Consequent.Expr)
	g.out.WriteString(" : ")
	g.gen(n.Alternate.Expr)
}

func (g *GenVisitor) VisitDebuggerStatement(n *ast.DebuggerStatement) {
	g.out.WriteString("debugger;")
}

func (g *GenVisitor) VisitDoWhileStatement(n *ast.DoWhileStatement) {
	g.gen(n.Test.Expr)
	g.gen(n.Body.Stmt)
}

func (g *GenVisitor) VisitMemberExpression(n *ast.MemberExpression) {
	if _, ok := n.Object.Expr.(*ast.AssignExpression); ok {
		g.out.WriteString("(")
		g.gen(n.Object.Expr)
		g.out.WriteString(")")
	} else {
		g.gen(n.Object.Expr)
	}
	if st, ok := n.Property.Expr.(*ast.StringLiteral); ok && valid(st.Value.String()) {
		g.out.WriteString(".")
		g.out.WriteString(st.Value.String())
	} else {
		g.out.WriteString("[")
		g.gen(n.Property.Expr)
		g.out.WriteString("]")
	}
}

func (g *GenVisitor) VisitEmptyStatement(n *ast.EmptyStatement) {
	g.out.WriteString(";")
}

func (g *GenVisitor) VisitExpressionStatement(n *ast.ExpressionStatement) {
	g.gen(n.Expression.Expr)
	g.out.WriteString(";")
	if len(n.Comment) > 0 {
		g.out.WriteString(" // " + n.Comment)
	}
}

func (g *GenVisitor) VisitExpressionBody(n *ast.ExpressionBody) {
	g.gen(n.Expression.Expr)
}

func (g *GenVisitor) VisitForInStatement(n *ast.ForInStatement) {
	g.out.WriteString("for (")
	g.gen(*n.Into)
	g.out.WriteString(" in ")
	g.gen(n.Source.Expr)
	g.out.WriteString(") ")
	g.gen(n.Body.Stmt)
}

func (g *GenVisitor) VisitForIntoExpression(n *ast.ForIntoExpression) {
	g.gen(n.Expression.Expr)
}

func (g *GenVisitor) VisitForStatement(n *ast.ForStatement) {
	g.out.WriteString("for (")
	if *n.Initializer != nil {
		g.gen(*n.Initializer)
	}
	g.out.WriteString("; ")
	if n.Test.Expr != nil {
		g.gen(n.Test.Expr)
	}
	g.out.WriteString("; ")
	if n.Update.Expr != nil {
		g.gen(n.Update.Expr)
	}
	g.out.WriteString(") ")

	switch n.Body.Stmt.(type) {
	case *ast.EmptyStatement, *ast.BlockStatement:
	default:
		n.Body = &ast.Statement{Stmt: &ast.BlockStatement{List: ast.Statements{*n.Body}}}
	}
	g.gen(n.Body.Stmt)
}

func (g *GenVisitor) VisitForLoopInitializerExpression(n *ast.ForLoopInitializerExpression) {
	g.gen(n.Expression.Expr)
}

func (g *GenVisitor) VisitForIntoVar(n *ast.ForIntoVar) {
	g.out.WriteString("var ")
	g.gen(n.Binding)
}

func (g *GenVisitor) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	g.out.WriteString("function ")
	g.gen(n.Name)
	g.out.WriteString("(")
	for i, p := range n.ParameterList.List {
		g.gen(p)
		if i < len(n.ParameterList.List)-1 {
			g.out.WriteString(", ")
		}
	}
	g.out.WriteString(") ")
	g.gen(n.Body)
}

func (g *GenVisitor) VisitIdentifier(n *ast.Identifier) {
	if n != nil {
		g.out.WriteString(n.Name.String())
	}
}

func (g *GenVisitor) VisitIfStatement(n *ast.IfStatement) {
	g.out.WriteString("if (")
	g.gen(n.Test.Expr)
	g.out.WriteString(") ")

	switch n.Consequent.Stmt.(type) {
	case *ast.EmptyStatement, *ast.BlockStatement:
	default:
		n.Consequent = &ast.Statement{Stmt: &ast.BlockStatement{List: ast.Statements{*n.Consequent}}}
	}
	g.gen(n.Consequent.Stmt)

	if n.Alternate != nil {
		g.out.WriteString(" else ")

		switch n.Alternate.Stmt.(type) {
		case *ast.EmptyStatement, *ast.BlockStatement, *ast.IfStatement:
		default:
			n.Alternate = &ast.Statement{Stmt: &ast.BlockStatement{List: ast.Statements{*n.Alternate}}}
		}
		g.gen(n.Alternate.Stmt)
	}
}

func (g *GenVisitor) VisitLabelledStatement(n *ast.LabelledStatement) {
	g.gen(n.Label)
	g.gen(n.Statement.Stmt)
}

func (g *GenVisitor) VisitNewExpression(n *ast.NewExpression) {
	g.out.WriteString("new ")
	g.gen(n.Callee.Expr)
	g.out.WriteString("(")
	for i, a := range n.ArgumentList {
		g.gen(a.Expr)
		if i < len(n.ArgumentList)-1 {
			g.out.WriteString(", ")
		}
	}
	g.out.WriteString(")")
}

func (g *GenVisitor) VisitNullLiteral(n *ast.NullLiteral) {
	g.out.WriteString(n.Literal)
}

func (g *GenVisitor) VisitNumberLiteral(n *ast.NumberLiteral) {
	g.out.WriteString(n.Literal)
}

func (g *GenVisitor) VisitObjectLiteral(n *ast.ObjectLiteral) {
	g.out.WriteString("{")

	g.s.indent++
	for i, p := range n.Value {
		g.lineAndPad()
		g.gen(p)
		if i < len(n.Value)-1 {
			g.out.WriteString(", ")
		}
	}
	g.s.indent--

	if len(n.Value) > 0 {
		g.lineAndPad()
	}
	g.out.WriteString("}")
}

func (g *GenVisitor) VisitPropertyKeyed(n *ast.PropertyKeyed) {
	g.gen(n.Key.Expr)
	g.out.WriteString(": ")
	g.gen(n.Value.Expr)
}

func (g *GenVisitor) VisitProgram(n *ast.Program) {
	for _, b := range n.Body {
		g.gen(b.Stmt)
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
		g.gen(n.Argument.Expr)
	}
	g.out.WriteString(";")
}

func (g *GenVisitor) VisitSequenceExpression(n *ast.SequenceExpression) {
	switch g.p.node.(type) {
	case *ast.PropertyKeyed, *ast.UnaryExpression, *ast.BinaryExpression, *ast.ConditionalExpression, *ast.AssignExpression, *ast.CallExpression:
		g.out.WriteString("(")
		defer g.out.WriteString(")")
	}
	for i, e := range n.Sequence {
		g.gen(e.Expr)
		if i < len(n.Sequence)-1 {
			g.out.WriteString(", ")
		}
	}
}

func (g *GenVisitor) VisitStringLiteral(n *ast.StringLiteral) {
	g.out.WriteString(n.Literal)
}

func (g *GenVisitor) VisitSwitchStatement(n *ast.SwitchStatement) {
	g.out.WriteString("switch (")
	g.gen(n.Discriminant.Expr)
	g.out.WriteString(") {")

	g.s.indent++
	for _, c := range n.Body {
		g.lineAndPad()
		g.gen(&c)
	}
	g.s.indent--

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
	g.gen(n.Argument.Expr)
	g.out.WriteString(";")
}

func (g *GenVisitor) VisitTryStatement(n *ast.TryStatement) {
	g.out.WriteString("try ")

	g.gen(n.Body)

	if n.Catch != nil {
		g.out.WriteString(" catch (")
		g.gen(*n.Catch.Parameter)
		g.out.WriteString(") ")
		g.gen(n.Catch.Body)
	}
	if n.Finally != nil {
		g.gen(n.Finally)
	}
}

func (g *GenVisitor) VisitUnaryExpression(n *ast.UnaryExpression) {
	if !n.Postfix {
		g.out.WriteString(n.Operator.String())
		if len(n.Operator.String()) > 2 {
			g.out.WriteString(" ")
		}
	}

	wrap := false
	switch n.Operand.Expr.(type) {
	case *ast.BinaryExpression, *ast.ConditionalExpression, *ast.AssignExpression, *ast.UnaryExpression:
		wrap = true
	}

	if wrap {
		g.out.WriteString("(")
	}
	g.gen(n.Operand.Expr)
	if wrap {
		g.out.WriteString(")")
	}

	if n.Postfix {
		g.out.WriteString(n.Operator.String())
	}
}

func (g *GenVisitor) VisitVariableStatement(n *ast.VariableStatement) {
	g.out.WriteString("var ")
	for i, v := range n.List {
		g.gen(v.Target)
		if v.Initializer != nil {
			g.out.WriteString(" = ")
			g.gen(v.Initializer.Expr)
		}
		if i < len(n.List)-1 {
			g.out.WriteString(",")
		}
	}
	g.out.WriteString(";")
}

func (g *GenVisitor) VisitWhileStatement(n *ast.WhileStatement) {
	g.out.WriteString("while (")
	g.gen(n.Test.Expr)
	g.out.WriteString(") ")
	g.gen(n.Body.Stmt)
}

func (g *GenVisitor) VisitWithStatement(n *ast.WithStatement) {
	g.gen(n.Object.Expr)
	g.gen(n.Body.Stmt)
}

func (g *GenVisitor) VisitVariableDeclarator(n *ast.VariableDeclarator) {
	g.gen(n.Target)
	if n.Initializer != nil {
		g.out.WriteString(" = ")
		g.gen(n.Initializer.Expr)
	}
}

func (g *GenVisitor) VisitForLoopInitializerVarDeclList(n *ast.ForLoopInitializerVarDeclList) {
	g.out.WriteString("var ")
	for i, decl := range n.List {
		g.gen(decl)
		if i < len(n.List)-1 {
			g.out.WriteString(", ")
		}
	}
}

func (g *GenVisitor) VisitLexicalDeclaration(n *ast.LexicalDeclaration) {
	g.out.WriteString(n.Token.String())
	g.out.WriteString(" ")
	for i, b := range n.List {
		g.gen(b)
		if i < len(n.List)-1 {
			g.out.WriteString(", ")
		}
	}
	g.out.WriteString(";")
	if len(n.Comment) > 0 {
		g.out.WriteString(" // " + n.Comment)
	}
}

func valid(s string) bool {
	for i, r := range s {
		if i == 0 && unicode.IsDigit(r) {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}
