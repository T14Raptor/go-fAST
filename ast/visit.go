package ast

type VisitableNode interface {
	Node
	VisitWith(v Visitor)
	VisitChildrenWith(v Visitor)
}

type Visitor interface {
	VisitProgram(node *Program)
	VisitVariableDeclarators(node *VariableDeclarators)
	VisitVariableDeclarator(node *VariableDeclarator)
	VisitYieldExpression(node *YieldExpression)
	VisitAwaitExpression(node *AwaitExpression)
	VisitArrayLiteral(node *ArrayLiteral)
	VisitStringLiteral(node *StringLiteral)
	VisitArrayPattern(node *ArrayPattern)
	VisitAssignExpression(node *AssignExpression)
	VisitBinaryExpression(node *BinaryExpression)
	VisitBooleanLiteral(node *BooleanLiteral)
	VisitMemberExpression(node *MemberExpression)
	VisitCallExpression(node *CallExpression)
	VisitConditionalExpression(node *ConditionalExpression)
	VisitPrivateDotExpression(node *PrivateDotExpression)
	VisitOptionalChain(node *OptionalChain)
	VisitOptional(node *Optional)
	VisitFunctionLiteral(node *FunctionLiteral)
	VisitClassLiteral(node *ClassLiteral)
	VisitExpressionBody(node *ExpressionBody)
	VisitArrowFunctionLiteral(node *ArrowFunctionLiteral)
	VisitIdentifier(node *Identifier)
	VisitPrivateIdentifier(node *PrivateIdentifier)
	VisitNewExpression(node *NewExpression)
	VisitNullLiteral(node *NullLiteral)
	VisitNumberLiteral(node *NumberLiteral)
	VisitObjectLiteral(node *ObjectLiteral)
	VisitObjectPattern(node *ObjectPattern)
	VisitParameterList(node *ParameterList)
	VisitPropertyShort(node *PropertyShort)
	VisitPropertyKeyed(node *PropertyKeyed)
	VisitSpreadElement(node *SpreadElement)
	VisitExpressions(node *Expressions)
	VisitStatements(node *Statements)
	VisitExpression(node *Expression)
	VisitStatement(node *Statement)
	VisitBlockStatement(node *BlockStatement)
	VisitCaseStatement(node *CaseStatement)
	VisitBranchStatement(node *BranchStatement)
	VisitCatchStatement(node *CatchStatement)
	VisitSwitchStatement(node *SwitchStatement)
	VisitWithStatement(node *WithStatement)
	VisitIfStatement(node *IfStatement)
	VisitThrowStatement(node *ThrowStatement)
	VisitWhileStatement(node *WhileStatement)
	VisitTryStatement(node *TryStatement)
	VisitForStatement(node *ForStatement)
	VisitVariableStatement(node *VariableStatement)
	VisitReturnStatement(node *ReturnStatement)
	VisitThisExpression(node *ThisExpression)
	VisitSequenceExpression(node *SequenceExpression)
	VisitExpressionStatement(node *ExpressionStatement)
	VisitUnaryExpression(node *UnaryExpression)
	VisitForLoopInitializerExpression(node *ForLoopInitializerExpression)
	VisitEmptyStatement(node *EmptyStatement)
	VisitRegExpLiteral(node *RegExpLiteral)
	VisitFunctionDeclaration(node *FunctionDeclaration)
	VisitSuperExpression(node *SuperExpression)
	VisitInvalidExpression(node *InvalidExpression)
	VisitTemplateLiteral(node *TemplateLiteral)
	VisitMetaProperty(node *MetaProperty)
	VisitLexicalDeclaration(node *LexicalDeclaration)
	VisitBadStatement(node *BadStatement)
	VisitDebuggerStatement(node *DebuggerStatement)
	VisitDoWhileStatement(node *DoWhileStatement)
	VisitLabelledStatement(node *LabelledStatement)
	VisitClassDeclaration(node *ClassDeclaration)
	VisitForInStatement(node *ForInStatement)
	VisitForOfStatement(node *ForOfStatement)
	VisitForIntoVar(node *ForIntoVar)
	VisitForDeclaration(node *ForDeclaration)
	VisitForIntoExpression(node *ForIntoExpression)
	VisitForLoopInitializerLexicalDecl(node *ForLoopInitializerLexicalDecl)
	VisitForLoopInitializerVarDeclList(node *ForLoopInitializerVarDeclList)
	VisitConciseBody(node *ConciseBody)
}

type NoopVisitor struct {
	V Visitor
}

func (nv *NoopVisitor) VisitProgram(node *Program) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitVariableDeclarators(node *VariableDeclarators) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitVariableDeclarator(node *VariableDeclarator) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitYieldExpression(node *YieldExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitAwaitExpression(node *AwaitExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitArrayLiteral(node *ArrayLiteral) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitStringLiteral(node *StringLiteral) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitArrayPattern(node *ArrayPattern) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitAssignExpression(node *AssignExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitBinaryExpression(node *BinaryExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitBooleanLiteral(node *BooleanLiteral) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitMemberExpression(node *MemberExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitCallExpression(node *CallExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitConditionalExpression(node *ConditionalExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitPrivateDotExpression(node *PrivateDotExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitOptionalChain(node *OptionalChain) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitOptional(node *Optional) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitFunctionLiteral(node *FunctionLiteral) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitClassLiteral(node *ClassLiteral) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitExpressionBody(node *ExpressionBody) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitBranchStatement(node *BranchStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitArrowFunctionLiteral(node *ArrowFunctionLiteral) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitIdentifier(node *Identifier) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitPrivateIdentifier(node *PrivateIdentifier) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitNewExpression(node *NewExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitNullLiteral(node *NullLiteral) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitNumberLiteral(node *NumberLiteral) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitObjectLiteral(node *ObjectLiteral) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitObjectPattern(node *ObjectPattern) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitParameterList(node *ParameterList) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitPropertyShort(node *PropertyShort) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitPropertyKeyed(node *PropertyKeyed) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitSpreadElement(node *SpreadElement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitExpressions(node *Expressions) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitStatements(node *Statements) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitExpression(node *Expression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitStatement(node *Statement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitBlockStatement(node *BlockStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitCaseStatement(node *CaseStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitCatchStatement(node *CatchStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitSwitchStatement(node *SwitchStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitWithStatement(node *WithStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitIfStatement(node *IfStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitThrowStatement(node *ThrowStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitWhileStatement(node *WhileStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitTryStatement(node *TryStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitForStatement(node *ForStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitVariableStatement(node *VariableStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitReturnStatement(node *ReturnStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitThisExpression(node *ThisExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitSequenceExpression(node *SequenceExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitExpressionStatement(node *ExpressionStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitUnaryExpression(node *UnaryExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitForLoopInitializerExpression(node *ForLoopInitializerExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitEmptyStatement(node *EmptyStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitRegExpLiteral(node *RegExpLiteral) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitFunctionDeclaration(node *FunctionDeclaration) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitSuperExpression(node *SuperExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitInvalidExpression(node *InvalidExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitTemplateLiteral(node *TemplateLiteral) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitMetaProperty(node *MetaProperty) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitLexicalDeclaration(node *LexicalDeclaration) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitBadStatement(node *BadStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitDebuggerStatement(node *DebuggerStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitDoWhileStatement(node *DoWhileStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitLabelledStatement(node *LabelledStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitClassDeclaration(node *ClassDeclaration) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitForInStatement(node *ForInStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitForOfStatement(node *ForOfStatement) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitForIntoVar(node *ForIntoVar) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitForDeclaration(node *ForDeclaration) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitForIntoExpression(node *ForIntoExpression) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitForLoopInitializerVarDeclList(node *ForLoopInitializerVarDeclList) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitForLoopInitializerLexicalDecl(node *ForLoopInitializerLexicalDecl) {
	node.VisitChildrenWith(nv.V)
}

func (nv *NoopVisitor) VisitConciseBody(node *ConciseBody) {
	node.VisitChildrenWith(nv.V)
}

func (n *ConciseBody) VisitWith(v Visitor) {
	v.VisitConciseBody(n)
}

func (n *ConciseBody) VisitChildrenWith(v Visitor) {
	n.Body.VisitWith(v)
}

func (n *ForIntoVar) VisitWith(v Visitor) {
	v.VisitForIntoVar(n)
}

func (n *ForIntoVar) VisitChildrenWith(v Visitor) {
	n.Binding.VisitWith(v)
}

func (n *ForDeclaration) VisitWith(v Visitor) {
	v.VisitForDeclaration(n)
}

func (n *ForDeclaration) VisitChildrenWith(v Visitor) {
	n.Target.VisitWith(v)
}

func (n *ForIntoExpression) VisitWith(v Visitor) {
	v.VisitForIntoExpression(n)
}

func (n *ForIntoExpression) VisitChildrenWith(v Visitor) {
	n.Expression.VisitWith(v)
}

func (n *ClassDeclaration) VisitWith(v Visitor) {
	v.VisitClassDeclaration(n)
}

func (n *ClassDeclaration) VisitChildrenWith(v Visitor) {
	n.Class.VisitWith(v)
}

func (n *BadStatement) VisitWith(v Visitor) {
	v.VisitBadStatement(n)
}

func (n *BadStatement) VisitChildrenWith(v Visitor) {}

func (n *MetaProperty) VisitWith(v Visitor) {
	v.VisitMetaProperty(n)
}

func (n *MetaProperty) VisitChildrenWith(v Visitor) {
	n.Meta.VisitWith(v)
	n.Property.VisitWith(v)
}

func (n *TemplateLiteral) VisitWith(v Visitor) {
	v.VisitTemplateLiteral(n)
}

func (n *TemplateLiteral) VisitChildrenWith(v Visitor) {
	n.Tag.VisitWith(v)
	n.Expressions.VisitWith(v)
}

func (n *SuperExpression) VisitWith(v Visitor) {
	v.VisitSuperExpression(n)
}

func (n *InvalidExpression) VisitWith(v Visitor) {
	v.VisitInvalidExpression(n)
}

func (n *InvalidExpression) VisitChildrenWith(v Visitor) {}

func (n *SuperExpression) VisitChildrenWith(v Visitor) {}

func (n *ParameterList) VisitWith(v Visitor) {
	v.VisitParameterList(n)
}

func (n *ParameterList) VisitChildrenWith(v Visitor) {
	n.List.VisitWith(v)
}

func (n *ObjectPattern) VisitWith(v Visitor) {
	v.VisitObjectPattern(n)
}

func (n *ObjectPattern) VisitChildrenWith(v Visitor) {
	for _, p := range n.Properties {
		p.(VisitableNode).VisitWith(v)
	}
}

func (n *SpreadElement) VisitWith(v Visitor) {
	v.VisitSpreadElement(n)
}

func (n *SpreadElement) VisitChildrenWith(v Visitor) {
	n.Expression.VisitWith(v)
}

func (n *PropertyShort) VisitWith(v Visitor) {
	v.VisitPropertyShort(n)
}

func (n *PropertyShort) VisitChildrenWith(v Visitor) {
	n.Name.VisitWith(v)
	n.Initializer.VisitWith(v)
}

func (b *VariableDeclarator) VisitWith(v Visitor) {
	v.VisitVariableDeclarator(b)
}

func (y *YieldExpression) VisitWith(v Visitor) {
	v.VisitYieldExpression(y)
}

func (y *YieldExpression) VisitChildrenWith(v Visitor) {
	y.Argument.VisitWith(v)
}

func (a *AwaitExpression) VisitWith(v Visitor) {
	v.VisitAwaitExpression(a)
}

func (a *AwaitExpression) VisitChildrenWith(v Visitor) {
	a.Argument.VisitWith(v)
}

func (a *ArrayLiteral) VisitWith(v Visitor) {
	v.VisitArrayLiteral(a)
}

func (a *ArrayPattern) VisitWith(v Visitor) {
	v.VisitArrayPattern(a)
}

func (a *ArrayPattern) VisitChildrenWith(v Visitor) {
	a.Rest.VisitWith(v)
	a.Elements.VisitWith(v)
}

func (a *AssignExpression) VisitWith(v Visitor) {
	v.VisitAssignExpression(a)
}

func (b *BinaryExpression) VisitWith(v Visitor) {
	v.VisitBinaryExpression(b)
}

func (b *BooleanLiteral) VisitWith(v Visitor) {
	v.VisitBooleanLiteral(b)
}

func (m *MemberExpression) VisitWith(v Visitor) {
	v.VisitMemberExpression(m)
}

func (c *CallExpression) VisitWith(v Visitor) {
	v.VisitCallExpression(c)
}

func (n *CallExpression) VisitChildrenWith(v Visitor) {
	n.Callee.VisitWith(v)
	n.ArgumentList.VisitWith(v)
}

func (c *ConditionalExpression) VisitWith(v Visitor) {
	v.VisitConditionalExpression(c)
}

func (p *PrivateDotExpression) VisitWith(v Visitor) {
	v.VisitPrivateDotExpression(p)
}

func (p *PrivateDotExpression) VisitChildrenWith(v Visitor) {
	p.Left.VisitWith(v)
	p.Identifier.VisitWith(v)
}

func (o *OptionalChain) VisitWith(v Visitor) {
	v.VisitOptionalChain(o)
}

func (o *OptionalChain) VisitChildrenWith(v Visitor) {
	o.Base.VisitWith(v)
}

func (o *Optional) VisitWith(v Visitor) {
	v.VisitOptional(o)
}

func (o *Optional) VisitChildrenWith(v Visitor) {
	o.VisitWith(v)
}

func (f *FunctionLiteral) VisitWith(v Visitor) {
	v.VisitFunctionLiteral(f)
}

func (c *ClassLiteral) VisitWith(v Visitor) {
	v.VisitClassLiteral(c)
	for _, element := range c.Body {
		if i, ok := (element).(MethodDefinition); ok {
			i.Key.VisitWith(v)
			i.Body.VisitWith(v)
		}
	}
}

func (c *ClassLiteral) VisitChildrenWith(v Visitor) {
	c.Name.VisitWith(v)
	c.SuperClass.VisitWith(v)
}

func (e *ExpressionBody) VisitWith(v Visitor) {
	v.VisitExpressionBody(e)
}

func (e *ExpressionBody) VisitChildrenWith(v Visitor) {
	e.Expression.VisitWith(v)
}

func (a *ArrowFunctionLiteral) VisitWith(v Visitor) {
	v.VisitArrowFunctionLiteral(a)
}

func (i *Identifier) VisitWith(v Visitor) {
	v.VisitIdentifier(i)
}

func (i *Identifier) VisitChildrenWith(v Visitor) {}

func (p *PrivateIdentifier) VisitWith(v Visitor) {
	v.VisitPrivateIdentifier(p)
}

func (n *NewExpression) VisitWith(v Visitor) {
	v.VisitNewExpression(n)
}

func (n *NullLiteral) VisitWith(v Visitor) {
	v.VisitNullLiteral(n)
}

func (n *NumberLiteral) VisitWith(v Visitor) {
	v.VisitNumberLiteral(n)
}

func (n *NumberLiteral) VisitChildrenWith(v Visitor) {}

func (n *Expression) VisitWith(v Visitor) {
	v.VisitExpression(n)
}

func (n *Expression) VisitChildrenWith(v Visitor) {
	if n == nil || n.Expr == nil {
		return
	}
	n.Expr.VisitWith(v)
}

func (n *Statement) VisitWith(v Visitor) {
	v.VisitStatement(n)
}

func (n *Statement) VisitChildrenWith(v Visitor) {
	if n != nil {
		n.Stmt.VisitWith(v)
	}
}

func (n *Expressions) VisitWith(v Visitor) {
	v.VisitExpressions(n)
}

func (n *Expressions) VisitChildrenWith(v Visitor) {
	for i := range *n {
		v.VisitExpression(&(*n)[i])
	}
}

func (n *Statements) VisitWith(v Visitor) {
	v.VisitStatements(n)
}

func (n *Statements) VisitChildrenWith(v Visitor) {
	for i := range *n {
		v.VisitStatement(&(*n)[i])
	}
}

func (n *VariableDeclarators) VisitWith(v Visitor) {
	v.VisitVariableDeclarators(n)
}

func (n *VariableDeclarators) VisitChildrenWith(v Visitor) {
	for i := range *n {
		v.VisitVariableDeclarator((*n)[i])
	}
}

func (n *MemberExpression) VisitChildrenWith(v Visitor) {
	n.Object.VisitWith(v)
	n.Property.VisitWith(v)
}

func (n *AssignExpression) VisitChildrenWith(v Visitor) {
	n.Left.VisitWith(v)
	n.Right.VisitWith(v)
}

func (n *VariableDeclarator) VisitChildrenWith(v Visitor) {
	if n.Initializer != nil {
		n.Initializer.VisitWith(v)
	}
	n.Target.VisitWith(v)
}

func (n *BinaryExpression) VisitChildrenWith(v Visitor) {
	n.Left.VisitWith(v)
	n.Right.VisitWith(v)
}

func (n *ThisExpression) VisitWith(v Visitor) {
	v.VisitThisExpression(n)
}

func (n *ThisExpression) VisitChildrenWith(v Visitor) {}

func (n *ArrayLiteral) VisitChildrenWith(v Visitor) {
	for i := range n.Value {
		n.Value[i].VisitWith(v)
	}
}

func (n *BlockStatement) VisitWith(v Visitor) {
	v.VisitBlockStatement(n)
}

func (n *BlockStatement) VisitChildrenWith(v Visitor) {
	n.List.VisitWith(v)
}

func (n *FunctionLiteral) VisitChildrenWith(v Visitor) {
	n.Name.VisitWith(v)
	n.ParameterList.List.VisitWith(v)
	n.Body.VisitWith(v)
}

func (n *StringLiteral) VisitWith(v Visitor) {
	v.VisitStringLiteral(n)
}

func (n *StringLiteral) VisitChildrenWith(v Visitor) {}

func (n *ReturnStatement) VisitWith(v Visitor) {
	v.VisitReturnStatement(n)
}

func (n *ReturnStatement) VisitChildrenWith(v Visitor) {
	n.Argument.VisitWith(v)
}

func (n *SequenceExpression) VisitWith(v Visitor) {
	v.VisitSequenceExpression(n)
}

func (n *SequenceExpression) VisitChildrenWith(v Visitor) {
	n.Sequence.VisitWith(v)
}

func (n *PropertyKeyed) VisitWith(v Visitor) {
	v.VisitPropertyKeyed(n)
}

func (n *PropertyKeyed) VisitChildrenWith(v Visitor) {
	n.Key.VisitWith(v)
	n.Value.VisitWith(v)
}

func (n *UnaryExpression) VisitWith(v Visitor) {
	v.VisitUnaryExpression(n)
}

func (n *UnaryExpression) VisitChildrenWith(v Visitor) {
	n.Operand.VisitWith(v)
}

func (n *ArrowFunctionLiteral) VisitChildrenWith(v Visitor) {}

func (n *BooleanLiteral) VisitChildrenWith(v Visitor) {}

func (n *BranchStatement) VisitWith(v Visitor) {
	v.VisitBranchStatement(n)
}

func (n *BranchStatement) VisitChildrenWith(v Visitor) {
	if n.Label != nil {
		n.Label.VisitWith(v)
	}
}

func (n *CaseStatement) VisitWith(v Visitor) {
	v.VisitCaseStatement(n)
}

func (n *CaseStatement) VisitChildrenWith(v Visitor) {
	n.Test.VisitWith(v)
	n.Consequent.VisitWith(v)
}

func (n *CatchStatement) VisitWith(v Visitor) {
	v.VisitCatchStatement(n)
}

func (n *CatchStatement) VisitChildrenWith(v Visitor) {
	(*n.Parameter).VisitWith(v)
	n.Body.VisitWith(v)
}

func (n *ConditionalExpression) VisitChildrenWith(v Visitor) {
	n.Test.VisitWith(v)
	n.Consequent.VisitWith(v)
	n.Alternate.VisitWith(v)
}

func (n *DebuggerStatement) VisitWith(v Visitor) {
	v.VisitDebuggerStatement(n)
}

func (n *DebuggerStatement) VisitChildrenWith(v Visitor) {}

func (n *DoWhileStatement) VisitWith(v Visitor) {
	v.VisitDoWhileStatement(n)
}

func (n *DoWhileStatement) VisitChildrenWith(v Visitor) {
	n.Test.VisitWith(v)
	n.Body.VisitWith(v)
}

func (n *EmptyStatement) VisitWith(v Visitor) {
	v.VisitEmptyStatement(n)
}

func (n *EmptyStatement) VisitChildrenWith(v Visitor) {}

func (n *ExpressionStatement) VisitWith(v Visitor) {
	v.VisitExpressionStatement(n)
}

func (n *ExpressionStatement) VisitChildrenWith(v Visitor) {
	n.Expression.VisitWith(v)
}

func (n *ForInStatement) VisitWith(v Visitor) {
	v.VisitForInStatement(n)
}

func (n *ForInStatement) VisitChildrenWith(v Visitor) {
	(*n.Into).(VisitableNode).VisitWith(v)
	n.Source.VisitWith(v)
	n.Body.VisitWith(v)
}

func (n *ForOfStatement) VisitWith(v Visitor) {
	v.VisitForOfStatement(n)
}

func (n *ForOfStatement) VisitChildrenWith(v Visitor) {
	(*n.Into).(VisitableNode).VisitWith(v)
	n.Source.VisitWith(v)
	n.Body.VisitWith(v)
}

func (n *ForLoopInitializerExpression) VisitWith(v Visitor) {
	v.VisitForLoopInitializerExpression(n)
}

func (n *ForLoopInitializerExpression) VisitChildrenWith(v Visitor) {
	n.Expression.VisitWith(v)
}

func (n *ForStatement) VisitWith(v Visitor) {
	v.VisitForStatement(n)
}

func (n *ForStatement) VisitChildrenWith(v Visitor) {
	if *n.Initializer != nil {
		(*n.Initializer).(VisitableNode).VisitWith(v)
	}
	n.Update.VisitWith(v)
	n.Test.VisitWith(v)
	n.Body.VisitWith(v)
}

func (n *FunctionDeclaration) VisitWith(v Visitor) {
	v.VisitFunctionDeclaration(n)
}

func (n *FunctionDeclaration) VisitChildrenWith(v Visitor) {
	n.Function.VisitWith(v)
}

func (n *IfStatement) VisitWith(v Visitor) {
	v.VisitIfStatement(n)
}

func (n *IfStatement) VisitChildrenWith(v Visitor) {
	n.Test.VisitWith(v)
	n.Consequent.VisitWith(v)
	n.Alternate.VisitWith(v)
}

func (n *LabelledStatement) VisitWith(v Visitor) {
	v.VisitLabelledStatement(n)
}

func (n *LabelledStatement) VisitChildrenWith(v Visitor) {
	n.Label.VisitWith(v)
	n.Statement.VisitWith(v)
}

func (n *NewExpression) VisitChildrenWith(v Visitor) {
	n.Callee.VisitWith(v)
	n.ArgumentList.VisitWith(v)
}

func (n *NullLiteral) VisitChildrenWith(v Visitor) {}

func (n *ObjectLiteral) VisitWith(v Visitor) {
	v.VisitObjectLiteral(n)
}

func (n *ObjectLiteral) VisitChildrenWith(v Visitor) {
	for _, p := range n.Value {
		p.VisitWith(v)
	}
}

func (n *Program) VisitWith(v Visitor) {
	v.VisitProgram(n)
}

func (n *Program) VisitChildrenWith(v Visitor) {
	n.Body.VisitWith(v)
}

func (n *RegExpLiteral) VisitWith(v Visitor) {
	v.VisitRegExpLiteral(n)
}

func (n *RegExpLiteral) VisitChildrenWith(v Visitor) {}

func (n *SwitchStatement) VisitWith(v Visitor) {
	v.VisitSwitchStatement(n)
}

func (n *SwitchStatement) VisitChildrenWith(v Visitor) {
	n.Discriminant.VisitWith(v)
	for _, c := range n.Body {
		c.VisitWith(v)
	}
}

func (n *ThrowStatement) VisitWith(v Visitor) {
	v.VisitThrowStatement(n)
}

func (n *ThrowStatement) VisitChildrenWith(v Visitor) {
	n.Argument.VisitWith(v)
}

func (n *TryStatement) VisitWith(v Visitor) {
	v.VisitTryStatement(n)
}

func (n *TryStatement) VisitChildrenWith(v Visitor) {
	n.Body.VisitWith(v)
	if n.Catch != nil {
		n.Catch.VisitWith(v)
	}
	if n.Finally != nil {
		n.Finally.VisitWith(v)
	}
}

func (n *VariableStatement) VisitWith(v Visitor) {
	v.VisitVariableStatement(n)
}

func (n *VariableStatement) VisitChildrenWith(v Visitor) {
	n.List.VisitWith(v)
}

func (n *WhileStatement) VisitWith(v Visitor) {
	v.VisitWhileStatement(n)
}

func (n *WhileStatement) VisitChildrenWith(v Visitor) {
	n.Test.VisitWith(v)
	n.Body.VisitWith(v)
}

func (n *WithStatement) VisitWith(v Visitor) {
	v.VisitWithStatement(n)
}

func (n *WithStatement) VisitChildrenWith(v Visitor) {
	n.Object.VisitWith(v)
	n.Body.VisitWith(v)
}

func (n *LexicalDeclaration) VisitWith(v Visitor) {
	v.VisitLexicalDeclaration(n)
}

func (n *LexicalDeclaration) VisitChildrenWith(v Visitor) {
	n.List.VisitWith(v)
}

func (n *ForLoopInitializerVarDeclList) VisitWith(v Visitor) {
	v.VisitForLoopInitializerVarDeclList(n)
}

func (n *ForLoopInitializerVarDeclList) VisitChildrenWith(v Visitor) {
	n.List.VisitWith(v)
}

func (n *ForLoopInitializerLexicalDecl) VisitWith(v Visitor) {
	v.VisitForLoopInitializerLexicalDecl(n)
}

func (n *ForLoopInitializerLexicalDecl) VisitChildrenWith(v Visitor) {
	n.LexicalDeclaration.VisitWith(v)
}
