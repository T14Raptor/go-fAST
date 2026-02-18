package parser

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

func (p *parser) parseBlockStatement() *ast.BlockStatement {
	node := p.alloc.BlockStatement()
	node.LeftBrace = p.expect(token.LeftBrace)
	node.List = p.parseStatementList()
	node.RightBrace = p.expect(token.RightBrace)

	return node
}

func (p *parser) parseEmptyStatement() ast.Stmt {
	idx := p.expect(token.Semicolon)
	return p.alloc.EmptyStatement(idx)
}

func (p *parser) parseStatementList() (list ast.Statements) {
	mark := len(p.stmtBuf)
	for p.currentKind() != token.RightBrace && p.currentKind() != token.Eof {
		p.scope.allowLet = true
		p.stmtBuf = append(p.stmtBuf, *p.parseStatement())
	}

	return p.finishStmtBuf(mark)
}

func (p *parser) parseStatement() *ast.Statement {
	tok := p.currentKind()
	if tok == token.Eof {
		p.errorUnexpectedToken(tok)
		return p.alloc.Statement(p.alloc.BadStatement(p.currentOffset(), p.currentOffset()+1))
	}

	switch tok {
	case token.Semicolon:
		return p.alloc.Statement(p.parseEmptyStatement())
	case token.LeftBrace:
		return p.alloc.Statement(p.parseBlockStatement())
	case token.If:
		return p.alloc.Statement(p.parseIfStatement())
	case token.Do:
		return p.alloc.Statement(p.parseDoWhileStatement())
	case token.While:
		return p.alloc.Statement(p.parseWhileStatement())
	case token.For:
		return p.parseForOrForInStatement()
	case token.Break:
		return p.alloc.Statement(p.parseBreakStatement())
	case token.Continue:
		return p.alloc.Statement(p.parseContinueStatement())
	case token.Debugger:
		return p.alloc.Statement(p.parseDebuggerStatement())
	case token.With:
		return p.alloc.Statement(p.parseWithStatement())
	case token.Var:
		return p.alloc.Statement(p.parseLexicalDeclaration(p.currentKind()))
	case token.Let:
		tok := p.peek().Kind
		if tok == token.LeftBracket || p.scope.allowLet && (token.ID(tok) || tok == token.LeftBrace) {
			return p.alloc.Statement(p.parseLexicalDeclaration(p.currentKind()))
		}
	case token.Const:
		return p.alloc.Statement(p.parseLexicalDeclaration(p.currentKind()))
	case token.Async:
		if f := p.parseMaybeAsyncFunction(true); f != nil {
			return p.alloc.Statement(p.alloc.FunctionDeclaration(f))
		}
	case token.Function:
		return p.alloc.Statement(p.alloc.FunctionDeclaration(p.parseFunction(true, false, p.currentOffset())))
	case token.Class:
		return p.alloc.Statement(p.alloc.ClassDeclaration(p.parseClass(true)))
	case token.Switch:
		return p.alloc.Statement(p.parseSwitchStatement())
	case token.Return:
		return p.alloc.Statement(p.parseReturnStatement())
	case token.Throw:
		return p.alloc.Statement(p.parseThrowStatement())
	case token.Try:
		return p.alloc.Statement(p.parseTryStatement())
	}

	expression := p.parseExpression()

	if identifier, isIdentifier := expression.Expr.(*ast.Identifier); isIdentifier && p.currentKind() == token.Colon {
		// LabelledStatement
		colon := p.currentOffset()
		p.next() // :
		label := identifier.Name
		for _, value := range p.scope.labels {
			if label == value {
				p.errorf("%s", label)
			}
		}
		p.scope.labels = append(p.scope.labels, label) // Push the label
		p.scope.allowLet = false
		statement := p.parseStatement()
		p.scope.labels = p.scope.labels[:len(p.scope.labels)-1] // Pop the label
		return p.alloc.Statement(p.alloc.LabelledStatement(identifier, colon, statement))
	}

	p.semicolon()

	return p.alloc.Statement(p.alloc.ExpressionStatement(expression))
}

func (p *parser) parseTryStatement() ast.Stmt {
	node := p.alloc.TryStatement(p.expect(token.Try), p.parseBlockStatement())

	if p.currentKind() == token.Catch {
		catch := p.currentOffset()
		p.next()
		var parameter *ast.BindingTarget
		if p.currentKind() == token.LeftParenthesis {
			p.next()
			parameter = p.alloc.BindingTarget(p.parseBindingTarget())
			p.expect(token.RightParenthesis)
		}
		node.Catch = p.alloc.CatchStatement(catch, parameter, p.parseBlockStatement())
	}

	if p.currentKind() == token.Finally {
		p.next()
		node.Finally = p.parseBlockStatement()
	}

	if node.Catch == nil && node.Finally == nil {
		p.errorf("Missing catch or finally after try")
		return p.alloc.BadStatement(node.Try, node.Body.Idx1())
	}

	return node
}

func (p *parser) parseFunctionParameterList() ast.ParameterList {
	opening := p.expect(token.LeftParenthesis)
	var list ast.VariableDeclarators
	var rest ast.Expr
	savedFuncParams := p.scope.inFuncParams
	if !savedFuncParams {
		p.scope.inFuncParams = true
	}
	for p.currentKind() != token.RightParenthesis && p.currentKind() != token.Eof {
		if p.currentKind() == token.Ellipsis {
			p.next()
			rest = p.reinterpretAsDestructBindingTarget(p.parseAssignmentExpression().Expr)
			break
		}
		p.parseVariableDeclaration(&list)
		if p.currentKind() != token.RightParenthesis {
			p.expect(token.Comma)
		}
	}
	closing := p.expect(token.RightParenthesis)
	p.scope.inFuncParams = savedFuncParams

	return ast.ParameterList{
		Opening: opening,
		List:    list,
		Rest:    rest,
		Closing: closing,
	}
}

func (p *parser) parseMaybeAsyncFunction(declaration bool) *ast.FunctionLiteral {
	if p.peek().Kind == token.Function {
		idx := p.currentOffset()
		p.next()
		fn := p.parseFunction(declaration, true, idx)
		return fn
	}
	return nil
}

func (p *parser) parseFunction(declaration, async bool, start ast.Idx) *ast.FunctionLiteral {
	node := p.alloc.FunctionLiteral(start, async)
	p.expect(token.Function)

	if p.currentKind() == token.Multiply {
		node.Generator = true
		p.next()
	}

	savedAwait := p.scope.allowAwait
	savedYield := p.scope.allowYield

	if !declaration {
		if async != savedAwait {
			p.scope.allowAwait = async
		}
		if node.Generator != savedYield {
			p.scope.allowYield = node.Generator
		}
	}

	p.tokenToBindingId()
	name := p.alloc.Identifier(0, "")
	if p.currentKind() == token.Identifier {
		name = p.parseIdentifier()
	} else if declaration {
		// Use expect errorf handling
		p.expect(token.Identifier)
	}
	node.Name = name

	if declaration {
		if async != p.scope.allowAwait {
			p.scope.allowAwait = async
		}
		if node.Generator != p.scope.allowYield {
			p.scope.allowYield = node.Generator
		}
	}

	node.ParameterList = p.parseFunctionParameterList()
	node.Body = p.parseFunctionBlock(async, async, p.scope.allowYield)

	p.scope.allowAwait = savedAwait
	p.scope.allowYield = savedYield
	return node
}

func (p *parser) parseFunctionBlock(async, allowAwait, allowYield bool) *ast.BlockStatement {
	p.openScope()
	p.scope.inFunction = true
	p.scope.inAsync = async
	p.scope.allowAwait = allowAwait
	p.scope.allowYield = allowYield
	body := p.parseBlockStatement()
	p.closeScope()
	return body
}

func (p *parser) parseArrowFunctionBody(async bool) *ast.ConciseBody {
	if p.currentKind() == token.LeftBrace {
		return p.alloc.ConciseBody(p.parseFunctionBlock(async, async, false))
	}
	if async != p.scope.inAsync || async != p.scope.allowAwait {
		inAsync := p.scope.inAsync
		allowAwait := p.scope.allowAwait
		allowYield := p.scope.allowYield
		p.scope.inAsync = async
		p.scope.allowAwait = async
		p.scope.allowYield = false
		result := p.alloc.ConciseBody(p.parseAssignmentExpression())
		p.scope.inAsync = inAsync
		p.scope.allowAwait = allowAwait
		p.scope.allowYield = allowYield
		return result
	}

	return p.alloc.ConciseBody(p.parseAssignmentExpression())
}

func (p *parser) parseClass(declaration bool) *ast.ClassLiteral {
	if !p.scope.allowLet && p.currentKind() == token.Class {
		p.errorUnexpectedToken(token.Class)
	}

	node := p.alloc.ClassLiteral(p.expect(token.Class))

	p.tokenToBindingId()
	name := p.alloc.Identifier(0, "")
	if p.currentKind() == token.Identifier {
		name = p.parseIdentifier()
	} else if declaration {
		// Use expect errorf handling
		p.expect(token.Identifier)
	}

	node.Name = name

	if p.currentKind() != token.LeftBrace {
		p.expect(token.Extends)
		node.SuperClass = p.alloc.Expression(p.parseLeftHandSideExpressionAllowCall())
	}

	p.expect(token.LeftBrace)

	elemMark := len(p.elemBuf)
	for p.currentKind() != token.RightBrace && p.currentKind() != token.Eof {
		if p.currentKind() == token.Semicolon {
			p.next()
			continue
		}
		start := p.currentOffset()
		static := false
		if p.currentKind() == token.Static {
			switch p.peek().Kind {
			case token.Assign, token.Semicolon, token.RightBrace, token.LeftParenthesis:
				// treat as identifier
			default:
				p.next()
				if p.currentKind() == token.LeftBrace {
					b := p.alloc.ClassStaticBlock(start)
					b.Block = p.parseFunctionBlock(false, true, false)
					p.elemBuf = append(p.elemBuf, ast.ClassElement{Element: b})
					continue
				}
				static = true
			}
		}

		var kind ast.PropertyKind
		var async bool
		methodBodyStart := p.currentOffset()
		if p.currentString() == "get" || p.currentString() == "set" {
			if tok := p.peek().Kind; tok != token.Semicolon && tok != token.LeftParenthesis {
				if p.currentString() == "get" {
					kind = ast.PropertyKindGet
				} else {
					kind = ast.PropertyKindSet
				}
				p.next()
			}
		} else if p.currentKind() == token.Async {
			if tok := p.peek().Kind; tok != token.Semicolon && tok != token.LeftParenthesis {
				async = true
				kind = ast.PropertyKindMethod
				p.next()
			}
		}
		generator := false
		if p.currentKind() == token.Multiply && (kind == "" || kind == ast.PropertyKindMethod) {
			generator = true
			kind = ast.PropertyKindMethod
			p.next()
		}

		_, keyName, value, tkn := p.parseObjectPropertyKey()
		if value == nil {
			continue
		}
		computed := tkn == token.Illegal
		_, private := value.(*ast.PrivateIdentifier)

		if static && !private && keyName == "prototype" {
			p.errorf("Classes may not have a static property named 'prototype'")
		}

		if kind == "" && p.currentKind() == token.LeftParenthesis {
			kind = ast.PropertyKindMethod
		}

		if kind != "" {
			// method
			if keyName == "constructor" && !computed {
				if !static {
					if kind != ast.PropertyKindMethod {
						p.errorf("Class constructor may not be an accessor")
					} else if async {
						p.errorf("Class constructor may not be an async method")
					} else if generator {
						p.errorf("Class constructor may not be a generator")
					}
				} else if private {
					p.errorf("Class constructor may not be a private method")
				}
			}
			md := p.alloc.MethodDefinition(start, p.alloc.Expression(value), kind,
				p.parseMethodDefinition(methodBodyStart, kind, generator, async),
				static, computed)
			p.elemBuf = append(p.elemBuf, ast.ClassElement{Element: md})
		} else {
			// field
			isCtor := !computed && keyName == "constructor"
			if !isCtor {
				if name, ok := value.(*ast.PrivateIdentifier); ok {
					isCtor = name.Identifier.Name == "constructor"
				}
			}
			if isCtor {
				p.errorf("Classes may not have a field named 'constructor'")
			}
			var initializer *ast.Expression
			if p.currentKind() == token.Assign {
				p.next()
				initializer = p.parseExpression()
			}

			if !p.canInsertSemicolon() {
				p.errorUnexpectedToken(p.currentKind())
				break
			}
			p.elemBuf = append(p.elemBuf, ast.ClassElement{Element: p.alloc.FieldDefinition(
				start, p.alloc.Expression(value), initializer, static, computed,
			)})
		}
	}

	node.Body = p.finishElemBuf(elemMark)
	node.RightBrace = p.expect(token.RightBrace)
	return node
}

func (p *parser) parseDebuggerStatement() ast.Stmt {
	idx := p.expect(token.Debugger)
	node := p.alloc.DebuggerStatement(idx)
	p.semicolon()
	return node
}

func (p *parser) parseReturnStatement() ast.Stmt {
	idx := p.expect(token.Return)

	if !p.scope.inFunction {
		p.errorf("Illegal return statement")
		p.nextStatement()
		return p.alloc.BadStatement(idx, p.currentOffset())
	}

	node := p.alloc.ReturnStatement(idx)

	if !p.canInsertSemicolon() {
		node.Argument = p.parseExpression()
	}

	p.semicolon()

	return node
}

func (p *parser) parseThrowStatement() ast.Stmt {
	idx := p.expect(token.Throw)

	if p.token.OnNewLine {
		p.errorf("Illegal newline after throw")
		p.nextStatement()
		return p.alloc.BadStatement(idx, p.currentOffset())
	}

	node := p.alloc.ThrowStatement(idx, p.parseExpression())

	p.semicolon()
	return node
}

func (p *parser) parseSwitchStatement() ast.Stmt {
	p.expect(token.Switch)
	p.expect(token.LeftParenthesis)
	node := p.alloc.SwitchStatement(p.parseExpression())
	p.expect(token.RightParenthesis)

	p.expect(token.LeftBrace)

	inSwitch := p.scope.inSwitch
	p.scope.inSwitch = true

	for index := 0; p.currentKind() != token.Eof; index++ {
		if p.currentKind() == token.RightBrace {
			p.next()
			break
		}

		clause := p.parseCaseStatement()
		if clause.Test == nil {
			if node.Default != -1 {
				p.errorf("Already saw a default in switch")
			}
			node.Default = index
		}
		node.Body = append(node.Body, clause)
	}

	p.scope.inSwitch = inSwitch
	return node
}

func (p *parser) parseWithStatement() ast.Stmt {
	p.expect(token.With)
	p.expect(token.LeftParenthesis)
	node := p.alloc.WithStatement(p.parseExpression())
	p.expect(token.RightParenthesis)
	p.scope.allowLet = false
	node.Body = p.parseStatement()

	return node
}

func (p *parser) parseCaseStatement() ast.CaseStatement {
	node := ast.CaseStatement{
		Case: p.currentOffset(),
	}
	if p.currentKind() == token.Default {
		p.next()
	} else {
		p.expect(token.Case)
		node.Test = p.parseExpression()
	}
	p.expect(token.Colon)

	mark := len(p.stmtBuf)
	for {
		if k := p.currentKind(); k == token.Eof ||
			k == token.RightBrace ||
			k == token.Case ||
			k == token.Default {
			break
		}
		p.scope.allowLet = true
		p.stmtBuf = append(p.stmtBuf, *p.parseStatement())
	}
	node.Consequent = p.finishStmtBuf(mark)

	return node
}

func (p *parser) parseIterationStatement() *ast.Statement {
	inIteration := p.scope.inIteration
	p.scope.inIteration = true
	p.scope.allowLet = false
	result := p.parseStatement()
	p.scope.inIteration = inIteration
	return result
}

func (p *parser) parseForIn(idx ast.Idx, into *ast.ForInto) *ast.ForInStatement {
	// Already have consumed "<into> in"

	source := p.parseExpression()
	p.expect(token.RightParenthesis)

	return p.alloc.ForInStatement(idx, into, source, p.parseIterationStatement())
}

func (p *parser) parseForOf(idx ast.Idx, into *ast.ForInto) *ast.ForOfStatement {
	// Already have consumed "<into> of"

	source := p.parseAssignmentExpression()
	p.expect(token.RightParenthesis)

	return p.alloc.ForOfStatement(idx, into, source, p.parseIterationStatement())
}

func (p *parser) parseFor(idx ast.Idx, initializer *ast.ForLoopInitializer) *ast.ForStatement {
	// Already have consumed "<initializer> ;"

	var test, update ast.Expr

	if p.currentKind() != token.Semicolon {
		test = p.parseExpression().Expr
	}
	p.expect(token.Semicolon)

	if p.currentKind() != token.RightParenthesis {
		update = p.parseExpression().Expr
	}
	p.expect(token.RightParenthesis)

	return p.alloc.ForStatement(idx, initializer, p.alloc.Expression(test), p.alloc.Expression(update), p.parseIterationStatement())
}

func (p *parser) parseForOrForInStatement() *ast.Statement {
	idx := p.expect(token.For)
	p.expect(token.LeftParenthesis)

	var initializer *ast.ForLoopInitializer

	forIn := false
	forOf := false
	var into *ast.ForInto
	if p.currentKind() != token.Semicolon {
		allowIn := p.scope.allowIn
		p.scope.allowIn = false
		tok := p.currentKind()
		if tok == token.Let {
			switch p.peek().Kind {
			case token.Identifier, token.LeftBracket, token.LeftBrace:
			default:
				tok = token.Identifier
			}
		}
		if tok == token.Var || tok == token.Let || tok == token.Const {
			idx := p.currentOffset()
			p.next()

			list := p.parseVariableDeclarationList()
			if len(list) == 1 {
				if p.currentKind() == token.In {
					p.next() // in
					forIn = true
				} else if p.currentKind() == token.Of {
					p.next()
					forOf = true
				}
			}
			if forIn || forOf {
				if list[0].Initializer != nil {
					p.errorf("for-in loop variable declaration may not have an initializer")
				}
				into = p.alloc.ForIntoPtr(p.alloc.VariableDeclaration(0, tok, ast.VariableDeclarators{list[0]}))
			} else {
				p.ensurePatternInit(list)

				initializer = p.alloc.ForLoopInitializer(p.alloc.VariableDeclaration(idx, tok, list))
			}
		} else {
			exprNode := p.parseExpression()
			if p.currentKind() == token.In {
				p.next()
				forIn = true
			} else if p.currentKind() == token.Of {
				p.next()
				forOf = true
			}
			if forIn || forOf {
				switch e := exprNode.Expr.(type) {
				case *ast.Identifier, *ast.PrivateDotExpression, *ast.VariableDeclarator, *ast.MemberExpression:
					// These are all acceptable
				case *ast.ObjectLiteral:
					exprNode.Expr = p.reinterpretAsObjectAssignmentPattern(e)
				case *ast.ArrayLiteral:
					exprNode.Expr = p.reinterpretAsArrayAssignmentPattern(e)
				default:
					p.errorf("Invalid left-hand side in for-in or for-of")
					p.nextStatement()
					return p.alloc.Statement(p.alloc.BadStatement(idx, p.currentOffset()))
				}
				into = p.alloc.ForIntoPtr(exprNode)
			} else {
				initializer = p.alloc.ForLoopInitializer(exprNode)
			}
		}
		p.scope.allowIn = allowIn
	}

	if forIn {
		return p.alloc.Statement(p.parseForIn(idx, into))
	}
	if forOf {
		return p.alloc.Statement(p.parseForOf(idx, into))
	}

	p.expect(token.Semicolon)
	return p.alloc.Statement(p.parseFor(idx, initializer))
}

func (p *parser) ensurePatternInit(list []ast.VariableDeclarator) {
	for _, item := range list {
		if _, ok := item.Target.Target.(ast.Pattern); ok {
			if item.Initializer == nil {
				p.errorf("Missing initializer in destructuring declaration")
				break
			}
		}
	}
}

func (p *parser) parseLexicalDeclaration(tok token.Token) *ast.VariableDeclaration {
	idx := p.expect(tok)
	if !p.scope.allowLet && tok != token.Var {
		p.errorf("Lexical declaration cannot appear in a single-statement context")
	}

	list := p.parseVariableDeclarationList()
	p.ensurePatternInit(list)
	p.semicolon()

	return p.alloc.VariableDeclaration(idx, tok, list)
}

func (p *parser) parseDoWhileStatement() ast.Stmt {
	inIteration := p.scope.inIteration
	p.scope.inIteration = true

	p.expect(token.Do)
	node := p.alloc.DoWhileStatement()
	if p.currentKind() == token.LeftBrace {
		node.Body = p.alloc.Statement(p.parseBlockStatement())
	} else {
		p.scope.allowLet = false
		node.Body = p.parseStatement()
	}

	p.expect(token.While)
	p.expect(token.LeftParenthesis)
	node.Test = p.parseExpression()
	p.expect(token.RightParenthesis)
	if p.currentKind() == token.Semicolon {
		p.next()
	}

	p.scope.inIteration = inIteration
	return node
}

func (p *parser) parseWhileStatement() ast.Stmt {
	p.expect(token.While)
	p.expect(token.LeftParenthesis)
	node := p.alloc.WhileStatement(p.parseExpression())
	p.expect(token.RightParenthesis)
	node.Body = p.parseIterationStatement()

	return node
}

func (p *parser) parseIfStatement() ast.Stmt {
	p.expect(token.If)
	p.expect(token.LeftParenthesis)
	node := p.alloc.IfStatement(p.parseExpression())
	p.expect(token.RightParenthesis)

	if p.currentKind() == token.LeftBrace {
		node.Consequent = p.alloc.Statement(p.parseBlockStatement())
	} else {
		p.scope.allowLet = false
		node.Consequent = p.parseStatement()
	}

	if p.currentKind() == token.Else {
		p.next()
		p.scope.allowLet = false
		node.Alternate = p.parseStatement()
	}

	return node
}

func (p *parser) parseSourceElements() (body ast.Statements) {
	mark := len(p.stmtBuf)
	for p.currentKind() != token.Eof {
		p.scope.allowLet = true
		p.stmtBuf = append(p.stmtBuf, *p.parseStatement())
	}

	return p.finishStmtBuf(mark)
}

func (p *parser) parseProgram() *ast.Program {
	return &ast.Program{
		Body: p.parseSourceElements(),
	}
}

func (p *parser) parseBreakStatement() ast.Stmt {
	idx := p.expect(token.Break)

	if p.canInsertSemicolon() {
		if p.currentKind() == token.Semicolon {
			p.next()
		}
		if !p.scope.inIteration && !p.scope.inSwitch {
			goto illegal
		}
		return p.alloc.BreakStatement(idx, nil)
	}

	p.tokenToBindingId()
	if p.currentKind() == token.Identifier {
		identifier := p.parseIdentifier()
		if !p.scope.hasLabel(identifier.Name) {
			p.errorf("%s", identifier.Name)
			return p.alloc.BadStatement(idx, identifier.Idx1())
		}
		p.semicolon()
		return p.alloc.BreakStatement(idx, identifier)
	}

	p.expect(token.Identifier)

illegal:
	p.errorf("Illegal break statement")
	p.nextStatement()
	return p.alloc.BadStatement(idx, p.currentOffset())
}

func (p *parser) parseContinueStatement() ast.Stmt {
	idx := p.expect(token.Continue)

	if p.canInsertSemicolon() {
		if p.currentKind() == token.Semicolon {
			p.next()
		}
		if !p.scope.inIteration {
			goto illegal
		}
		return p.alloc.ContinueStatement(idx, nil)
	}

	p.tokenToBindingId()
	if p.currentKind() == token.Identifier {
		identifier := p.parseIdentifier()
		if !p.scope.hasLabel(identifier.Name) {
			p.errorf("%s", identifier.Name)
			return p.alloc.BadStatement(idx, identifier.Idx1())
		}
		if !p.scope.inIteration {
			goto illegal
		}
		p.semicolon()
		return p.alloc.ContinueStatement(idx, identifier)
	}

	p.expect(token.Identifier)

illegal:
	p.errorf("Illegal continue statement")
	p.nextStatement()
	return p.alloc.BadStatement(idx, p.currentOffset())
}

// Find the next statement after an errorf (recover)
func (p *parser) nextStatement() {
	for {
		switch p.currentKind() {
		case token.Break, token.Continue,
			token.For, token.If, token.Return, token.Switch,
			token.Var, token.Do, token.Try, token.With,
			token.While, token.Throw, token.Catch, token.Finally:
			// Return only if parser made some progress since last
			// sync or if it has not reached 10 next calls without
			// progress. Otherwise consume at least one token to
			// avoid an endless parser loop
			if p.currentOffset() == p.recover.idx && p.recover.count < 10 {
				p.recover.count++
				return
			}
			if p.currentOffset() > p.recover.idx {
				p.recover.idx = p.currentOffset()
				p.recover.count = 0
				return
			}
			// Reaching here indicates a parser bug, likely an
			// incorrect token list in this function, but it only
			// leads to skipping of possibly correct code if a
			// previous errorf is present, and thus is preferred
			// over a non-terminating parse.
		case token.Eof:
			return
		}
		p.next()
	}
}
