// Code generated by gen_clone.go; DO NOT EDIT.
package ast

func (n *ArrayLiteral) Clone() *ArrayLiteral {
	return &ArrayLiteral{LeftBracket: n.LeftBracket, RightBracket: n.RightBracket, Value: *n.Value.Clone()}
}
func (n *ArrayPattern) Clone() *ArrayPattern {
	return &ArrayPattern{LeftBracket: n.LeftBracket, RightBracket: n.RightBracket, Elements: *n.Elements.Clone(), Rest: n.Rest.Clone()}
}
func (n *ArrowFunctionLiteral) Clone() *ArrowFunctionLiteral {
	return &ArrowFunctionLiteral{Start: n.Start, ParameterList: *n.ParameterList.Clone(), Body: n.Body.Clone(), Async: n.Async, ScopeContext: n.ScopeContext}
}
func (n *AssignExpression) Clone() *AssignExpression {
	return &AssignExpression{Operator: n.Operator, Left: n.Left.Clone(), Right: n.Right.Clone()}
}
func (n *AwaitExpression) Clone() *AwaitExpression {
	return &AwaitExpression{Await: n.Await, Argument: n.Argument.Clone()}
}
func (n *BadStatement) Clone() *BadStatement {
	return &BadStatement{From: n.From, To: n.To}
}
func (n *BinaryExpression) Clone() *BinaryExpression {
	return &BinaryExpression{Operator: n.Operator, Left: n.Left.Clone(), Right: n.Right.Clone()}
}
func (n *BindingTarget) Clone() *BindingTarget {
	var clonedTarget Target
	switch target := n.Target.(type) {
	case *ArrayPattern:
		clonedTarget = target.Clone()
	case *Identifier:
		clonedTarget = target.Clone()
	case *InvalidExpression:
		clonedTarget = target.Clone()
	case *MemberExpression:
		clonedTarget = target.Clone()
	case *ObjectPattern:
		clonedTarget = target.Clone()
	}
	return &BindingTarget{Target: clonedTarget}
}
func (n *BlockStatement) Clone() *BlockStatement {
	return &BlockStatement{LeftBrace: n.LeftBrace, List: *n.List.Clone(), RightBrace: n.RightBrace, ScopeContext: n.ScopeContext}
}
func (n *BooleanLiteral) Clone() *BooleanLiteral {
	return &BooleanLiteral{Idx: n.Idx, Value: n.Value}
}
func (n *BreakStatement) Clone() *BreakStatement {
	var label *Identifier
	if n.Label != nil {
		label = n.Label.Clone()
	}
	return &BreakStatement{Idx: n.Idx, Label: label}
}
func (n *CallExpression) Clone() *CallExpression {
	return &CallExpression{Callee: n.Callee.Clone(), LeftParenthesis: n.LeftParenthesis, ArgumentList: *n.ArgumentList.Clone(), RightParenthesis: n.RightParenthesis}
}
func (n *CaseStatement) Clone() *CaseStatement {
	var test *Expression
	if n.Test != nil {
		test = n.Test.Clone()
	}
	return &CaseStatement{Case: n.Case, Test: test, Consequent: *n.Consequent.Clone()}
}
func (n *CaseStatements) Clone() *CaseStatements {
	ns := make(CaseStatements, len(*n))
	for i := range *n {
		ns[i] = *(*n)[i].Clone()
	}
	return &ns
}
func (n *CatchStatement) Clone() *CatchStatement {
	var parameter *BindingTarget
	if n.Parameter != nil {
		parameter = n.Parameter.Clone()
	}
	return &CatchStatement{Catch: n.Catch, Parameter: parameter, Body: n.Body.Clone()}
}
func (n *ClassDeclaration) Clone() *ClassDeclaration {
	return &ClassDeclaration{Class: n.Class.Clone()}
}
func (n *ClassElement) Clone() *ClassElement {
	var clonedElement Element
	switch element := n.Element.(type) {
	case *ClassStaticBlock:
		clonedElement = element.Clone()
	case *FieldDefinition:
		clonedElement = element.Clone()
	case *MethodDefinition:
		clonedElement = element.Clone()
	}
	return &ClassElement{Element: clonedElement}
}
func (n *ClassElements) Clone() *ClassElements {
	ns := make(ClassElements, len(*n))
	for i := range *n {
		ns[i] = *(*n)[i].Clone()
	}
	return &ns
}
func (n *ClassLiteral) Clone() *ClassLiteral {
	var name *Identifier
	if n.Name != nil {
		name = n.Name.Clone()
	}
	var superclass *Expression
	if n.SuperClass != nil {
		superclass = n.SuperClass.Clone()
	}
	return &ClassLiteral{Class: n.Class, RightBrace: n.RightBrace, Name: name, SuperClass: superclass, Body: *n.Body.Clone()}
}
func (n *ClassStaticBlock) Clone() *ClassStaticBlock {
	return &ClassStaticBlock{Static: n.Static, Block: n.Block.Clone()}
}
func (n *ConciseBody) Clone() *ConciseBody {
	var clonedBody Body
	switch body := n.Body.(type) {
	case *BlockStatement:
		clonedBody = body.Clone()
	case *Expression:
		clonedBody = body.Clone()
	}
	return &ConciseBody{Body: clonedBody}
}
func (n *ConditionalExpression) Clone() *ConditionalExpression {
	return &ConditionalExpression{Test: n.Test.Clone(), Consequent: n.Consequent.Clone(), Alternate: n.Alternate.Clone()}
}
func (n *ContinueStatement) Clone() *ContinueStatement {
	var label *Identifier
	if n.Label != nil {
		label = n.Label.Clone()
	}
	return &ContinueStatement{Idx: n.Idx, Label: label}
}
func (n *DebuggerStatement) Clone() *DebuggerStatement {
	return &DebuggerStatement{Debugger: n.Debugger}
}
func (n *DoWhileStatement) Clone() *DoWhileStatement {
	return &DoWhileStatement{Do: n.Do, Test: n.Test.Clone(), Body: n.Body.Clone()}
}
func (n *EmptyStatement) Clone() *EmptyStatement {
	return &EmptyStatement{Semicolon: n.Semicolon}
}
func (n *Expression) Clone() *Expression {
	var clonedExpr Expr
	switch expr := n.Expr.(type) {
	case *ArrayLiteral:
		clonedExpr = expr.Clone()
	case *ArrayPattern:
		clonedExpr = expr.Clone()
	case *ArrowFunctionLiteral:
		clonedExpr = expr.Clone()
	case *AssignExpression:
		clonedExpr = expr.Clone()
	case *AwaitExpression:
		clonedExpr = expr.Clone()
	case *BinaryExpression:
		clonedExpr = expr.Clone()
	case *BooleanLiteral:
		clonedExpr = expr.Clone()
	case *CallExpression:
		clonedExpr = expr.Clone()
	case *ClassLiteral:
		clonedExpr = expr.Clone()
	case *ConditionalExpression:
		clonedExpr = expr.Clone()
	case *FunctionLiteral:
		clonedExpr = expr.Clone()
	case *Identifier:
		clonedExpr = expr.Clone()
	case *InvalidExpression:
		clonedExpr = expr.Clone()
	case *MemberExpression:
		clonedExpr = expr.Clone()
	case *MetaProperty:
		clonedExpr = expr.Clone()
	case *NewExpression:
		clonedExpr = expr.Clone()
	case *NullLiteral:
		clonedExpr = expr.Clone()
	case *NumberLiteral:
		clonedExpr = expr.Clone()
	case *ObjectLiteral:
		clonedExpr = expr.Clone()
	case *ObjectPattern:
		clonedExpr = expr.Clone()
	case *Optional:
		clonedExpr = expr.Clone()
	case *OptionalChain:
		clonedExpr = expr.Clone()
	case *PrivateDotExpression:
		clonedExpr = expr.Clone()
	case *PrivateIdentifier:
		clonedExpr = expr.Clone()
	case *PropertyKeyed:
		clonedExpr = expr.Clone()
	case *PropertyShort:
		clonedExpr = expr.Clone()
	case *RegExpLiteral:
		clonedExpr = expr.Clone()
	case *SequenceExpression:
		clonedExpr = expr.Clone()
	case *SpreadElement:
		clonedExpr = expr.Clone()
	case *StringLiteral:
		clonedExpr = expr.Clone()
	case *SuperExpression:
		clonedExpr = expr.Clone()
	case *TemplateLiteral:
		clonedExpr = expr.Clone()
	case *ThisExpression:
		clonedExpr = expr.Clone()
	case *UnaryExpression:
		clonedExpr = expr.Clone()
	case *UpdateExpression:
		clonedExpr = expr.Clone()
	case *VariableDeclarator:
		clonedExpr = expr.Clone()
	case *YieldExpression:
		clonedExpr = expr.Clone()
	}
	return &Expression{Expr: clonedExpr}
}
func (n *ExpressionStatement) Clone() *ExpressionStatement {
	return &ExpressionStatement{Expression: n.Expression.Clone(), Comment: n.Comment}
}
func (n *Expressions) Clone() *Expressions {
	ns := make(Expressions, len(*n))
	for i := range *n {
		ns[i] = *(*n)[i].Clone()
	}
	return &ns
}
func (n *FieldDefinition) Clone() *FieldDefinition {
	var initializer *Expression
	if n.Initializer != nil {
		initializer = n.Initializer.Clone()
	}
	return &FieldDefinition{Idx: n.Idx, Key: n.Key.Clone(), Initializer: initializer, Computed: n.Computed, Static: n.Static}
}
func (n *ForInStatement) Clone() *ForInStatement {
	return &ForInStatement{For: n.For, Into: n.Into.Clone(), Source: n.Source.Clone(), Body: n.Body.Clone()}
}
func (n *ForInto) Clone() *ForInto {
	var clonedInto Into
	switch into := n.Into.(type) {
	case *Expression:
		clonedInto = into.Clone()
	case *VariableDeclaration:
		clonedInto = into.Clone()
	}
	return &ForInto{Into: clonedInto}
}
func (n *ForLoopInitializer) Clone() *ForLoopInitializer {
	var clonedForLoopInit ForLoopInit
	switch forLoopInit := n.Initializer.(type) {
	case *Expression:
		clonedForLoopInit = forLoopInit.Clone()
	case *VariableDeclaration:
		clonedForLoopInit = forLoopInit.Clone()
	}
	return &ForLoopInitializer{Initializer: clonedForLoopInit}
}
func (n *ForOfStatement) Clone() *ForOfStatement {
	return &ForOfStatement{For: n.For, Into: n.Into.Clone(), Source: n.Source.Clone(), Body: n.Body.Clone()}
}
func (n *ForStatement) Clone() *ForStatement {
	var initializer *ForLoopInitializer
	if n.Initializer != nil {
		initializer = n.Initializer.Clone()
	}
	return &ForStatement{For: n.For, Initializer: initializer, Update: n.Update.Clone(), Test: n.Test.Clone(), Body: n.Body.Clone()}
}
func (n *FunctionDeclaration) Clone() *FunctionDeclaration {
	return &FunctionDeclaration{Function: n.Function.Clone()}
}
func (n *FunctionLiteral) Clone() *FunctionLiteral {
	return &FunctionLiteral{Function: n.Function, Name: n.Name.Clone(), ParameterList: *n.ParameterList.Clone(), Body: n.Body.Clone(), Async: n.Async, ScopeContext: n.ScopeContext}
}
func (n *Identifier) Clone() *Identifier {
	return &Identifier{Idx: n.Idx, Name: n.Name, ScopeContext: n.ScopeContext}
}
func (n *IfStatement) Clone() *IfStatement {
	var alternate *Statement
	if n.Alternate != nil {
		alternate = n.Alternate.Clone()
	}
	return &IfStatement{If: n.If, Test: n.Test.Clone(), Consequent: n.Consequent.Clone(), Alternate: alternate}
}
func (n *InvalidExpression) Clone() *InvalidExpression {
	return &InvalidExpression{From: n.From, To: n.To}
}
func (n *LabelledStatement) Clone() *LabelledStatement {
	return &LabelledStatement{Label: n.Label.Clone(), Colon: n.Colon, Statement: n.Statement.Clone()}
}
func (n *MemberExpression) Clone() *MemberExpression {
	return &MemberExpression{Object: n.Object.Clone(), Property: n.Property.Clone()}
}
func (n *MetaProperty) Clone() *MetaProperty {
	return &MetaProperty{Meta: n.Meta.Clone(), Idx: n.Idx}
}
func (n *MethodDefinition) Clone() *MethodDefinition {
	return &MethodDefinition{Idx: n.Idx, Key: n.Key.Clone(), Kind: n.Kind, Body: n.Body.Clone(), Computed: n.Computed, Static: n.Static}
}
func (n *NewExpression) Clone() *NewExpression {
	return &NewExpression{New: n.New, Callee: n.Callee.Clone(), LeftParenthesis: n.LeftParenthesis, ArgumentList: *n.ArgumentList.Clone(), RightParenthesis: n.RightParenthesis}
}
func (n *NullLiteral) Clone() *NullLiteral {
	return &NullLiteral{Idx: n.Idx}
}
func (n *NumberLiteral) Clone() *NumberLiteral {
	return &NumberLiteral{Idx: n.Idx, Literal: n.Literal, Value: n.Value}
}
func (n *ObjectLiteral) Clone() *ObjectLiteral {
	return &ObjectLiteral{LeftBrace: n.LeftBrace, RightBrace: n.RightBrace, Value: *n.Value.Clone()}
}
func (n *ObjectPattern) Clone() *ObjectPattern {
	var clonedExpr Expr
	switch expr := n.Rest.(type) {
	case *ArrayLiteral:
		clonedExpr = expr.Clone()
	case *ArrayPattern:
		clonedExpr = expr.Clone()
	case *ArrowFunctionLiteral:
		clonedExpr = expr.Clone()
	case *AssignExpression:
		clonedExpr = expr.Clone()
	case *AwaitExpression:
		clonedExpr = expr.Clone()
	case *BinaryExpression:
		clonedExpr = expr.Clone()
	case *BooleanLiteral:
		clonedExpr = expr.Clone()
	case *CallExpression:
		clonedExpr = expr.Clone()
	case *ClassLiteral:
		clonedExpr = expr.Clone()
	case *ConditionalExpression:
		clonedExpr = expr.Clone()
	case *FunctionLiteral:
		clonedExpr = expr.Clone()
	case *Identifier:
		clonedExpr = expr.Clone()
	case *InvalidExpression:
		clonedExpr = expr.Clone()
	case *MemberExpression:
		clonedExpr = expr.Clone()
	case *MetaProperty:
		clonedExpr = expr.Clone()
	case *NewExpression:
		clonedExpr = expr.Clone()
	case *NullLiteral:
		clonedExpr = expr.Clone()
	case *NumberLiteral:
		clonedExpr = expr.Clone()
	case *ObjectLiteral:
		clonedExpr = expr.Clone()
	case *ObjectPattern:
		clonedExpr = expr.Clone()
	case *Optional:
		clonedExpr = expr.Clone()
	case *OptionalChain:
		clonedExpr = expr.Clone()
	case *PrivateDotExpression:
		clonedExpr = expr.Clone()
	case *PrivateIdentifier:
		clonedExpr = expr.Clone()
	case *PropertyKeyed:
		clonedExpr = expr.Clone()
	case *PropertyShort:
		clonedExpr = expr.Clone()
	case *RegExpLiteral:
		clonedExpr = expr.Clone()
	case *SequenceExpression:
		clonedExpr = expr.Clone()
	case *SpreadElement:
		clonedExpr = expr.Clone()
	case *StringLiteral:
		clonedExpr = expr.Clone()
	case *SuperExpression:
		clonedExpr = expr.Clone()
	case *TemplateLiteral:
		clonedExpr = expr.Clone()
	case *ThisExpression:
		clonedExpr = expr.Clone()
	case *UnaryExpression:
		clonedExpr = expr.Clone()
	case *UpdateExpression:
		clonedExpr = expr.Clone()
	case *VariableDeclarator:
		clonedExpr = expr.Clone()
	case *YieldExpression:
		clonedExpr = expr.Clone()
	}
	return &ObjectPattern{LeftBrace: n.LeftBrace, RightBrace: n.RightBrace, Properties: *n.Properties.Clone(), Rest: clonedExpr}
}
func (n *Optional) Clone() *Optional {
	return &Optional{Expr: n.Expr.Clone()}
}
func (n *OptionalChain) Clone() *OptionalChain {
	return &OptionalChain{Base: n.Base.Clone()}
}
func (n *ParameterList) Clone() *ParameterList {
	var clonedExpr Expr
	switch expr := n.Rest.(type) {
	case *ArrayLiteral:
		clonedExpr = expr.Clone()
	case *ArrayPattern:
		clonedExpr = expr.Clone()
	case *ArrowFunctionLiteral:
		clonedExpr = expr.Clone()
	case *AssignExpression:
		clonedExpr = expr.Clone()
	case *AwaitExpression:
		clonedExpr = expr.Clone()
	case *BinaryExpression:
		clonedExpr = expr.Clone()
	case *BooleanLiteral:
		clonedExpr = expr.Clone()
	case *CallExpression:
		clonedExpr = expr.Clone()
	case *ClassLiteral:
		clonedExpr = expr.Clone()
	case *ConditionalExpression:
		clonedExpr = expr.Clone()
	case *FunctionLiteral:
		clonedExpr = expr.Clone()
	case *Identifier:
		clonedExpr = expr.Clone()
	case *InvalidExpression:
		clonedExpr = expr.Clone()
	case *MemberExpression:
		clonedExpr = expr.Clone()
	case *MetaProperty:
		clonedExpr = expr.Clone()
	case *NewExpression:
		clonedExpr = expr.Clone()
	case *NullLiteral:
		clonedExpr = expr.Clone()
	case *NumberLiteral:
		clonedExpr = expr.Clone()
	case *ObjectLiteral:
		clonedExpr = expr.Clone()
	case *ObjectPattern:
		clonedExpr = expr.Clone()
	case *Optional:
		clonedExpr = expr.Clone()
	case *OptionalChain:
		clonedExpr = expr.Clone()
	case *PrivateDotExpression:
		clonedExpr = expr.Clone()
	case *PrivateIdentifier:
		clonedExpr = expr.Clone()
	case *PropertyKeyed:
		clonedExpr = expr.Clone()
	case *PropertyShort:
		clonedExpr = expr.Clone()
	case *RegExpLiteral:
		clonedExpr = expr.Clone()
	case *SequenceExpression:
		clonedExpr = expr.Clone()
	case *SpreadElement:
		clonedExpr = expr.Clone()
	case *StringLiteral:
		clonedExpr = expr.Clone()
	case *SuperExpression:
		clonedExpr = expr.Clone()
	case *TemplateLiteral:
		clonedExpr = expr.Clone()
	case *ThisExpression:
		clonedExpr = expr.Clone()
	case *UnaryExpression:
		clonedExpr = expr.Clone()
	case *UpdateExpression:
		clonedExpr = expr.Clone()
	case *VariableDeclarator:
		clonedExpr = expr.Clone()
	case *YieldExpression:
		clonedExpr = expr.Clone()
	}
	return &ParameterList{Opening: n.Opening, List: *n.List.Clone(), Rest: clonedExpr, Closing: n.Closing}
}
func (n *PrivateDotExpression) Clone() *PrivateDotExpression {
	return &PrivateDotExpression{Left: n.Left.Clone(), Identifier: n.Identifier.Clone()}
}
func (n *PrivateIdentifier) Clone() *PrivateIdentifier {
	return &PrivateIdentifier{Identifier: n.Identifier.Clone()}
}
func (n *Program) Clone() *Program {
	return &Program{Body: *n.Body.Clone()}
}
func (n *Properties) Clone() *Properties {
	ns := make(Properties, len(*n))
	for i := range *n {
		ns[i] = *(*n)[i].Clone()
	}
	return &ns
}
func (n *Property) Clone() *Property {
	var clonedProp Prop
	switch prop := n.Prop.(type) {
	case *PropertyKeyed:
		clonedProp = prop.Clone()
	case *PropertyShort:
		clonedProp = prop.Clone()
	case *SpreadElement:
		clonedProp = prop.Clone()
	}
	return &Property{Prop: clonedProp}
}
func (n *PropertyKeyed) Clone() *PropertyKeyed {
	return &PropertyKeyed{Key: n.Key.Clone(), Kind: n.Kind, Value: n.Value.Clone(), Computed: n.Computed}
}
func (n *PropertyShort) Clone() *PropertyShort {
	return &PropertyShort{Name: n.Name.Clone(), Initializer: n.Initializer.Clone()}
}
func (n *RegExpLiteral) Clone() *RegExpLiteral {
	return &RegExpLiteral{Idx: n.Idx, Literal: n.Literal, Pattern: n.Pattern, Flags: n.Flags}
}
func (n *ReturnStatement) Clone() *ReturnStatement {
	var argument *Expression
	if n.Argument != nil {
		argument = n.Argument.Clone()
	}
	return &ReturnStatement{Return: n.Return, Argument: argument}
}
func (n *SequenceExpression) Clone() *SequenceExpression {
	return &SequenceExpression{Sequence: *n.Sequence.Clone()}
}
func (n *SpreadElement) Clone() *SpreadElement {
	return &SpreadElement{Expression: n.Expression.Clone()}
}
func (n *Statement) Clone() *Statement {
	var clonedStmt Stmt
	switch stmt := n.Stmt.(type) {
	case *BadStatement:
		clonedStmt = stmt.Clone()
	case *BlockStatement:
		clonedStmt = stmt.Clone()
	case *BreakStatement:
		clonedStmt = stmt.Clone()
	case *CaseStatement:
		clonedStmt = stmt.Clone()
	case *CatchStatement:
		clonedStmt = stmt.Clone()
	case *ClassDeclaration:
		clonedStmt = stmt.Clone()
	case *ContinueStatement:
		clonedStmt = stmt.Clone()
	case *DebuggerStatement:
		clonedStmt = stmt.Clone()
	case *DoWhileStatement:
		clonedStmt = stmt.Clone()
	case *EmptyStatement:
		clonedStmt = stmt.Clone()
	case *ExpressionStatement:
		clonedStmt = stmt.Clone()
	case *ForInStatement:
		clonedStmt = stmt.Clone()
	case *ForOfStatement:
		clonedStmt = stmt.Clone()
	case *ForStatement:
		clonedStmt = stmt.Clone()
	case *FunctionDeclaration:
		clonedStmt = stmt.Clone()
	case *IfStatement:
		clonedStmt = stmt.Clone()
	case *LabelledStatement:
		clonedStmt = stmt.Clone()
	case *ReturnStatement:
		clonedStmt = stmt.Clone()
	case *SwitchStatement:
		clonedStmt = stmt.Clone()
	case *ThrowStatement:
		clonedStmt = stmt.Clone()
	case *TryStatement:
		clonedStmt = stmt.Clone()
	case *VariableDeclaration:
		clonedStmt = stmt.Clone()
	case *WhileStatement:
		clonedStmt = stmt.Clone()
	case *WithStatement:
		clonedStmt = stmt.Clone()
	}
	return &Statement{Stmt: clonedStmt}
}
func (n *Statements) Clone() *Statements {
	ns := make(Statements, len(*n))
	for i := range *n {
		ns[i] = *(*n)[i].Clone()
	}
	return &ns
}
func (n *StringLiteral) Clone() *StringLiteral {
	return &StringLiteral{Idx: n.Idx, Literal: n.Literal, Value: n.Value}
}
func (n *SuperExpression) Clone() *SuperExpression {
	return &SuperExpression{Idx: n.Idx}
}
func (n *SwitchStatement) Clone() *SwitchStatement {
	return &SwitchStatement{Switch: n.Switch, Discriminant: n.Discriminant.Clone(), Default: n.Default, Body: *n.Body.Clone()}
}
func (n *TemplateElement) Clone() *TemplateElement {
	return &TemplateElement{Idx: n.Idx, Literal: n.Literal, Parsed: n.Parsed, Valid: n.Valid}
}
func (n *TemplateElements) Clone() *TemplateElements {
	ns := make(TemplateElements, len(*n))
	for i := range *n {
		ns[i] = *(*n)[i].Clone()
	}
	return &ns
}
func (n *TemplateLiteral) Clone() *TemplateLiteral {
	var tag *Expression
	if n.Tag != nil {
		tag = n.Tag.Clone()
	}
	return &TemplateLiteral{OpenQuote: n.OpenQuote, CloseQuote: n.CloseQuote, Tag: tag, Elements: *n.Elements.Clone(), Expressions: *n.Expressions.Clone()}
}
func (n *ThisExpression) Clone() *ThisExpression {
	return &ThisExpression{Idx: n.Idx}
}
func (n *ThrowStatement) Clone() *ThrowStatement {
	return &ThrowStatement{Throw: n.Throw, Argument: n.Argument.Clone()}
}
func (n *TryStatement) Clone() *TryStatement {
	var catch *CatchStatement
	if n.Catch != nil {
		catch = n.Catch.Clone()
	}
	var finally *BlockStatement
	if n.Finally != nil {
		finally = n.Finally.Clone()
	}
	return &TryStatement{Try: n.Try, Body: n.Body.Clone(), Catch: catch, Finally: finally}
}
func (n *UnaryExpression) Clone() *UnaryExpression {
	return &UnaryExpression{Operator: n.Operator, Idx: n.Idx, Operand: n.Operand.Clone()}
}
func (n *UpdateExpression) Clone() *UpdateExpression {
	return &UpdateExpression{Operator: n.Operator, Idx: n.Idx, Operand: n.Operand.Clone(), Postfix: n.Postfix}
}
func (n *VariableDeclaration) Clone() *VariableDeclaration {
	return &VariableDeclaration{Idx: n.Idx, Token: n.Token, List: *n.List.Clone(), Comment: n.Comment}
}
func (n *VariableDeclarator) Clone() *VariableDeclarator {
	var initializer *Expression
	if n.Initializer != nil {
		initializer = n.Initializer.Clone()
	}
	return &VariableDeclarator{Target: n.Target.Clone(), Initializer: initializer}
}
func (n *VariableDeclarators) Clone() *VariableDeclarators {
	ns := make(VariableDeclarators, len(*n))
	for i := range *n {
		ns[i] = *(*n)[i].Clone()
	}
	return &ns
}
func (n *WhileStatement) Clone() *WhileStatement {
	return &WhileStatement{While: n.While, Test: n.Test.Clone(), Body: n.Body.Clone()}
}
func (n *WithStatement) Clone() *WithStatement {
	return &WithStatement{With: n.With, Object: n.Object.Clone(), Body: n.Body.Clone()}
}
func (n *YieldExpression) Clone() *YieldExpression {
	return &YieldExpression{Yield: n.Yield, Argument: n.Argument.Clone(), Delegate: n.Delegate}
}
