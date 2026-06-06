package ast

import "unsafe"

type (
	Patterns []Pattern

	// Pattern is the recursive binding/assignment target union. It is used for
	// variable-declarator targets, function parameters, catch parameters, and the
	// left-hand side of assignments. Its children are themselves Patterns, so a
	// whole destructuring tree is typed without falling back to Expression.
	//
	// The Member/PrivDot variants only occur in assignment-target position
	// (e.g. `obj.x = 1`); declaration patterns never produce them. The validity
	// of a given variant in a given position is enforced by the parser.
	//
	//union:Identifier,ArrayPattern,ObjectPattern,AssignmentPattern,MemberExpression,PrivateDotExpression,InvalidExpression
	Pattern struct {
		kind PatternKind

		ptr unsafe.Pointer
	}

	PatternProperties []PatternProperty

	// PatternProperty is a single property inside an ObjectPattern. Unlike the
	// object-literal Property union, its value is a Pattern.
	//
	//union:PatternKeyValue,PatternShorthand
	PatternProperty struct {
		kind PatPropKind

		ptr unsafe.Pointer
	}

	// PatternKeyValue is `{ key: value }` / `{ [key]: value }` inside a pattern.
	PatternKeyValue struct {
		Key   PropertyName
		Value *Pattern
	}

	// PatternShorthand is `{ a }` or `{ a = default }` inside a pattern.
	PatternShorthand struct {
		Name        *Identifier
		Initializer *Expression `optional:"true"`
	}

	// AssignmentPattern is a defaulted binding element, e.g. `[a = 1]` or
	// `{ a: b = 1 }`. It replaces the old practice of storing an AssignExpression
	// inside patterns.
	AssignmentPattern struct {
		Left  *Pattern
		Right *Expression
	}

	ObjectPattern struct {
		Properties PatternProperties
		Rest       *Pattern `optional:"true"`

		LeftBrace  Idx
		RightBrace Idx
	}

	ArrayPattern struct {
		Elements Patterns
		Rest     *Pattern `optional:"true"`

		LeftBracket  Idx
		RightBracket Idx
	}
)

// IsPattern reports whether the target is a destructuring pattern (array or
// object), as opposed to a simple binding/assignment target.
func (p *Pattern) IsPattern() bool {
	return p.kind == PatternArrPat || p.kind == PatternObjPat
}

func (n *AssignmentPattern) Idx0() Idx { return n.Left.Idx0() }
func (n *AssignmentPattern) Idx1() Idx { return n.Right.Idx1() }

func (n *PatternKeyValue) Idx0() Idx { return n.Key.Idx0() }
func (n *PatternKeyValue) Idx1() Idx { return n.Value.Idx1() }

func (n *PatternShorthand) Idx0() Idx { return n.Name.Idx0() }
func (n *PatternShorthand) Idx1() Idx {
	if n.Initializer != nil {
		return n.Initializer.Idx1()
	}
	return n.Name.Idx1()
}
