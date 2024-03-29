package injast

type Pos int

// All node types implement the Node interface.
type Node interface {
	Pos() Pos // position of first character belonging to the node
	End() Pos // position of first character immediately after the node
}

type Expr interface {
	Node
	exprNode()
}

type (
	// A BadExpr node is a placeholder for expressions containing
	// syntax errors for which no correct expression nodes can be
	// created.
	//
	BadExpr struct {
		From, To Pos // position range of bad expression
	}

	// An Ident node represents an identifier.
	Ident struct {
		NamePos Pos // identifier position
		Name    string    // identifier name
	}

	// An IndexExpr node represents an expression followed by an index.
	IndexExpr struct {
		X      Expr      // expression
		Lbrack Pos // position of "["
		Index  Expr      // index expression
		Rbrack Pos // position of "]"
	}

	// A KeyValueExpr node represents (key : value) pairs
	// in composite literals.
	//
	KeyValueExpr struct {
		Key   Expr
		Colon Pos // position of ":"
		Value Expr
	}
)

func (*Ident) exprNode()          {}
func (x *Ident) Pos() Pos    { return x.NamePos }
func (x *Ident) End() Pos   { return Pos(int(x.NamePos) + len(x.Name)) }
func (id *Ident) String() string {
	if id != nil {
		return id.Name
	}
	return "<nil>"
}

func (*IndexExpr) exprNode()      {}
func (x *IndexExpr) Pos() Pos      { return x.X.Pos() }
func (x *IndexExpr) End() Pos      { return x.Rbrack + 1 }

func (*KeyValueExpr) exprNode()   {}
func (x *KeyValueExpr) Pos() Pos   { return x.Key.Pos() }
func (x *KeyValueExpr) End() Pos   { return x.Value.End() }