package parser

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

func (p *parser) parseBlockStatement() *ast.BlockStatement {
	node := &ast.BlockStatement{}
	node.LeftBrace = p.expect(token.LeftBrace)
	node.List = p.parseStatementList()
	node.RightBrace = p.expect(token.RightBrace)

	return node
}

func (p *parser) parseEmptyStatement() ast.Stmt {
	idx := p.expect(token.Semicolon)
	return &ast.EmptyStatement{Semicolon: idx}
}

func (p *parser) parseStatementList() (list ast.Statements) {
	for p.token != token.RightBrace && p.token != token.Eof {
		p.scope.allowLet = true
		list = append(list, ast.Statement{Stmt: p.parseStatement()})
	}

	return
}

func (p *parser) parseStatement() ast.Stmt {
	if p.token == token.Eof {
		p.errorUnexpectedToken(p.token)
		return &ast.BadStatement{From: p.idx, To: p.idx + 1}
	}

	switch p.token {
	case token.Semicolon:
		return p.parseEmptyStatement()
	case token.LeftBrace:
		return p.parseBlockStatement()
	case token.If:
		return p.parseIfStatement()
	case token.Do:
		return p.parseDoWhileStatement()
	case token.While:
		return p.parseWhileStatement()
	case token.For:
		return p.parseForOrForInStatement()
	case token.Break:
		return p.parseBreakStatement()
	case token.Continue:
		return p.parseContinueStatement()
	case token.Debugger:
		return p.parseDebuggerStatement()
	case token.With:
		return p.parseWithStatement()
	case token.Var:
		return p.parseLexicalDeclaration(p.token)
	case token.Let:
		tok := p.peek()
		if tok == token.LeftBracket || p.scope.allowLet && (token.ID(tok) || tok == token.LeftBrace) {
			return p.parseLexicalDeclaration(p.token)
		}
		p.insertSemicolon = true
	case token.Const:
		return p.parseLexicalDeclaration(p.token)
	case token.Async:
		if f := p.parseMaybeAsyncFunction(true); f != nil {
			return &ast.FunctionDeclaration{
				Function: f,
			}
		}
	case token.Function:
		return &ast.FunctionDeclaration{
			Function: p.parseFunction(true, false, p.idx),
		}
	case token.Class:
		return &ast.ClassDeclaration{
			Class: p.parseClass(true),
		}
	case token.Switch:
		return p.parseSwitchStatement()
	case token.Return:
		return p.parseReturnStatement()
	case token.Throw:
		return p.parseThrowStatement()
	case token.Try:
		return p.parseTryStatement()
	}

	expression := p.parseExpression()

	if identifier, isIdentifier := expression.(*ast.Identifier); isIdentifier && p.token == token.Colon {
		// LabelledStatement
		colon := p.idx
		p.next() // :
		label := identifier.Name
		for _, value := range p.scope.labels {
			if label == value {
				p.error(label)
			}
		}
		p.scope.labels = append(p.scope.labels, label) // Push the label
		p.scope.allowLet = false
		statement := p.parseStatement()
		p.scope.labels = p.scope.labels[:len(p.scope.labels)-1] // Pop the label
		return &ast.LabelledStatement{
			Label:     identifier,
			Colon:     colon,
			Statement: refStmt(statement),
		}
	}

	p.optionalSemicolon()

	return &ast.ExpressionStatement{
		Expression: ptrExpr(expression),
	}
}

func (p *parser) parseTryStatement() ast.Stmt {
	node := &ast.TryStatement{
		Try:  p.expect(token.Try),
		Body: p.parseBlockStatement(),
	}

	if p.token == token.Catch {
		catch := p.idx
		p.next()
		var parameter ast.Target
		if p.token == token.LeftParenthesis {
			p.next()
			parameter = p.parseBindingTarget()
			p.expect(token.RightParenthesis)
		}
		node.Catch = &ast.CatchStatement{
			Catch:     catch,
			Parameter: &ast.BindingTarget{Target: parameter},
			Body:      p.parseBlockStatement(),
		}
	}

	if p.token == token.Finally {
		p.next()
		node.Finally = p.parseBlockStatement()
	}

	if node.Catch == nil && node.Finally == nil {
		p.error("Missing catch or finally after try")
		return &ast.BadStatement{From: node.Try, To: node.Body.Idx1()}
	}

	return node
}

func (p *parser) parseFunctionParameterList() ast.ParameterList {
	opening := p.expect(token.LeftParenthesis)
	var list ast.VariableDeclarators
	var rest ast.Expr
	if !p.scope.inFuncParams {
		p.scope.inFuncParams = true
		defer func() {
			p.scope.inFuncParams = false
		}()
	}
	for p.token != token.RightParenthesis && p.token != token.Eof {
		if p.token == token.Ellipsis {
			p.next()
			rest = p.reinterpretAsDestructBindingTarget(p.parseAssignmentExpression())
			break
		}
		p.parseVariableDeclaration(&list)
		if p.token != token.RightParenthesis {
			p.expect(token.Comma)
		}
	}
	closing := p.expect(token.RightParenthesis)

	return ast.ParameterList{
		Opening: opening,
		List:    list,
		Rest:    rest,
		Closing: closing,
	}
}

func (p *parser) parseMaybeAsyncFunction(declaration bool) *ast.FunctionLiteral {
	if p.peek() == token.Function {
		idx := p.idx
		p.next()
		fn := p.parseFunction(declaration, true, idx)
		return fn
	}
	return nil
}

func (p *parser) parseFunction(declaration, async bool, start ast.Idx) *ast.FunctionLiteral {
	node := &ast.FunctionLiteral{
		Function: start,
		Async:    async,
	}
	p.expect(token.Function)

	if p.token == token.Multiply {
		node.Generator = true
		p.next()
	}

	if !declaration {
		if async != p.scope.allowAwait {
			p.scope.allowAwait = async
			defer func() {
				p.scope.allowAwait = !async
			}()
		}
		if node.Generator != p.scope.allowYield {
			p.scope.allowYield = node.Generator
			defer func() {
				p.scope.allowYield = !node.Generator
			}()
		}
	}

	p.tokenToBindingId()
	name := &ast.Identifier{}
	if p.token == token.Identifier {
		name = p.parseIdentifier()
	} else if declaration {
		// Use expect error handling
		p.expect(token.Identifier)
	}
	node.Name = name

	if declaration {
		if async != p.scope.allowAwait {
			p.scope.allowAwait = async
			defer func() {
				p.scope.allowAwait = !async
			}()
		}
		if node.Generator != p.scope.allowYield {
			p.scope.allowYield = node.Generator
			defer func() {
				p.scope.allowYield = !node.Generator
			}()
		}
	}

	node.ParameterList = p.parseFunctionParameterList()
	node.Body = p.parseFunctionBlock(async, async, p.scope.allowYield)

	return node
}

func (p *parser) parseFunctionBlock(async, allowAwait, allowYield bool) (body *ast.BlockStatement) {
	p.openScope()
	p.scope.inFunction = true
	p.scope.inAsync = async
	p.scope.allowAwait = allowAwait
	p.scope.allowYield = allowYield
	defer p.closeScope()
	body = p.parseBlockStatement()
	return
}

func (p *parser) parseArrowFunctionBody(async bool) *ast.ConciseBody {
	if p.token == token.LeftBrace {
		return &ast.ConciseBody{Body: p.parseFunctionBlock(async, async, false)}
	}
	if async != p.scope.inAsync || async != p.scope.allowAwait {
		inAsync := p.scope.inAsync
		allowAwait := p.scope.allowAwait
		p.scope.inAsync = async
		p.scope.allowAwait = async
		allowYield := p.scope.allowYield
		p.scope.allowYield = false
		defer func() {
			p.scope.inAsync = inAsync
			p.scope.allowAwait = allowAwait
			p.scope.allowYield = allowYield
		}()
	}

	return &ast.ConciseBody{
		Body: &ast.Expression{Expr: p.parseAssignmentExpression()},
	}
}

func (p *parser) parseClass(declaration bool) *ast.ClassLiteral {
	if !p.scope.allowLet && p.token == token.Class {
		p.errorUnexpectedToken(token.Class)
	}

	node := &ast.ClassLiteral{
		Class: p.expect(token.Class),
	}

	p.tokenToBindingId()
	name := &ast.Identifier{}
	if p.token == token.Identifier {
		name = p.parseIdentifier()
	} else if declaration {
		// Use expect error handling
		p.expect(token.Identifier)
	}

	node.Name = name

	if p.token != token.LeftBrace {
		p.expect(token.Extends)
		node.SuperClass = ptrExpr(p.parseLeftHandSideExpressionAllowCall())
	}

	p.expect(token.LeftBrace)

	for p.token != token.RightBrace && p.token != token.Eof {
		if p.token == token.Semicolon {
			p.next()
			continue
		}
		start := p.idx
		static := false
		if p.token == token.Static {
			switch p.peek() {
			case token.Assign, token.Semicolon, token.RightBrace, token.LeftParenthesis:
				// treat as identifier
			default:
				p.next()
				if p.token == token.LeftBrace {
					b := &ast.ClassStaticBlock{
						Static: start,
					}
					b.Block = p.parseFunctionBlock(false, true, false)
					node.Body = append(node.Body, ast.ClassElement{Element: b})
					continue
				}
				static = true
			}
		}

		var kind ast.PropertyKind
		var async bool
		methodBodyStart := p.idx
		if p.literal == "get" || p.literal == "set" {
			if tok := p.peek(); tok != token.Semicolon && tok != token.LeftParenthesis {
				if p.literal == "get" {
					kind = ast.PropertyKindGet
				} else {
					kind = ast.PropertyKindSet
				}
				p.next()
			}
		} else if p.token == token.Async {
			if tok := p.peek(); tok != token.Semicolon && tok != token.LeftParenthesis {
				async = true
				kind = ast.PropertyKindMethod
				p.next()
			}
		}
		generator := false
		if p.token == token.Multiply && (kind == "" || kind == ast.PropertyKindMethod) {
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
			p.error("Classes may not have a static property named 'prototype'")
		}

		if kind == "" && p.token == token.LeftParenthesis {
			kind = ast.PropertyKindMethod
		}

		if kind != "" {
			// method
			if keyName == "constructor" && !computed {
				if !static {
					if kind != ast.PropertyKindMethod {
						p.error("Class constructor may not be an accessor")
					} else if async {
						p.error("Class constructor may not be an async method")
					} else if generator {
						p.error("Class constructor may not be a generator")
					}
				} else if private {
					p.error("Class constructor may not be a private method")
				}
			}
			md := &ast.MethodDefinition{
				Idx:      start,
				Key:      ptrExpr(value),
				Kind:     kind,
				Body:     p.parseMethodDefinition(methodBodyStart, kind, generator, async),
				Static:   static,
				Computed: computed,
			}
			node.Body = append(node.Body, ast.ClassElement{Element: md})
		} else {
			// field
			isCtor := !computed && keyName == "constructor"
			if !isCtor {
				if name, ok := value.(*ast.PrivateIdentifier); ok {
					isCtor = name.Identifier.Name == "constructor"
				}
			}
			if isCtor {
				p.error("Classes may not have a field named 'constructor'")
			}
			var initializer ast.Expr
			if p.token == token.Assign {
				p.next()
				initializer = p.parseExpression()
			}

			if !p.implicitSemicolon && p.token != token.Semicolon && p.token != token.RightBrace {
				p.errorUnexpectedToken(p.token)
				break
			}
			node.Body = append(node.Body, ast.ClassElement{Element: &ast.FieldDefinition{
				Idx:         start,
				Key:         ptrExpr(value),
				Initializer: ptrExpr(initializer),
				Static:      static,
				Computed:    computed,
			}})
		}
	}

	node.RightBrace = p.expect(token.RightBrace)

	return node
}

func (p *parser) parseDebuggerStatement() ast.Stmt {
	idx := p.expect(token.Debugger)

	node := &ast.DebuggerStatement{
		Debugger: idx,
	}

	p.semicolon()

	return node
}

func (p *parser) parseReturnStatement() ast.Stmt {
	idx := p.expect(token.Return)

	if !p.scope.inFunction {
		p.error("Illegal return statement")
		p.nextStatement()
		return &ast.BadStatement{From: idx, To: p.idx}
	}

	node := &ast.ReturnStatement{
		Return: idx,
	}

	if !p.implicitSemicolon && p.token != token.Semicolon && p.token != token.RightBrace && p.token != token.Eof {
		node.Argument = ptrExpr(p.parseExpression())
	}

	p.semicolon()

	return node
}

func (p *parser) parseThrowStatement() ast.Stmt {
	idx := p.expect(token.Throw)

	if p.implicitSemicolon {
		if p.chr == -1 { // Hackish
			p.error("Unexpected end of input")
		} else {
			p.error("Illegal newline after throw")
		}
		p.nextStatement()
		return &ast.BadStatement{From: idx, To: p.idx}
	}

	node := &ast.ThrowStatement{
		Throw:    idx,
		Argument: ptrExpr(p.parseExpression()),
	}

	p.semicolon()

	return node
}

func (p *parser) parseSwitchStatement() ast.Stmt {
	p.expect(token.Switch)
	p.expect(token.LeftParenthesis)
	node := &ast.SwitchStatement{
		Discriminant: ptrExpr(p.parseExpression()),
		Default:      -1,
	}
	p.expect(token.RightParenthesis)

	p.expect(token.LeftBrace)

	inSwitch := p.scope.inSwitch
	p.scope.inSwitch = true
	defer func() {
		p.scope.inSwitch = inSwitch
	}()

	for index := 0; p.token != token.Eof; index++ {
		if p.token == token.RightBrace {
			p.next()
			break
		}

		clause := p.parseCaseStatement()
		if clause.Test == nil {
			if node.Default != -1 {
				p.error("Already saw a default in switch")
			}
			node.Default = index
		}
		node.Body = append(node.Body, clause)
	}

	return node
}

func (p *parser) parseWithStatement() ast.Stmt {
	p.expect(token.With)
	p.expect(token.LeftParenthesis)
	node := &ast.WithStatement{
		Object: ptrExpr(p.parseExpression()),
	}
	p.expect(token.RightParenthesis)
	p.scope.allowLet = false
	node.Body = refStmt(p.parseStatement())

	return node
}

func (p *parser) parseCaseStatement() ast.CaseStatement {
	node := ast.CaseStatement{
		Case: p.idx,
	}
	if p.token == token.Default {
		p.next()
	} else {
		p.expect(token.Case)
		node.Test = ptrExpr(p.parseExpression())
	}
	p.expect(token.Colon)

	for {
		if p.token == token.Eof ||
			p.token == token.RightBrace ||
			p.token == token.Case ||
			p.token == token.Default {
			break
		}
		p.scope.allowLet = true
		node.Consequent = append(node.Consequent, ast.Statement{Stmt: p.parseStatement()})
	}

	return node
}

func (p *parser) parseIterationStatement() ast.Stmt {
	inIteration := p.scope.inIteration
	p.scope.inIteration = true
	defer func() {
		p.scope.inIteration = inIteration
	}()
	p.scope.allowLet = false
	return p.parseStatement()
}

func (p *parser) parseForIn(idx ast.Idx, into ast.ForInto) *ast.ForInStatement {
	// Already have consumed "<into> in"

	source := p.parseExpression()
	p.expect(token.RightParenthesis)

	return &ast.ForInStatement{
		For:    idx,
		Into:   &into,
		Source: ptrExpr(source),
		Body:   refStmt(p.parseIterationStatement()),
	}
}

func (p *parser) parseForOf(idx ast.Idx, into ast.ForInto) *ast.ForOfStatement {
	// Already have consumed "<into> of"

	source := p.parseAssignmentExpression()
	p.expect(token.RightParenthesis)

	return &ast.ForOfStatement{
		For:    idx,
		Into:   &into,
		Source: ptrExpr(source),
		Body:   refStmt(p.parseIterationStatement()),
	}
}

func (p *parser) parseFor(idx ast.Idx, initializer *ast.ForLoopInitializer) *ast.ForStatement {
	// Already have consumed "<initializer> ;"

	var test, update ast.Expr

	if p.token != token.Semicolon {
		test = p.parseExpression()
	}
	p.expect(token.Semicolon)

	if p.token != token.RightParenthesis {
		update = p.parseExpression()
	}
	p.expect(token.RightParenthesis)

	return &ast.ForStatement{
		For:         idx,
		Initializer: initializer,
		Test:        ptrExpr(test),
		Update:      ptrExpr(update),
		Body:        refStmt(p.parseIterationStatement()),
	}
}

func (p *parser) parseForOrForInStatement() ast.Stmt {
	idx := p.expect(token.For)
	p.expect(token.LeftParenthesis)

	var initializer *ast.ForLoopInitializer

	forIn := false
	forOf := false
	var into ast.ForInto
	if p.token != token.Semicolon {

		allowIn := p.scope.allowIn
		p.scope.allowIn = false
		tok := p.token
		if tok == token.Let {
			switch p.peek() {
			case token.Identifier, token.LeftBracket, token.LeftBrace:
			default:
				tok = token.Identifier
			}
		}
		if tok == token.Var || tok == token.Let || tok == token.Const {
			idx := p.idx
			p.next()

			list := p.parseVariableDeclarationList()
			if len(list) == 1 {
				if p.token == token.In {
					p.next() // in
					forIn = true
				} else if p.token == token.Identifier && p.literal == "of" {
					p.next()
					forOf = true
				}
			}
			if forIn || forOf {
				if list[0].Initializer != nil {
					p.error("for-in loop variable declaration may not have an initializer")
				}
				into = ast.ForInto{Into: &ast.VariableDeclaration{
					Token: tok,
					List:  ast.VariableDeclarators{list[0]},
				}}
			} else {
				p.ensurePatternInit(list)

				initializer = &ast.ForLoopInitializer{Initializer: &ast.VariableDeclaration{
					Idx:   idx,
					Token: tok,
					List:  list,
				}}
			}
		} else {
			expr := p.parseExpression()
			if p.token == token.In {
				p.next()
				forIn = true
			} else if p.token == token.Identifier && p.literal == "of" {
				p.next()
				forOf = true
			}
			if forIn || forOf {
				switch e := expr.(type) {
				case *ast.Identifier, *ast.PrivateDotExpression, *ast.VariableDeclarator:
					// These are all acceptable
				case *ast.ObjectLiteral:
					expr = p.reinterpretAsObjectAssignmentPattern(e)
				case *ast.ArrayLiteral:
					expr = p.reinterpretAsArrayAssignmentPattern(e)
				default:
					p.error("Invalid left-hand side in for-in or for-of")
					p.nextStatement()
					return &ast.BadStatement{From: idx, To: p.idx}
				}
				into = ast.ForInto{Into: ptrExpr(expr)}
			} else {
				initializer = &ast.ForLoopInitializer{Initializer: ptrExpr(expr)}
			}
		}
		p.scope.allowIn = allowIn
	}

	if forIn {
		return p.parseForIn(idx, into)
	}
	if forOf {
		return p.parseForOf(idx, into)
	}

	p.expect(token.Semicolon)
	return p.parseFor(idx, initializer)
}

func (p *parser) ensurePatternInit(list []ast.VariableDeclarator) {
	for _, item := range list {
		if _, ok := item.Target.Target.(ast.Pattern); ok {
			if item.Initializer == nil {
				p.error("Missing initializer in destructuring declaration")
				break
			}
		}
	}
}

func (p *parser) parseLexicalDeclaration(tok token.Token) *ast.VariableDeclaration {
	idx := p.expect(tok)
	if !p.scope.allowLet && tok != token.Var {
		p.error("Lexical declaration cannot appear in a single-statement context")
	}

	list := p.parseVariableDeclarationList()
	p.ensurePatternInit(list)
	p.semicolon()

	return &ast.VariableDeclaration{
		Idx:   idx,
		Token: tok,
		List:  list,
	}
}

func (p *parser) parseDoWhileStatement() ast.Stmt {
	inIteration := p.scope.inIteration
	p.scope.inIteration = true
	defer func() {
		p.scope.inIteration = inIteration
	}()

	p.expect(token.Do)
	node := &ast.DoWhileStatement{}
	if p.token == token.LeftBrace {
		node.Body = refStmt(p.parseBlockStatement())
	} else {
		p.scope.allowLet = false
		node.Body = refStmt(p.parseStatement())
	}

	p.expect(token.While)
	p.expect(token.LeftParenthesis)
	node.Test = ptrExpr(p.parseExpression())
	p.expect(token.RightParenthesis)
	if p.token == token.Semicolon {
		p.next()
	}

	return node
}

func (p *parser) parseWhileStatement() ast.Stmt {
	p.expect(token.While)
	p.expect(token.LeftParenthesis)
	node := &ast.WhileStatement{
		Test: ptrExpr(p.parseExpression()),
	}
	p.expect(token.RightParenthesis)
	node.Body = refStmt(p.parseIterationStatement())

	return node
}

func (p *parser) parseIfStatement() ast.Stmt {
	p.expect(token.If)
	p.expect(token.LeftParenthesis)
	node := &ast.IfStatement{
		Test: ptrExpr(p.parseExpression()),
	}
	p.expect(token.RightParenthesis)

	if p.token == token.LeftBrace {
		node.Consequent = refStmt(p.parseBlockStatement())
	} else {
		p.scope.allowLet = false
		node.Consequent = refStmt(p.parseStatement())
	}

	if p.token == token.Else {
		p.next()
		p.scope.allowLet = false
		node.Alternate = refStmt(p.parseStatement())
	}

	return node
}

func (p *parser) parseSourceElements() (body ast.Statements) {
	for p.token != token.Eof {
		p.scope.allowLet = true
		body = append(body, ast.Statement{Stmt: p.parseStatement()})
	}

	return body
}

func (p *parser) parseProgram() *ast.Program {
	return &ast.Program{
		Body: p.parseSourceElements(),
	}
}

func (p *parser) parseBreakStatement() ast.Stmt {
	idx := p.expect(token.Break)
	semicolon := p.implicitSemicolon
	if p.token == token.Semicolon {
		semicolon = true
		p.next()
	}

	if semicolon || p.token == token.RightBrace {
		p.implicitSemicolon = false
		if !p.scope.inIteration && !p.scope.inSwitch {
			goto illegal
		}
		return &ast.BreakStatement{
			Idx: idx,
		}
	}

	p.tokenToBindingId()
	if p.token == token.Identifier {
		identifier := p.parseIdentifier()
		if !p.scope.hasLabel(identifier.Name) {
			p.error(identifier.Name)
			return &ast.BadStatement{From: idx, To: identifier.Idx1()}
		}
		p.semicolon()
		return &ast.BreakStatement{
			Idx:   idx,
			Label: identifier,
		}
	}

	p.expect(token.Identifier)

illegal:
	p.error("Illegal break statement")
	p.nextStatement()
	return &ast.BadStatement{From: idx, To: p.idx}
}

func (p *parser) parseContinueStatement() ast.Stmt {
	idx := p.expect(token.Continue)
	semicolon := p.implicitSemicolon
	if p.token == token.Semicolon {
		semicolon = true
		p.next()
	}

	if semicolon || p.token == token.RightBrace {
		p.implicitSemicolon = false
		if !p.scope.inIteration {
			goto illegal
		}
		return &ast.ContinueStatement{
			Idx: idx,
		}
	}

	p.tokenToBindingId()
	if p.token == token.Identifier {
		identifier := p.parseIdentifier()
		if !p.scope.hasLabel(identifier.Name) {
			p.error(identifier.Name)
			return &ast.BadStatement{From: idx, To: identifier.Idx1()}
		}
		if !p.scope.inIteration {
			goto illegal
		}
		p.semicolon()
		return &ast.ContinueStatement{
			Idx:   idx,
			Label: identifier,
		}
	}

	p.expect(token.Identifier)

illegal:
	p.error("Illegal continue statement")
	p.nextStatement()
	return &ast.BadStatement{From: idx, To: p.idx}
}

// Find the next statement after an error (recover)
func (p *parser) nextStatement() {
	for {
		switch p.token {
		case token.Break, token.Continue,
			token.For, token.If, token.Return, token.Switch,
			token.Var, token.Do, token.Try, token.With,
			token.While, token.Throw, token.Catch, token.Finally:
			// Return only if parser made some progress since last
			// sync or if it has not reached 10 next calls without
			// progress. Otherwise consume at least one token to
			// avoid an endless parser loop
			if p.idx == p.recover.idx && p.recover.count < 10 {
				p.recover.count++
				return
			}
			if p.idx > p.recover.idx {
				p.recover.idx = p.idx
				p.recover.count = 0
				return
			}
			// Reaching here indicates a parser bug, likely an
			// incorrect token list in this function, but it only
			// leads to skipping of possibly correct code if a
			// previous error is present, and thus is preferred
			// over a non-terminating parse.
		case token.Eof:
			return
		}
		p.next()
	}
}

func refStmt(stmt ast.Stmt) *ast.Statement {
	return &ast.Statement{Stmt: stmt}
}
