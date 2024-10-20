// Code generated by gen_clone.go; DO NOT EDIT.
package ast

func (n *ArrayLiteral) Clone() *ArrayLiteral {
	return &ArrayLiteral{LeftBracket: n.LeftBracket, RightBracket: n.RightBracket, Value: n.Value.Clone()}
}
func (n *ArrayPattern) Clone() *ArrayPattern {
	return &ArrayPattern{LeftBracket: n.LeftBracket, RightBracket: n.RightBracket, Elements: n.Elements.Clone(), Rest: n.Rest.Clone()}
}
func (n *ArrowFunctionLiteral) Clone() *ArrowFunctionLiteral {
	return &ArrowFunctionLiteral{Start: n.Start, ParameterList: n.ParameterList.Clone(), Body: n.Body.Clone(), Async: n.Async}
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
	return &BindingTarget{Target: n.Target.Clone()}
}
func (n *BlockStatement) Clone() *BlockStatement {
	return &BlockStatement{LeftBrace: n.LeftBrace, List: n.List.Clone(), RightBrace: n.RightBrace}
}
func (n *BooleanLiteral) Clone() *BooleanLiteral {
	return &BooleanLiteral{Idx: n.Idx, Value: n.Value}
}
func (n *BreakStatement) Clone() *BreakStatement {
	return &BreakStatement{Idx: n.Idx, Label: n.Label.Clone()}
}
func (n *CallExpression) Clone() *CallExpression {
	return &CallExpression{Callee: n.Callee.Clone(), LeftParenthesis: n.LeftParenthesis, ArgumentList: n.ArgumentList.Clone(), RightParenthesis: n.RightParenthesis}
}
func (n *CaseStatement) Clone() *CaseStatement {
	return &CaseStatement{Case: n.Case, Test: n.Test.Clone(), Consequent: n.Consequent.Clone()}
}
func (n *CaseStatements) Clone() *CaseStatements {
	ns := make(CaseStatements, len(*n))
	for i := range *n {
		ns[i] = *(*n)[i].Clone()
	}
	return &ns
}
func (n *CatchStatement) Clone() *CatchStatement {
	return &CatchStatement{Catch: n.Catch, Parameter: n.Parameter.Clone(), Body: n.Body.Clone()}
}
func (n *ClassDeclaration) Clone() *ClassDeclaration {
	return &ClassDeclaration{Class: n.Class.Clone()}
}
func (n *ClassElement) Clone() *ClassElement {
	return &ClassElement{Element: n.Element.Clone()}
}
func (n *ClassElements) Clone() *ClassElements {
	ns := make(ClassElements, len(*n))
	for i := range *n {
		ns[i] = *(*n)[i].Clone()
	}
	return &ns
}
func (n *ClassLiteral) Clone() *ClassLiteral {
	return &ClassLiteral{Class: n.Class, RightBrace: n.RightBrace, Name: n.Name.Clone(), SuperClass: n.SuperClass.Clone(), Body: n.Body.Clone()}
}
func (n *ClassStaticBlock) Clone() *ClassStaticBlock {
	return &ClassStaticBlock{Static: n.Static, Block: n.Block.Clone()}
}
func (n *ConciseBody) Clone() *ConciseBody {
	return &ConciseBody{Body: n.Body.Clone()}
}
func (n *ConditionalExpression) Clone() *ConditionalExpression {
	return &ConditionalExpression{Test: n.Test.Clone(), Consequent: n.Consequent.Clone(), Alternate: n.Alternate.Clone()}
}
func (n *ContinueStatement) Clone() *ContinueStatement {
	return &ContinueStatement{Idx: n.Idx, Label: n.Label.Clone()}
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
	return &Expression{Expr: n.Expr.Clone()}
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
	return &FieldDefinition{Idx: n.Idx, Key: n.Key.Clone(), Initializer: n.Initializer.Clone(), Computed: n.Computed, Static: n.Static}
}
func (n *ForInStatement) Clone() *ForInStatement {
	return &ForInStatement{For: n.For, Into: n.Into.Clone(), Source: n.Source.Clone(), Body: n.Body.Clone()}
}
func (n *ForInto) Clone() *ForInto {
	return &ForInto{Into: n.Into.Clone()}
}
func (n *ForLoopInitializer) Clone() *ForLoopInitializer {
	return &ForLoopInitializer{Initializer: n.Initializer.Clone()}
}
func (n *ForOfStatement) Clone() *ForOfStatement {
	return &ForOfStatement{For: n.For, Into: n.Into.Clone(), Source: n.Source.Clone(), Body: n.Body.Clone()}
}
func (n *ForStatement) Clone() *ForStatement {
	return &ForStatement{For: n.For, Initializer: n.Initializer.Clone(), Update: n.Update.Clone(), Test: n.Test.Clone(), Body: n.Body.Clone()}
}
func (n *FunctionDeclaration) Clone() *FunctionDeclaration {
	return &FunctionDeclaration{Function: n.Function.Clone()}
}
func (n *FunctionLiteral) Clone() *FunctionLiteral {
	return &FunctionLiteral{Function: n.Function, Name: n.Name.Clone(), ParameterList: n.ParameterList.Clone(), Body: n.Body.Clone(), Async: n.Async}
}
func (n *Identifier) Clone() *Identifier {
	return &Identifier{Idx: n.Idx, Name: n.Name, ScopeContext: n.ScopeContext}
}
func (n *IfStatement) Clone() *IfStatement {
	return &IfStatement{If: n.If, Test: n.Test.Clone(), Consequent: n.Consequent.Clone(), Alternate: n.Alternate.Clone()}
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
	return &NewExpression{New: n.New, Callee: n.Callee.Clone(), LeftParenthesis: n.LeftParenthesis, ArgumentList: n.ArgumentList.Clone(), RightParenthesis: n.RightParenthesis}
}
func (n *NullLiteral) Clone() *NullLiteral {
	return &NullLiteral{Idx: n.Idx}
}
func (n *NumberLiteral) Clone() *NumberLiteral {
	return &NumberLiteral{Idx: n.Idx, Literal: n.Literal, Value: n.Value}
}
func (n *ObjectLiteral) Clone() *ObjectLiteral {
	return &ObjectLiteral{LeftBrace: n.LeftBrace, RightBrace: n.RightBrace, Value: n.Value.Clone()}
}
func (n *ObjectPattern) Clone() *ObjectPattern {
	return &ObjectPattern{LeftBrace: n.LeftBrace, RightBrace: n.RightBrace, Properties: n.Properties.Clone(), Rest: n.Rest.Clone()}
}
func (n *Optional) Clone() *Optional {
	return &Optional{Expr: n.Expr.Clone()}
}
func (n *OptionalChain) Clone() *OptionalChain {
	return &OptionalChain{Base: n.Base.Clone()}
}
func (n *ParameterList) Clone() *ParameterList {
	return &ParameterList{Opening: n.Opening, List: n.List.Clone(), Rest: n.Rest.Clone(), Closing: n.Closing}
}
func (n *PrivateDotExpression) Clone() *PrivateDotExpression {
	return &PrivateDotExpression{Left: n.Left.Clone(), Identifier: n.Identifier.Clone()}
}
func (n *PrivateIdentifier) Clone() *PrivateIdentifier {
	return &PrivateIdentifier{Identifier: n.Identifier.Clone()}
}
func (n *Program) Clone() *Program {
	return &Program{Body: n.Body.Clone()}
}
func (n *Properties) Clone() *Properties {
	ns := make(Properties, len(*n))
	for i := range *n {
		ns[i] = *(*n)[i].Clone()
	}
	return &ns
}
func (n *Property) Clone() *Property {
	return &Property{Prop: n.Prop.Clone()}
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
	return &ReturnStatement{Return: n.Return, Argument: n.Argument.Clone()}
}
func (n *SequenceExpression) Clone() *SequenceExpression {
	return &SequenceExpression{Sequence: n.Sequence.Clone()}
}
func (n *SpreadElement) Clone() *SpreadElement {
	return &SpreadElement{Expression: n.Expression.Clone()}
}
func (n *Statement) Clone() *Statement {
	return &Statement{Stmt: n.Stmt.Clone()}
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
	return &SwitchStatement{Switch: n.Switch, Discriminant: n.Discriminant.Clone(), Default: n.Default, Body: n.Body.Clone()}
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
	return &TemplateLiteral{OpenQuote: n.OpenQuote, CloseQuote: n.CloseQuote, Tag: n.Tag.Clone(), Elements: n.Elements.Clone(), Expressions: n.Expressions.Clone()}
}
func (n *ThisExpression) Clone() *ThisExpression {
	return &ThisExpression{Idx: n.Idx}
}
func (n *ThrowStatement) Clone() *ThrowStatement {
	return &ThrowStatement{Throw: n.Throw, Argument: n.Argument.Clone()}
}
func (n *TryStatement) Clone() *TryStatement {
	return &TryStatement{Try: n.Try, Body: n.Body.Clone(), Catch: n.Catch.Clone(), Finally: n.Finally.Clone()}
}
func (n *UnaryExpression) Clone() *UnaryExpression {
	return &UnaryExpression{Operator: n.Operator, Idx: n.Idx, Operand: n.Operand.Clone()}
}
func (n *UpdateExpression) Clone() *UpdateExpression {
	return &UpdateExpression{Operator: n.Operator, Idx: n.Idx, Operand: n.Operand.Clone(), Postfix: n.Postfix}
}
func (n *VariableDeclaration) Clone() *VariableDeclaration {
	return &VariableDeclaration{Idx: n.Idx, Token: n.Token, List: n.List.Clone(), Comment: n.Comment}
}
func (n *VariableDeclarator) Clone() *VariableDeclarator {
	return &VariableDeclarator{Target: n.Target.Clone(), Initializer: n.Initializer.Clone()}
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