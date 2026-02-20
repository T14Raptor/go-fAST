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

func (p *parser) parsePrimaryExpression() *ast.Expression {
	idx := p.currentOffset()
	switch p.currentKind() {
	case token.Identifier:
		parsedLiteral := p.currentString()
		p.next()
		return p.alloc.Expression(ast.NewIdentExpr(p.alloc.Identifier(idx, parsedLiteral)))
	case token.Null:
		p.next()
		return p.alloc.Expression(ast.NewNullLitExpr(ast.NullLiteral{Idx: idx}))
	case token.Boolean:
		value := p.token.Idx1-p.token.Idx0 == 4
		p.next()
		return p.alloc.Expression(ast.NewBoolLitExpr(ast.BooleanLiteral{Idx: idx, Value: value}))
	case token.String:
		parsedLiteral := p.currentString()
		raw := p.token.Raw(p.scanner)
		p.next()
		return p.alloc.Expression(ast.NewStrLitExpr(p.alloc.StringLiteral(idx, parsedLiteral, raw)))
	case token.Number:
		parsedLiteral := p.currentString()
		raw := p.token.Raw(p.scanner)
		p.next()
		value, err := parseNumberLiteral(parsedLiteral)
		if err != nil {
			p.errorf("%s", err.Error())
			value = 0
		}
		return p.alloc.Expression(ast.NewNumLitExpr(p.alloc.NumberLiteral(idx, value, raw)))
	case token.Slash, token.QuotientAssign:
		pat, flags, lit := p.scanner.ParseRegExp()
		p.next()
		return p.alloc.Expression(ast.NewRegExpLitExpr(p.alloc.RegExpLiteral(idx, lit, pat, flags)))
	case token.LeftBrace:
		return p.alloc.Expression(ast.NewObjLitExpr(p.parseObjectLiteral()))
	case token.LeftBracket:
		return p.alloc.Expression(ast.NewArrLitExpr(p.parseArrayLiteral()))
	case token.LeftParenthesis:
		return p.parseParenthesisedExpression()
	case token.NoSubstitutionTemplate, token.TemplateHead:
		return p.alloc.Expression(ast.NewTmplLitExpr(p.parseTemplateLiteral(false)))
	case token.This:
		p.next()
		return p.alloc.Expression(ast.NewThisExpr(ast.ThisExpression{Idx: idx}))
	case token.Super:
		return p.parseSuperProperty()
	case token.Async:
		if f := p.parseMaybeAsyncFunction(false); f != nil {
			return p.alloc.Expression(ast.NewFuncLitExpr(f))
		}
	case token.Function:
		return p.alloc.Expression(ast.NewFuncLitExpr(p.parseFunction(false, false, idx)))
	case token.Class:
		return p.alloc.Expression(ast.NewClassLitExpr(p.parseClass(false)))
	}

	if p.isBindingId(p.currentKind()) {
		p.next()
		return p.alloc.Expression(ast.NewIdentExpr(p.alloc.Identifier(idx, "")))
	}

	p.errorUnexpectedToken(p.currentKind())
	p.nextStatement()
	return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: idx, To: p.currentOffset()}))
}

func (p *parser) parseSuperProperty() *ast.Expression {
	idx := p.currentOffset()
	p.next()
	switch p.currentKind() {
	case token.Period:
		p.next()
		if !token.ID(p.currentKind()) {
			p.expect(token.Identifier)
			p.nextStatement()
			return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: idx, To: p.currentOffset()}))
		}
		idIdx := p.currentOffset()
		parsedLiteral := p.currentString()
		p.next()
		return p.alloc.Expression(ast.NewMemberExpr(p.alloc.MemberExpression(
			p.alloc.Expression(ast.NewSuperExpr(ast.SuperExpression{Idx: idx})),
			p.alloc.MemberProperty(ast.NewIdentMemProp(p.alloc.Identifier(idIdx, parsedLiteral))),
		)))
	case token.LeftBracket:
		return p.parseBracketMember(p.alloc.Expression(ast.NewSuperExpr(ast.SuperExpression{Idx: idx})))
	case token.LeftParenthesis:
		return p.parseCallExpression(p.alloc.Expression(ast.NewSuperExpr(ast.SuperExpression{Idx: idx})))
	default:
		p.errorf("'super' keyword unexpected here")
		p.nextStatement()
		return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: idx, To: p.currentOffset()}))
	}
}

func (p *parser) reinterpretSequenceAsArrowFuncParams(list ast.Expressions) *ast.ParameterList {
	firstRestIdx := -1
	mark := len(p.declBuf)
	for i := range list {
		if list[i].IsSpread() {
			if firstRestIdx == -1 {
				firstRestIdx = i
				continue
			}
		}
		if firstRestIdx != -1 {
			p.errorf("Rest parameter must be last formal parameter")
			p.declBuf = p.declBuf[:mark]
			return p.alloc.ParameterList(ast.ParameterList{})
		}
		p.declBuf = append(p.declBuf, p.reinterpretAsBinding(&list[i]))
	}
	var rest *ast.Expression
	if firstRestIdx != -1 {
		rest = p.reinterpretAsBindingRestElement(&list[firstRestIdx])
	}
	return p.alloc.ParameterList(ast.ParameterList{
		List: p.finishDeclBuf(mark),
		Rest: rest,
	})
}

func (p *parser) parseParenthesisedExpression() *ast.Expression {
	opening := p.currentOffset()
	p.expect(token.LeftParenthesis)
	mark := len(p.exprBuf)
	if p.currentKind() != token.RightParenthesis {
		for {
			if p.currentKind() == token.Ellipsis {
				start := p.currentOffset()
				p.errorUnexpectedToken(token.Ellipsis)
				p.next()
				expr := p.parseAssignmentExpression()
				p.exprBuf = append(p.exprBuf, ast.NewInvalidExpr(ast.InvalidExpression{From: start, To: expr.Idx1()}))
			} else {
				p.exprBuf = append(p.exprBuf, *p.parseAssignmentExpression())
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
	n := len(p.exprBuf) - mark
	if n == 1 && p.errors == nil {
		result := p.exprBuf[mark]
		p.exprBuf = p.exprBuf[:mark]
		return p.alloc.Expression(result)
	}
	if n == 0 {
		p.exprBuf = p.exprBuf[:mark]
		p.errorUnexpectedToken(token.RightParenthesis)
		return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: opening, To: p.currentOffset()}))
	}
	return p.alloc.Expression(ast.NewSequenceExpr(p.alloc.SequenceExpression(p.finishExprBuf(mark))))
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

func (p *parser) parseBindingTarget() (target ast.BindingTarget) {
	p.tokenToBindingId()
	switch p.currentKind() {
	case token.Identifier:
		target = ast.NewIdentBindingTarget(p.alloc.Identifier(p.currentOffset(), p.currentString()))
		p.next()
	case token.LeftBracket:
		target = p.parseArrayBindingPattern()
	case token.LeftBrace:
		target = p.parseObjectBindingPattern()
	default:
		idx := p.expect(token.Identifier)
		p.nextStatement()
		target = ast.NewInvalidBindingTarget(ast.InvalidExpression{From: idx, To: p.currentOffset()})
	}

	return
}

func (p *parser) parseVariableDeclaration(declarationList *ast.VariableDeclarators) ast.VariableDeclarator {
	node := p.alloc.VariableDeclarator(p.alloc.BindingTarget(p.parseBindingTarget()))

	if p.currentKind() == token.Assign {
		p.next()
		node.Initializer = p.parseAssignmentExpression()
	}

	if declarationList != nil {
		*declarationList = append(*declarationList, *node)
	}

	return *node
}

func (p *parser) parseVariableDeclarationList() ast.VariableDeclarators {
	mark := len(p.declBuf)
	for {
		p.declBuf = append(p.declBuf, p.parseVariableDeclaration(nil))
		if p.currentKind() != token.Comma {
			break
		}
		p.next()
	}
	return p.finishDeclBuf(mark)
}

func (p *parser) parseObjectPropertyKey() (string, string, *ast.Expression, token.Token) {
	if p.currentKind() == token.LeftBracket {
		p.next()
		expr := p.parseAssignmentExpression()
		p.expect(token.RightBracket)
		return "", "", expr, token.Illegal
	}
	idx, tkn, literal, parsedLiteral := p.currentOffset(), p.currentKind(), p.token.Raw(p.scanner), p.currentString()
	var value *ast.Expression
	p.next()
	switch tkn {
	case token.Identifier, token.String, token.Keyword, token.EscapedReservedWord:
		value = p.alloc.Expression(ast.NewStrLitExpr(p.alloc.StringLiteral(idx, parsedLiteral, literal)))
	case token.Number:
		num, err := parseNumberLiteral(literal)
		if err != nil {
			p.errorf("%s", err.Error())
		} else {
			value = p.alloc.Expression(ast.NewNumLitExpr(p.alloc.NumberLiteral(idx, num, literal)))
		}
	case token.PrivateIdentifier:
		value = p.alloc.Expression(ast.NewPrivIdentExpr(p.alloc.PrivateIdentifier(p.alloc.Identifier(idx, parsedLiteral))))
	default:
		if token.ID(tkn) {
			value = p.alloc.Expression(ast.NewStrLitExpr(p.alloc.StringLiteral(idx, literal, literal)))
		} else {
			p.errorUnexpectedToken(tkn)
		}
	}
	return literal, parsedLiteral, value, tkn
}

func (p *parser) parseObjectProperty() ast.Property {
	if p.currentKind() == token.Ellipsis {
		p.next()
		return ast.NewSpreadProp(p.alloc.SpreadElement(p.parseAssignmentExpression()))
	}
	keyStartIdx := p.currentOffset()
	generator := false
	if p.currentKind() == token.Multiply {
		generator = true
		p.next()
	}
	literal, parsedLiteral, value, tkn := p.parseObjectPropertyKey()
	if value == nil {
		return ast.Property{}
	}
	if token.ID(tkn) || tkn == token.String || tkn == token.Number || tkn == token.Illegal {
		if generator {
			return ast.NewKeyedProp(p.alloc.PropertyKeyed(
				value,
				ast.PropertyKindMethod,
				p.alloc.Expression(ast.NewFuncLitExpr(p.parseMethodDefinition(keyStartIdx, ast.PropertyKindMethod, true, false))),
				tkn == token.Illegal,
			))
		}
		switch {
		case p.currentKind() == token.LeftParenthesis:
			return ast.NewKeyedProp(p.alloc.PropertyKeyed(
				value,
				ast.PropertyKindMethod,
				p.alloc.Expression(ast.NewFuncLitExpr(p.parseMethodDefinition(keyStartIdx, ast.PropertyKindMethod, false, false))),
				tkn == token.Illegal,
			))
		case p.currentKind() == token.Comma || p.currentKind() == token.RightBrace || p.currentKind() == token.Assign:
			if p.isBindingId(tkn) {
				var initializer *ast.Expression
				if p.currentKind() == token.Assign {
					p.next()
					initializer = p.parseAssignmentExpression()
				}
				return ast.NewShortProp(p.alloc.PropertyShort(
					p.alloc.Identifier(value.Idx0(), parsedLiteral),
					initializer,
				))
			} else {
				p.errorUnexpectedToken(p.currentKind())
			}
		case (literal == "get" || literal == "set" || tkn == token.Async) && p.currentKind() != token.Colon:
			_, _, keyValue, tkn1 := p.parseObjectPropertyKey()
			if keyValue == nil {
				return ast.Property{}
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

			return ast.NewKeyedProp(p.alloc.PropertyKeyed(
				keyValue,
				kind,
				p.alloc.Expression(ast.NewFuncLitExpr(p.parseMethodDefinition(keyStartIdx, kind, false, async))),
				tkn1 == token.Illegal,
			))
		}
	}

	p.expect(token.Colon)
	return ast.NewKeyedProp(p.alloc.PropertyKeyed(
		value,
		ast.PropertyKindValue,
		p.parseAssignmentExpression(),
		tkn == token.Illegal,
	))
}

func (p *parser) parseMethodDefinition(keyStartIdx ast.Idx, kind ast.PropertyKind, generator, async bool) *ast.FunctionLiteral {
	savedYield := p.scope.allowYield
	savedAwait := p.scope.allowAwait
	if generator != savedYield {
		p.scope.allowYield = generator
	}
	if async != savedAwait {
		p.scope.allowAwait = async
	}
	parameterList := p.parseFunctionParameterList()
	switch kind {
	case ast.PropertyKindGet:
		if len(parameterList.List) > 0 || parameterList.Rest != nil {
			p.errorf("Getter must not have any formal parameters.")
		}
	case ast.PropertyKindSet:
		if len(parameterList.List) != 1 || parameterList.Rest != nil {
			p.errorf("Setter must have exactly one formal parameter.")
		}
	}
	node := p.alloc.FunctionLiteral(keyStartIdx, async)
	node.ParameterList = parameterList
	node.Generator = generator
	node.Body = p.parseFunctionBlock(async, async, generator)
	p.scope.allowYield = savedYield
	p.scope.allowAwait = savedAwait
	return node
}

func (p *parser) parseObjectLiteral() *ast.ObjectLiteral {
	mark := len(p.propBuf)
	idx0 := p.expect(token.LeftBrace)
	for p.currentKind() != token.RightBrace && p.currentKind() != token.Eof {
		property := p.parseObjectProperty()
		if !property.IsNone() {
			p.propBuf = append(p.propBuf, property)
		}
		if p.currentKind() != token.RightBrace {
			p.expect(token.Comma)
		} else {
			break
		}
	}
	idx1 := p.expect(token.RightBrace)

	return p.alloc.ObjectLiteral(idx0, idx1, p.finishPropBuf(mark))
}

func (p *parser) parseArrayLiteral() *ast.ArrayLiteral {
	idx0 := p.expect(token.LeftBracket)
	mark := len(p.exprBuf)
	for p.currentKind() != token.RightBracket && p.currentKind() != token.Eof {
		if p.currentKind() == token.Comma {
			p.next()
			p.exprBuf = append(p.exprBuf, ast.Expression{})
			continue
		}
		if p.currentKind() == token.Ellipsis {
			p.next()
			p.exprBuf = append(p.exprBuf, ast.NewSpreadExpr(p.alloc.SpreadElement(
				p.parseAssignmentExpression(),
			)))
		} else {
			p.exprBuf = append(p.exprBuf, *p.parseAssignmentExpression())
		}
		if p.currentKind() != token.RightBracket {
			p.expect(token.Comma)
		}
	}
	idx1 := p.expect(token.RightBracket)

	return p.alloc.ArrayLiteral(idx0, idx1, p.finishExprBuf(mark))
}

func (p *parser) parseTemplateLiteral(tagged bool) *ast.TemplateLiteral {
	res := p.alloc.TemplateLiteral(p.currentOffset())
	mark := len(p.exprBuf)

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

		p.next()
		expr := p.parseExpression()
		p.exprBuf = append(p.exprBuf, *expr)

		if p.currentKind() != token.RightBrace {
			p.errorUnexpectedToken(p.currentKind())
			break
		}
		p.token = p.scanner.NextTemplatePart()
	}
	res.Expressions = p.finishExprBuf(mark)
	return res
}

func (p *parser) parseTaggedTemplateLiteral(tag *ast.Expression) *ast.TemplateLiteral {
	l := p.parseTemplateLiteral(true)
	l.Tag = tag
	return l
}

func (p *parser) parseArgumentList() (argumentList ast.Expressions, idx0, idx1 ast.Idx) {
	idx0 = p.expect(token.LeftParenthesis)
	mark := len(p.exprBuf)
	for p.currentKind() != token.RightParenthesis {
		if p.currentKind() == token.Ellipsis {
			p.next()
			p.exprBuf = append(p.exprBuf, ast.NewSpreadExpr(p.alloc.SpreadElement(p.parseAssignmentExpression())))
		} else {
			p.exprBuf = append(p.exprBuf, *p.parseAssignmentExpression())
		}
		if p.currentKind() != token.Comma {
			break
		}
		p.next()
	}
	idx1 = p.expect(token.RightParenthesis)
	argumentList = p.finishExprBuf(mark)
	return
}

func (p *parser) parseCallExpression(left *ast.Expression) *ast.Expression {
	argumentList, idx0, idx1 := p.parseArgumentList()
	return p.alloc.Expression(ast.NewCallExpr(p.alloc.CallExpression(left, idx0, argumentList, idx1)))
}

func (p *parser) parseDotMember(left *ast.Expression) *ast.Expression {
	period := p.currentOffset()
	p.next()

	literal := p.currentString()
	idx := p.currentOffset()

	if p.currentKind() == token.PrivateIdentifier {
		p.next()
		return p.alloc.Expression(ast.NewPrivDotExpr(p.alloc.PrivateDotExpression(
			left,
			p.alloc.PrivateIdentifier(p.alloc.Identifier(idx, literal)),
		)))
	}

	if !token.ID(p.currentKind()) {
		p.expect(token.Identifier)
		p.nextStatement()
		return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: period, To: p.currentOffset()}))
	}

	p.next()

	return p.alloc.Expression(ast.NewMemberExpr(p.alloc.MemberExpression(
		left,
		p.alloc.MemberProperty(ast.NewIdentMemProp(p.alloc.Identifier(idx, literal))),
	)))
}

func (p *parser) parseBracketMember(left *ast.Expression) *ast.Expression {
	p.expect(token.LeftBracket)
	member := p.parseExpression()
	p.expect(token.RightBracket)
	return p.alloc.Expression(ast.NewMemberExpr(p.alloc.MemberExpression(
		left,
		p.alloc.MemberProperty(ast.NewComputedMemProp(p.alloc.ComputedProperty(member))),
	)))
}

func (p *parser) parseNewExpression() *ast.Expression {
	idx := p.expect(token.New)
	if p.currentKind() == token.Period {
		p.next()
		if p.currentString() == "target" {
			return p.alloc.Expression(ast.NewMetaPropExpr(p.alloc.MetaProperty(
				p.alloc.Identifier(idx, token.New.String()),
				p.parseIdentifier(),
				idx,
			)))
		}
		p.errorUnexpectedToken(token.Identifier)
	}
	callee := p.parseLeftHandSideExpression()
	if bad, ok := callee.Invalid(); ok {
		bad.From = idx
		return callee
	}
	node := p.alloc.NewExpression(idx, callee)
	if p.currentKind() == token.LeftParenthesis {
		argumentList, idx0, idx1 := p.parseArgumentList()
		node.ArgumentList = argumentList
		node.LeftParenthesis = idx0
		node.RightParenthesis = idx1
	}
	return p.alloc.Expression(ast.NewNewExpr(node))
}

func (p *parser) parseLeftHandSideExpression() *ast.Expression {
	var left *ast.Expression
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
			left = p.alloc.Expression(ast.NewTmplLitExpr(p.parseTaggedTemplateLiteral(left)))
		default:
			break L
		}
	}

	return left
}

func (p *parser) parseLeftHandSideExpressionAllowCall() *ast.Expression {
	allowIn := p.scope.allowIn
	p.scope.allowIn = true

	var left *ast.Expression
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
				p.errorf("Invalid template literal on optional chain")
				p.nextStatement()
				p.scope.allowIn = allowIn
				return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: start, To: p.currentOffset()}))
			}
			left = p.alloc.Expression(ast.NewTmplLitExpr(p.parseTaggedTemplateLiteral(left)))
		case token.QuestionDot:
			optionalChain = true
			left = p.alloc.Expression(ast.NewOptionalExpr(p.alloc.Optional(left)))

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
		left = p.alloc.Expression(ast.NewOptChainExpr(p.alloc.OptionalChain(left)))
	}
	p.scope.allowIn = allowIn
	return left
}

func (p *parser) parseUpdateExpression() *ast.Expression {
	switch p.currentKind() {
	case token.Increment, token.Decrement:
		tkn := p.currentKind()
		idx := p.currentOffset()
		p.next()
		operand := p.parseUnaryExpression()
		switch operand.Kind() {
		case ast.ExprIdent, ast.ExprPrivDot, ast.ExprMember:
		default:
			p.errorf("Invalid left-hand side in assignment")
			p.nextStatement()
			return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: idx, To: p.currentOffset()}))
		}
		return p.alloc.Expression(ast.NewUpdateExpr(p.alloc.UpdateExpression(tkn, idx, operand, false)))
	default:
		operand := p.parseLeftHandSideExpressionAllowCall()
		if p.currentKind() == token.Increment || p.currentKind() == token.Decrement {
			if p.token.OnNewLine {
				return operand
			}
			tkn := p.currentKind()
			idx := p.currentOffset()
			p.next()
			switch operand.Kind() {
			case ast.ExprIdent, ast.ExprPrivDot, ast.ExprMember:
			default:
				p.errorf("Invalid left-hand side in assignment")
				p.nextStatement()
				return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: idx, To: p.currentOffset()}))
			}
			return p.alloc.Expression(ast.NewUpdateExpr(p.alloc.UpdateExpression(tkn, idx, operand, true)))
		}
		return operand
	}
}

func (p *parser) parseUnaryExpression() *ast.Expression {
	switch p.currentKind() {
	case token.Plus, token.Minus, token.Not, token.BitwiseNot:
		fallthrough
	case token.Delete, token.Void, token.Typeof:
		tkn := p.currentKind()
		idx := p.currentOffset()
		p.next()
		return p.alloc.Expression(ast.NewUnaryExpr(p.alloc.UnaryExpression(tkn, idx, p.parseUnaryExpression())))
	case token.Await:
		if p.scope.allowAwait {
			idx := p.currentOffset()
			p.next()
			if !p.scope.inAsync {
				p.errorUnexpectedToken(token.Await)
				return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: idx, To: p.currentOffset()}))
			}
			if p.scope.inFuncParams {
				p.errorf("Illegal await-expression in formal parameters of async function")
			}
			return p.alloc.Expression(ast.NewAwaitExpr(p.alloc.AwaitExpression(idx, p.parseUnaryExpression())))
		}
	}

	return p.parseUpdateExpression()
}

func (p *parser) parseBinaryExpressionOrHigher(minPrecedence Precedence) *ast.Expression {
	lhsParenthesized := p.currentKind() == token.LeftParenthesis

	var lhs *ast.Expression
	if p.scope.allowIn && p.currentKind() == token.PrivateIdentifier {
		lhs = p.parsePrivateInExpression(minPrecedence)
	} else {
		lhs = p.parseUnaryExpression()
	}

	return p.parseBinaryExpressionRest(lhs, lhsParenthesized, minPrecedence)
}

func (p *parser) parseBinaryExpressionRest(lhs *ast.Expression, lhsParenthesized bool, minPrecedence Precedence) *ast.Expression {
	for {
		kind := p.currentKind()

		lbp := kindToPrecedence(kind)

		if lbp <= minPrecedence {
			break
		}

		if kind == token.In && !p.scope.allowIn {
			break
		}

		p.next()

		rhsParenthesized := p.currentKind() == token.LeftParenthesis
		rhs := p.parseBinaryExpressionOrHigher(lbp ^ 1)

		if isLogicalOperator(kind) {
			if kind == token.Coalesce {
				if bexp, ok := rhs.Binary(); ok && !rhsParenthesized {
					if bexp.Operator == token.LogicalAnd || bexp.Operator == token.LogicalOr {
						p.errorf("Logical expressions and coalesce expressions cannot be mixed. Wrap either by parentheses")
					}
				}
				if bexp, ok := lhs.Binary(); ok && !lhsParenthesized {
					if bexp.Operator == token.LogicalAnd || bexp.Operator == token.LogicalOr {
						p.errorf("Logical expressions and coalesce expressions cannot be mixed. Wrap either by parentheses")
					}
				}
			}
		} else {
			if kind == token.Exponent && !lhsParenthesized {
				switch lhs.Kind() {
				case ast.ExprUnary, ast.ExprAwait:
					p.errorf("Unary operator used immediately before exponentiation expression. Parenthesis must be used to disambiguate operator precedence")
				}
			}
		}
		lhs = p.alloc.Expression(ast.NewBinaryExpr(p.alloc.BinaryExpression(kind, lhs, rhs)))

		lhsParenthesized = false
	}

	return lhs
}

func (p *parser) parsePrivateInExpression(minPrecedence Precedence) *ast.Expression {
	left := p.alloc.Expression(ast.NewPrivIdentExpr(p.alloc.PrivateIdentifier(p.alloc.Identifier(p.currentOffset(), p.currentString()))))
	p.next()

	if p.currentKind() != token.In || PrecedenceCompare <= minPrecedence {
		return left
	}

	p.next()
	rhs := p.parseBinaryExpressionOrHigher(PrecedenceCompare)
	return p.alloc.Expression(ast.NewBinaryExpr(p.alloc.BinaryExpression(token.In, left, rhs)))
}

func (p *parser) parseConditionalExpression() *ast.Expression {
	left := p.parseBinaryExpressionOrHigher(PrecedenceLowest)

	if p.currentKind() == token.QuestionMark {
		p.next()
		allowIn := p.scope.allowIn
		p.scope.allowIn = true
		consequent := p.parseAssignmentExpression()
		p.scope.allowIn = allowIn
		p.expect(token.Colon)
		return p.alloc.Expression(ast.NewConditionalExpr(p.alloc.ConditionalExpression(
			left,
			consequent,
			p.parseAssignmentExpression(),
		)))
	}

	return left
}

func (p *parser) parseArrowFunction(start ast.Idx, paramList *ast.ParameterList, async bool) *ast.Expression {
	p.expect(token.Arrow)
	node := p.alloc.ArrowFunctionLiteral(start, paramList, async)
	node.Body = p.parseArrowFunctionBody(async)
	return p.alloc.Expression(ast.NewArrowFuncLitExpr(node))
}

func (p *parser) parseSingleArgArrowFunction(start ast.Idx, async bool) *ast.Expression {
	savedAwait := p.scope.allowAwait
	if async != savedAwait {
		p.scope.allowAwait = async
	}
	p.tokenToBindingId()
	if p.currentKind() != token.Identifier {
		p.errorUnexpectedToken(p.currentKind())
		p.next()
		p.scope.allowAwait = savedAwait
		return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: start, To: p.currentOffset()}))
	}

	id := p.parseIdentifier()

	paramList := p.alloc.ParameterList(ast.ParameterList{
		Opening: id.Idx,
		Closing: id.Idx1(),
		List: ast.VariableDeclarators{{
			Target: p.alloc.BindingTarget(ast.NewIdentBindingTarget(id)),
		}},
	})

	result := p.parseArrowFunction(start, paramList, async)
	p.scope.allowAwait = savedAwait
	return result
}

func (p *parser) parseAssignmentExpression() *ast.Expression {
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
			p.next()
			return p.parseSingleArgArrowFunction(start, true)
		} else if tok == token.LeftParenthesis {
			state = p.mark()
			async = true
		}
	case token.Yield:
		if p.scope.allowYield {
			return p.alloc.Expression(ast.NewYieldExpr(p.parseYieldExpression()))
		}
		fallthrough
	default:
		p.tokenToBindingId()
	}
	left := p.parseConditionalExpression()
	kind := p.currentKind()
	operator := assignToOperator[kind]
	if operator == 0 && kind == token.Arrow {
		var paramList *ast.ParameterList
		if id, ok := left.Ident(); ok {
			paramList = p.alloc.ParameterList(ast.ParameterList{
				Opening: id.Idx,
				Closing: id.Idx1() - 1,
				List: ast.VariableDeclarators{{
					Target: p.alloc.BindingTarget(ast.NewIdentBindingTarget(id)),
				}},
			})
		} else if parenthesis {
			if seq, ok := left.Sequence(); ok && p.errors == nil {
				paramList = p.reinterpretSequenceAsArrowFuncParams(seq.Sequence)
			} else {
				p.restore(state)
				paramList = p.parseFunctionParameterList()
			}
		} else if async {
			savedAwait := p.scope.allowAwait
			if !savedAwait {
				p.scope.allowAwait = true
			}
			if left.IsCall() {
				p.restore(state)
				p.next()
				paramList = p.parseFunctionParameterList()
			}
			if paramList == nil {
				p.errorf("Malformed arrow function parameter list")
				p.scope.allowAwait = savedAwait
				return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: left.Idx0(), To: left.Idx1()}))
			}
			result := p.parseArrowFunction(start, paramList, async)
			p.scope.allowAwait = savedAwait
			return result
		}
		if paramList == nil {
			p.errorf("Malformed arrow function parameter list")
			return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: left.Idx0(), To: left.Idx1()}))
		}
		return p.parseArrowFunction(start, paramList, async)
	}

	if operator != 0 {
		idx := p.currentOffset()
		p.next()
		ok := false
		switch left.Kind() {
		case ast.ExprIdent, ast.ExprPrivDot, ast.ExprMember:
			ok = true
		case ast.ExprArrLit:
			if !parenthesis && operator == token.Assign {
				left = p.reinterpretAsArrayAssignmentPattern(left.MustArrLit())
				ok = true
			}
		case ast.ExprObjLit:
			if !parenthesis && operator == token.Assign {
				left = p.reinterpretAsObjectAssignmentPattern(left.MustObjLit())
				ok = true
			}
		}
		if ok {
			return p.alloc.Expression(ast.NewAssignExpr(p.alloc.AssignExpression(operator, left, p.parseAssignmentExpression())))
		}
		p.errorf("Invalid left-hand side in assignment")
		p.nextStatement()
		return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: idx, To: p.currentOffset()}))
	}

	return left
}

func (p *parser) parseYieldExpression() *ast.YieldExpression {
	idx := p.expect(token.Yield)

	if p.scope.inFuncParams {
		p.errorf("Yield expression not allowed in formal parameter")
	}

	node := p.alloc.YieldExpression(idx)

	if !p.token.OnNewLine && p.currentKind() == token.Multiply {
		node.Delegate = true
		p.next()
	}

	if !p.canInsertSemicolon() {
		state := p.mark()
		expr := p.parseAssignmentExpression()
		if expr.IsInvalid() {
			expr = p.alloc.Expression(ast.Expression{})
			p.restore(state)
		}
		node.Argument = expr
	}

	return node
}

func (p *parser) parseExpression() *ast.Expression {
	left := p.parseAssignmentExpression()

	if p.currentKind() == token.Comma {
		mark := len(p.exprBuf)
		p.exprBuf = append(p.exprBuf, *left)
		for {
			if p.currentKind() != token.Comma {
				break
			}
			p.next()
			p.exprBuf = append(p.exprBuf, *p.parseAssignmentExpression())
		}
		return p.alloc.Expression(ast.NewSequenceExpr(p.alloc.SequenceExpression(p.finishExprBuf(mark))))
	}

	return left
}

func (p *parser) checkComma(from, to ast.Idx) {
	if pos := strings.IndexByte(p.str[int(from)-1:int(to)-1], ','); pos >= 0 {
		p.errorf("Comma is not allowed here")
	}
}

func (p *parser) reinterpretAsArrayAssignmentPattern(left *ast.ArrayLiteral) *ast.Expression {
	value := left.Value
	var rest *ast.Expression
	for i := range value {
		if spread, ok := value[i].Spread(); ok {
			if i != len(value)-1 {
				p.errorf("Rest element must be last element")
				return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: left.Idx0(), To: left.Idx1()}))
			}
			p.checkComma(spread.Idx1(), left.RightBracket)
			rest = p.reinterpretAsDestructAssignTarget(spread.Expression)
			value = value[:len(value)-1]
		} else {
			result := p.reinterpretAsAssignmentElement(&value[i])
			value[i] = *result
		}
	}
	return p.alloc.Expression(ast.NewArrPatExpr(p.alloc.ArrayPattern(left.LeftBracket, left.RightBracket, value, rest)))
}

func (p *parser) reinterpretArrayAssignPatternAsBinding(pattern *ast.ArrayPattern) *ast.ArrayPattern {
	for i := range pattern.Elements {
		result := p.reinterpretAsDestructBindingTarget(&pattern.Elements[i])
		pattern.Elements[i] = *result
	}
	if pattern.Rest != nil {
		pattern.Rest = p.reinterpretAsDestructBindingTarget(pattern.Rest)
	}
	return pattern
}

func (p *parser) reinterpretAsArrayBindingPattern(left *ast.ArrayLiteral) *ast.Expression {
	value := left.Value
	var rest *ast.Expression
	for i := range value {
		if spread, ok := value[i].Spread(); ok {
			if i != len(value)-1 {
				p.errorf("Rest element must be last element")
				return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: left.Idx0(), To: left.Idx1()}))
			}
			p.checkComma(spread.Idx1(), left.RightBracket)
			rest = p.reinterpretAsDestructBindingTarget(spread.Expression)
			value = value[:len(value)-1]
		} else {
			result := p.reinterpretAsBindingElement(&value[i])
			value[i] = *result
		}
	}
	return p.alloc.Expression(ast.NewArrPatExpr(p.alloc.ArrayPattern(left.LeftBracket, left.RightBracket, value, rest)))
}

func (p *parser) parseArrayBindingPattern() ast.BindingTarget {
	result := p.reinterpretAsArrayBindingPattern(p.parseArrayLiteral())
	return ast.BindingTargetFromExpression(result)
}

func (p *parser) parseObjectBindingPattern() ast.BindingTarget {
	result := p.reinterpretAsObjectBindingPattern(p.parseObjectLiteral())
	return ast.BindingTargetFromExpression(result)
}

func (p *parser) reinterpretArrayObjectPatternAsBinding(pattern *ast.ObjectPattern) *ast.ObjectPattern {
	for _, prop := range pattern.Properties {
		if keyed, ok := prop.Keyed(); ok {
			keyed.Value = p.reinterpretAsBindingElement(keyed.Value)
		}
	}
	if pattern.Rest != nil {
		pattern.Rest = p.reinterpretAsBindingRestElement(pattern.Rest)
	}
	return pattern
}

func (p *parser) reinterpretAsObjectBindingPattern(expr *ast.ObjectLiteral) *ast.Expression {
	var rest *ast.Expression
	value := expr.Value
	for i, prop := range value {
		ok := false
		switch prop.Kind() {
		case ast.PropKeyed:
			keyed := prop.MustKeyed()
			if keyed.Kind == ast.PropertyKindValue {
				keyed.Value = p.reinterpretAsBindingElement(keyed.Value)
				ok = true
			}
		case ast.PropShort:
			ok = true
		case ast.PropSpread:
			spread := prop.MustSpread()
			if i != len(expr.Value)-1 {
				p.errorf("Rest element must be last element")
				return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()}))
			}
			rest = p.reinterpretAsBindingRestElement(spread.Expression)
			value = value[:i]
			ok = true
		}
		if !ok {
			p.errorf("Invalid destructuring binding target")
			return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()}))
		}
	}
	return p.alloc.Expression(ast.NewObjPatExpr(p.alloc.ObjectPattern(expr.LeftBrace, expr.RightBrace, value, rest)))
}

func (p *parser) reinterpretAsObjectAssignmentPattern(l *ast.ObjectLiteral) *ast.Expression {
	var rest *ast.Expression
	value := l.Value
	for i, prop := range value {
		ok := false
		switch prop.Kind() {
		case ast.PropKeyed:
			keyed := prop.MustKeyed()
			if keyed.Kind == ast.PropertyKindValue {
				keyed.Value = p.reinterpretAsAssignmentElement(keyed.Value)
				ok = true
			}
		case ast.PropShort:
			ok = true
		case ast.PropSpread:
			spread := prop.MustSpread()
			if i != len(l.Value)-1 {
				p.errorf("Rest element must be last element")
				return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: l.Idx0(), To: l.Idx1()}))
			}
			rest = spread.Expression
			value = value[:i]
			ok = true
		}
		if !ok {
			p.errorf("Invalid destructuring assignment target")
			return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: l.Idx0(), To: l.Idx1()}))
		}
	}
	return p.alloc.Expression(ast.NewObjPatExpr(p.alloc.ObjectPattern(l.LeftBrace, l.RightBrace, value, rest)))
}

func (p *parser) reinterpretAsAssignmentElement(expr *ast.Expression) *ast.Expression {
	switch expr.Kind() {
	case ast.ExprAssign:
		e := expr.MustAssign()
		if e.Operator == token.Assign {
			e.Left = p.reinterpretAsDestructAssignTarget(e.Left)
			return expr
		}
		p.errorf("Invalid destructuring assignment target")
		return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()}))
	default:
		return p.reinterpretAsDestructAssignTarget(expr)
	}
}

func (p *parser) reinterpretAsBindingElement(expr *ast.Expression) *ast.Expression {
	switch expr.Kind() {
	case ast.ExprAssign:
		e := expr.MustAssign()
		if e.Operator == token.Assign {
			e.Left = p.reinterpretAsDestructBindingTarget(e.Left)
			return expr
		}
		p.errorf("Invalid destructuring assignment target")
		return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()}))
	default:
		return p.reinterpretAsDestructBindingTarget(expr)
	}
}

func (p *parser) reinterpretAsBinding(expr *ast.Expression) ast.VariableDeclarator {
	switch expr.Kind() {
	case ast.ExprAssign:
		e := expr.MustAssign()
		if e.Operator == token.Assign {
			target := p.reinterpretAsDestructBindingTarget(e.Left)
			return ast.VariableDeclarator{
				Target:      p.alloc.BindingTarget(ast.BindingTargetFromExpression(target)),
				Initializer: e.Right,
			}
		}
		p.errorf("Invalid destructuring assignment target")
		return ast.VariableDeclarator{
			Target: p.alloc.BindingTarget(ast.NewInvalidBindingTarget(ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()})),
		}
	default:
		target := p.reinterpretAsDestructBindingTarget(expr)
		return ast.VariableDeclarator{
			Target: p.alloc.BindingTarget(ast.BindingTargetFromExpression(target)),
		}
	}
}

func (p *parser) reinterpretAsDestructAssignTarget(item *ast.Expression) *ast.Expression {
	if item == nil || item.IsNone() {
		return nil
	}
	switch item.Kind() {
	case ast.ExprArrLit:
		return p.reinterpretAsArrayAssignmentPattern(item.MustArrLit())
	case ast.ExprObjLit:
		return p.reinterpretAsObjectAssignmentPattern(item.MustObjLit())
	case ast.ExprArrPat, ast.ExprObjPat, ast.ExprIdent, ast.ExprPrivDot, ast.ExprMember:
		return item
	}
	p.errorf("Invalid destructuring assignment target")
	return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: item.Idx0(), To: item.Idx1()}))
}

func (p *parser) reinterpretAsDestructBindingTarget(item *ast.Expression) *ast.Expression {
	if item == nil || item.IsNone() {
		return item
	}
	switch item.Kind() {
	case ast.ExprArrPat:
		p.reinterpretArrayAssignPatternAsBinding(item.MustArrPat())
		return item
	case ast.ExprObjPat:
		p.reinterpretArrayObjectPatternAsBinding(item.MustObjPat())
		return item
	case ast.ExprArrLit:
		return p.reinterpretAsArrayBindingPattern(item.MustArrLit())
	case ast.ExprObjLit:
		return p.reinterpretAsObjectBindingPattern(item.MustObjLit())
	case ast.ExprIdent:
		id := item.MustIdent()
		if !p.scope.allowAwait || id.Name != "await" {
			return item
		}
	}
	p.errorf("Invalid destructuring binding target")
	return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: item.Idx0(), To: item.Idx1()}))
}

func (p *parser) reinterpretAsBindingRestElement(expr *ast.Expression) *ast.Expression {
	if expr.IsIdent() {
		return expr
	}
	p.errorf("Invalid binding rest")
	return p.alloc.Expression(ast.NewInvalidExpr(ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()}))
}
