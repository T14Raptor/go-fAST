package generator

import (
	"fmt"
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
	"strings"
	"unicode"
)

func Generate(node ast.Node) string {
	s := &state{
		out:    &strings.Builder{},
		node:   node,
		parent: &state{},
	}
	gen(s)
	return s.out.String()
}

func gen(s *state) {
	switch n := s.node.(type) {
	case nil:
	case *ast.ArrowFunctionLiteral:
		s.out.WriteString(n.Source)
	case *ast.ArrayLiteral:
		s.out.WriteString("[")
		for i, ex := range n.Value {
			if ex.Expr != nil {
				gen(s.wrap(ex.Expr))
				if i < len(n.Value)-1 {
					s.out.WriteString(", ")
				}
			}
		}
		s.out.WriteString("]")
	case *ast.AssignExpression:
		if _, ok := s.parent.node.(*ast.BinaryExpression); ok {
			s.out.WriteString("(")
			defer s.out.WriteString(")")
		}
		gen(s.wrap(n.Left.Expr))

		s.out.WriteString(" ")
		s.out.WriteString(n.Operator.String())
		if n.Operator != token.Assign {
			s.out.WriteString("=")
		}
		s.out.WriteString(" ")

		gen(s.wrap(n.Right.Expr))
	case *ast.InvalidExpression:
	case *ast.BadStatement:
	case *ast.BinaryExpression:
		if pn, ok := s.parent.node.(*ast.BinaryExpression); ok {
			operatorPrecedence := n.Operator.Precedence(true)
			parentOperatorPrecedence := pn.Operator.Precedence(true)
			if operatorPrecedence < parentOperatorPrecedence || operatorPrecedence == parentOperatorPrecedence && pn.Right.Expr == n {
				s.out.WriteString("(")
				defer s.out.WriteString(")")
			}
		}
		gen(s.wrap(n.Left.Expr))
		s.out.WriteString(" " + n.Operator.String() + " ")
		gen(s.wrap(n.Right.Expr))
	case *ast.BlockStatement:
		s.out.WriteString("{")

		s.indent++
		for _, st := range n.List {
			s.lineAndPad()
			gen(s.wrap(st.Stmt))
		}
		s.indent--

		s.lineAndPad()
		s.out.WriteString("}")
	case *ast.BooleanLiteral:
		s.out.WriteString(n.Literal)
	case *ast.BranchStatement:
		s.out.WriteString(n.Token.String())
		if n.Label != nil {
			s.out.WriteString(" ")
			gen(s.wrap(n.Label))
		}
		s.out.WriteString(";")
	case *ast.CallExpression:
		if _, ok := n.Callee.Expr.(*ast.FunctionLiteral); ok {
			s.out.WriteString("(")
			gen(s.wrap(n.Callee.Expr))
			s.out.WriteString(")")
		} else {
			gen(s.wrap(n.Callee.Expr))
		}
		s.out.WriteString("(")
		for i, a := range n.ArgumentList {
			gen(s.wrap(a.Expr))
			if i < len(n.ArgumentList)-1 {
				s.out.WriteString(", ")
			}
		}
		s.out.WriteString(")")
	case *ast.CaseStatement:
		if n.Test != nil {
			s.out.WriteString("case ")
			gen(s.wrap(n.Test.Expr))
			s.out.WriteString(": ")
		} else {
			s.out.WriteString("default: ")
		}
		gen(s.wrap(&ast.BlockStatement{List: n.Consequent}))
	case *ast.CatchStatement:
		gen(s.wrap(*n.Parameter))
		gen(s.wrap(n.Body))
	case *ast.FunctionDeclaration:
		s.lineAndPad()
		gen(s.wrap(n.Function))
	case *ast.ConditionalExpression:
		if _, ok := s.parent.node.(*ast.BinaryExpression); ok {
			s.out.WriteString("(")
			defer s.out.WriteString(")")
		}
		gen(s.wrap(n.Test.Expr))
		s.out.WriteString(" ? ")
		gen(s.wrap(n.Consequent.Expr))
		s.out.WriteString(" : ")
		gen(s.wrap(n.Alternate.Expr))
	case *ast.DebuggerStatement:
		s.out.WriteString("debugger;")
	case *ast.DoWhileStatement:
		gen(s.wrap(n.Test.Expr))
		gen(s.wrap(n.Body.Stmt))
	case *ast.MemberExpression:
		gen(s.wrap(n.Object.Expr))
		if st, ok := n.Property.Expr.(*ast.StringLiteral); ok && valid(st.Value.String()) {
			s.out.WriteString(".")
			s.out.WriteString(st.Value.String())
		} else {
			s.out.WriteString("[")
			gen(s.wrap(n.Property.Expr))
			s.out.WriteString("]")
		}
	case *ast.DotExpression:
		gen(s.wrap(n.Left.Expr))
		s.out.WriteString(".")
		s.out.WriteString(n.Identifier.Name.String())
	case *ast.EmptyStatement:
		s.out.WriteString(";")
	case *ast.ExpressionStatement:
		gen(s.wrap(n.Expression.Expr))
		s.out.WriteString(";")
		if len(n.Comment) > 0 {
			s.out.WriteString(" // " + n.Comment)
		}
	case *ast.ExpressionBody:
		gen(s.wrap(n.Expression.Expr))
	case *ast.ForInStatement:
		s.out.WriteString("for (")
		gen(s.wrap(*n.Into))
		s.out.WriteString(" in ")
		gen(s.wrap(n.Source.Expr))
		s.out.WriteString(") ")
		gen(s.wrap(n.Body.Stmt))
	case *ast.ForIntoExpression:
		gen(s.wrap(n.Expression.Expr))
	case *ast.ForStatement:
		s.out.WriteString("for (")
		gen(s.wrap(*n.Initializer))
		s.out.WriteString("; ")
		gen(s.wrap(n.Test.Expr))
		s.out.WriteString("; ")
		gen(s.wrap(n.Update.Expr))
		s.out.WriteString(") ")

		switch n.Body.Stmt.(type) {
		case *ast.EmptyStatement, *ast.BlockStatement:
		default:
			n.Body = &ast.Statement{&ast.BlockStatement{List: ast.Statements{*n.Body}}}
		}
		gen(s.wrap(n.Body.Stmt))
	case *ast.ForLoopInitializerExpression:
		gen(s.wrap(n.Expression.Expr))
	case *ast.ForIntoVar:
		gen(s.wrap(n.Binding))
	case *ast.FunctionLiteral:
		s.out.WriteString("function ")
		gen(s.wrap(n.Name))
		s.out.WriteString("(")
		for i, p := range n.ParameterList.List {
			gen(s.wrap(p))
			if i < len(n.ParameterList.List)-1 {
				s.out.WriteString(", ")
			}
		}
		s.out.WriteString(") ")
		gen(s.wrap(n.Body))
	case *ast.Identifier:
		if n != nil {
			s.out.WriteString(n.Name.String())
		}
	case *ast.IfStatement:
		s.out.WriteString("if (")
		gen(s.wrap(n.Test.Expr))
		s.out.WriteString(") ")

		switch n.Consequent.Stmt.(type) {
		case *ast.EmptyStatement, *ast.BlockStatement:
		default:
			n.Consequent = &ast.Statement{Stmt: &ast.BlockStatement{List: ast.Statements{*n.Consequent}}}
		}
		gen(s.wrap(n.Consequent.Stmt))

		if n.Alternate != nil {
			s.out.WriteString(" else ")

			switch n.Alternate.Stmt.(type) {
			case *ast.EmptyStatement, *ast.BlockStatement, *ast.IfStatement:
			default:
				n.Alternate = &ast.Statement{Stmt: &ast.BlockStatement{List: ast.Statements{*n.Alternate}}}
			}
			gen(s.wrap(n.Alternate.Stmt))
		}
	case *ast.LabelledStatement:
		gen(s.wrap(n.Label))
		gen(s.wrap(n.Statement.Stmt))
	case *ast.NewExpression:
		s.out.WriteString("new ")
		gen(s.wrap(n.Callee.Expr))
		s.out.WriteString("(")
		for i, a := range n.ArgumentList {
			gen(s.wrap(a.Expr))
			if i < len(n.ArgumentList)-1 {
				s.out.WriteString(", ")
			}
		}
		s.out.WriteString(")")
	case *ast.NullLiteral:
		s.out.WriteString(n.Literal)
	case *ast.NumberLiteral:
		s.out.WriteString(n.Literal)
	case *ast.ObjectLiteral:
		s.out.WriteString("{")

		s.indent++
		for i, p := range n.Value {
			s.lineAndPad()
			gen(s.wrap(p))
			if i < len(n.Value)-1 {
				s.out.WriteString(", ")
			}
		}
		s.indent--

		if len(n.Value) > 0 {
			s.lineAndPad()
		}
		s.out.WriteString("}")
	case *ast.PropertyKeyed:
		gen(s.wrap(n.Key.Expr))
		s.out.WriteString(": ")
		gen(s.wrap(n.Value.Expr))
	case *ast.Program:
		if n != nil {
			for _, b := range n.Body {
				gen(s.wrap(b.Stmt))
				s.line()
			}
		}
	case *ast.RegExpLiteral:
		s.out.WriteString(n.Literal)
	case *ast.ReturnStatement:
		if n != nil {
			s.out.WriteString("return")
			if n.Argument != nil {
				s.out.WriteString(" ")
				gen(s.wrap(n.Argument.Expr))
			}
			s.out.WriteString(";")
		}
	case *ast.SequenceExpression:
		switch s.parent.node.(type) {
		case *ast.BinaryExpression, *ast.ConditionalExpression, *ast.AssignExpression, *ast.CallExpression:
			s.out.WriteString("(")
			defer s.out.WriteString(")")
		}
		for i, e := range n.Sequence {
			gen(s.wrap(e.Expr))
			if i < len(n.Sequence)-1 {
				s.out.WriteString(", ")
			}
		}
	case *ast.StringLiteral:
		s.out.WriteString(n.Literal)
	case *ast.SwitchStatement:
		s.out.WriteString("switch (")
		gen(s.wrap(n.Discriminant.Expr))
		s.out.WriteString(") {")

		s.indent++
		for _, c := range n.Body {
			s.lineAndPad()
			gen(s.wrap(&c))
		}
		s.indent--

		if len(n.Body) > 0 {
			s.lineAndPad()
		}
		s.out.WriteString("}")
	case *ast.ThisExpression:
		s.out.WriteString("this")
	case *ast.ThrowStatement:
		s.out.WriteString("throw ")
		gen(s.wrap(n.Argument.Expr))
		s.out.WriteString(";")
	case *ast.TryStatement:
		s.out.WriteString("try ")

		gen(s.wrap(n.Body))

		if n.Catch != nil {
			s.out.WriteString(" catch (")
			gen(s.wrap(*n.Catch.Parameter))
			s.out.WriteString(") ")
			gen(s.wrap(n.Catch.Body))
		}
		if n.Finally != nil {
			gen(s.wrap(n.Finally))
		}
	case *ast.UnaryExpression:
		if !n.Postfix {
			s.out.WriteString(n.Operator.String())
			if len(n.Operator.String()) > 2 {
				s.out.WriteString(" ")
			}
		}

		wrap := false
		switch n.Operand.Expr.(type) {
		case *ast.BinaryExpression, *ast.ConditionalExpression, *ast.AssignExpression, *ast.UnaryExpression:
			wrap = true
		}

		if wrap {
			s.out.WriteString("(")
		}
		gen(s.wrap(n.Operand.Expr))
		if wrap {
			s.out.WriteString(")")
		}

		if n.Postfix {
			s.out.WriteString(n.Operator.String())
		}
	case *ast.VariableStatement:
		s.out.WriteString("var ")
		for _, v := range n.List {
			gen(s.wrap(v.Target))
			if v.Initializer != nil {
				s.out.WriteString(" = ")
				gen(s.wrap(v.Initializer.Expr))
			}
		}
		s.out.WriteString(";")
	case *ast.WhileStatement:
		s.out.WriteString("while (")
		gen(s.wrap(n.Test.Expr))
		s.out.WriteString(") ")
		gen(s.wrap(n.Body.Stmt))
	case *ast.WithStatement:
		gen(s.wrap(n.Object.Expr))
		gen(s.wrap(n.Body.Stmt))
	case *ast.VariableDeclarator:
		gen(s.wrap(n.Target))
		if n.Initializer != nil {
			s.out.WriteString(" = ")
			gen(s.wrap(n.Initializer.Expr))
		}
	case *ast.ForLoopInitializerVarDeclList:
		s.out.WriteString("var ")
		for i, decl := range n.List {
			gen(s.wrap(decl))
			if i < len(n.List)-1 {
				s.out.WriteString(", ")
			}
		}
	case *ast.LexicalDeclaration:
		s.out.WriteString(n.Token.String())
		s.out.WriteString(" ")
		for i, b := range n.List {
			gen(s.wrap(b))
			if i < len(n.List)-1 {
				s.out.WriteString(", ")
			}
		}
		s.out.WriteString(";")
		if len(n.Comment) > 0 {
			s.out.WriteString(" // " + n.Comment)
		}
	default:
		panic(fmt.Sprintf("gen: unexpected node type %T", n))
	}
}

func valid(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}
