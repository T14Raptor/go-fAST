package ast

// AST matching is implemented via generated, type-safe code.

type Any struct{}

func AnyNode() *Any { return &Any{} }

type Capture struct {
	Out any
}

func CaptureNode(out any) *Capture {
	return &Capture{Out: out}
}

func Match(actual, pattern any) bool {
	return matchAny(actual, pattern)
}

func (*Any) Idx0() Idx                 { return 0 }
func (*Any) Idx1() Idx                 { return 0 }
func (*Any) VisitWith(Visitor)         {}
func (*Any) VisitChildrenWith(Visitor) {}
func (*Any) _expr()                    {}
func (*Any) _stmt()                    {}

func (*Capture) Idx0() Idx                 { return 0 }
func (*Capture) Idx1() Idx                 { return 0 }
func (*Capture) VisitWith(Visitor)         {}
func (*Capture) VisitChildrenWith(Visitor) {}
func (*Capture) _expr()                    {}
func (*Capture) _stmt()                    {}
