package parser

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

// nodeAllocator encapsulates typed arenas for all frequently allocated AST
// node types. Constructor methods arena-allocate and initialize nodes in a
// single call, replacing scattered &ast.X{} heap allocations throughout the
// parser.
type nodeAllocator struct {
	// Wrapper types (previously on the parser struct directly).
	expr miniArena[ast.Expression]
	stmt miniArena[ast.Statement]

	// Slice backing arenas — separate from the per-element arenas so that
	// contiguous slice allocations don't fragment with individual node allocs.
	exprSlice miniArena[ast.Expression]
	stmtSlice miniArena[ast.Statement]

	// Concrete expression nodes.
	ident     miniArena[ast.Identifier]
	strLit    miniArena[ast.StringLiteral]
	numLit    miniArena[ast.NumberLiteral]
	boolLit   miniArena[ast.BooleanLiteral]
	nullLit   miniArena[ast.NullLiteral]
	regexpLit miniArena[ast.RegExpLiteral]
	binExpr   miniArena[ast.BinaryExpression]
	unaryExpr miniArena[ast.UnaryExpression]
	updateExp miniArena[ast.UpdateExpression]
	assignExp miniArena[ast.AssignExpression]
	condExpr  miniArena[ast.ConditionalExpression]
	seqExpr   miniArena[ast.SequenceExpression]
	memberExp miniArena[ast.MemberExpression]
	memberPrp miniArena[ast.MemberProperty]
	compProp  miniArena[ast.ComputedProperty]
	callExpr  miniArena[ast.CallExpression]
	newExpr   miniArena[ast.NewExpression]
	spread    miniArena[ast.SpreadElement]
	privIdent miniArena[ast.PrivateIdentifier]
	privDot   miniArena[ast.PrivateDotExpression]
	metaProp  miniArena[ast.MetaProperty]
	optional  miniArena[ast.Optional]
	optChain  miniArena[ast.OptionalChain]
	objLit    miniArena[ast.ObjectLiteral]
	arrLit    miniArena[ast.ArrayLiteral]
	arrPat    miniArena[ast.ArrayPattern]
	objPat    miniArena[ast.ObjectPattern]
	tmplLit   miniArena[ast.TemplateLiteral]
	thisExpr  miniArena[ast.ThisExpression]
	superExpr miniArena[ast.SuperExpression]
	awaitExpr miniArena[ast.AwaitExpression]
	yieldExpr miniArena[ast.YieldExpression]
	arrowFn   miniArena[ast.ArrowFunctionLiteral]
	funcLit   miniArena[ast.FunctionLiteral]
	invalidEx miniArena[ast.InvalidExpression]

	// Property nodes.
	propKeyed miniArena[ast.PropertyKeyed]
	propShort miniArena[ast.PropertyShort]

	// Statement nodes.
	exprStmt  miniArena[ast.ExpressionStatement]
	blockStmt miniArena[ast.BlockStatement]
	retStmt   miniArena[ast.ReturnStatement]
	ifStmt    miniArena[ast.IfStatement]
	throwStmt miniArena[ast.ThrowStatement]
	switchStm miniArena[ast.SwitchStatement]
	withStmt  miniArena[ast.WithStatement]
	tryStmt   miniArena[ast.TryStatement]
	catchStmt miniArena[ast.CatchStatement]
	forStmt   miniArena[ast.ForStatement]
	forInStmt miniArena[ast.ForInStatement]
	forOfStmt miniArena[ast.ForOfStatement]
	whileStmt miniArena[ast.WhileStatement]
	doWhile   miniArena[ast.DoWhileStatement]
	debugStmt miniArena[ast.DebuggerStatement]
	emptyStmt miniArena[ast.EmptyStatement]
	badStmt   miniArena[ast.BadStatement]
	labelStmt miniArena[ast.LabelledStatement]
	breakStmt miniArena[ast.BreakStatement]
	contStmt  miniArena[ast.ContinueStatement]

	// Declaration nodes.
	varDecl  miniArena[ast.VariableDeclaration]
	varDeclr miniArena[ast.VariableDeclarator]
	funcDecl miniArena[ast.FunctionDeclaration]
	classLit miniArena[ast.ClassLiteral]
	classDcl miniArena[ast.ClassDeclaration]
	methDef  miniArena[ast.MethodDefinition]
	fieldDef miniArena[ast.FieldDefinition]
	staticBl miniArena[ast.ClassStaticBlock]

	// Wrapper/helper types.
	bindTgt  miniArena[ast.BindingTarget]
	concBody miniArena[ast.ConciseBody]
	forInit  miniArena[ast.ForLoopInitializer]
	forInto  miniArena[ast.ForInto]

	// String pointers (for Raw fields on StringLiteral/NumberLiteral).
	str miniArena[string]

	// Scopes.
	scopes miniArena[scope]
}

func newNodeAllocator() nodeAllocator {
	return nodeAllocator{
		// Wrapper types — high volume.
		expr: *newArena[ast.Expression](1024),
		stmt: *newArena[ast.Statement](1024),

		// Slice backing arenas.
		exprSlice: *newArena[ast.Expression](1024),
		stmtSlice: *newArena[ast.Statement](1024),

		// Identifiers are the most frequent node.
		ident: *newArena[ast.Identifier](1024),

		// Literals.
		strLit:    *newArena[ast.StringLiteral](256),
		numLit:    *newArena[ast.NumberLiteral](256),
		boolLit:   *newArena[ast.BooleanLiteral](64),
		nullLit:   *newArena[ast.NullLiteral](32),
		regexpLit: *newArena[ast.RegExpLiteral](32),

		// Expressions.
		binExpr:   *newArena[ast.BinaryExpression](256),
		unaryExpr: *newArena[ast.UnaryExpression](64),
		updateExp: *newArena[ast.UpdateExpression](64),
		assignExp: *newArena[ast.AssignExpression](64),
		condExpr:  *newArena[ast.ConditionalExpression](64),
		seqExpr:   *newArena[ast.SequenceExpression](32),
		memberExp: *newArena[ast.MemberExpression](256),
		memberPrp: *newArena[ast.MemberProperty](256),
		compProp:  *newArena[ast.ComputedProperty](64),
		callExpr:  *newArena[ast.CallExpression](256),
		newExpr:   *newArena[ast.NewExpression](32),
		spread:    *newArena[ast.SpreadElement](64),
		privIdent: *newArena[ast.PrivateIdentifier](32),
		privDot:   *newArena[ast.PrivateDotExpression](32),
		metaProp:  *newArena[ast.MetaProperty](8),
		optional:  *newArena[ast.Optional](32),
		optChain:  *newArena[ast.OptionalChain](32),
		objLit:    *newArena[ast.ObjectLiteral](64),
		arrLit:    *newArena[ast.ArrayLiteral](64),
		arrPat:    *newArena[ast.ArrayPattern](32),
		objPat:    *newArena[ast.ObjectPattern](32),
		tmplLit:   *newArena[ast.TemplateLiteral](32),
		thisExpr:  *newArena[ast.ThisExpression](32),
		superExpr: *newArena[ast.SuperExpression](8),
		awaitExpr: *newArena[ast.AwaitExpression](32),
		yieldExpr: *newArena[ast.YieldExpression](32),
		arrowFn:   *newArena[ast.ArrowFunctionLiteral](64),
		funcLit:   *newArena[ast.FunctionLiteral](64),
		invalidEx: *newArena[ast.InvalidExpression](32),

		// Properties.
		propKeyed: *newArena[ast.PropertyKeyed](128),
		propShort: *newArena[ast.PropertyShort](64),

		// Statements.
		exprStmt:  *newArena[ast.ExpressionStatement](256),
		blockStmt: *newArena[ast.BlockStatement](128),
		retStmt:   *newArena[ast.ReturnStatement](64),
		ifStmt:    *newArena[ast.IfStatement](64),
		throwStmt: *newArena[ast.ThrowStatement](32),
		switchStm: *newArena[ast.SwitchStatement](16),
		withStmt:  *newArena[ast.WithStatement](8),
		tryStmt:   *newArena[ast.TryStatement](16),
		catchStmt: *newArena[ast.CatchStatement](16),
		forStmt:   *newArena[ast.ForStatement](32),
		forInStmt: *newArena[ast.ForInStatement](16),
		forOfStmt: *newArena[ast.ForOfStatement](16),
		whileStmt: *newArena[ast.WhileStatement](32),
		doWhile:   *newArena[ast.DoWhileStatement](16),
		debugStmt: *newArena[ast.DebuggerStatement](8),
		emptyStmt: *newArena[ast.EmptyStatement](16),
		badStmt:   *newArena[ast.BadStatement](8),
		labelStmt: *newArena[ast.LabelledStatement](16),
		breakStmt: *newArena[ast.BreakStatement](16),
		contStmt:  *newArena[ast.ContinueStatement](16),

		// Declarations.
		varDecl:  *newArena[ast.VariableDeclaration](64),
		varDeclr: *newArena[ast.VariableDeclarator](128),
		funcDecl: *newArena[ast.FunctionDeclaration](32),
		classLit: *newArena[ast.ClassLiteral](16),
		classDcl: *newArena[ast.ClassDeclaration](16),
		methDef:  *newArena[ast.MethodDefinition](32),
		fieldDef: *newArena[ast.FieldDefinition](32),
		staticBl: *newArena[ast.ClassStaticBlock](8),

		// Wrappers.
		bindTgt:  *newArena[ast.BindingTarget](128),
		concBody: *newArena[ast.ConciseBody](64),
		forInit:  *newArena[ast.ForLoopInitializer](32),
		forInto:  *newArena[ast.ForInto](16),

		// String pointers.
		str: *newArena[string](256),

		// Scopes.
		scopes: *newArena[scope](64),
	}
}

// ---------------------------------------------------------------------------
// Wrapper constructors
// ---------------------------------------------------------------------------

func (a *nodeAllocator) Expression(expr ast.Expr) *ast.Expression {
	n := a.expr.make()
	n.Expr = expr
	return n
}

func (a *nodeAllocator) Statement(stmt ast.Stmt) *ast.Statement {
	n := a.stmt.make()
	n.Stmt = stmt
	return n
}

// CopyExpressions allocates a contiguous []Expression from the arena and copies
// src into it. The returned slice's backing array lives in arena memory.
func (a *nodeAllocator) CopyExpressions(src []ast.Expression) ast.Expressions {
	if len(src) == 0 {
		return nil
	}
	dst := a.exprSlice.makeSlice(len(src))
	copy(dst, src)
	return dst
}

// CopyStatements allocates a contiguous []Statement from the arena and copies
// src into it. The returned slice's backing array lives in arena memory.
func (a *nodeAllocator) CopyStatements(src []ast.Statement) ast.Statements {
	if len(src) == 0 {
		return nil
	}
	dst := a.stmtSlice.makeSlice(len(src))
	copy(dst, src)
	return dst
}

// stringPtr arena-allocates a *string, avoiding the heap escape of &localVar.
func (a *nodeAllocator) stringPtr(s string) *string {
	p := a.str.make()
	*p = s
	return p
}

// ---------------------------------------------------------------------------
// Identifier / literals
// ---------------------------------------------------------------------------

func (a *nodeAllocator) Identifier(idx ast.Idx, name string) *ast.Identifier {
	n := a.ident.make()
	*n = ast.Identifier{Idx: idx, Name: name}
	return n
}

func (a *nodeAllocator) StringLiteral(idx ast.Idx, value, raw string) *ast.StringLiteral {
	n := a.strLit.make()
	*n = ast.StringLiteral{Idx: idx, Value: value, Raw: a.stringPtr(raw)}
	return n
}

func (a *nodeAllocator) NumberLiteral(idx ast.Idx, value float64, raw string) *ast.NumberLiteral {
	n := a.numLit.make()
	*n = ast.NumberLiteral{Idx: idx, Value: value, Raw: a.stringPtr(raw)}
	return n
}

func (a *nodeAllocator) BooleanLiteral(idx ast.Idx, value bool) *ast.BooleanLiteral {
	n := a.boolLit.make()
	*n = ast.BooleanLiteral{Idx: idx, Value: value}
	return n
}

func (a *nodeAllocator) NullLiteral(idx ast.Idx) *ast.NullLiteral {
	n := a.nullLit.make()
	*n = ast.NullLiteral{Idx: idx}
	return n
}

func (a *nodeAllocator) RegExpLiteral(idx ast.Idx, literal, pattern, flags string) *ast.RegExpLiteral {
	n := a.regexpLit.make()
	*n = ast.RegExpLiteral{Idx: idx, Literal: literal, Pattern: pattern, Flags: flags}
	return n
}

func (a *nodeAllocator) BinaryExpression(op token.Token, left, right *ast.Expression) *ast.BinaryExpression {
	n := a.binExpr.make()
	*n = ast.BinaryExpression{Operator: op, Left: left, Right: right}
	return n
}

func (a *nodeAllocator) UnaryExpression(op token.Token, idx ast.Idx, operand *ast.Expression) *ast.UnaryExpression {
	n := a.unaryExpr.make()
	*n = ast.UnaryExpression{Operator: op, Idx: idx, Operand: operand}
	return n
}

func (a *nodeAllocator) UpdateExpression(op token.Token, idx ast.Idx, operand *ast.Expression, postfix bool) *ast.UpdateExpression {
	n := a.updateExp.make()
	*n = ast.UpdateExpression{Operator: op, Idx: idx, Operand: operand, Postfix: postfix}
	return n
}

func (a *nodeAllocator) AssignExpression(op token.Token, left, right *ast.Expression) *ast.AssignExpression {
	n := a.assignExp.make()
	*n = ast.AssignExpression{Operator: op, Left: left, Right: right}
	return n
}

func (a *nodeAllocator) ConditionalExpression(test, consequent, alternate *ast.Expression) *ast.ConditionalExpression {
	n := a.condExpr.make()
	*n = ast.ConditionalExpression{Test: test, Consequent: consequent, Alternate: alternate}
	return n
}

func (a *nodeAllocator) SequenceExpression(seq ast.Expressions) *ast.SequenceExpression {
	n := a.seqExpr.make()
	*n = ast.SequenceExpression{Sequence: seq}
	return n
}

func (a *nodeAllocator) MemberExpression(object *ast.Expression, property *ast.MemberProperty) *ast.MemberExpression {
	n := a.memberExp.make()
	*n = ast.MemberExpression{Object: object, Property: property}
	return n
}

func (a *nodeAllocator) MemberProperty(prop ast.MemberProp) *ast.MemberProperty {
	n := a.memberPrp.make()
	*n = ast.MemberProperty{Prop: prop}
	return n
}

func (a *nodeAllocator) ComputedProperty(expr *ast.Expression) *ast.ComputedProperty {
	n := a.compProp.make()
	*n = ast.ComputedProperty{Expr: expr}
	return n
}

func (a *nodeAllocator) CallExpression(callee *ast.Expression, lp ast.Idx, args ast.Expressions, rp ast.Idx) *ast.CallExpression {
	n := a.callExpr.make()
	*n = ast.CallExpression{Callee: callee, LeftParenthesis: lp, ArgumentList: args, RightParenthesis: rp}
	return n
}

func (a *nodeAllocator) NewExpression(idx ast.Idx, callee *ast.Expression) *ast.NewExpression {
	n := a.newExpr.make()
	*n = ast.NewExpression{New: idx, Callee: callee}
	return n
}

func (a *nodeAllocator) SpreadElement(expr *ast.Expression) *ast.SpreadElement {
	n := a.spread.make()
	*n = ast.SpreadElement{Expression: expr}
	return n
}

func (a *nodeAllocator) PrivateIdentifier(ident *ast.Identifier) *ast.PrivateIdentifier {
	n := a.privIdent.make()
	*n = ast.PrivateIdentifier{Identifier: ident}
	return n
}

func (a *nodeAllocator) PrivateDotExpression(left *ast.Expression, ident *ast.PrivateIdentifier) *ast.PrivateDotExpression {
	n := a.privDot.make()
	*n = ast.PrivateDotExpression{Left: left, Identifier: ident}
	return n
}

func (a *nodeAllocator) MetaProperty(meta, property *ast.Identifier, idx ast.Idx) *ast.MetaProperty {
	n := a.metaProp.make()
	*n = ast.MetaProperty{Meta: meta, Property: property, Idx: idx}
	return n
}

func (a *nodeAllocator) Optional(expr *ast.Expression) *ast.Optional {
	n := a.optional.make()
	*n = ast.Optional{Expr: expr}
	return n
}

func (a *nodeAllocator) OptionalChain(base *ast.Expression) *ast.OptionalChain {
	n := a.optChain.make()
	*n = ast.OptionalChain{Base: base}
	return n
}

func (a *nodeAllocator) ObjectLiteral(lb, rb ast.Idx, value []ast.Property) *ast.ObjectLiteral {
	n := a.objLit.make()
	*n = ast.ObjectLiteral{LeftBrace: lb, RightBrace: rb, Value: value}
	return n
}

func (a *nodeAllocator) ArrayLiteral(lb, rb ast.Idx, value ast.Expressions) *ast.ArrayLiteral {
	n := a.arrLit.make()
	*n = ast.ArrayLiteral{LeftBracket: lb, RightBracket: rb, Value: value}
	return n
}

func (a *nodeAllocator) ArrayPattern(lb, rb ast.Idx, elems ast.Expressions, rest *ast.Expression) *ast.ArrayPattern {
	n := a.arrPat.make()
	*n = ast.ArrayPattern{LeftBracket: lb, RightBracket: rb, Elements: elems, Rest: rest}
	return n
}

func (a *nodeAllocator) ObjectPattern(lb, rb ast.Idx, props ast.Properties, rest ast.Expr) *ast.ObjectPattern {
	n := a.objPat.make()
	*n = ast.ObjectPattern{LeftBrace: lb, RightBrace: rb, Properties: props, Rest: rest}
	return n
}

func (a *nodeAllocator) TemplateLiteral(openQuote ast.Idx) *ast.TemplateLiteral {
	n := a.tmplLit.make()
	*n = ast.TemplateLiteral{OpenQuote: openQuote}
	return n
}

func (a *nodeAllocator) ThisExpression(idx ast.Idx) *ast.ThisExpression {
	n := a.thisExpr.make()
	*n = ast.ThisExpression{Idx: idx}
	return n
}

func (a *nodeAllocator) SuperExpression(idx ast.Idx) *ast.SuperExpression {
	n := a.superExpr.make()
	*n = ast.SuperExpression{Idx: idx}
	return n
}

func (a *nodeAllocator) AwaitExpression(idx ast.Idx, argument *ast.Expression) *ast.AwaitExpression {
	n := a.awaitExpr.make()
	*n = ast.AwaitExpression{Await: idx, Argument: argument}
	return n
}

func (a *nodeAllocator) YieldExpression(idx ast.Idx) *ast.YieldExpression {
	n := a.yieldExpr.make()
	*n = ast.YieldExpression{Yield: idx}
	return n
}

func (a *nodeAllocator) ArrowFunctionLiteral(start ast.Idx, params ast.ParameterList, async bool) *ast.ArrowFunctionLiteral {
	n := a.arrowFn.make()
	*n = ast.ArrowFunctionLiteral{Start: start, ParameterList: params, Async: async}
	return n
}

func (a *nodeAllocator) FunctionLiteral(start ast.Idx, async bool) *ast.FunctionLiteral {
	n := a.funcLit.make()
	*n = ast.FunctionLiteral{Function: start, Async: async}
	return n
}

func (a *nodeAllocator) InvalidExpression(from, to ast.Idx) *ast.InvalidExpression {
	n := a.invalidEx.make()
	*n = ast.InvalidExpression{From: from, To: to}
	return n
}

func (a *nodeAllocator) PropertyKeyed(key *ast.Expression, kind ast.PropertyKind, value *ast.Expression, computed bool) *ast.PropertyKeyed {
	n := a.propKeyed.make()
	*n = ast.PropertyKeyed{Key: key, Kind: kind, Value: value, Computed: computed}
	return n
}

func (a *nodeAllocator) PropertyShort(name *ast.Identifier, initializer *ast.Expression) *ast.PropertyShort {
	n := a.propShort.make()
	*n = ast.PropertyShort{Name: name, Initializer: initializer}
	return n
}

func (a *nodeAllocator) ExpressionStatement(expr *ast.Expression) *ast.ExpressionStatement {
	n := a.exprStmt.make()
	*n = ast.ExpressionStatement{Expression: expr}
	return n
}

func (a *nodeAllocator) BlockStatement() *ast.BlockStatement {
	return a.blockStmt.make()
}

func (a *nodeAllocator) ReturnStatement(idx ast.Idx) *ast.ReturnStatement {
	n := a.retStmt.make()
	*n = ast.ReturnStatement{Return: idx}
	return n
}

func (a *nodeAllocator) IfStatement(test *ast.Expression) *ast.IfStatement {
	n := a.ifStmt.make()
	*n = ast.IfStatement{Test: test}
	return n
}

func (a *nodeAllocator) ThrowStatement(idx ast.Idx, argument *ast.Expression) *ast.ThrowStatement {
	n := a.throwStmt.make()
	*n = ast.ThrowStatement{Throw: idx, Argument: argument}
	return n
}

func (a *nodeAllocator) SwitchStatement(discriminant *ast.Expression) *ast.SwitchStatement {
	n := a.switchStm.make()
	*n = ast.SwitchStatement{Discriminant: discriminant, Default: -1}
	return n
}

func (a *nodeAllocator) WithStatement(object *ast.Expression) *ast.WithStatement {
	n := a.withStmt.make()
	*n = ast.WithStatement{Object: object}
	return n
}

func (a *nodeAllocator) TryStatement(idx ast.Idx, body *ast.BlockStatement) *ast.TryStatement {
	n := a.tryStmt.make()
	*n = ast.TryStatement{Try: idx, Body: body}
	return n
}

func (a *nodeAllocator) CatchStatement(idx ast.Idx, param *ast.BindingTarget, body *ast.BlockStatement) *ast.CatchStatement {
	n := a.catchStmt.make()
	*n = ast.CatchStatement{Catch: idx, Parameter: param, Body: body}
	return n
}

func (a *nodeAllocator) ForStatement(idx ast.Idx, init *ast.ForLoopInitializer, test, update *ast.Expression, body *ast.Statement) *ast.ForStatement {
	n := a.forStmt.make()
	*n = ast.ForStatement{For: idx, Initializer: init, Test: test, Update: update, Body: body}
	return n
}

func (a *nodeAllocator) ForInStatement(idx ast.Idx, into *ast.ForInto, source *ast.Expression, body *ast.Statement) *ast.ForInStatement {
	n := a.forInStmt.make()
	*n = ast.ForInStatement{For: idx, Into: into, Source: source, Body: body}
	return n
}

func (a *nodeAllocator) ForOfStatement(idx ast.Idx, into *ast.ForInto, source *ast.Expression, body *ast.Statement) *ast.ForOfStatement {
	n := a.forOfStmt.make()
	*n = ast.ForOfStatement{For: idx, Into: into, Source: source, Body: body}
	return n
}

func (a *nodeAllocator) WhileStatement(test *ast.Expression) *ast.WhileStatement {
	n := a.whileStmt.make()
	*n = ast.WhileStatement{Test: test}
	return n
}

func (a *nodeAllocator) DoWhileStatement() *ast.DoWhileStatement {
	return a.doWhile.make()
}

func (a *nodeAllocator) DebuggerStatement(idx ast.Idx) *ast.DebuggerStatement {
	n := a.debugStmt.make()
	*n = ast.DebuggerStatement{Debugger: idx}
	return n
}

func (a *nodeAllocator) EmptyStatement(idx ast.Idx) *ast.EmptyStatement {
	n := a.emptyStmt.make()
	*n = ast.EmptyStatement{Semicolon: idx}
	return n
}

func (a *nodeAllocator) BadStatement(from, to ast.Idx) *ast.BadStatement {
	n := a.badStmt.make()
	*n = ast.BadStatement{From: from, To: to}
	return n
}

func (a *nodeAllocator) LabelledStatement(label *ast.Identifier, colon ast.Idx, stmt *ast.Statement) *ast.LabelledStatement {
	n := a.labelStmt.make()
	*n = ast.LabelledStatement{Label: label, Colon: colon, Statement: stmt}
	return n
}

func (a *nodeAllocator) BreakStatement(idx ast.Idx, label *ast.Identifier) *ast.BreakStatement {
	n := a.breakStmt.make()
	*n = ast.BreakStatement{Idx: idx, Label: label}
	return n
}

func (a *nodeAllocator) ContinueStatement(idx ast.Idx, label *ast.Identifier) *ast.ContinueStatement {
	n := a.contStmt.make()
	*n = ast.ContinueStatement{Idx: idx, Label: label}
	return n
}

func (a *nodeAllocator) VariableDeclaration(idx ast.Idx, tok token.Token, list ast.VariableDeclarators) *ast.VariableDeclaration {
	n := a.varDecl.make()
	*n = ast.VariableDeclaration{Idx: idx, Token: tok, List: list}
	return n
}

func (a *nodeAllocator) VariableDeclarator(target *ast.BindingTarget) *ast.VariableDeclarator {
	n := a.varDeclr.make()
	*n = ast.VariableDeclarator{Target: target}
	return n
}

func (a *nodeAllocator) FunctionDeclaration(fn *ast.FunctionLiteral) *ast.FunctionDeclaration {
	n := a.funcDecl.make()
	*n = ast.FunctionDeclaration{Function: fn}
	return n
}

func (a *nodeAllocator) ClassLiteral(idx ast.Idx) *ast.ClassLiteral {
	n := a.classLit.make()
	*n = ast.ClassLiteral{Class: idx}
	return n
}

func (a *nodeAllocator) ClassDeclaration(class *ast.ClassLiteral) *ast.ClassDeclaration {
	n := a.classDcl.make()
	*n = ast.ClassDeclaration{Class: class}
	return n
}

func (a *nodeAllocator) MethodDefinition(idx ast.Idx, key *ast.Expression, kind ast.PropertyKind, body *ast.FunctionLiteral, static, computed bool) *ast.MethodDefinition {
	n := a.methDef.make()
	*n = ast.MethodDefinition{Idx: idx, Key: key, Kind: kind, Body: body, Static: static, Computed: computed}
	return n
}

func (a *nodeAllocator) FieldDefinition(idx ast.Idx, key, initializer *ast.Expression, static, computed bool) *ast.FieldDefinition {
	n := a.fieldDef.make()
	*n = ast.FieldDefinition{Idx: idx, Key: key, Initializer: initializer, Static: static, Computed: computed}
	return n
}

func (a *nodeAllocator) ClassStaticBlock(idx ast.Idx) *ast.ClassStaticBlock {
	n := a.staticBl.make()
	*n = ast.ClassStaticBlock{Static: idx}
	return n
}

func (a *nodeAllocator) BindingTarget(target ast.Target) *ast.BindingTarget {
	n := a.bindTgt.make()
	*n = ast.BindingTarget{Target: target}
	return n
}

func (a *nodeAllocator) ConciseBody(body ast.Body) *ast.ConciseBody {
	n := a.concBody.make()
	*n = ast.ConciseBody{Body: body}
	return n
}

func (a *nodeAllocator) ForLoopInitializer(init ast.ForLoopInit) *ast.ForLoopInitializer {
	n := a.forInit.make()
	*n = ast.ForLoopInitializer{Initializer: init}
	return n
}

func (a *nodeAllocator) ForIntoPtr(into ast.Into) *ast.ForInto {
	n := a.forInto.make()
	*n = ast.ForInto{Into: into}
	return n
}
