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
	return &ast.Identifier{
		Idx:  idx,
		Name: literal,
	}
}

func (p *parser) parsePrimaryExpression() ast.Expr {
	idx := p.currentOffset()
	switch p.currentKind() {
	case token.Identifier:
		parsedLiteral := p.currentString()
		p.next()
		return &ast.Identifier{
			Idx:  idx,
			Name: parsedLiteral,
		}
	case token.Null:
		p.next()
		return &ast.NullLiteral{
			Idx: idx,
		}
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
		return &ast.BooleanLiteral{
			Idx:   idx,
			Value: value,
		}
	case token.String:
		parsedLiteral := p.currentString()
		raw := p.token.Raw(p.scanner)
		p.next()
		return &ast.StringLiteral{
			Idx:   idx,
			Value: parsedLiteral,

			Raw: &raw,
		}
	case token.Number:
		parsedLiteral := p.currentString()
		raw := p.token.Raw(p.scanner)
		p.next()
		value, err := parseNumberLiteral(parsedLiteral)
		if err != nil {
			p.error(err.Error())
			value = 0
		}
		return &ast.NumberLiteral{
			Idx:   idx,
			Value: value,

			Raw: &raw,
		}
	case token.Slash, token.QuotientAssign:
		return p.parseRegExpLiteral()
	case token.LeftBrace:
		return p.parseObjectLiteral()
	case token.LeftBracket:
		return p.parseArrayLiteral()
	case token.LeftParenthesis:
		return p.parseParenthesisedExpression()
	case token.Backtick:
		return p.parseTemplateLiteral(false)
	case token.This:
		p.next()
		return &ast.ThisExpression{Idx: idx}
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
		return &ast.Identifier{Idx: idx}
	}

	p.errorUnexpectedToken(p.currentKind())
	p.nextStatement()
	return &ast.InvalidExpression{From: idx, To: p.currentOffset()}
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
			return &ast.InvalidExpression{From: idx, To: p.currentOffset()}
		}
		idIdx := p.currentOffset()
		parsedLiteral := p.currentString()
		p.next()
		return &ast.MemberExpression{
			Object: p.makeExpr(&ast.SuperExpression{
				Idx: idx,
			}),
			Property: &ast.MemberProperty{Prop: &ast.Identifier{
				Idx:  idIdx,
				Name: parsedLiteral,
			}},
		}
	case token.LeftBracket:
		return p.parseBracketMember(&ast.SuperExpression{
			Idx: idx,
		})
	case token.LeftParenthesis:
		return p.parseCallExpression(&ast.SuperExpression{
			Idx: idx,
		})
	default:
		p.error("'super' keyword unexpected here")
		p.nextStatement()
		return &ast.InvalidExpression{From: idx, To: p.currentOffset()}
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
				list = append(list, ast.Expression{Expr: &ast.InvalidExpression{
					From: start,
					To:   expr.Idx1(),
				}})
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
	if len(list) == 1 && len(p.errors) == 0 {
		return list[0].Expr
	}
	if len(list) == 0 {
		p.errorUnexpectedToken(token.RightParenthesis)
		return &ast.InvalidExpression{
			From: opening,
			To:   p.currentOffset(),
		}
	}
	return &ast.SequenceExpression{
		Sequence: list,
	}
}

func (p *parser) parseRegExpLiteral() *ast.RegExpLiteral {
	offset := p.scanner.Offset() - 1 // Opening slash already gotten
	if p.currentKind() == token.QuotientAssign {
		offset -= 1 // =
	}
	idx := offset

	var (
		inEscape    bool
		inCharClass bool
		chr         rune
	)
	for {
		chr, ok := p.scanner.NextRune()
		if !ok || isLineTerminator(chr) {
			p.error(errUnexpectedEndOfInput)
			return nil
		}

		if inEscape {
			inEscape = false
		} else if chr == '/' && !inCharClass {
			break
		} else if chr == '[' {
			inCharClass = true
		} else if chr == '\\' {
			inEscape = true
		} else if chr == ']' {
			inCharClass = false
		}
	}

	endOffset := p.currentOffset()
	pattern := p.str[offset+1 : endOffset]

	flags := ""
	if !isLineTerminator(chr) && !isLineWhiteSpace(p.chr) {
		p.next()

		if p.currentKind() == token.Identifier { // gim
			flags = p.currentString()
			p.next()
			endOffset = p.currentOffset() - 1
		}
	} else {
		p.next()
	}

	literal := p.str[offset:endOffset]
	return &ast.RegExpLiteral{
		Idx:     idx,
		Literal: literal,
		Pattern: pattern,
		Flags:   flags,
	}
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
		target = &ast.Identifier{
			Name: p.currentString(),
			Idx:  p.currentOffset(),
		}
		p.next()
	case token.LeftBracket:
		target = p.parseArrayBindingPattern()
	case token.LeftBrace:
		target = p.parseObjectBindingPattern()
	default:
		idx := p.expect(token.Identifier)
		p.nextStatement()
		target = &ast.InvalidExpression{From: idx, To: p.currentOffset()}
	}

	return
}

func (p *parser) parseVariableDeclaration(declarationList *ast.VariableDeclarators) ast.VariableDeclarator {
	node := &ast.VariableDeclarator{
		Target: &ast.BindingTarget{Target: p.parseBindingTarget()},
	}

	if p.currentKind() == token.Assign {
		p.next()
		node.Initializer = p.makeExpr(p.parseAssignmentExpression())
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
		value = &ast.StringLiteral{
			Idx:   idx,
			Value: parsedLiteral,

			Raw: &literal,
		}
	case token.Number:
		num, err := parseNumberLiteral(literal)
		if err != nil {
			p.error(err.Error())
		} else {
			value = &ast.NumberLiteral{
				Idx:   idx,
				Value: num,

				Raw: &literal,
			}
		}
	case token.PrivateIdentifier:
		value = &ast.PrivateIdentifier{
			Identifier: &ast.Identifier{
				Idx:  idx,
				Name: parsedLiteral,
			},
		}
	default:
		// null, false, class, etc.
		if token.ID(tkn) {
			value = &ast.StringLiteral{
				Idx:   idx,
				Value: literal,

				Raw: &literal,
			}
		} else {
			p.errorUnexpectedToken(tkn)
		}
	}
	return literal, parsedLiteral, value, tkn
}

func (p *parser) parseObjectProperty() ast.Prop {
	if p.currentKind() == token.Ellipsis {
		p.next()
		return &ast.SpreadElement{
			Expression: p.makeExpr(p.parseAssignmentExpression()),
		}
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
			return &ast.PropertyKeyed{
				Key:      p.makeExpr(value),
				Kind:     ast.PropertyKindMethod,
				Value:    p.makeExpr(p.parseMethodDefinition(keyStartIdx, ast.PropertyKindMethod, true, false)),
				Computed: tkn == token.Illegal,
			}
		}
		switch {
		case p.currentKind() == token.LeftParenthesis:
			return &ast.PropertyKeyed{
				Key:      p.makeExpr(value),
				Kind:     ast.PropertyKindMethod,
				Value:    p.makeExpr(p.parseMethodDefinition(keyStartIdx, ast.PropertyKindMethod, false, false)),
				Computed: tkn == token.Illegal,
			}
		case p.currentKind() == token.Comma || p.currentKind() == token.RightBrace || p.currentKind() == token.Assign: // shorthand property
			if p.isBindingId(tkn) {
				var initializer ast.Expr
				if p.currentKind() == token.Assign {
					// allow the initializer syntax here in case the object literal
					// needs to be reinterpreted as an assignment pattern, enforce later if it doesn't.
					p.next()
					initializer = p.parseAssignmentExpression()
				}
				return &ast.PropertyShort{
					Name: &ast.Identifier{
						Name: parsedLiteral,
						Idx:  value.Idx0(),
					},
					Initializer: p.makeExpr(initializer),
				}
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

			return &ast.PropertyKeyed{
				Key:      p.makeExpr(keyValue),
				Kind:     kind,
				Value:    p.makeExpr(p.parseMethodDefinition(keyStartIdx, kind, false, async)),
				Computed: tkn1 == token.Illegal,
			}
		}
	}

	p.expect(token.Colon)
	return &ast.PropertyKeyed{
		Key:      p.makeExpr(value),
		Kind:     ast.PropertyKindValue,
		Value:    p.makeExpr(p.parseAssignmentExpression()),
		Computed: tkn == token.Illegal,
	}
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
	node := &ast.FunctionLiteral{
		Function:      keyStartIdx,
		ParameterList: parameterList,
		Generator:     generator,
		Async:         async,
	}
	node.Body = p.parseFunctionBlock(async, async, generator)
	return node
}

func (p *parser) parseObjectLiteral() *ast.ObjectLiteral {
	var value []ast.Property
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

	return &ast.ObjectLiteral{
		LeftBrace:  idx0,
		RightBrace: idx1,
		Value:      value,
	}
}

func (p *parser) parseArrayLiteral() *ast.ArrayLiteral {
	idx0 := p.expect(token.LeftBracket)
	var value ast.Expressions
	for p.currentKind() != token.RightBracket && p.currentKind() != token.Eof {
		if p.currentKind() == token.Comma {
			p.next()
			value = append(value, ast.Expression{})
			continue
		}
		if p.currentKind() == token.Ellipsis {
			p.next()
			value = append(value, ast.Expression{Expr: &ast.SpreadElement{
				Expression: p.makeExpr(p.parseAssignmentExpression()),
			}})
		} else {
			value = append(value, ast.Expression{Expr: p.parseAssignmentExpression()})
		}
		if p.currentKind() != token.RightBracket {
			p.expect(token.Comma)
		}
	}
	idx1 := p.expect(token.RightBracket)

	return &ast.ArrayLiteral{
		LeftBracket:  idx0,
		RightBracket: idx1,
		Value:        value,
	}
}

func (p *parser) parseTemplateLiteral(tagged bool) *ast.TemplateLiteral {
	res := &ast.TemplateLiteral{
		OpenQuote: p.currentOffset(),
	}
	for {
		start := p.offset
		literal, parsed, finished, parseErr, err := p.parseTemplateCharacters()
		if err != "" {
			p.error(err)
		}
		res.Elements = append(res.Elements, ast.TemplateElement{
			Idx:     p.idxOf(start),
			Literal: literal,
			Parsed:  parsed,
			Valid:   parseErr == "",
		})
		if !tagged && parseErr != "" {
			p.error(parseErr)
		}
		end := p.chrOffset - 1
		p.next()
		if finished {
			res.CloseQuote = p.idxOf(end)
			break
		}
		expr := p.parseExpression()
		res.Expressions = append(res.Expressions, ast.Expression{Expr: expr})
		if p.currentKind() != token.RightBrace {
			p.errorUnexpectedToken(p.currentKind())
		}
	}
	return res
}

func (p *parser) parseTaggedTemplateLiteral(tag ast.Expr) *ast.TemplateLiteral {
	l := p.parseTemplateLiteral(true)
	l.Tag = p.makeExpr(tag)
	return l
}

func (p *parser) parseArgumentList() (argumentList ast.Expressions, idx0, idx1 ast.Idx) {
	idx0 = p.expect(token.LeftParenthesis)
	for p.currentKind() != token.RightParenthesis {
		var item ast.Expr
		if p.currentKind() == token.Ellipsis {
			p.next()
			item = &ast.SpreadElement{
				Expression: p.makeExpr(p.parseAssignmentExpression()),
			}
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
	return &ast.CallExpression{
		Callee:           p.makeExpr(left),
		LeftParenthesis:  idx0,
		ArgumentList:     argumentList,
		RightParenthesis: idx1,
	}
}

func (p *parser) parseDotMember(left ast.Expr) ast.Expr {
	period := p.currentOffset()
	p.next()

	literal := p.currentString()
	idx := p.currentOffset()

	if p.currentKind() == token.PrivateIdentifier {
		p.next()
		return &ast.PrivateDotExpression{
			Left: p.makeExpr(left),
			Identifier: &ast.PrivateIdentifier{
				Identifier: &ast.Identifier{
					Idx:  idx,
					Name: literal,
				},
			},
		}
	}

	if !token.ID(p.currentKind()) {
		p.expect(token.Identifier)
		p.nextStatement()
		return &ast.InvalidExpression{From: period, To: p.currentOffset()}
	}

	p.next()

	return &ast.MemberExpression{
		Object: p.makeExpr(left),
		Property: &ast.MemberProperty{
			Prop: &ast.Identifier{
				Idx:  idx,
				Name: literal,
			},
		},
	}
}

func (p *parser) parseBracketMember(left ast.Expr) *ast.MemberExpression {
	p.expect(token.LeftBracket)
	member := p.parseExpression()
	p.expect(token.RightBracket)
	return &ast.MemberExpression{
		Object: p.makeExpr(left),
		Property: &ast.MemberProperty{
			Prop: &ast.ComputedProperty{
				Expr: p.makeExpr(member),
			},
		},
	}
}

func (p *parser) parseNewExpression() ast.Expr {
	idx := p.expect(token.New)
	if p.currentKind() == token.Period {
		p.next()
		if p.currentString() == "target" {
			return &ast.MetaProperty{
				Meta: &ast.Identifier{
					Name: token.New.String(),
					Idx:  idx,
				},
				Property: p.parseIdentifier(),
			}
		}
		p.errorUnexpectedToken(token.Identifier)
	}
	callee := p.parseLeftHandSideExpression()
	if bad, ok := callee.(*ast.InvalidExpression); ok {
		bad.From = idx
		return bad
	}
	node := &ast.NewExpression{
		New:    idx,
		Callee: p.makeExpr(callee),
	}
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
		case token.Backtick:
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
		case token.Backtick:
			if optionalChain {
				p.error("Invalid template literal on optional chain")
				p.nextStatement()
				return &ast.InvalidExpression{From: start, To: p.currentOffset()}
			}
			left = p.parseTaggedTemplateLiteral(left)
		case token.QuestionDot:
			optionalChain = true
			left = &ast.Optional{Expr: p.makeExpr(left)}

			switch p.peek().Kind {
			case token.LeftBracket, token.LeftParenthesis, token.Backtick:
				p.next()
			default:
				left = p.parseDotMember(left)
			}
		default:
			break L
		}
	}

	if optionalChain {
		left = &ast.OptionalChain{Base: p.makeExpr(left)}
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
			return &ast.InvalidExpression{From: idx, To: p.currentOffset()}
		}
		return &ast.UpdateExpression{
			Operator: tkn,
			Idx:      idx,
			Operand:  p.makeExpr(operand),
		}
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
				return &ast.InvalidExpression{From: idx, To: p.currentOffset()}
			}
			return &ast.UpdateExpression{
				Operator: tkn,
				Idx:      idx,
				Operand:  p.makeExpr(operand),
				Postfix:  true,
			}
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
		return &ast.UnaryExpression{
			Operator: tkn,
			Idx:      idx,
			Operand:  p.makeExpr(p.parseUnaryExpression()),
		}
	case token.Await:
		if p.scope.allowAwait {
			idx := p.currentOffset()
			p.next()
			if !p.scope.inAsync {
				p.errorUnexpectedToken(token.Await)
				return &ast.InvalidExpression{
					From: idx,
					To:   p.currentOffset(),
				}
			}
			if p.scope.inFuncParams {
				p.error("Illegal await-expression in formal parameters of async function")
			}
			return &ast.AwaitExpression{
				Await:    idx,
				Argument: p.makeExpr(p.parseUnaryExpression()),
			}
		}
	}

	return p.parseUpdateExpression()
}

func (p *parser) parseExponentiationExpression() ast.Expr {
	parenthesis := p.currentKind() == token.LeftParenthesis

	left := p.parseUnaryExpression()

	if p.currentKind() == token.Exponent {
		if !parenthesis {
			if _, isUnary := left.(*ast.UnaryExpression); isUnary {
				p.error("Unary operator used immediately before exponentiation expression. Parenthesis must be used to disambiguate operator precedence")
			}
		}
		for {
			p.next()
			left = &ast.BinaryExpression{
				Operator: token.Exponent,
				Left:     p.makeExpr(left),
				Right:    p.makeExpr(p.parseExponentiationExpression()),
			}
			if p.currentKind() != token.Exponent {
				break
			}
		}
	}

	return left
}

func (p *parser) parseMultiplicativeExpression() ast.Expr {
	left := p.parseExponentiationExpression()

	for p.currentKind() == token.Multiply || p.currentKind() == token.Slash ||
		p.currentKind() == token.Remainder {
		tkn := p.currentKind()
		p.next()
		left = &ast.BinaryExpression{
			Operator: tkn,
			Left:     p.makeExpr(left),
			Right:    p.makeExpr(p.parseExponentiationExpression()),
		}
	}

	return left
}

func (p *parser) parseAdditiveExpression() ast.Expr {
	left := p.parseMultiplicativeExpression()

	for p.currentKind() == token.Plus || p.currentKind() == token.Minus {
		tkn := p.currentKind()
		p.next()
		left = &ast.BinaryExpression{
			Operator: tkn,
			Left:     p.makeExpr(left),
			Right:    p.makeExpr(p.parseMultiplicativeExpression()),
		}
	}

	return left
}

func (p *parser) parseShiftExpression() ast.Expr {
	left := p.parseAdditiveExpression()

	for p.currentKind() == token.ShiftLeft || p.currentKind() == token.ShiftRight ||
		p.currentKind() == token.UnsignedShiftRight {
		tkn := p.currentKind()
		p.next()
		left = &ast.BinaryExpression{
			Operator: tkn,
			Left:     p.makeExpr(left),
			Right:    p.makeExpr(p.parseAdditiveExpression()),
		}
	}

	return left
}

func (p *parser) parseRelationalExpression() ast.Expr {
	if p.scope.allowIn && p.currentKind() == token.PrivateIdentifier {
		left := &ast.PrivateIdentifier{
			Identifier: &ast.Identifier{
				Idx:  p.currentOffset(),
				Name: p.currentString(),
			},
		}
		p.next()
		if p.currentKind() == token.In {
			p.next()
			return &ast.BinaryExpression{
				Operator: p.currentKind(),
				Left:     p.makeExpr(left),
				Right:    p.makeExpr(p.parseShiftExpression()),
			}
		}
		return left
	}
	left := p.parseShiftExpression()

	allowIn := p.scope.allowIn
	p.scope.allowIn = true
	defer func() {
		p.scope.allowIn = allowIn
	}()

	switch p.currentKind() {
	case token.Less, token.LessOrEqual, token.Greater, token.GreaterOrEqual, token.InstanceOf:
		tkn := p.currentKind()
		p.next()
		return &ast.BinaryExpression{
			Operator: tkn,
			Left:     p.makeExpr(left),
			Right:    p.makeExpr(p.parseRelationalExpression()),
		}
	case token.In:
		if !allowIn {
			return left
		}
		tkn := p.currentKind()
		p.next()
		return &ast.BinaryExpression{
			Operator: tkn,
			Left:     p.makeExpr(left),
			Right:    p.makeExpr(p.parseRelationalExpression()),
		}
	}

	return left
}

func (p *parser) parseEqualityExpression() ast.Expr {
	left := p.parseRelationalExpression()

	for p.currentKind() == token.Equal || p.currentKind() == token.NotEqual ||
		p.currentKind() == token.StrictEqual || p.currentKind() == token.StrictNotEqual {
		tkn := p.currentKind()
		p.next()
		left = &ast.BinaryExpression{
			Operator: tkn,
			Left:     p.makeExpr(left),
			Right:    p.makeExpr(p.parseRelationalExpression()),
		}
	}

	return left
}

func (p *parser) parseBitwiseAndExpression() ast.Expr {
	left := p.parseEqualityExpression()

	for p.currentKind() == token.And {
		tkn := p.currentKind()
		p.next()
		left = &ast.BinaryExpression{
			Operator: tkn,
			Left:     p.makeExpr(left),
			Right:    p.makeExpr(p.parseEqualityExpression()),
		}
	}

	return left
}

func (p *parser) parseBitwiseExclusiveOrExpression() ast.Expr {
	left := p.parseBitwiseAndExpression()

	for p.currentKind() == token.ExclusiveOr {
		tkn := p.currentKind()
		p.next()
		left = &ast.BinaryExpression{
			Operator: tkn,
			Left:     p.makeExpr(left),
			Right:    p.makeExpr(p.parseBitwiseAndExpression()),
		}
	}

	return left
}

func (p *parser) parseBitwiseOrExpression() ast.Expr {
	left := p.parseBitwiseExclusiveOrExpression()

	for p.currentKind() == token.Or {
		tkn := p.currentKind()
		p.next()
		left = &ast.BinaryExpression{
			Operator: tkn,
			Left:     p.makeExpr(left),
			Right:    p.makeExpr(p.parseBitwiseExclusiveOrExpression()),
		}
	}

	return left
}

func (p *parser) parseLogicalAndExpression() ast.Expr {
	left := p.parseBitwiseOrExpression()

	for p.currentKind() == token.LogicalAnd {
		tkn := p.currentKind()
		p.next()
		left = &ast.BinaryExpression{
			Operator: tkn,
			Left:     p.makeExpr(left),
			Right:    p.makeExpr(p.parseBitwiseOrExpression()),
		}
	}

	return left
}

func isLogicalAndExpr(expr ast.Expr) bool {
	if bexp, ok := expr.(*ast.BinaryExpression); ok && bexp.Operator == token.LogicalAnd {
		return true
	}
	return false
}

func (p *parser) parseLogicalOrExpression() ast.Expr {
	parenthesis := p.currentKind() == token.LeftParenthesis
	left := p.parseLogicalAndExpression()

	if p.currentKind() == token.LogicalOr || !parenthesis && isLogicalAndExpr(left) {
		for {
			switch p.currentKind() {
			case token.LogicalOr:
				p.next()
				left = &ast.BinaryExpression{
					Operator: token.LogicalOr,
					Left:     p.makeExpr(left),
					Right:    p.makeExpr(p.parseLogicalAndExpression()),
				}
			case token.Coalesce:
				goto mixed
			default:
				return left
			}
		}
	} else {
		for {
			switch p.currentKind() {
			case token.Coalesce:
				p.next()

				parenthesis := p.currentKind() == token.LeftParenthesis
				right := p.parseLogicalAndExpression()
				if !parenthesis && isLogicalAndExpr(right) {
					goto mixed
				}

				left = &ast.BinaryExpression{
					Operator: token.Coalesce,
					Left:     p.makeExpr(left),
					Right:    p.makeExpr(right),
				}
			case token.LogicalOr:
				goto mixed
			default:
				return left
			}
		}
	}

mixed:
	p.error("Logical expressions and coalesce expressions cannot be mixed. Wrap either by parentheses")
	return left
}

func (p *parser) parseConditionalExpression() ast.Expr {
	left := p.parseLogicalOrExpression()

	if p.currentKind() == token.QuestionMark {
		p.next()
		allowIn := p.scope.allowIn
		p.scope.allowIn = true
		consequent := p.parseAssignmentExpression()
		p.scope.allowIn = allowIn
		p.expect(token.Colon)
		return &ast.ConditionalExpression{
			Test:       p.makeExpr(left),
			Consequent: p.makeExpr(consequent),
			Alternate:  p.makeExpr(p.parseAssignmentExpression()),
		}
	}

	return left
}

func (p *parser) parseArrowFunction(start ast.Idx, paramList ast.ParameterList, async bool) ast.Expr {
	p.expect(token.Arrow)
	node := &ast.ArrowFunctionLiteral{
		Start:         start,
		ParameterList: paramList,
		Async:         async,
	}
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
		return &ast.InvalidExpression{
			From: start,
			To:   p.currentOffset(),
		}
	}

	id := p.parseIdentifier()

	paramList := ast.ParameterList{
		Opening: id.Idx,
		Closing: id.Idx1(),
		List: ast.VariableDeclarators{{
			Target: &ast.BindingTarget{Target: id},
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
	case token.Arrow:
		var paramList *ast.ParameterList
		if id, ok := left.(*ast.Identifier); ok {
			paramList = &ast.ParameterList{
				Opening: id.Idx,
				Closing: id.Idx1() - 1,
				List: ast.VariableDeclarators{{
					Target: &ast.BindingTarget{Target: id},
				}},
			}
		} else if parenthesis {
			if seq, ok := left.(*ast.SequenceExpression); ok && len(p.errors) == 0 {
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
			return &ast.InvalidExpression{From: left.Idx0(), To: left.Idx1()}
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
			return &ast.AssignExpression{
				Left:     p.makeExpr(left),
				Operator: operator,
				Right:    p.makeExpr(p.parseAssignmentExpression()),
			}
		}
		p.error("Invalid left-hand side in assignment")
		p.nextStatement()
		return &ast.InvalidExpression{From: idx, To: p.currentOffset()}
	}

	return left
}

func (p *parser) parseYieldExpression() *ast.YieldExpression {
	idx := p.expect(token.Yield)

	if p.scope.inFuncParams {
		p.error("Yield expression not allowed in formal parameter")
	}

	node := &ast.YieldExpression{
		Yield: idx,
	}

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
		node.Argument = p.makeExpr(expr)
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
		return &ast.SequenceExpression{
			Sequence: sequence,
		}
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
				return &ast.InvalidExpression{From: left.Idx0(), To: left.Idx1()}
			}
			p.checkComma(spread.Idx1(), left.RightBracket)
			rest = p.reinterpretAsDestructAssignTarget(spread.Expression.Expr)
			value = value[:len(value)-1]
		} else {
			value[i] = ast.Expression{Expr: p.reinterpretAsAssignmentElement(item.Expr)}
		}
	}
	return &ast.ArrayPattern{
		LeftBracket:  left.LeftBracket,
		RightBracket: left.RightBracket,
		Elements:     value,
		Rest:         p.makeExpr(rest),
	}
}

func (p *parser) reinterpretArrayAssignPatternAsBinding(pattern *ast.ArrayPattern) *ast.ArrayPattern {
	for i, item := range pattern.Elements {
		pattern.Elements[i] = ast.Expression{Expr: p.reinterpretAsDestructBindingTarget(item.Expr)}
	}
	if pattern.Rest != nil {
		pattern.Rest = p.makeExpr(p.reinterpretAsDestructBindingTarget(pattern.Rest.Expr))
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
				return &ast.InvalidExpression{From: left.Idx0(), To: left.Idx1()}
			}
			p.checkComma(spread.Idx1(), left.RightBracket)
			rest = p.reinterpretAsDestructBindingTarget(spread.Expression.Expr)
			value = value[:len(value)-1]
		} else {
			value[i] = ast.Expression{Expr: p.reinterpretAsBindingElement(item.Expr)}
		}
	}
	return &ast.ArrayPattern{
		LeftBracket:  left.LeftBracket,
		RightBracket: left.RightBracket,
		Elements:     value,
		Rest:         p.makeExpr(rest),
	}
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
			keyed.Value = p.makeExpr(p.reinterpretAsBindingElement(keyed.Value.Expr))
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
				prop.Value = p.makeExpr(p.reinterpretAsBindingElement(prop.Value.Expr))
				ok = true
			}
		case *ast.PropertyShort:
			ok = true
		case *ast.SpreadElement:
			if i != len(expr.Value)-1 {
				p.error("Rest element must be last element")
				return &ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()}
			}
			// TODO make sure there is no trailing comma
			rest = p.reinterpretAsBindingRestElement(prop.Expression.Expr)
			value = value[:i]
			ok = true
		}
		if !ok {
			p.error("Invalid destructuring binding target")
			return &ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()}
		}
	}
	return &ast.ObjectPattern{
		LeftBrace:  expr.LeftBrace,
		RightBrace: expr.RightBrace,
		Properties: value,
		Rest:       rest,
	}
}

func (p *parser) reinterpretAsObjectAssignmentPattern(l *ast.ObjectLiteral) ast.Expr {
	var rest ast.Expr
	value := l.Value
	for i, prop := range value {
		ok := false
		switch prop := prop.Prop.(type) {
		case *ast.PropertyKeyed:
			if prop.Kind == ast.PropertyKindValue {
				prop.Value = p.makeExpr(p.reinterpretAsAssignmentElement(prop.Value.Expr))
				ok = true
			}
		case *ast.PropertyShort:
			ok = true
		case *ast.SpreadElement:
			if i != len(l.Value)-1 {
				p.error("Rest element must be last element")
				return &ast.InvalidExpression{From: l.Idx0(), To: l.Idx1()}
			}
			// TODO make sure there is no trailing comma
			rest = prop.Expression.Expr
			value = value[:i]
			ok = true
		}
		if !ok {
			p.error("Invalid destructuring assignment target")
			return &ast.InvalidExpression{From: l.Idx0(), To: l.Idx1()}
		}
	}
	return &ast.ObjectPattern{
		LeftBrace:  l.LeftBrace,
		RightBrace: l.RightBrace,
		Properties: value,
		Rest:       rest,
	}
}

func (p *parser) reinterpretAsAssignmentElement(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {
	case *ast.AssignExpression:
		if expr.Operator == token.Assign {
			expr.Left = p.makeExpr(p.reinterpretAsDestructAssignTarget(expr.Left.Expr))
			return expr
		} else {
			p.error("Invalid destructuring assignment target")
			return &ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()}
		}
	default:
		return p.reinterpretAsDestructAssignTarget(expr)
	}
}

func (p *parser) reinterpretAsBindingElement(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {
	case *ast.AssignExpression:
		if expr.Operator == token.Assign {
			expr.Left = p.makeExpr(p.reinterpretAsDestructBindingTarget(expr.Left.Expr))
			return expr
		} else {
			p.error("Invalid destructuring assignment target")
			return &ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()}
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
				Target:      &ast.BindingTarget{Target: p.reinterpretAsDestructBindingTarget(expr.Left.Expr)},
				Initializer: expr.Right,
			}
		} else {
			p.error("Invalid destructuring assignment target")
			return ast.VariableDeclarator{
				Target: &ast.BindingTarget{Target: &ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()}},
			}
		}
	default:
		return ast.VariableDeclarator{
			Target: &ast.BindingTarget{Target: p.reinterpretAsDestructBindingTarget(expr)},
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
	return &ast.InvalidExpression{From: item.Idx0(), To: item.Idx1()}
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
	return &ast.InvalidExpression{From: item.Idx0(), To: item.Idx1()}
}

func (p *parser) reinterpretAsBindingRestElement(expr ast.Expr) ast.Expr {
	if _, ok := expr.(*ast.Identifier); ok {
		return expr
	}
	p.error("Invalid binding rest")
	return &ast.InvalidExpression{From: expr.Idx0(), To: expr.Idx1()}
}
