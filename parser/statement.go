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
	for p.currentKind() != token.RightBrace && p.currentKind() != token.Eof {
		p.scope.allowLet = true
		list = append(list, ast.Statement{Stmt: p.parseStatement()})
	}

	return list
}

func (p *parser) parseStatement() ast.Stmt {
	if p.currentKind() == token.Eof {
		p.errorUnexpectedToken(p.currentKind())
		return &ast.BadStatement{From: p.currentOffset(), To: p.currentOffset() + 1}
	}

	switch p.currentKind() {
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
		return p.parseLexicalDeclaration(p.currentKind())
	case token.Let:
		tok := p.peek().Kind
		if tok == token.LeftBracket || p.scope.allowLet && (token.ID(tok) || tok == token.LeftBrace) {
			return p.parseLexicalDeclaration(p.currentKind())
		}
		p.insertSemicolon = true
	case token.Const:
		return p.parseLexicalDeclaration(p.currentKind())
	case token.Async:
		if f := p.parseMaybeAsyncFunction(true); f != nil {
			return &ast.FunctionDeclaration{
				Function: f,
			}
		}
	case token.Function:
		return &ast.FunctionDeclaration{
			Function: p.parseFunction(true, false, p.currentOffset()),
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

	if identifier, isIdentifier := expression.(*ast.Identifier); isIdentifier && p.currentKind() == token.Colon {
		// LabelledStatement
		colon := p.currentOffset()
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
			Statement: p.makeStmt(statement),
		}
	}

	p.semicolon()

	return &ast.ExpressionStatement{
		Expression: p.makeExpr(expression),
	}
}

func (p *parser) parseTryStatement() ast.Stmt {
	node := &ast.TryStatement{
		Try:  p.expect(token.Try),
		Body: p.parseBlockStatement(),
	}

	if p.currentKind() == token.Catch {
		catch := p.currentOffset()
		p.next()
		var parameter *ast.BindingTarget
		if p.currentKind() == token.LeftParenthesis {
			p.next()
			parameter = &ast.BindingTarget{Target: p.parseBindingTarget()}
			p.expect(token.RightParenthesis)
		}
		node.Catch = &ast.CatchStatement{
			Catch:     catch,
			Parameter: parameter,
			Body:      p.parseBlockStatement(),
		}
	}

	if p.currentKind() == token.Finally {
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
	for p.currentKind() != token.RightParenthesis && p.currentKind() != token.Eof {
		if p.currentKind() == token.Ellipsis {
			p.next()
			rest = p.reinterpretAsDestructBindingTarget(p.parseAssignmentExpression())
			break
		}
		p.parseVariableDeclaration(&list)
		if p.currentKind() != token.RightParenthesis {
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
	if p.peek().Kind == token.Function {
		idx := p.currentOffset()
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

	if p.currentKind() == token.Multiply {
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
	if p.currentKind() == token.Identifier {
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
	if p.currentKind() == token.LeftBrace {
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
		Body: p.makeExpr(p.parseAssignmentExpression()),
	}
}

func (p *parser) parseClass(declaration bool) *ast.ClassLiteral {
	if !p.scope.allowLet && p.currentKind() == token.Class {
		p.errorUnexpectedToken(token.Class)
	}

	node := &ast.ClassLiteral{
		Class: p.expect(token.Class),
	}

	p.tokenToBindingId()
	name := &ast.Identifier{}
	if p.currentKind() == token.Identifier {
		name = p.parseIdentifier()
	} else if declaration {
		// Use expect error handling
		p.expect(token.Identifier)
	}

	node.Name = name

	if p.currentKind() != token.LeftBrace {
		p.expect(token.Extends)
		node.SuperClass = p.makeExpr(p.parseLeftHandSideExpressionAllowCall())
	}

	p.expect(token.LeftBrace)

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
			p.error("Classes may not have a static property named 'prototype'")
		}

		if kind == "" && p.currentKind() == token.LeftParenthesis {
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
				Key:      p.makeExpr(value),
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
			if p.currentKind() == token.Assign {
				p.next()
				initializer = p.parseExpression()
			}

			if !p.implicitSemicolon && p.currentKind() != token.Semicolon && p.currentKind() != token.RightBrace {
				p.errorUnexpectedToken(p.currentKind())
				break
			}
			node.Body = append(node.Body, ast.ClassElement{Element: &ast.FieldDefinition{
				Idx:         start,
				Key:         p.makeExpr(value),
				Initializer: p.makeExpr(initializer),
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
		return &ast.BadStatement{From: idx, To: p.currentOffset()}
	}

	node := &ast.ReturnStatement{
		Return: idx,
	}

	if !p.implicitSemicolon && p.currentKind() != token.Semicolon && p.currentKind() != token.RightBrace && p.currentKind() != token.Eof {
		node.Argument = p.makeExpr(p.parseExpression())
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
		return &ast.BadStatement{From: idx, To: p.currentOffset()}
	}

	node := &ast.ThrowStatement{
		Throw:    idx,
		Argument: p.makeExpr(p.parseExpression()),
	}

	p.semicolon()

	return node
}

func (p *parser) parseSwitchStatement() ast.Stmt {
	p.expect(token.Switch)
	p.expect(token.LeftParenthesis)
	node := &ast.SwitchStatement{
		Discriminant: p.makeExpr(p.parseExpression()),
		Default:      -1,
	}
	p.expect(token.RightParenthesis)

	p.expect(token.LeftBrace)

	inSwitch := p.scope.inSwitch
	p.scope.inSwitch = true
	defer func() {
		p.scope.inSwitch = inSwitch
	}()

	for index := 0; p.currentKind() != token.Eof; index++ {
		if p.currentKind() == token.RightBrace {
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
		Object: p.makeExpr(p.parseExpression()),
	}
	p.expect(token.RightParenthesis)
	p.scope.allowLet = false
	node.Body = p.makeStmt(p.parseStatement())

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
		node.Test = p.makeExpr(p.parseExpression())
	}
	p.expect(token.Colon)

	for {
		if p.currentKind() == token.Eof ||
			p.currentKind() == token.RightBrace ||
			p.currentKind() == token.Case ||
			p.currentKind() == token.Default {
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
		Source: p.makeExpr(source),
		Body:   p.makeStmt(p.parseIterationStatement()),
	}
}

func (p *parser) parseForOf(idx ast.Idx, into ast.ForInto) *ast.ForOfStatement {
	// Already have consumed "<into> of"

	source := p.parseAssignmentExpression()
	p.expect(token.RightParenthesis)

	return &ast.ForOfStatement{
		For:    idx,
		Into:   &into,
		Source: p.makeExpr(source),
		Body:   p.makeStmt(p.parseIterationStatement()),
	}
}

func (p *parser) parseFor(idx ast.Idx, initializer *ast.ForLoopInitializer) *ast.ForStatement {
	// Already have consumed "<initializer> ;"

	var test, update ast.Expr

	if p.currentKind() != token.Semicolon {
		test = p.parseExpression()
	}
	p.expect(token.Semicolon)

	if p.currentKind() != token.RightParenthesis {
		update = p.parseExpression()
	}
	p.expect(token.RightParenthesis)

	return &ast.ForStatement{
		For:         idx,
		Initializer: initializer,
		Test:        p.makeExpr(test),
		Update:      p.makeExpr(update),
		Body:        p.makeStmt(p.parseIterationStatement()),
	}
}

func (p *parser) parseForOrForInStatement() ast.Stmt {
	idx := p.expect(token.For)
	p.expect(token.LeftParenthesis)

	var initializer *ast.ForLoopInitializer

	forIn := false
	forOf := false
	var into ast.ForInto
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
			if p.currentKind() == token.In {
				p.next()
				forIn = true
			} else if p.currentKind() == token.Of {
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
					return &ast.BadStatement{From: idx, To: p.currentOffset()}
				}
				into = ast.ForInto{Into: p.makeExpr(expr)}
			} else {
				initializer = &ast.ForLoopInitializer{Initializer: p.makeExpr(expr)}
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
	if p.currentKind() == token.LeftBrace {
		node.Body = p.makeStmt(p.parseBlockStatement())
	} else {
		p.scope.allowLet = false
		node.Body = p.makeStmt(p.parseStatement())
	}

	p.expect(token.While)
	p.expect(token.LeftParenthesis)
	node.Test = p.makeExpr(p.parseExpression())
	p.expect(token.RightParenthesis)
	if p.currentKind() == token.Semicolon {
		p.next()
	}

	return node
}

func (p *parser) parseWhileStatement() ast.Stmt {
	p.expect(token.While)
	p.expect(token.LeftParenthesis)
	node := &ast.WhileStatement{
		Test: p.makeExpr(p.parseExpression()),
	}
	p.expect(token.RightParenthesis)
	node.Body = p.makeStmt(p.parseIterationStatement())

	return node
}

func (p *parser) parseIfStatement() ast.Stmt {
	p.expect(token.If)
	p.expect(token.LeftParenthesis)
	node := &ast.IfStatement{
		Test: p.makeExpr(p.parseExpression()),
	}
	p.expect(token.RightParenthesis)

	if p.currentKind() == token.LeftBrace {
		node.Consequent = p.makeStmt(p.parseBlockStatement())
	} else {
		p.scope.allowLet = false
		node.Consequent = p.makeStmt(p.parseStatement())
	}

	if p.currentKind() == token.Else {
		p.next()
		p.scope.allowLet = false
		node.Alternate = p.makeStmt(p.parseStatement())
	}

	return node
}

func (p *parser) parseSourceElements() (body ast.Statements) {
	for p.currentKind() != token.Eof {
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
	if p.currentKind() == token.Semicolon {
		semicolon = true
		p.next()
	}

	if semicolon || p.currentKind() == token.RightBrace {
		p.implicitSemicolon = false
		if !p.scope.inIteration && !p.scope.inSwitch {
			goto illegal
		}
		return &ast.BreakStatement{
			Idx: idx,
		}
	}

	p.tokenToBindingId()
	if p.currentKind() == token.Identifier {
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
	return &ast.BadStatement{From: idx, To: p.currentOffset()}
}

func (p *parser) parseContinueStatement() ast.Stmt {
	idx := p.expect(token.Continue)
	semicolon := p.implicitSemicolon
	if p.currentKind() == token.Semicolon {
		semicolon = true
		p.next()
	}

	if semicolon || p.currentKind() == token.RightBrace {
		p.implicitSemicolon = false
		if !p.scope.inIteration {
			goto illegal
		}
		return &ast.ContinueStatement{
			Idx: idx,
		}
	}

	p.tokenToBindingId()
	if p.currentKind() == token.Identifier {
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
	return &ast.BadStatement{From: idx, To: p.currentOffset()}
}

// Find the next statement after an error (recover)
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
			// previous error is present, and thus is preferred
			// over a non-terminating parse.
		case token.Eof:
			return
		}
		p.next()
	}
}
