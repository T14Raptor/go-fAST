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
		return ast.NewIdentifierExpr(p.alloc.Identifier(idx, parsedLiteral))
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
		return ast.NewStringLitExpr(p.alloc.StringLiteral(idx, parsedLiteral, raw))
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
		return ast.NewNumberLitExpr(p.alloc.NumberLiteral(idx, value, raw))
	case token.Slash, token.QuotientAssign:
		pat, flags, lit := p.scanner.ParseRegExp()
		p.next()
		return ast.NewRegExpLitExpr(p.alloc.RegExpLiteral(idx, lit, pat, flags))
	case token.LeftBrace:
		return ast.NewObjectLitExpr(p.parseObjectLiteral())
	case token.LeftBracket:
		return ast.NewArrayLitExpr(p.parseArrayLiteral())
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
		return ast.NewIdentifierExpr(p.alloc.Identifier(idx, ""))
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
			p.alloc.MemberProperty(ast.NewIdentifierMemProp(p.alloc.Identifier(idIdx, parsedLiteral))),
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
		p.declBuf = append(p.declBuf, p.declaratorFromExpression(&list[i]))
	}
	var rest *ast.Pattern
	if firstRestIdx != -1 {
		if spread, ok := list[firstRestIdx].Spread(); ok {
			rest = p.alloc.Pattern(p.restPattern(spread.Expression, patBinding))
		}
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

// parsePattern parses a binding pattern in declaration/parameter/catch
// position and returns an arena-allocated *Pattern.
func (p *parser) parsePattern() *ast.Pattern {
	return p.alloc.Pattern(p.parseBindingTarget())
}

// parseBindingTarget parses a binding target — an identifier, array binding
// pattern, or object binding pattern — directly, without round-tripping
// through the expression cover grammar. This avoids allocating an intermediate
// ArrayLiteral/ObjectLiteral (plus its Expression nodes) only to convert it to
// a pattern, which is the dominant cost of binding-pattern parsing.
//
// A trailing `= default` is NOT consumed here; see parseBindingElement.
func (p *parser) parseBindingTarget() ast.Pattern {
	p.tokenToBindingId()
	switch p.currentKind() {
	case token.Identifier:
		pat := ast.NewIdentifierPattern(p.alloc.Identifier(p.currentOffset(), p.currentString()))
		p.next()
		return pat
	case token.LeftBracket:
		return p.parseArrayBindingPattern()
	case token.LeftBrace:
		return p.parseObjectBindingPattern()
	default:
		idx := p.expect(token.Identifier)
		p.nextStatement()
		return ast.NewInvalidPattern(p.alloc.InvalidExpression(idx, p.currentOffset()))
	}
}

// parseBindingElement parses a binding target followed by an optional default
// initializer (`target = expr`), as found in array binding patterns and the
// value position of object binding properties.
func (p *parser) parseBindingElement() ast.Pattern {
	target := p.parseBindingTarget()
	if p.currentKind() == token.Assign {
		p.next()
		left := p.alloc.Pattern(target)
		right := p.alloc.Expression(p.parseAssignmentExpression())
		return ast.NewAssignPattern(p.alloc.AssignmentPattern(left, right))
	}
	return target
}

// parseArrayBindingPattern parses `[ ... ]` directly into an ArrayPattern.
// Elisions become zero-value Pattern holes; a `...` rest element (which may be
// a nested pattern but never has a default) must be last.
func (p *parser) parseArrayBindingPattern() ast.Pattern {
	lb := p.expect(token.LeftBracket)
	mark := len(p.patBuf)
	var rest *ast.Pattern
	for p.currentKind() != token.RightBracket && p.currentKind() != token.Eof {
		if p.currentKind() == token.Comma {
			p.next()
			p.patBuf = append(p.patBuf, ast.Pattern{}) // elision hole
			continue
		}
		if p.currentKind() == token.Ellipsis {
			p.next()
			rest = p.alloc.Pattern(p.parseBindingTarget())
			break
		}
		p.patBuf = append(p.patBuf, p.parseBindingElement())
		if p.currentKind() != token.RightBracket {
			p.expect(token.Comma)
		}
	}
	rb := p.expect(token.RightBracket)
	return ast.NewArrayPatPattern(p.alloc.ArrayPattern(lb, rb, p.finishPatBuf(mark), rest))
}

// parseObjectBindingPattern parses `{ ... }` directly into an ObjectPattern.
// A `...` rest property must be a plain identifier in binding position and must
// be last (a single trailing comma after it is tolerated, matching the prior
// cover-grammar behavior).
func (p *parser) parseObjectBindingPattern() ast.Pattern {
	lb := p.expect(token.LeftBrace)
	mark := len(p.patPropBuf)
	var rest *ast.Pattern
	for p.currentKind() != token.RightBrace && p.currentKind() != token.Eof {
		if p.currentKind() == token.Ellipsis {
			p.next()
			rest = p.alloc.Pattern(p.parseObjectRestBinding())
			if p.currentKind() == token.Comma {
				p.next()
			}
			break
		}
		prop := p.parseBindingProperty()
		if !prop.IsNone() {
			p.patPropBuf = append(p.patPropBuf, prop)
		}
		if p.currentKind() != token.RightBrace {
			p.expect(token.Comma)
		}
	}
	rb := p.expect(token.RightBrace)
	return ast.NewObjectPatPattern(p.alloc.ObjectPattern(lb, rb, p.finishPatPropBuf(mark), rest))
}

// parseBindingProperty parses a single object-binding property: either a
// shorthand (`a` / `a = default`) or a keyed element (`key: target` /
// `[computed]: target`). Key parsing is shared with object literals so that
// numeric, string, computed, and private keys behave identically.
func (p *parser) parseBindingProperty() ast.PatternProperty {
	keyStartIdx := p.currentOffset()
	_, parsedLiteral, key, tkn := p.parseObjectPropertyKey()
	if key.IsNone() {
		return ast.PatternProperty{}
	}
	if p.currentKind() == token.Colon {
		p.next()
		val := p.alloc.Pattern(p.parseBindingElement())
		return ast.NewKeyValuePatProp(p.alloc.PatternKeyValue(key, val))
	}
	kind := p.currentKind()
	if p.isBindingId(tkn) && (kind == token.Comma || kind == token.RightBrace || kind == token.Assign) {
		var initializer *ast.Expression
		if kind == token.Assign {
			p.next()
			initializer = p.alloc.Expression(p.parseAssignmentExpression())
		}
		return ast.NewShorthandPatProp(p.alloc.PatternShorthand(
			p.alloc.Identifier(keyStartIdx, parsedLiteral),
			initializer,
		))
	}
	p.errorUnexpectedToken(kind)
	return ast.PatternProperty{}
}

// parseObjectRestBinding parses the target of an object `...rest` in binding
// position, which must be a plain identifier.
func (p *parser) parseObjectRestBinding() ast.Pattern {
	p.tokenToBindingId()
	if p.currentKind() == token.Identifier {
		pat := ast.NewIdentifierPattern(p.alloc.Identifier(p.currentOffset(), p.currentString()))
		p.next()
		return pat
	}
	idx := p.currentOffset()
	p.errorf("Invalid destructuring binding target")
	return ast.NewInvalidPattern(p.alloc.InvalidExpression(idx, p.currentOffset()))
}

func (p *parser) parseVariableDeclaration() ast.VariableDeclarator {
	node := ast.VariableDeclarator{Target: p.parsePattern()}

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

func (p *parser) parseObjectPropertyKey() (string, string, *ast.PropertyName, token.Token) {
	if p.currentKind() == token.LeftBracket {
		lb := p.currentOffset()
		p.next()
		expr := p.alloc.Expression(p.parseAssignmentExpression())
		rb := p.expect(token.RightBracket)
		return "", "", p.alloc.PropertyName(ast.NewComputedPropName(p.alloc.ComputedProperty(lb, expr, rb))), token.Illegal
	}
	idx, tkn, literal, parsedLiteral := p.currentOffset(), p.currentKind(), p.scanner.Token.Raw(p.scanner), p.currentString()
	var value ast.PropertyName
	p.next()
	switch tkn {
	case token.Identifier, token.String, token.Keyword, token.EscapedReservedWord:
		value = ast.NewStringLitPropName(p.alloc.StringLiteral(idx, parsedLiteral, literal))
	case token.Number:
		if isBigIntLiteral(literal) {
			bi, err := parseBigIntLiteral(literal)
			if err != nil {
				p.errorf("%s", err.Error())
			} else {
				value = ast.NewBigIntLitPropName(p.alloc.BigIntLiteral(idx, bi, literal))
			}
		} else {
			num, err := parseNumberLiteral(literal)
			if err != nil {
				p.errorf("%s", err.Error())
			} else {
				value = ast.NewNumberLitPropName(p.alloc.NumberLiteral(idx, num, literal))
			}
		}
	case token.PrivateIdentifier:
		value = ast.NewPrivIdentifierPropName(p.alloc.PrivateIdentifier(p.alloc.Identifier(idx, parsedLiteral)))
	default:
		if token.ID(tkn) {
			value = ast.NewStringLitPropName(p.alloc.StringLiteral(idx, literal, literal))
		} else {
			p.errorUnexpectedToken(tkn)
		}
	}
	return literal, parsedLiteral, p.alloc.PropertyName(value), tkn
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
	if value.IsNone() {
		return ast.Property{}
	}
	if token.ID(tkn) || tkn == token.String || tkn == token.Number || tkn == token.Illegal {
		if generator {
			return ast.NewMethodProp(p.alloc.PropertyMethod(value,
				p.parseMethodDefinition(keyStartIdx, ast.MethodKindMethod, true, false)))
		}
		switch {
		case p.currentKind() == token.LeftParenthesis:
			return ast.NewMethodProp(p.alloc.PropertyMethod(value,
				p.parseMethodDefinition(keyStartIdx, ast.MethodKindMethod, false, false)))
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
			async := tkn == token.Async
			methodGen := false
			if async && p.currentKind() == token.Multiply {
				methodGen = true
				p.next()
			}
			_, _, keyValue, _ := p.parseObjectPropertyKey()
			if keyValue.IsNone() {
				return ast.Property{}
			}

			if async {
				return ast.NewMethodProp(p.alloc.PropertyMethod(keyValue,
					p.parseMethodDefinition(keyStartIdx, ast.MethodKindMethod, methodGen, true)))
			}
			if literal == "get" {
				return ast.NewGetterProp(p.alloc.PropertyGetter(keyValue,
					p.parseMethodDefinition(keyStartIdx, ast.MethodKindGet, false, false)))
			}
			return ast.NewSetterProp(p.alloc.PropertySetter(keyValue,
				p.parseMethodDefinition(keyStartIdx, ast.MethodKindSet, false, false)))
		}
	}

	p.expect(token.Colon)
	return ast.NewKeyValueProp(p.alloc.PropertyKeyValue(value,
		p.alloc.Expression(p.parseAssignmentExpression())))
}

func (p *parser) parseMethodDefinition(keyStartIdx ast.Idx, kind ast.MethodKind, generator, async bool) *ast.FunctionLiteral {
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
	case ast.MethodKindGet:
		if len(parameterList.List) > 0 || parameterList.Rest != nil {
			p.errorf("Getter must not have any formal parameters.")
		}
	case ast.MethodKindSet:
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
		p.alloc.MemberProperty(ast.NewIdentifierMemProp(p.alloc.Identifier(idx, literal))),
	))
}

func (p *parser) parseBracketMember(left *ast.Expression) ast.Expression {
	leftBracket := p.expect(token.LeftBracket)
	member := p.alloc.Expression(p.parseExpression())
	rightBracket := p.expect(token.RightBracket)
	return ast.NewMemberExpr(p.alloc.MemberExpression(
		left,
		p.alloc.MemberProperty(ast.NewComputedMemProp(p.alloc.ComputedProperty(leftBracket, member, rightBracket))),
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
		left = ast.NewOptionalChainExpr(p.alloc.OptionalChain(p.alloc.Expression(left)))
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
		case ast.ExprIdentifier, ast.ExprPrivDot, ast.ExprMember:
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
		case ast.ExprIdentifier, ast.ExprPrivDot, ast.ExprMember:
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
	left := ast.NewPrivIdentifierExpr(p.alloc.PrivateIdentifier(p.alloc.Identifier(p.currentOffset(), p.currentString())))
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
			Target: p.alloc.Pattern(ast.NewIdentifierPattern(id)),
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
		if id, ok := left.Identifier(); ok {
			paramList = p.alloc.ParameterList(ast.ParameterList{
				Opening: id.Idx,
				Closing: id.Idx1() - 1,
				List: ast.VariableDeclarators{{
					Target: p.alloc.Pattern(ast.NewIdentifierPattern(id)),
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
		var target *ast.Pattern
		switch left.Kind() {
		case ast.ExprIdentifier, ast.ExprPrivDot, ast.ExprMember:
			target = p.alloc.Pattern(p.patternFromExpression(&left, patAssign))
		case ast.ExprArrayLit, ast.ExprObjectLit:
			if !parenthesis && operator == ast.AssignmentAssign {
				target = p.alloc.Pattern(p.patternFromExpression(&left, patAssign))
			}
		}
		if target != nil {
			return ast.NewAssignExpr(p.alloc.AssignExpression(operator, target, p.alloc.Expression(p.parseAssignmentExpression())))
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
	if pos := strings.IndexByte(p.str[int(from):int(to)], ','); pos >= 0 {
		p.errorf("Comma is not allowed here")
	}
}

// patternMode selects which simple targets are permitted when converting an
// already-parsed expression (cover grammar) into a Pattern.
type patternMode uint8

const (
	patBinding patternMode = iota // declaration targets: ident / array / object
	patAssign                     // assignment targets: + member / private-dot
)

func (p *parser) invalidPattern(node ast.Node, mode patternMode) ast.Pattern {
	if mode == patBinding {
		p.errorf("Invalid destructuring binding target")
	} else {
		p.errorf("Invalid destructuring assignment target")
	}
	return ast.NewInvalidPattern(p.alloc.InvalidExpression(node.Idx0(), node.Idx1()))
}

// patternFromExpression converts a cover-grammar expression into a Pattern.
func (p *parser) patternFromExpression(expr *ast.Expression, mode patternMode) ast.Pattern {
	if expr == nil || expr.IsNone() {
		return ast.Pattern{} // elision hole
	}
	switch expr.Kind() {
	case ast.ExprIdentifier:
		id := expr.MustIdentifier()
		if mode == patBinding && p.scope.allowAwait && id.Name == "await" {
			break
		}
		return ast.NewIdentifierPattern(id)
	case ast.ExprMember:
		if mode == patAssign {
			return ast.NewMemberPattern(expr.MustMember())
		}
	case ast.ExprPrivDot:
		if mode == patAssign {
			return ast.NewPrivDotPattern(expr.MustPrivDot())
		}
	case ast.ExprArrayLit:
		return p.arrayPatternFromLiteral(expr.MustArrayLit(), mode)
	case ast.ExprObjectLit:
		return p.objectPatternFromLiteral(expr.MustObjectLit(), mode)
	case ast.ExprAssign:
		e := expr.MustAssign()
		if e.Operator == ast.AssignmentAssign {
			// e.Left was already converted to a Pattern (in assign mode) when the
			// assignment expression was parsed. In binding position, re-validate it.
			if mode == patBinding && !p.ensureBindingTarget(e.Left) {
				return ast.NewInvalidPattern(p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1()))
			}
			return ast.NewAssignPattern(p.alloc.AssignmentPattern(e.Left, e.Right))
		}
	}
	return p.invalidPattern(expr, mode)
}

func (p *parser) arrayPatternFromLiteral(lit *ast.ArrayLiteral, mode patternMode) ast.Pattern {
	value := lit.Value
	var rest *ast.Pattern
	mark := len(p.patBuf)
	for i := range value {
		if spread, ok := value[i].Spread(); ok {
			if i != len(value)-1 {
				p.patBuf = p.patBuf[:mark]
				p.errorf("Rest element must be last element")
				return ast.NewInvalidPattern(p.alloc.InvalidExpression(lit.Idx0(), lit.Idx1()))
			}
			p.checkComma(spread.Idx1(), lit.RightBracket)
			rest = p.alloc.Pattern(p.patternFromExpression(spread.Expression, mode))
			break
		}
		p.patBuf = append(p.patBuf, p.patternFromExpression(&value[i], mode))
	}
	elems := p.finishPatBuf(mark)
	return ast.NewArrayPatPattern(p.alloc.ArrayPattern(lit.LeftBracket, lit.RightBracket, elems, rest))
}

func (p *parser) objectPatternFromLiteral(lit *ast.ObjectLiteral, mode patternMode) ast.Pattern {
	value := lit.Value
	var rest *ast.Pattern
	mark := len(p.patPropBuf)
	for i := range value {
		switch value[i].Kind() {
		case ast.PropKeyValue:
			keyed := value[i].MustKeyValue()
			val := p.alloc.Pattern(p.patternFromExpression(keyed.Value, mode))
			p.patPropBuf = append(p.patPropBuf, ast.NewKeyValuePatProp(p.alloc.PatternKeyValue(keyed.Key, val)))
		case ast.PropShort:
			short := value[i].MustShort()
			p.patPropBuf = append(p.patPropBuf, ast.NewShorthandPatProp(p.alloc.PatternShorthand(short.Name, short.Initializer)))
		case ast.PropSpread:
			spread := value[i].MustSpread()
			if i != len(value)-1 {
				p.patPropBuf = p.patPropBuf[:mark]
				p.errorf("Rest element must be last element")
				return ast.NewInvalidPattern(p.alloc.InvalidExpression(lit.Idx0(), lit.Idx1()))
			}
			rest = p.alloc.Pattern(p.restPattern(spread.Expression, mode))
		default:
			p.patPropBuf = p.patPropBuf[:mark]
			return p.invalidPattern(lit, mode)
		}
	}
	props := p.finishPatPropBuf(mark)
	return ast.NewObjectPatPattern(p.alloc.ObjectPattern(lit.LeftBrace, lit.RightBrace, props, rest))
}

// restPattern converts an object/parameter rest target, which must be a simple
// binding/assignment target (no defaults).
func (p *parser) restPattern(expr *ast.Expression, mode patternMode) ast.Pattern {
	switch expr.Kind() {
	case ast.ExprIdentifier:
		return ast.NewIdentifierPattern(expr.MustIdentifier())
	case ast.ExprMember:
		if mode == patAssign {
			return ast.NewMemberPattern(expr.MustMember())
		}
	case ast.ExprPrivDot:
		if mode == patAssign {
			return ast.NewPrivDotPattern(expr.MustPrivDot())
		}
	}
	return p.invalidPattern(expr, mode)
}

// ensureBindingTarget reports whether a pattern (built in assignment mode) is a
// valid binding target, i.e. contains no member/private-dot leaves.
func (p *parser) ensureBindingTarget(pat *ast.Pattern) bool {
	switch pat.Kind() {
	case ast.PatternNone, ast.PatternIdentifier, ast.PatternInvalid:
		return true
	case ast.PatternMember, ast.PatternPrivDot:
		return false
	case ast.PatternAssign:
		return p.ensureBindingTarget(pat.MustAssign().Left)
	case ast.PatternArrayPat:
		ap := pat.MustArrayPat()
		for i := range ap.Elements {
			if !p.ensureBindingTarget(&ap.Elements[i]) {
				return false
			}
		}
		return ap.Rest == nil || p.ensureBindingTarget(ap.Rest)
	case ast.PatternObjectPat:
		op := pat.MustObjectPat()
		for i := range op.Properties {
			if kv, ok := op.Properties[i].KeyValue(); ok && !p.ensureBindingTarget(kv.Value) {
				return false
			}
		}
		return op.Rest == nil || p.ensureBindingTarget(op.Rest)
	}
	return false
}

// declaratorFromExpression converts an arrow-function parameter (parsed as an
// expression) into a VariableDeclarator. Defaults are stored in Initializer.
func (p *parser) declaratorFromExpression(expr *ast.Expression) ast.VariableDeclarator {
	if e, ok := expr.Assign(); ok && e.Operator == ast.AssignmentAssign {
		if !p.ensureBindingTarget(e.Left) {
			return ast.VariableDeclarator{Target: p.alloc.Pattern(ast.NewInvalidPattern(p.alloc.InvalidExpression(expr.Idx0(), expr.Idx1())))}
		}
		return ast.VariableDeclarator{Target: e.Left, Initializer: e.Right}
	}
	return ast.VariableDeclarator{Target: p.alloc.Pattern(p.patternFromExpression(expr, patBinding))}
}
