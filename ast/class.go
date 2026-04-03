package ast

type (
	ClassLiteral struct {
		Name       *Identifier `optional:"true"`
		SuperClass *Expression `optional:"true"`
		Body       ClassElements

		Class      Idx
		RightBrace Idx
	}

	ClassElements []ClassElement

	ClassElement struct {
		Element Element
	}

	Element interface {
		VisitableNode
		_classElement()
	}

	FieldDefinition struct {
		Key         *Expression
		Initializer *Expression `optional:"true"`

		Idx Idx

		Computed bool
		Static   bool
	}

	MethodDefinition struct {
		Key  *Expression
		Kind PropertyKind // "method", "get" or "set"
		Body *FunctionLiteral

		Idx      Idx
		Computed bool
		Static   bool
	}

	ClassStaticBlock struct {
		Block *BlockStatement

		Static Idx
	}
)

func (*ClassLiteral) _expr()  {}
func (*PropertyShort) _expr() {}
func (*PropertyKeyed) _expr() {}

func (*FieldDefinition) _classElement()  {}
func (*MethodDefinition) _classElement() {}
func (*ClassStaticBlock) _classElement() {}
