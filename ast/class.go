package ast

type (
	ClassLiteral struct {
		Class      Idx
		RightBrace Idx
		Name       *Identifier `optional:"true"`
		SuperClass *Expression
		Body       ClassElements
	}

	ClassElements []ClassElement

	ClassElement struct {
		Element
	}

	Element interface {
		VisitableNode
		_classElement()
	}

	FieldDefinition struct {
		Idx         Idx
		Key         *Expression
		Initializer *Expression `optional:"true"`
		Computed    bool
		Static      bool
	}

	MethodDefinition struct {
		Idx      Idx
		Key      *Expression
		Kind     PropertyKind // "method", "get" or "set"
		Body     *FunctionLiteral
		Computed bool
		Static   bool
	}

	ClassStaticBlock struct {
		Static Idx
		Block  *BlockStatement
	}
)

func (*ClassLiteral) _expr()  {}
func (*PropertyShort) _expr() {}
func (*PropertyKeyed) _expr() {}

func (*FieldDefinition) _classElement()  {}
func (*MethodDefinition) _classElement() {}
func (*ClassStaticBlock) _classElement() {}
