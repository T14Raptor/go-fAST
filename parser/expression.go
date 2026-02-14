package parser

import (
	"strings"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

func (p *parser) parseIdentifier() *ast.Identifier {
	literal := p.currentString()
	idx := p.currentOffset()
	p.next()
	return p.alloc.Identifier(idx, literal)
}

func (p *parser) parsePrimaryExpression() ast.Expr {
	idx := p.currentOffset()
	switch p.currentKind() {
	case token.Identifier:
		parsedLiteral := p.currentString()
		p.next()
		return p.alloc.Identifier(idx, parsedLiteral)
	case token.Null:
		p.next()
		return p.alloc.NullLiteral(idx)
	case token.Boolean:
		parsedLiteral := p.currentString()
		p.next()
		value := false
		switch parsedLiteral {
		case "true":
			value = true
		case "false":
			value = false
		default:
			p.error("Illegal boolean literal")
		}
		return p.alloc.BooleanLiteral(idx, value)
	case token.String:
		parsedLiteral := p.currentString()
		raw := p.token.Raw(p.scanner)
		p.next()
		return p.alloc.StringLiteral(idx, parsedLiteral, raw)
	case token.Number:
		parsedLiteral := p.currentString()
		raw := p.token.Raw(p.scanner)
		p.next()
		value, err := parseNumberLiteral(parsedLiteral)
		if err != nil {
			p.error(err.Error())
			value = 0
		}
		return p.alloc.NumberLiteral(idx, value, raw)
	case token.Slash, token.QuotientAssign:
		pat, flags, lit := p.scanner.ParseRegExp()
		p.next()
		return p.alloc.RegExpLiteral(idx, lit, pat, flags)
	case token.LeftBrace:
		return p.parseObjectLiteral()
	case token.LeftBracket:
		return p.parseArrayLiteral()
	case token.LeftParenthesis:
		return p.parseParenthesisedExpression()
	case token.NoSubstitutionTemplate, token.TemplateHead:
		return p.parseTemplateLiteral(false)
	case token.This:
		p.next()
		return p.alloc.ThisExpression(idx)
	case token.Super:
		return p.parseSuperProperty()
	case token.Async:
		if f := p.parseMaybeAsyncFunction(false); f != nil {
			return f
		}
	case token.Function:
		return p.parseFunction(false, false, idx)
	case token.Class:
		return p.parseClass(false)
	}

	if p.isBindingId(p.currentKind()) {
		p.next()
		return p.alloc.Identifier(idx, "")
	}

	p.errorUnexpectedToken(p.currentKind())
	p.nextStatement()
	return p.alloc.InvalidExpression(idx, p.currentOffset())
}

func (p *parser) parseSuperProperty() ast.Expr {
	idx := p.currentOffset()
	p.next()
	switch p.currentKind() {
	case token.Period:
		p.next()
		if !token.ID(p.currentKind()) {
			p.expect(token.Identifier)
			p.nextStatement()
			return p.alloc.InvalidExpression(idx, p.currentOffset())
		}
		idIdx := p.currentOffset()
		parsedLiteral := p.currentString()
		p.next()
		return p.alloc.MemberExpression(
			p.alloc.Expression(p.alloc.SuperExpression(idx)),
			p.alloc.MemberProperty(p.alloc.Identifier(idIdx, parsedLiteral)),
		)
	case token.LeftBracket:
		return p.parseBracketMember(p.alloc.SuperExpression(idx))
	case token.LeftParenthesis:
		return p.parseCallExpression(p.alloc.SuperExpression(idx))
	default:
		p.error("'super' keyword unexpected here")
		p.nextStatement()
		return p.alloc.InvalidExpression(idx, p.currentOffset())
	}
}

func (p *parser) reinterpretSequenceAsArrowFuncParams(list ast.Expressions) ast.ParameterList {
	firstRestIdx := -1
	params := make([]ast.VariableDeclarator, 0, len(list))
	for i, item := range list {
		if _, ok := item.Expr.(*ast.SpreadElement); ok {
			if firstRestIdx == -1 {
				firstRestIdx = i
				continue
			}
		}
		if firstRestIdx != -1 {
			p.error("Rest parameter must be last formal parameter")
			return ast.ParameterList{}
		}
		params = append(params, p.reinterpretAsBinding(item.Expr))
	}
	var rest ast.Expr
	if firstRestIdx != -1 {
		rest = p.reinterpretAsBindingRestElement(list[firstRestIdx].Expr)
	}
	return ast.ParameterList{
		List: params,
		Rest: rest,
	}
}

func (p *parser) parseParenthesisedExpression() ast.Expr {
	opening := p.currentOffset()
	p.expect(token.LeftParenthesis)
	var list ast.Expressions
	if p.currentKind() != token.RightParenthesis {
		for {
			if p.currentKind() == token.Ellipsis {
				start := p.currentOffset()
				p.errorUnexpectedToken(token.Ellipsis)
				p.next()
				expr := p.parseAssignmentExpression()
				list = append(list, ast.Expression{Expr: p.alloc.InvalidExpression(start, expr.Idx1())})
			} else {
				list = append(list, ast.Expression{Expr: p.parseAssignmentExpression()})
			}
			if p.currentKind() != token.Comma {
				break
			}
			p.next()
			if p.currentKind() == token.RightParenthesis {
				p.errorUnexpectedToken(token.RightParenthesis)
				break
			}
		}
	}
	p.expect(token.RightParenthesis)
	if len(list) == 1 && p.errors == nil {
		return list[0].Expr
	}
	if len(list) == 0 {
		p.errorUnexpectedToken(token.RightParenthesis)
		return p.alloc.InvalidExpression(opening, p.currentOffset())
	}
	return p.alloc.SequenceExpression(list)
}

func (p *parser) isBindingId(tok token.Token) bool {
	if tok == token.Identifier {
		return true
	}

	if tok == token.Await {
		return !p.scope.allowAwait
	}
	if tok == token.Yield {
		return !p.scope.allowYield
	}

	if token.UnreservedWord(tok) {
		return true
	}
	return false
}

func (p *parser) tokenToBindingId() {
	if p.isBindingId(p.currentKind()) {
		p.token.Kind = token.Identifier
	}
}

func (p *parser) parseBindingTarget() (target ast.Target) {
	p.tokenToBindingId()
	switch p.currentKind() {
	case token.Identifier:
		target = p.alloc.Identifier(p.currentOffset(), p.currentString())
		p.next()
	case token.LeftBracket:
		target = p.parseArrayBindingPattern()
	case token.LeftBrace:
		target = p.parseObjectBindingPattern()
	default:
		idx := p.expect(token.Identifier)
		p.nextStatement()
		target = p.alloc.InvalidExpression(idx, p.currentOffset())
	}

	return
}

func (p *parser) parseVariableDeclaration(declarationList *ast.VariableDeclarators) ast.VariableDeclarator {
	node := p.alloc.VariableDeclarator(p.alloc.BindingTarget(p.parseBindingTarget()))

	if p.currentKind() == token.Assign {
		p.next()
		node.Initializer = p.alloc.Expression(p.parseAssignmentExpression())
	}

	if declarationList != nil {
		*declarationList = append(*declarationList, *node)
	}

	return *node
}

func (p *parser) parseVariableDeclarationList() (declarationList ast.VariableDeclarators) {
	for {
		p.parseVariableDeclaration(&declarationList)
		if p.currentKind() != token.Comma {
			break
		}
		p.next()
	}
	return
}

func (p *parser) parseObjectPropertyKey() (string, string, ast.Expr, token.Token) {
	if p.currentKind() == token.LeftBracket {
		p.next()
		expr := p.parseAssignmentExpression()
		p.expect(token.RightBracket)
		return "", "", expr, token.Illegal
	}
	idx, tkn, literal, parsedLiteral := p.currentOffset(), p.currentKind(), p.token.Raw(p.scanner), p.currentString()
	var value ast.Expr
	p.next()
	switch tkn {
	case token.Identifier, token.String, token.Keyword, token.EscapedReservedWord:
		value = p.alloc.StringLiteral(idx, parsedLiteral, literal)
	case token.Number:
		num, err := parseNumberLiteral(literal)
		if err != nil {
			p.error(err.Error())
		} else {
			value = p.alloc.NumberLiteral(idx, num, literal)
		}
	case token.PrivateIdentifier:
		value = p.alloc.PrivateIdentifier(p.alloc.Identifier(idx, parsedLiteral))
	default:
		// null, false, class, etc.
		if token.ID(tkn) {
			value = p.alloc.StringLiteral(idx, literal, literal)
		} else {
			p.errorUnexpectedToken(tkn)
		}
	}
	return literal, parsedLiteral, value, tkn
}

func (p *parser) parseObjectProperty() ast.Prop {
	if p.currentKind() == token.Ellipsis {
		p.next()
		return p.alloc.SpreadElement(p.alloc.Expression(p.parseAssignmentExpression()))
	}
	keyStartIdx := p.currentOffset()
	generator := false
	if p.currentKind() == token.Multiply {
		generator = true
		p.next()
	}
	literal, parsedLiteral, value, tkn := p.parseObjectPropertyKey()
	if value == nil {
		return nil
	}
	if token.ID(tkn) || tkn == token.String || tkn == token.Number || tkn == token.Illegal {
		if generator {
			return p.alloc.PropertyKeyed(
				p.alloc.Expression(value),
				ast.PropertyKindMethod,
				p.alloc.Expression(p.parseMethodDefinition(keyStartIdx, ast.PropertyKindMethod, true, false)),
				tkn == token.Illegal,
			)
		}
		switch {
		case p.currentKind() == token.LeftParenthesis:
			return p.alloc.PropertyKeyed(
				p.alloc.Expression(value),
				ast.PropertyKindMethod,
				p.alloc.Expression(p.parseMethodDefinition(keyStartIdx, ast.PropertyKindMethod, false, false)),
				tkn == token.Illegal,
			)
		case p.currentKind() == token.Comma || p.currentKind() == token.RightBrace || p.currentKind() == token.Assign: // shorthand property
			if p.isBindingId(tkn) {
				var initializer ast.Expr
				if p.currentKind() == token.Assign {
					// allow the initializer syntax here in case the object literal
					// needs to be reinterpreted as an assignment pattern, enforce later if it doesn't.
					p.next()
					initializer = p.parseAssignmentExpression()
				}
				return p.alloc.PropertyShort(
					p.alloc.Identifier(value.Idx0(), parsedLiteral),
					p.alloc.Expression(initializer),
				)
			} else {
				p.errorUnexpectedToken(p.currentKind())
			}
		case (literal == "get" || literal == "set" || tkn == token.Async) && p.currentKind() != token.Colon:
			_, _, keyValue, tkn1 := p.parseObjectPropertyKey()
			if keyValue == nil {
				return nil
			}

			var kind ast.PropertyKind
			var async bool
			if tkn == token.Async {
				async = true
				kind = ast.PropertyKindMethod
			} else if literal == "get" {
				kind = ast.PropertyKindGet
			} else {
				kind = ast.PropertyKindSet
			}

			return p.alloc.PropertyKeyed(
				p.alloc.Expression(keyValue),
				kind,
				p.alloc.Expression(p.parseMethodDefinition(keyStartIdx, kind, false, async)),
				tkn1 == token.Illegal,
			)
		}
	}

	p.expect(token.Colon)
	return p.alloc.PropertyKeyed(
		p.alloc.Expression(value),
		ast.PropertyKindValue,
		p.alloc.Expression(p.parseAssignmentExpression()),
		tkn == token.Illegal,
	)
}

func (p *parser) parseMethodDefinition(keyStartIdx ast.Idx, kind ast.PropertyKind, generator, async bool) *ast.FunctionLiteral {
	if generator != p.scope.allowYield {
		p.scope.allowYield = generator
		defer func() {
			p.scope.allowYield = !generator
		}()
	}
	if async != p.scope.allowAwait {
		p.scope.allowAwait = async
		defer func() {
			p.scope.allowAwait = !async
		}()
	}
	parameterList := p.parseFunctionParameterList()
	switch kind {
	case ast.PropertyKindGet:
		if len(parameterList.List) > 0 || parameterList.Rest != nil {
			p.error("Getter must not have any formal parameters.")
		}
	case ast.PropertyKindSet:
		if len(parameterList.List) != 1 || parameterList.Rest != nil {
			p.error("Setter must have exactly one formal parameter.")
		}
	}
	node := p.alloc.FunctionLiteral(keyStartIdx, async)
	node.ParameterList = parameterList
	node.Generator = generator
	node.Body = p.parseFunctionBlock(async, async, generator)
	return node
}

func (p *parser) parseObjectLiteral() *ast.ObjectLiteral {
	value := make([]ast.Property, 0, 4)
	idx0 := p.expect(token.LeftBrace)
	for p.currentKind() != token.RightBrace && p.currentKind() != token.Eof {
		property := p.parseObjectProperty()
		if property != nil {
			value = append(value, ast.Property{Prop: property})
		}
		if p.currentKind() != token.RightBrace {
			p.expect(token.Comma)
		} else {
			break
		}
	}
	idx1 := p.expect(token.RightBrace)

	return p.alloc.ObjectLiteral(idx0, idx1, value)
}

func (p *parser) parseArrayLiteral() *ast.ArrayLiteral {
	idx0 := p.expect(token.LeftBracket)
	value := make(ast.Expressions, 0, 4)
	for p.currentKind() != token.RightBracket && p.currentKind() != token.Eof {
		if p.currentKind() == token.Comma {
			p.next()
			value = append(value, ast.Expression{})
			continue
		}
		if p.currentKind() == token.Ellipsis {
			p.next()
			value = append(value, ast.Expression{Expr: p.alloc.SpreadElement(
				p.alloc.Expression(p.parseAssignmentExpression()),
			)})
		} else {
			value = append(value, ast.Expression{Expr: p.parseAssignmentExpression()})
		}
		if p.currentKind() != token.RightBracket {
			p.expect(token.Comma)
		}
	}
	idx1 := p.expect(token.RightBracket)

	return p.alloc.ArrayLiteral(idx0, idx1, value)
}

func (p *parser) parseTemplateLiteral(tagged bool) *ast.TemplateLiteral {
	res := p.alloc.TemplateLiteral(p.currentOffset())

	for {
		start := p.currentOffset()
		literal := p.token.TemplateLiteral(p.scanner)
		parsed := p.token.TemplateParsed(p.scanner)
		kind := p.currentKind()

		res.Elements = append(res.Elements, ast.TemplateElement{
			Idx:     start,
			Literal: literal,
			Parsed:  parsed,
			Valid:   true,
		})

		if kind == token.NoSubstitutionTemplate || kind == token.TemplateTail {
			res.CloseQuote = p.token.Idx1 - 1
			p.next()
			break
		}

		// TemplateHead or TemplateMiddle: parse the substitution expression
		p.next()
		expr := p.parseExpression()
		res.Expressions = append(res.Expressions, ast.Expression{Expr: expr})

		if p.currentKind() != token.RightBrace {
			p.errorUnexpectedToken(p.currentKind())
			break
		}
		// Re-tokenize the `}` as the start of the next template part
		p.token = p.scanner.NextTemplatePart()
	}
	return res
}

func (p *parser) parseTaggedTemplateLiteral(tag ast.Expr) *ast.TemplateLiteral {
	l := p.parseTemplateLiteral(true)
	l.Tag = p.alloc.Expression(tag)
	return l
}

func (p *parser) parseArgumentList() (argumentList ast.Expressions, idx0, idx1 ast.Idx) {
	idx0 = p.expect(token.LeftParenthesis)
	argumentList = make(ast.Expressions, 0, 4)
	for p.currentKind() != token.RightParenthesis {
		var item ast.Expr
		if p.currentKind() == token.Ellipsis {
			p.next()
			item = p.alloc.SpreadElement(p.alloc.Expression(p.parseAssignmentExpression()))
		} else {
			item = p.parseAssignmentExpression()
		}
		argumentList = append(argumentList, ast.Expression{Expr: item})
		if p.currentKind() != token.Comma {
			break
		}
		p.next()
	}
	idx1 = p.expect(token.RightParenthesis)
	return
}

func (p *parser) parseCallExpression(left ast.Expr) ast.Expr {
	argumentList, idx0, idx1 := p.parseArgumentList()
	return p.alloc.CallExpression(p.alloc.Expression(left), idx0, argumentList, idx1)
}

func (p *parser) parseDotMember(left ast.Expr) ast.Expr {
	period := p.currentOffset()
	p.next()

	literal := p.currentString()
	idx := p.currentOffset()

	if p.currentKind() == token.PrivateIdentifier {
		p.next()
		return p.alloc.PrivateDotExpression(
			p.alloc.Expression(left),
			p.alloc.PrivateIdentifier(p.alloc.Identifier(idx, literal)),
		)
	}

	if !token.ID(p.currentKind()) {
		p.expect(token.Identifier)
		p.nextStatement()
		return p.alloc.InvalidExpression(period, p.currentOffset())
	}

	p.next()

	return p.alloc.MemberExpression(
		p.alloc.Expression(left),
		p.alloc.MemberProperty(p.alloc.Identifier(idx, literal)),
	)
}

func (p *parser) parseBracketMember(left ast.Expr) *ast.MemberExpression {
	p.expect(token.LeftBracket)
	member := p.parseExpression()
	p.expect(token.RightBracket)
	return p.alloc.MemberExpression(
		p.alloc.Expression(left),
		p.alloc.MemberProperty(p.alloc.ComputedProperty(p.alloc.Expression(member))),
	)
}

func (p *parser) parseNewExpression() ast.Expr {
	idx := p.expect(token.New)
	if p.currentKind() == token.Period {
		p.next()
		if p.currentString() == "target" {
			return p.alloc.MetaProperty(
				p.alloc.Identifier(idx, token.New.String()),
				p.parseIdentifier(),
				idx,
			)
		}
		p.errorUnexpectedToken(token.Identifier)
	}
	callee := p.parseLeftHandSideExpression()
	if bad, ok := callee.(*ast.InvalidExpression); ok {
		bad.From = idx
		return bad
	}
	node := p.alloc.NewExpression(idx, p.alloc.Expression(callee))
	if p.currentKind() == token.LeftParenthesis {
		argumentList, idx0, idx1 := p.parseArgumentList()
		node.ArgumentList = argumentList
		node.LeftParenthesis = idx0
		node.RightParenthesis = idx1
	}
	return node
}

func (p *parser) parseLeftHandSideExpression() ast.Expr {
	var left ast.Expr
	if p.currentKind() == token.New {
		left = p.parseNewExpression()
	} else {
		left = p.parsePrimaryExpression()
	}
L:
	for {
		switch p.currentKind() {
		case token.Period:
			left = p.parseDotMember(left)
		case token.LeftBracket:
			left = p.parseBracketMember(left)
		case token.NoSubstitutionTemplate, token.TemplateHead:
			left = p.parseTaggedTemplateLiteral(left)
		default:
			break L
		}
	}

	return left
}

func (p *parser) parseLeftHandSideExpressionAllowCall() ast.Expr {
	allowIn := p.scope.allowIn
	p.scope.allowIn = true
	defer func() {
		p.scope.allowIn = allowIn
	}()

	var left ast.Expr
	start := p.currentOffset()
	if p.currentKind() == token.New {
		left = p.parseNewExpression()
	} else {
		left = p.parsePrimaryExpression()
	}

	optionalChain := false
L:
	for {
		switch p.currentKind() {
		case token.Period:
			left = p.parseDotMember(left)
		case token.LeftBracket:
			left = p.parseBracketMember(left)
		case token.LeftParenthesis:
			left = p.parseCallExpression(left)
		case token.NoSubstitutionTemplate, token.TemplateHead:
			if optionalChain {
				p.error("Invalid template literal on optional chain")
				p.nextStatement()
				return p.alloc.InvalidExpression(start, p.currentOffset())
			}
			left = p.parseTaggedTemplateLiteral(left)
		case token.QuestionDot:
			optionalChain = true
			left = p.alloc.Optional(p.alloc.Expression(left))

			switch p.peek().Kind {
			case token.LeftBracket, token.LeftParenthesis, token.NoSubstitutionTemplate, token.TemplateHead:
				p.next()
			default:
				left = p.parseDotMember(left)
			}
		default:
			break L
		}
	}

	if optionalChain {
		left = p.alloc.OptionalChain(p.alloc.Expression(left))
	}
	return left
}

func (p *parser) parseUpdateExpression() ast.Expr {
	switch p.currentKind() {
	case token.Increment, token.Decrement:
		tkn := p.currentKind()
		idx := p.currentOffset()
		p.next()
		operand := p.parseUnaryExpression()
		switch operand.(type) {
		case *ast.Identifier, *ast.PrivateDotExpression, *ast.MemberExpression:
		default:
			p.error("Invalid left-hand side in assignment")
			p.nextStatement()
			return p.alloc.InvalidExpression(idx, p.currentOffset())
		}
		return p.alloc.UpdateExpression(tkn, idx, p.alloc.Expression(operand), false)
	default:
		operand := p.parseLeftHandSideExpressionAllowCall()
		if p.currentKind() == token.Increment || p.currentKind() == token.Decrement {
			// Make sure there is no line terminator here
			if p.implicitSemicolon {
				return operand
			}
			tkn := p.currentKind()
			idx := p.currentOffset()
			p.next()
			switch operand.(type) {
			case *ast.Identifier, *ast.PrivateDotExpression, *ast.MemberExpression:
			default:
				p.error("Invalid left-hand side in assignment")
				p.nextStatement()
				return p.alloc.InvalidExpression(idx, p.currentOffset())
			}
			return p.alloc.UpdateExpression(tkn, idx, p.alloc.Expression(operand), true)
		}
		return operand
	}
}

func (p *parser) parseUnaryExpression() ast.Expr {
	switch p.currentKind() {
	case token.Plus, token.Minus, token.Not, token.BitwiseNot:
		fallthrough
	case token.Delete, token.Void, token.Typeof:
		tkn := p.currentKind()
		idx := p.currentOffset()
		p.next()
		return p.alloc.UnaryExpression(tkn, idx, p.alloc.Expression(p.parseUnaryExpression()))
	case token.Await:
		if p.scope.allowAwait {
			idx := p.currentOffset()
			p.next()
			if !p.scope.inAsync {
				p.errorUnexpectedToken(token.Await)
				return p.alloc.InvalidExpression(idx, p.currentOffset())
			}
			if p.scope.inFuncParams {
				p.error("Illegal await-expression in formal parameters of async function")
			}
			return p.alloc.AwaitExpression(idx, p.alloc.Expression(p.parseUnaryExpression()))
		}
	}

	return p.parseUpdateExpression()
}

// parseBinaryExpressionOrHigher parses a binary expression using the Pratt parsing algorithm.
// minPrecedence is the minimum precedence level to parse (operators with lower
// or equal precedence will stop the loop, depending on associativity).
//
// This is a 1:1 port of oxc's parse_binary_expression_or_higher().
// See: https://matklad.github.io/2020/04/13/simple-but-powerful-pratt-parsing.html
func (p *parser) parseBinaryExpressionOrHigher(minPrecedence Precedence) ast.Expr {
	lhsParenthesized := p.currentKind() == token.LeftParenthesis

	// [+In] PrivateIdentifier in ShiftExpression[?Yield, ?Await]
	var lhs ast.Expr
	if p.scope.allowIn && p.currentKind() == token.PrivateIdentifier {
		lhs = p.parsePrivateInExpression(minPrecedence)
	} else {
		lhs = p.parseUnaryExpression()
	}

	return p.parseBinaryExpressionRest(lhs, lhsParenthesized, minPrecedence)
}

// parseBinaryExpressionRest is the core Pratt parsing loop.
// It consumes binary/logical operators and their right-hand sides as long as
// the operator's binding power exceeds minPrecedence.
//
// The loop is branchless with respect to associativity: the even/odd encoding
// of Precedence values combined with the XOR in the recursive call handles
// both left- and right-associative operators with a single <= comparison.
func (p *parser) parseBinaryExpressionRest(lhs ast.Expr, lhsParenthesized bool, minPrecedence Precedence) ast.Expr {
	for {
		kind := p.currentKind()

		// Single indexed table load — no branches. Returns 0 for non-operators.
		lbp := kindToPrecedence(kind)

		// Unified break: covers non-operators (lbp=0) and precedence check.
		// For left-assoc (even lbp):  lbp <= min breaks on same-or-lower precedence.
		// For right-assoc (odd lbp): lbp <= min only breaks on strictly lower
		//   (because the recursive call passed lbp-1 as min, not lbp+1).
		if lbp <= minPrecedence {
			break
		}

		// Omit the `in` keyword when not in [+In] context.
		if kind == token.In && !p.scope.allowIn {
			break
		}

		p.next() // consume operator

		// XOR flips even↔odd: left-assoc passes lbp+1 (same-level breaks),
		// right-assoc passes lbp-1 (same-level continues).
		rhsParenthesized := p.currentKind() == token.LeftParenthesis
		rhs := p.parseBinaryExpressionOrHigher(lbp ^ 1)

		if isLogicalOperator(kind) {
			// Mixed coalesce check: ?? cannot be mixed with && or || without parentheses.
			if kind == token.Coalesce {
				if bexp, isBin := rhs.(*ast.BinaryExpression); isBin && !rhsParenthesized &&
					(bexp.Operator == token.LogicalAnd || bexp.Operator == token.LogicalOr) {
					p.error("Logical expressions and coalesce expressions cannot be mixed. Wrap either by parentheses")
				}
				if bexp, isBin := lhs.(*ast.BinaryExpression); isBin && !lhsParenthesized &&
					(bexp.Operator == token.LogicalAnd || bexp.Operator == token.LogicalOr) {
					p.error("Logical expressions and coalesce expressions cannot be mixed. Wrap either by parentheses")
				}
			}
		} else {
			// Check for unparenthesized unary/await before **
			if kind == token.Exponent && !lhsParenthesized {
				switch lhs.(type) {
				case *ast.UnaryExpression, *ast.AwaitExpression:
					p.error("Unary operator used immediately before exponentiation expression. Parenthesis must be used to disambiguate operator precedence")
				}
			}
		}
		lhs = p.alloc.BinaryExpression(kind, p.alloc.Expression(lhs), p.alloc.Expression(rhs))

		// After first iteration, lhs is a BinaryExpression we just built — not parenthesized.
		lhsParenthesized = false
	}

	return lhs
}

// parsePrivateInExpression handles the `#identifier in expr` syntax.
// This is a 1:1 port of oxc's parse_private_in_expression().
func (p *parser) parsePrivateInExpression(minPrecedence Precedence) ast.Expr {
	left := p.alloc.PrivateIdentifier(p.alloc.Identifier(p.currentOffset(), p.currentString()))
	p.next()

	// If next token is not `in`, or `in`'s precedence (Compare) is too low, just return the identifier.
	if p.currentKind() != token.In || PrecedenceCompare <= minPrecedence {
		return left
	}

	p.next() // consume `in`
	rhs := p.parseBinaryExpressionOrHigher(PrecedenceCompare)
	return p.alloc.BinaryExpression(token.In, p.alloc.Expression(left), p.alloc.Expression(rhs))
}

func (p *parser) parseConditionalExpression() ast.Expr {
	left := p.parseBinaryExpressionOrHigher(PrecedenceLowest)

	if p.currentKind() == token.QuestionMark {
		p.next()
		allowIn := p.scope.allowIn
		p.scope.allowIn = true
		consequent := p.parseAssignmentExpression()
		p.scope.allowIn = allowIn
		p.expect(token.Colon)
		return p.alloc.ConditionalExpression(
			p.alloc.Expression(left),
			p.alloc.Expression(consequent),
			p.alloc.Expression(p.parseAssignmentExpression()),
		)
	}

	return left
}

func (p *parser) parseArrowFunction(start ast.Idx, paramList ast.ParameterList, async bool) ast.Expr {
	p.expect(token.Arrow)
	node := p.alloc.ArrowFunctionLiteral(start, paramList, async)
	node.Body = p.parseArrowFunctionBody(async)
	return node
}

func (p *parser) parseSingleArgArrowFunction(start ast.Idx, async bool) ast.Expr {
	if async != p.scope.allowAwait {
		p.scope.allowAwait = async
		defer func() {
			p.scope.allowAwait = !async
		}()
	}
	p.tokenToBindingId()
	if p.currentKind() != token.Identifier {
		p.errorUnexpectedToken(p.currentKind())
		p.next()
		return p.alloc.InvalidExpression(start, p.currentOffset())
	}

	id := p.parseIdentifier()

	paramList := ast.ParameterList{
		Opening: id.Idx,
		Closing: id.Idx1(),
		List: ast.VariableDeclarators{{
			Target: p.alloc.BindingTarget(id),
		}},
	}

	return p.parseArrowFunction(start, paramList, async)
}

func (p *parser) parseAssignmentExpression() ast.Expr {
	start := p.currentOffset()
	parenthesis := false
	async := false
	var state parserState
	switch p.currentKind() {
	case token.LeftParenthesis:
		state = p.mark()
		parenthesis = true
	case token.Async:
		tok := p.peek().Kind
		if p.isBindingId(tok) {
			// async x => ...
			p.next()
			return p.parseSingleArgArrowFunction(start, true)
		} else if tok == token.LeftParenthesis {
			state = p.mark()
			async = true
		}
	case token.Yield:
		if p.scope.allowYield {
			return p.parseYieldExpression()
		}
		fallthrough
	default:
		p.tokenToBindingId()
	}
	left := p.parseConditionalExpression()
	var operator token.Token
	switch p.currentKind() {
	case token.Assign:
		operator = p.currentKind()
	case token.AddAssign:
		operator = token.Plus
	case token.SubtractAssign:
		operator = token.Minus
	case token.MultiplyAssign:
		operator = token.Multiply
	case token.ExponentAssign:
		operator = token.Exponent
	case token.QuotientAssign:
		operator = token.Slash
	case token.RemainderAssign:
		operator = token.Remainder
	case token.AndAssign:
		operator = token.And
	case token.OrAssign:
		operator = token.Or
	case token.ExclusiveOrAssign:
		operator = token.ExclusiveOr
	case token.ShiftLeftAssign:
		operator = token.ShiftLeft
	case token.ShiftRightAssign:
		operator = token.ShiftRight
	case token.UnsignedShiftRightAssign:
		operator = token.UnsignedShiftRight
	case token.LogicalAndAssign:
		operator = token.LogicalAnd
	case token.LogicalOrAssign:
		operator = token.LogicalOr
	case token.CoalesceAssign:
		operator = token.Coalesce
	case token.Arrow:
		var paramList *ast.ParameterList
		if id, ok := left.(*ast.Identifier); ok {
			paramList = &ast.ParameterList{
				Opening: id.Idx,
				Closing: id.Idx1() - 1,
				List: ast.VariableDeclarators{{
					Target: p.alloc.BindingTarget(id),
				}},
			}
		} else if parenthesis {
			if seq, ok := left.(*ast.SequenceExpression); ok && p.errors == nil {
				paramL := p.reinterpretSequenceAsArrowFuncParams(seq.Sequence)
				paramList = &paramL
			} else {
				p.restore(state)
				paramL := p.parseFunctionParameterList()
				paramList = &paramL
			}
		} else if async {
			// async (x, y) => ...
			if !p.scope.allowAwait {
				p.scope.allowAwait = true
				defer func() {
					p.scope.allowAwait = false
				}()
			}
			if _, ok := left.(*ast.CallExpression); ok {
				p.restore(state)
				p.next() // skip "async"
				paramL := p.parseFunctionParameterList()
				paramList = &paramL
			}
		}
		if paramList == nil {
			p.error("Malformed arrow function parameter list")
			return p.alloc.InvalidExpression(left.Idx0(), left.Idx1())
		}
		return p.parseArrowFunction(start, *paramList, async)
	}

	if operator != 0 {
		idx := p.currentOffset()
		p.next()
		ok := false
		switch l := left.(type) {
		case *ast.Identifier, *ast.PrivateDotExpression, *ast.MemberExpression:
			ok = true
		case *ast.ArrayLiteral:
			if !parenthesis && operator == token.Assign {
				left = p.reinterpretAsArrayAssignmentPattern(l)
				ok = true
			}
		case *ast.ObjectLiteral:
			if !parenthesis && operator == token.Assign {
				left = p.reinterpretAsObjectAssignmentPattern(l)
				ok = true
			}
		}
		if ok {
			return p.alloc.AssignExpression(operator, p.alloc.Expression(left), p.alloc.Expression(p.parseAssignmentExpression()))
		}
		p.error("Invalid left-hand side in assignment")
		p.nextStatement()
		return p.alloc.InvalidExpression(idx, p.currentOffset())
	}

	return left
}

func (p *parser) parseYieldExpression() *ast.YieldExpression {
	idx := p.expect(token.Yield)

	if p.scope.inFuncParams {
		p.error("Yield expression not allowed in formal parameter")
	}

	node := p.alloc.YieldExpression(idx)

	if !p.implicitSemicolon && p.currentKind() == token.Multiply {
		node.Delegate = true
		p.next()
	}

	if !p.implicitSemicolon && p.currentKind() != token.Semicolon && p.currentKind() != token.RightBrace && p.currentKind() != token.Eof {
		state := p.mark()
		expr := p.parseAssignmentExpression()
		if _, bad := expr.(*ast.InvalidExpression); bad {
			expr = nil
			p.restore(state)
		}
		node.Argument = p.alloc.Expression(expr)
	}

	return node
}

func (p *parser) parseExpression() ast.Expr {
	left := p.parseAssignmentExpression()

	if p.currentKind() == token.Comma {
		sequence := ast.Expressions{ast.Expression{Expr: left}}
		for {
			if p.currentKind() != token.Comma {
				break
			}
			p.next()
			sequence = append(sequence, ast.Expression{Expr: p.parseAssignmentExpression()})
		}
		return p.alloc.SequenceExpression(sequence)
	}

	return left
}

func (p *parser) checkComma(from, to ast.Idx) {
	if pos := strings.IndexByte(p.str[int(from)-1:int(to)-1], ','); pos >= 0 {
		p.error("Comma is not allowed here")
	}
}

func (p *parser) reinterpretAsArrayAssignmentPattern(left *ast.ArrayLiteral) ast.Expr {
	value := left.Value
	var rest ast.Expr
	for i, item := range value {
		if spread, ok := item.Expr.(*ast.SpreadElement); ok {
			if i != len(value)-1 {
				p.error("Rest element must be last element")
				return p.alloc.InvalidExpression(left.Idx0(), left.Idx1())
			}
			p.checkComma(spread.Idx1(), left.RightBracket)
			rest = p.reinterpretAsDestructAssignTarget(spread.Expression.Expr)
			value = value[:len(value)-1]
		} else {
			value[i] = ast.Expression{Expr: p.reinterpretAsAssignmentElement(item.Expr)}
		}
	}
	return p.alloc.ArrayPattern(left.LeftBracket, left.RightBracket, value, p.alloc.Expression(rest))
}

func (p *parser) reinterpretArrayAssignPatternAsBinding(pattern *ast.ArrayPattern) *ast.ArrayPattern {
	for i, item := range pattern.Elements {
		pattern.Elements[i] = ast.Expression{Expr: p.reinterpretAsDestructBindingTarget(item.Expr)}
	}
	if pattern.Rest != nil {
		pattern.Rest = p.alloc.Expression(p.reinterpretAsDestructBindingTarget(pattern.Rest.Expr))
	}
	return pattern
}

func (p *parser) reinterpretAsArrayBindingPattern(left *ast.ArrayLiteral) ast.Target {
	value := left.Value
	var rest ast.Expr
	for i, item := range value {
		if spread, ok := item.Expr.(*ast.SpreadElement); ok {
			if i != len(value)-1 {
				p.error("Rest element must be last element")
				return p.alloc.InvalidExpression(left.Idx0(), left.Idx1())
			}
			p.checkComma(spread.Idx1(), left.RightBracket)
			rest = p.reinterpretAsDestructBindingTarget(spread.Expression.Expr)
			value = value[:len(value)-1]
		} else {
			value[i] = ast.Expression{Expr: p.reinterpretAsBindingElement(item.Expr)}
		}
	}
	return p.alloc.ArrayPattern(left.LeftBracket, left.RightBracket, value, p.alloc.Expression(rest))
}

func (p *parser) parseArrayBindingPattern() ast.Target {
	return p.reinterpretAsArrayBindingPattern(p.parseArrayLiteral())
}

func (p *parser) parseObjectBindingPattern() ast.Target {
	return p.reinterpretAsObjectBindingPattern(p.parseObjectLiteral())
}

func (p *parser) reinterpretArrayObjectPatternAsBinding(pattern *ast.ObjectPattern) *ast.ObjectPattern {
	for _, prop := range pattern.Properties {
		if keyed, ok := prop.Prop.(*ast.PropertyKeyed); ok {
			keyed.Value = p.alloc.Expression(p.reinterpretAsBindingElement(keyed.Value.Expr))
		}
	}
	if pattern.Rest != nil {
		pattern.Rest = p.reinterpretAsBindingRestElement(pattern.Rest)
	}
	return pattern
}

func (p *parser) reinterpretAsObjectBindingPattern(expr *ast.ObjectLiteral) ast.Target {
	var rest ast.Expr
	value := expr.Value
	for i, prop := range value {
		ok := false
		switch prop := prop.Prop.(type) {
		case *ast.PropertyKeyed:
			if prop.Kind == ast.PropertyKindValue {
				prop.Value = p.alloc.Expression(p.reinterpretAsBindingElement(prop.Value.Expr))
				ok = true
			}
		case *ast.PropertyShort:
			ok = true
		case *ast.SpreadElement:
			if i != len(expr.Value)-1 {
				p.error("Rest element must be last element")
				return p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1())
			}
			// TODO make sure there is no trailing comma
			rest = p.reinterpretAsBindingRestElement(prop.Expression.Expr)
			value = value[:i]
			ok = true
		}
		if !ok {
			p.error("Invalid destructuring binding target")
			return p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1())
		}
	}
	return p.alloc.ObjectPattern(expr.LeftBrace, expr.RightBrace, value, rest)
}

func (p *parser) reinterpretAsObjectAssignmentPattern(l *ast.ObjectLiteral) ast.Expr {
	var rest ast.Expr
	value := l.Value
	for i, prop := range value {
		ok := false
		switch prop := prop.Prop.(type) {
		case *ast.PropertyKeyed:
			if prop.Kind == ast.PropertyKindValue {
				prop.Value = p.alloc.Expression(p.reinterpretAsAssignmentElement(prop.Value.Expr))
				ok = true
			}
		case *ast.PropertyShort:
			ok = true
		case *ast.SpreadElement:
			if i != len(l.Value)-1 {
				p.error("Rest element must be last element")
				return p.alloc.InvalidExpression(l.Idx0(), l.Idx1())
			}
			// TODO make sure there is no trailing comma
			rest = prop.Expression.Expr
			value = value[:i]
			ok = true
		}
		if !ok {
			p.error("Invalid destructuring assignment target")
			return p.alloc.InvalidExpression(l.Idx0(), l.Idx1())
		}
	}
	return p.alloc.ObjectPattern(l.LeftBrace, l.RightBrace, value, rest)
}

func (p *parser) reinterpretAsAssignmentElement(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {
	case *ast.AssignExpression:
		if expr.Operator == token.Assign {
			expr.Left = p.alloc.Expression(p.reinterpretAsDestructAssignTarget(expr.Left.Expr))
			return expr
		} else {
			p.error("Invalid destructuring assignment target")
			return p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1())
		}
	default:
		return p.reinterpretAsDestructAssignTarget(expr)
	}
}

func (p *parser) reinterpretAsBindingElement(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {
	case *ast.AssignExpression:
		if expr.Operator == token.Assign {
			expr.Left = p.alloc.Expression(p.reinterpretAsDestructBindingTarget(expr.Left.Expr))
			return expr
		} else {
			p.error("Invalid destructuring assignment target")
			return p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1())
		}
	default:
		return p.reinterpretAsDestructBindingTarget(expr)
	}
}

func (p *parser) reinterpretAsBinding(expr ast.Expr) ast.VariableDeclarator {
	switch expr := expr.(type) {
	case *ast.AssignExpression:
		if expr.Operator == token.Assign {
			return ast.VariableDeclarator{
				Target:      p.alloc.BindingTarget(p.reinterpretAsDestructBindingTarget(expr.Left.Expr)),
				Initializer: expr.Right,
			}
		} else {
			p.error("Invalid destructuring assignment target")
			return ast.VariableDeclarator{
				Target: p.alloc.BindingTarget(p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1())),
			}
		}
	default:
		return ast.VariableDeclarator{
			Target: p.alloc.BindingTarget(p.reinterpretAsDestructBindingTarget(expr)),
		}
	}
}

func (p *parser) reinterpretAsDestructAssignTarget(item ast.Expr) ast.Expr {
	switch item := item.(type) {
	case nil:
		return nil
	case *ast.ArrayLiteral:
		return p.reinterpretAsArrayAssignmentPattern(item)
	case *ast.ObjectLiteral:
		return p.reinterpretAsObjectAssignmentPattern(item)
	case ast.Pattern, *ast.Identifier, *ast.PrivateDotExpression, *ast.MemberExpression:
		return item
	}
	p.error("Invalid destructuring assignment target")
	return p.alloc.InvalidExpression(item.Idx0(), item.Idx1())
}

func (p *parser) reinterpretAsDestructBindingTarget(item ast.Expr) ast.Target {
	switch item := item.(type) {
	case nil:
		return nil
	case *ast.ArrayPattern:
		return p.reinterpretArrayAssignPatternAsBinding(item)
	case *ast.ObjectPattern:
		return p.reinterpretArrayObjectPatternAsBinding(item)
	case *ast.ArrayLiteral:
		return p.reinterpretAsArrayBindingPattern(item)
	case *ast.ObjectLiteral:
		return p.reinterpretAsObjectBindingPattern(item)
	case *ast.Identifier:
		if !p.scope.allowAwait || item.Name != "await" {
			return item
		}
	}
	p.error("Invalid destructuring binding target")
	return p.alloc.InvalidExpression(item.Idx0(), item.Idx1())
}

func (p *parser) reinterpretAsBindingRestElement(expr ast.Expr) ast.Expr {
	if _, ok := expr.(*ast.Identifier); ok {
		return expr
	}
	p.error("Invalid binding rest")
	return p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1())
}
