package parser

import (
	"math/big"
	"strings"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser/scanner/token"
)

func (p *parser) parseIdentifier() *ast.Identifier {
	literal := p.currentString()
	idx := p.currentOffset()
	p.next()
	return p.alloc.Identifier(idx, literal)
}

func (p *parser) parsePrimaryExpression() ast.Expression {
	idx := p.currentOffset()
	switch p.currentKind() {
	case token.Identifier:
		parsedLiteral := p.currentString()
		p.next()
		return ast.NewIdentExpr(p.alloc.Identifier(idx, parsedLiteral))
	case token.Null:
		p.next()
		return ast.NewNullLitExpr(p.alloc.NullLiteral(idx))
	case token.Boolean:
		value := p.scanner.Token.Idx1-p.scanner.Token.Idx0 == 4 // "true" = 4 chars, "false" = 5
		p.next()
		return ast.NewBoolLitExpr(p.alloc.BooleanLiteral(idx, value))
	case token.String:
		parsedLiteral := p.currentString()
		raw := p.scanner.Token.Raw(p.scanner)
		p.next()
		return ast.NewStrLitExpr(p.alloc.StringLiteral(idx, parsedLiteral, raw))
	case token.Number:
		parsedLiteral := p.currentString()
		raw := p.scanner.Token.Raw(p.scanner)
		p.next()
		if isBigIntLiteral(parsedLiteral) {
			value, err := parseBigIntLiteral(parsedLiteral)
			if err != nil {
				p.errorf("%s", err.Error())
				value = new(big.Int)
			}
			return ast.NewBigIntLitExpr(p.alloc.BigIntLiteral(idx, value, raw))
		}
		value, err := parseNumberLiteral(parsedLiteral)
		if err != nil {
			p.errorf("%s", err.Error())
			value = 0
		}
		return ast.NewNumLitExpr(p.alloc.NumberLiteral(idx, value, raw))
	case token.Slash, token.QuotientAssign:
		pat, flags, lit := p.scanner.ParseRegExp()
		p.next()
		return ast.NewRegExpLitExpr(p.alloc.RegExpLiteral(idx, lit, pat, flags))
	case token.LeftBrace:
		return ast.NewObjLitExpr(p.parseObjectLiteral())
	case token.LeftBracket:
		return ast.NewArrLitExpr(p.parseArrayLiteral())
	case token.LeftParenthesis:
		return p.parseParenthesisedExpression()
	case token.NoSubstitutionTemplate, token.TemplateHead:
		return ast.NewTmplLitExpr(p.parseTemplateLiteral(false))
	case token.This:
		p.next()
		return ast.NewThisExpr(p.alloc.ThisExpression(idx))
	case token.Super:
		return p.parseSuperProperty()
	case token.Async:
		if f := p.parseMaybeAsyncFunction(false); f != nil {
			return ast.NewFuncLitExpr(f)
		}
	case token.Function:
		return ast.NewFuncLitExpr(p.parseFunction(false, false, idx))
	case token.Class:
		return ast.NewClassLitExpr(p.parseClass(false))
	}

	if p.isBindingId(p.currentKind()) {
		p.next()
		return ast.NewIdentExpr(p.alloc.Identifier(idx, ""))
	}

	p.errorUnexpectedToken(p.currentKind())
	p.nextStatement()
	return ast.NewInvalidExpr(p.alloc.InvalidExpression(idx, p.currentOffset()))
}

func (p *parser) parseSuperProperty() ast.Expression {
	idx := p.currentOffset()
	p.next()
	switch p.currentKind() {
	case token.Period:
		p.next()
		if !token.ID(p.currentKind()) {
			p.expect(token.Identifier)
			p.nextStatement()
			return ast.NewInvalidExpr(p.alloc.InvalidExpression(idx, p.currentOffset()))
		}
		idIdx := p.currentOffset()
		parsedLiteral := p.currentString()
		p.next()
		return ast.NewMemberExpr(p.alloc.MemberExpression(
			p.alloc.Expression(ast.NewSuperExpr(p.alloc.SuperExpression(idx))),
			p.alloc.MemberProperty(ast.NewIdentMemProp(p.alloc.Identifier(idIdx, parsedLiteral))),
		))
	case token.LeftBracket:
		return p.parseBracketMember(p.alloc.Expression(ast.NewSuperExpr(p.alloc.SuperExpression(idx))))
	case token.LeftParenthesis:
		return p.parseCallExpression(p.alloc.Expression(ast.NewSuperExpr(p.alloc.SuperExpression(idx))))
	default:
		p.errorf("'super' keyword unexpected here")
		p.nextStatement()
		return ast.NewInvalidExpr(p.alloc.InvalidExpression(idx, p.currentOffset()))
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

func (p *parser) parseParenthesisedExpression() ast.Expression {
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
				p.exprBuf = append(p.exprBuf, ast.NewInvalidExpr(p.alloc.InvalidExpression(start, expr.Idx1())))
			} else {
				p.exprBuf = append(p.exprBuf, p.parseAssignmentExpression())
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
		return result
	}
	if n == 0 {
		p.exprBuf = p.exprBuf[:mark]
		p.errorUnexpectedToken(token.RightParenthesis)
		return ast.NewInvalidExpr(p.alloc.InvalidExpression(opening, p.currentOffset()))
	}
	return ast.NewSequenceExpr(p.alloc.SequenceExpression(p.finishExprBuf(mark)))
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
		p.scanner.Token.Kind = token.Identifier
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
		target = ast.NewInvalidBindingTarget(p.alloc.InvalidExpression(idx, p.currentOffset()))
	}

	return
}

func (p *parser) parseVariableDeclaration() ast.VariableDeclarator {
	node := ast.VariableDeclarator{Target: p.alloc.BindingTarget(p.parseBindingTarget())}

	if p.currentKind() == token.Assign {
		p.next()
		node.Initializer = p.alloc.Expression(p.parseAssignmentExpression())
	}

	return node
}

func (p *parser) parseVariableDeclarationList() ast.VariableDeclarators {
	mark := len(p.declBuf)
	for {
		p.declBuf = append(p.declBuf, p.parseVariableDeclaration())
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
		expr := p.alloc.Expression(p.parseAssignmentExpression())
		p.expect(token.RightBracket)
		return "", "", expr, token.Illegal
	}
	idx, tkn, literal, parsedLiteral := p.currentOffset(), p.currentKind(), p.scanner.Token.Raw(p.scanner), p.currentString()
	var value *ast.Expression
	p.next()
	switch tkn {
	case token.Identifier, token.String, token.Keyword, token.EscapedReservedWord:
		value = p.alloc.Expression(ast.NewStrLitExpr(p.alloc.StringLiteral(idx, parsedLiteral, literal)))
	case token.Number:
		if isBigIntLiteral(literal) {
			bi, err := parseBigIntLiteral(literal)
			if err != nil {
				p.errorf("%s", err.Error())
			} else {
				value = p.alloc.Expression(ast.NewBigIntLitExpr(p.alloc.BigIntLiteral(idx, bi, literal)))
			}
		} else {
			num, err := parseNumberLiteral(literal)
			if err != nil {
				p.errorf("%s", err.Error())
			} else {
				value = p.alloc.Expression(ast.NewNumLitExpr(p.alloc.NumberLiteral(idx, num, literal)))
			}
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
		return ast.NewSpreadProp(p.alloc.SpreadElement(p.alloc.Expression(p.parseAssignmentExpression())))
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
					initializer = p.alloc.Expression(p.parseAssignmentExpression())
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
		p.alloc.Expression(p.parseAssignmentExpression()),
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
				p.alloc.Expression(p.parseAssignmentExpression()),
			)))
		} else {
			p.exprBuf = append(p.exprBuf, p.parseAssignmentExpression())
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
		literal := p.scanner.Token.TemplateLiteral(p.scanner)
		parsed := p.scanner.Token.TemplateParsed(p.scanner)
		kind := p.currentKind()

		res.Elements = append(res.Elements, ast.TemplateElement{
			Idx:     start,
			Literal: literal,
			Parsed:  parsed,
		})

		if kind == token.NoSubstitutionTemplate || kind == token.TemplateTail {
			res.CloseQuote = p.scanner.Token.Idx1 - 1
			p.next()
			break
		}

		p.next()
		p.exprBuf = append(p.exprBuf, p.parseExpression())

		if p.currentKind() != token.RightBrace {
			p.errorUnexpectedToken(p.currentKind())
			break
		}
		// Re-tokenize the `}` as the start of the next template part
		p.scanner.NextTemplatePart()
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
			p.exprBuf = append(p.exprBuf, ast.NewSpreadExpr(p.alloc.SpreadElement(p.alloc.Expression(p.parseAssignmentExpression()))))
		} else {
			p.exprBuf = append(p.exprBuf, p.parseAssignmentExpression())
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

func (p *parser) parseCallExpression(left *ast.Expression) ast.Expression {
	argumentList, idx0, idx1 := p.parseArgumentList()
	return ast.NewCallExpr(p.alloc.CallExpression(left, idx0, argumentList, idx1))
}

func (p *parser) parseDotMember(left *ast.Expression) ast.Expression {
	period := p.currentOffset()
	p.next()

	literal := p.currentString()
	idx := p.currentOffset()

	if p.currentKind() == token.PrivateIdentifier {
		p.next()
		return ast.NewPrivDotExpr(p.alloc.PrivateDotExpression(
			left,
			p.alloc.PrivateIdentifier(p.alloc.Identifier(idx, literal)),
		))
	}

	if !token.ID(p.currentKind()) {
		p.expect(token.Identifier)
		p.nextStatement()
		return ast.NewInvalidExpr(p.alloc.InvalidExpression(period, p.currentOffset()))
	}

	p.next()

	return ast.NewMemberExpr(p.alloc.MemberExpression(
		left,
		p.alloc.MemberProperty(ast.NewIdentMemProp(p.alloc.Identifier(idx, literal))),
	))
}

func (p *parser) parseBracketMember(left *ast.Expression) ast.Expression {
	p.expect(token.LeftBracket)
	member := p.alloc.Expression(p.parseExpression())
	p.expect(token.RightBracket)
	return ast.NewMemberExpr(p.alloc.MemberExpression(
		left,
		p.alloc.MemberProperty(ast.NewComputedMemProp(p.alloc.ComputedProperty(member))),
	))
}

func (p *parser) parseNewExpression() ast.Expression {
	idx := p.expect(token.New)
	if p.currentKind() == token.Period {
		p.next()
		if p.currentString() == "target" {
			return ast.NewMetaPropExpr(p.alloc.MetaProperty(
				p.alloc.Identifier(idx, token.New.String()),
				p.parseIdentifier(),
				idx,
			))
		}
		p.errorUnexpectedToken(token.Identifier)
	}
	calleeVal := p.parseLeftHandSideExpression()
	if bad, ok := calleeVal.Invalid(); ok {
		bad.From = idx
		return calleeVal
	}
	callee := p.alloc.Expression(calleeVal)
	node := p.alloc.NewExpression(idx, callee)
	if p.currentKind() == token.LeftParenthesis {
		argumentList, idx0, idx1 := p.parseArgumentList()
		node.ArgumentList = argumentList
		node.LeftParenthesis = idx0
		node.RightParenthesis = idx1
	}
	return ast.NewNewExpr(node)
}

func (p *parser) parseLeftHandSideExpression() ast.Expression {
	var left ast.Expression
	if p.currentKind() == token.New {
		left = p.parseNewExpression()
	} else {
		left = p.parsePrimaryExpression()
	}
L:
	for {
		switch p.currentKind() {
		case token.Period:
			left = p.parseDotMember(p.alloc.Expression(left))
		case token.LeftBracket:
			left = p.parseBracketMember(p.alloc.Expression(left))
		case token.NoSubstitutionTemplate, token.TemplateHead:
			left = ast.NewTmplLitExpr(p.parseTaggedTemplateLiteral(p.alloc.Expression(left)))
		default:
			break L
		}
	}

	return left
}

func (p *parser) parseLeftHandSideExpressionAllowCall() ast.Expression {
	allowIn := p.scope.allowIn
	p.scope.allowIn = true

	var left ast.Expression
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
			left = p.parseDotMember(p.alloc.Expression(left))
		case token.LeftBracket:
			left = p.parseBracketMember(p.alloc.Expression(left))
		case token.LeftParenthesis:
			left = p.parseCallExpression(p.alloc.Expression(left))
		case token.NoSubstitutionTemplate, token.TemplateHead:
			if optionalChain {
				p.errorf("Invalid template literal on optional chain")
				p.nextStatement()
				p.scope.allowIn = allowIn
				return ast.NewInvalidExpr(p.alloc.InvalidExpression(start, p.currentOffset()))
			}
			left = ast.NewTmplLitExpr(p.parseTaggedTemplateLiteral(p.alloc.Expression(left)))
		case token.QuestionDot:
			optionalChain = true
			left = ast.NewOptionalExpr(p.alloc.Optional(p.alloc.Expression(left)))

			switch p.peek().Kind {
			case token.LeftBracket, token.LeftParenthesis, token.NoSubstitutionTemplate, token.TemplateHead:
				p.next()
			default:
				left = p.parseDotMember(p.alloc.Expression(left))
			}
		default:
			break L
		}
	}

	if optionalChain {
		left = ast.NewOptChainExpr(p.alloc.OptionalChain(p.alloc.Expression(left)))
	}
	p.scope.allowIn = allowIn
	return left
}

func (p *parser) parseUpdateExpression() ast.Expression {
	kind := p.currentKind()
	if isUpdateOperator(kind) {
		idx := p.currentOffset()
		p.next()
		operand := p.parseUnaryExpression()
		switch operand.Kind() {
		case ast.ExprIdent, ast.ExprPrivDot, ast.ExprMember:
		default:
			p.errorf("Invalid left-hand side in assignment")
			p.nextStatement()
			return ast.NewInvalidExpr(p.alloc.InvalidExpression(idx, p.currentOffset()))
		}
		return ast.NewUpdateExpr(p.alloc.UpdateExpression(toUpdateOperator(kind), idx, p.alloc.Expression(operand), false))
	}

	operand := p.parseLeftHandSideExpressionAllowCall()
	postKind := p.currentKind()
	if isUpdateOperator(postKind) && !p.scanner.Token.OnNewLine {
		idx := p.currentOffset()
		p.next()
		switch operand.Kind() {
		case ast.ExprIdent, ast.ExprPrivDot, ast.ExprMember:
		default:
			p.errorf("Invalid left-hand side in assignment")
			p.nextStatement()
			return ast.NewInvalidExpr(p.alloc.InvalidExpression(idx, p.currentOffset()))
		}
		return ast.NewUpdateExpr(p.alloc.UpdateExpression(toUpdateOperator(postKind), idx, p.alloc.Expression(operand), true))
	}
	return operand
}

func (p *parser) parseUnaryExpression() ast.Expression {
	kind := p.currentKind()
	if isUnaryOperator(kind) {
		idx := p.currentOffset()
		p.next()
		return ast.NewUnaryExpr(p.alloc.UnaryExpression(toUnaryOperator(kind), idx, p.alloc.Expression(p.parseUnaryExpression())))
	}

	if kind == token.Await {
		if p.scope.allowAwait {
			idx := p.currentOffset()
			p.next()
			if !p.scope.inAsync {
				p.errorUnexpectedToken(token.Await)
				return ast.NewInvalidExpr(p.alloc.InvalidExpression(idx, p.currentOffset()))
			}
			if p.scope.inFuncParams {
				p.errorf("Illegal await-expression in formal parameters of async function")
			}
			return ast.NewAwaitExpr(p.alloc.AwaitExpression(idx, p.alloc.Expression(p.parseUnaryExpression())))
		}
	}

	return p.parseUpdateExpression()
}

// parseBinaryExpressionOrHigher parses a binary expression using the Pratt parsing algorithm.
// minPrecedence is the minimum precedence level to parse (operators with lower
// or equal precedence will stop the loop, depending on associativity).
//
// See: https://matklad.github.io/2020/04/13/simple-but-powerful-pratt-parsing.html
func (p *parser) parseBinaryExpressionOrHigher(minPrecedence Precedence) ast.Expression {
	lhsParenthesized := p.currentKind() == token.LeftParenthesis

	var lhs ast.Expression
	if p.scope.allowIn && p.currentKind() == token.PrivateIdentifier {
		lhs = p.parsePrivateInExpression(minPrecedence)
	} else {
		lhs = p.parseUnaryExpression()
	}

	return p.parseBinaryExpressionRest(lhs, lhsParenthesized, minPrecedence)
}

func (p *parser) parseBinaryExpressionRest(lhs ast.Expression, lhsParenthesized bool, minPrecedence Precedence) ast.Expression {
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
				if lexp, ok := rhs.Logical(); ok && !rhsParenthesized {
					if lexp.Operator == ast.LogicalAnd || lexp.Operator == ast.LogicalOr {
						p.errorf("Logical expressions and coalesce expressions cannot be mixed. Wrap either by parentheses")
					}
				}
				if lexp, ok := lhs.Logical(); ok && !lhsParenthesized {
					if lexp.Operator == ast.LogicalAnd || lexp.Operator == ast.LogicalOr {
						p.errorf("Logical expressions and coalesce expressions cannot be mixed. Wrap either by parentheses")
					}
				}
			}
			lhs = ast.NewLogicalExpr(p.alloc.LogicalExpression(toLogicalOperator(kind), p.alloc.Expression(lhs), p.alloc.Expression(rhs)))
		} else if isBinaryOperator(kind) {
			// Check for unparenthesized unary/await before **
			if kind == token.Exponent && !lhsParenthesized {
				switch lhs.Kind() {
				case ast.ExprUnary, ast.ExprAwait:
					p.errorf("Unary operator used immediately before exponentiation expression. Parenthesis must be used to disambiguate operator precedence")
				}
			}
			lhs = ast.NewBinaryExpr(p.alloc.BinaryExpression(toBinaryOperator(kind), p.alloc.Expression(lhs), p.alloc.Expression(rhs)))
		} else {
			break
		}

		lhsParenthesized = false
	}

	return lhs
}

// parsePrivateInExpression handles the `#identifier in expr` syntax.
func (p *parser) parsePrivateInExpression(minPrecedence Precedence) ast.Expression {
	left := ast.NewPrivIdentExpr(p.alloc.PrivateIdentifier(p.alloc.Identifier(p.currentOffset(), p.currentString())))
	p.next()

	// If next token is not `in`, or `in`'s precedence (Compare) is too low, just return the identifier.
	if p.currentKind() != token.In || PrecedenceCompare <= minPrecedence {
		return left
	}

	p.next() // consume `in`
	rhs := p.parseBinaryExpressionOrHigher(PrecedenceCompare)
	return ast.NewBinaryExpr(p.alloc.BinaryExpression(ast.BinaryIn, p.alloc.Expression(left), p.alloc.Expression(rhs)))
}

func (p *parser) parseConditionalExpression() ast.Expression {
	left := p.parseBinaryExpressionOrHigher(PrecedenceLowest)

	if p.currentKind() == token.QuestionMark {
		p.next()
		allowIn := p.scope.allowIn
		p.scope.allowIn = true
		consequent := p.parseAssignmentExpression()
		p.scope.allowIn = allowIn
		p.expect(token.Colon)
		alternate := p.parseAssignmentExpression()
		return ast.NewConditionalExpr(p.alloc.ConditionalExpression(
			p.alloc.Expression(left),
			p.alloc.Expression(consequent),
			p.alloc.Expression(alternate),
		))
	}

	return left
}

func (p *parser) parseArrowFunction(start ast.Idx, paramList *ast.ParameterList, async bool) ast.Expression {
	p.expect(token.Arrow)
	node := p.alloc.ArrowFunctionLiteral(start, paramList, async)
	node.Body = p.parseArrowFunctionBody(async)
	return ast.NewArrowFuncLitExpr(node)
}

func (p *parser) parseSingleArgArrowFunction(start ast.Idx, async bool) ast.Expression {
	savedAwait := p.scope.allowAwait
	if async != savedAwait {
		p.scope.allowAwait = async
	}
	p.tokenToBindingId()
	if p.currentKind() != token.Identifier {
		p.errorUnexpectedToken(p.currentKind())
		p.next()
		p.scope.allowAwait = savedAwait
		return ast.NewInvalidExpr(p.alloc.InvalidExpression(start, p.currentOffset()))
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

func (p *parser) parseAssignmentExpression() ast.Expression {
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
			return ast.NewYieldExpr(p.parseYieldExpression())
		}
		fallthrough
	default:
		p.tokenToBindingId()
	}
	left := p.parseConditionalExpression()
	kind := p.currentKind()

	if kind == token.Arrow {
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
			// async (x, y) => ...
			savedAwait := p.scope.allowAwait
			if !savedAwait {
				p.scope.allowAwait = true
			}
			if left.IsCall() {
				p.restore(state)
				p.next() // skip "async"
				paramList = p.parseFunctionParameterList()
			}
			if paramList == nil {
				p.errorf("Malformed arrow function parameter list")
				p.scope.allowAwait = savedAwait
				return ast.NewInvalidExpr(p.alloc.InvalidExpression(left.Idx0(), left.Idx1()))
			}
			result := p.parseArrowFunction(start, paramList, async)
			p.scope.allowAwait = savedAwait
			return result
		}
		if paramList == nil {
			p.errorf("Malformed arrow function parameter list")
			return ast.NewInvalidExpr(p.alloc.InvalidExpression(left.Idx0(), left.Idx1()))
		}
		return p.parseArrowFunction(start, paramList, async)
	}

	if isAssignOperator(kind) {
		operator := toAssignOperator(kind)

		idx := p.currentOffset()
		p.next()
		ok := false
		switch left.Kind() {
		case ast.ExprIdent, ast.ExprPrivDot, ast.ExprMember:
			ok = true
		case ast.ExprArrLit:
			if !parenthesis && operator == ast.AssignmentAssign {
				left = p.reinterpretAsArrayAssignmentPattern(left.MustArrLit())
				ok = true
			}
		case ast.ExprObjLit:
			if !parenthesis && operator == ast.AssignmentAssign {
				left = p.reinterpretAsObjectAssignmentPattern(left.MustObjLit())
				ok = true
			}
		}
		if ok {
			return ast.NewAssignExpr(p.alloc.AssignExpression(operator, p.alloc.Expression(left), p.alloc.Expression(p.parseAssignmentExpression())))
		}
		p.errorf("Invalid left-hand side in assignment")
		p.nextStatement()
		return ast.NewInvalidExpr(p.alloc.InvalidExpression(idx, p.currentOffset()))
	}

	return left
}

func (p *parser) parseYieldExpression() *ast.YieldExpression {
	idx := p.expect(token.Yield)

	if p.scope.inFuncParams {
		p.errorf("Yield expression not allowed in formal parameter")
	}

	node := p.alloc.YieldExpression(idx)

	if !p.scanner.Token.OnNewLine && p.currentKind() == token.Multiply {
		node.Delegate = true
		p.next()
	}

	if !p.canInsertSemicolon() {
		state := p.mark()
		expr := p.parseAssignmentExpression()
		if expr.IsInvalid() {
			expr = ast.Expression{}
			p.restore(state)
		}
		node.Argument = p.alloc.Expression(expr)
	}

	return node
}

func (p *parser) parseExpression() ast.Expression {
	left := p.parseAssignmentExpression()

	if p.currentKind() == token.Comma {
		mark := len(p.exprBuf)
		p.exprBuf = append(p.exprBuf, left)
		for {
			if p.currentKind() != token.Comma {
				break
			}
			p.next()
			p.exprBuf = append(p.exprBuf, p.parseAssignmentExpression())
		}
		return ast.NewSequenceExpr(p.alloc.SequenceExpression(p.finishExprBuf(mark)))
	}

	return left
}

func (p *parser) checkComma(from, to ast.Idx) {
	if pos := strings.IndexByte(p.str[int(from)-1:int(to)-1], ','); pos >= 0 {
		p.errorf("Comma is not allowed here")
	}
}

func (p *parser) reinterpretAsArrayAssignmentPattern(left *ast.ArrayLiteral) ast.Expression {
	value := left.Value
	var rest *ast.Expression
	for i := range value {
		if spread, ok := value[i].Spread(); ok {
			if i != len(value)-1 {
				p.errorf("Rest element must be last element")
				return ast.NewInvalidExpr(p.alloc.InvalidExpression(left.Idx0(), left.Idx1()))
			}
			p.checkComma(spread.Idx1(), left.RightBracket)
			rest = p.reinterpretAsDestructAssignTarget(spread.Expression)
			value = value[:len(value)-1]
		} else {
			result := p.reinterpretAsAssignmentElement(&value[i])
			value[i] = *result
		}
	}
	return ast.NewArrPatExpr(p.alloc.ArrayPattern(left.LeftBracket, left.RightBracket, value, rest))
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

func (p *parser) reinterpretAsArrayBindingPattern(left *ast.ArrayLiteral) ast.Expression {
	value := left.Value
	var rest *ast.Expression
	for i := range value {
		if spread, ok := value[i].Spread(); ok {
			if i != len(value)-1 {
				p.errorf("Rest element must be last element")
				return ast.NewInvalidExpr(p.alloc.InvalidExpression(left.Idx0(), left.Idx1()))
			}
			p.checkComma(spread.Idx1(), left.RightBracket)
			rest = p.reinterpretAsDestructBindingTarget(spread.Expression)
			value = value[:len(value)-1]
		} else {
			result := p.reinterpretAsBindingElement(&value[i])
			value[i] = *result
		}
	}
	return ast.NewArrPatExpr(p.alloc.ArrayPattern(left.LeftBracket, left.RightBracket, value, rest))
}

func (p *parser) parseArrayBindingPattern() ast.BindingTarget {
	result := p.reinterpretAsArrayBindingPattern(p.parseArrayLiteral())
	return ast.BindingTargetFromExpression(&result)
}

func (p *parser) parseObjectBindingPattern() ast.BindingTarget {
	result := p.reinterpretAsObjectBindingPattern(p.parseObjectLiteral())
	return ast.BindingTargetFromExpression(&result)
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

func (p *parser) reinterpretAsObjectBindingPattern(expr *ast.ObjectLiteral) ast.Expression {
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
				return ast.NewInvalidExpr(p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1()))
			}
			rest = p.reinterpretAsBindingRestElement(spread.Expression)
			value = value[:i]
			ok = true
		}
		if !ok {
			p.errorf("Invalid destructuring binding target")
			return ast.NewInvalidExpr(p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1()))
		}
	}
	return ast.NewObjPatExpr(p.alloc.ObjectPattern(expr.LeftBrace, expr.RightBrace, value, rest))
}

func (p *parser) reinterpretAsObjectAssignmentPattern(l *ast.ObjectLiteral) ast.Expression {
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
				return ast.NewInvalidExpr(p.alloc.InvalidExpression(l.Idx0(), l.Idx1()))
			}
			rest = spread.Expression
			value = value[:i]
			ok = true
		}
		if !ok {
			p.errorf("Invalid destructuring assignment target")
			return ast.NewInvalidExpr(p.alloc.InvalidExpression(l.Idx0(), l.Idx1()))
		}
	}
	return ast.NewObjPatExpr(p.alloc.ObjectPattern(l.LeftBrace, l.RightBrace, value, rest))
}

func (p *parser) reinterpretAsAssignmentElement(expr *ast.Expression) *ast.Expression {
	switch expr.Kind() {
	case ast.ExprAssign:
		e := expr.MustAssign()
		if e.Operator == ast.AssignmentAssign {
			e.Left = p.reinterpretAsDestructAssignTarget(e.Left)
			return expr
		}
		p.errorf("Invalid destructuring assignment target")
		return p.alloc.Expression(ast.NewInvalidExpr(p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1())))
	default:
		return p.reinterpretAsDestructAssignTarget(expr)
	}
}

func (p *parser) reinterpretAsBindingElement(expr *ast.Expression) *ast.Expression {
	switch expr.Kind() {
	case ast.ExprAssign:
		e := expr.MustAssign()
		if e.Operator == ast.AssignmentAssign {
			e.Left = p.reinterpretAsDestructBindingTarget(e.Left)
			return expr
		}
		p.errorf("Invalid destructuring assignment target")
		return p.alloc.Expression(ast.NewInvalidExpr(p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1())))
	default:
		return p.reinterpretAsDestructBindingTarget(expr)
	}
}

func (p *parser) reinterpretAsBinding(expr *ast.Expression) ast.VariableDeclarator {
	switch expr.Kind() {
	case ast.ExprAssign:
		e := expr.MustAssign()
		if e.Operator == ast.AssignmentAssign {
			target := p.reinterpretAsDestructBindingTarget(e.Left)
			return ast.VariableDeclarator{
				Target:      p.alloc.BindingTarget(ast.BindingTargetFromExpression(target)),
				Initializer: e.Right,
			}
		}
		p.errorf("Invalid destructuring assignment target")
		return ast.VariableDeclarator{
			Target: p.alloc.BindingTarget(ast.NewInvalidBindingTarget(p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1()))),
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
		v := p.reinterpretAsArrayAssignmentPattern(item.MustArrLit())
		return p.alloc.Expression(v)
	case ast.ExprObjLit:
		v := p.reinterpretAsObjectAssignmentPattern(item.MustObjLit())
		return p.alloc.Expression(v)
	case ast.ExprArrPat, ast.ExprObjPat, ast.ExprIdent, ast.ExprPrivDot, ast.ExprMember:
		return item
	}
	p.errorf("Invalid destructuring assignment target")
	return p.alloc.Expression(ast.NewInvalidExpr(p.alloc.InvalidExpression(item.Idx0(), item.Idx1())))
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
		v := p.reinterpretAsArrayBindingPattern(item.MustArrLit())
		return p.alloc.Expression(v)
	case ast.ExprObjLit:
		v := p.reinterpretAsObjectBindingPattern(item.MustObjLit())
		return p.alloc.Expression(v)
	case ast.ExprIdent:
		id := item.MustIdent()
		if !p.scope.allowAwait || id.Name != "await" {
			return item
		}
	}
	p.errorf("Invalid destructuring binding target")
	return p.alloc.Expression(ast.NewInvalidExpr(p.alloc.InvalidExpression(item.Idx0(), item.Idx1())))
}

func (p *parser) reinterpretAsBindingRestElement(expr *ast.Expression) *ast.Expression {
	if expr.IsIdent() {
		return expr
	}
	p.errorf("Invalid binding rest")
	return p.alloc.Expression(ast.NewInvalidExpr(p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1())))
}
