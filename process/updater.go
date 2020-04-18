package process

// set
// num_{add,mul,div)

// XXX: We have very naive implementation. We must optimize by creating incremental operation tree.

type Manipulator struct {
	activeKey     string
	next          DocumentWriter
	active        Manipulation
	manipulations map[int64]Manipulation
}

type Manipulation interface {
	DocumentProcessor
}

func NewManipulator(next DocumentWriter) *Manipulator {
	return &Manipulator{next: next}
}

func (m *Manipulator) AddManipulation(path []string, man Manipulation) {
}

func (m *Manipulator) PushString(ctx *Context, s string) error {
}

func (m *Manipulator) PushBlob(ctx *Context, b []byte) error {
}

func (m *Manipulator) PushInt(ctx *Context, i int64) error {
}

func (m *Manipulator) PushUint(ctx *Context, u uint64) error {
}

func (m *Manipulator) PushFloat(ctx *Context, f float64) error {
}

func (m *Manipulator) PushBool(ctx *Context, b bool) error {
}

func (m *Manipulator) PushNull(*Context) error {
}

func (m *Manipulator) BeginArray(*Context) (int, error) {
	panic("implement me")
}

func (m *Manipulator) EndArray(*Context, int) error {
	panic("implement me")
}

func (m *Manipulator) BeginObject(*Context) (int, error) {
	panic("implement me")
}

func (m *Manipulator) EndObject(*Context, int) error {
	panic("implement me")
}

func (m *Manipulator) PushObjectKey(ctx *Context, k string) error {
	m.activeKey = k
	return nil
}

type ManipulationTree struct {
	root ManipulationTreeNode
}

type ManipulationTreeElem interface {
	Key() string
}

type ManipulationTreeNode struct {
	key         string // "" at root
	childNodes  []*ManipulationTreeNode
	childLeaves []*ManipulationTreeLeaf
	parent      *ManipulationTreeNode // nil at root
}

type ManipulationTreeLeaf struct {
	key string
	man Manipulation
}

func (n *ManipulationTreeNode) Lookup(path []string) *ManipulationTreeLeaf {
	if len(path) == 0 {
		panic("path must be have one or more elements")
	}
	nextKey := path[0]
	tailKeys := path[1:]
	if len(tailKeys) == 0 {
		for _, leaf := range n.childLeaves {
			if leaf.key == nextKey {
				return leaf
			}
		}
	} else {
		var nextNode *ManipulationTreeNode
		for _, node := range n.childNodes {
			if node.key == nextKey {
				nextNode = node
			}
		}
		if nextNode != nil {
			return nextNode.Lookup(tailKeys)
		}
	}
	return nil
}

func (n *ManipulationTreeNode) Push(path []string, man Manipulation) {
	if len(path) == 0 {
		panic("path must be have one or more elements")
	}
	nextKey := path[0]
	tailKeys := path[1:]
	if len(tailKeys) == 0 {
		newLeaf := &ManipulationTreeLeaf{
			key: nextKey,
			man: man,
		}
		for i, leaf := range n.childLeaves {
			if leaf.key == nextKey {
				// replace leaf
				n.childLeaves[i] = newLeaf
				break
			}
		}
		n.childLeaves = append(n.childLeaves, newLeaf)
	} else {
		var nextNode *ManipulationTreeNode
		for _, node := range n.childNodes {
			if node.key == nextKey {
				nextNode = node
			}
		}
		if nextNode == nil {
			// create new node
			nn := &ManipulationTreeNode{
				key:        nextKey,
				childNodes: nil,
				parent:     n,
			}
			n.childNodes = append(n.childNodes, nn)
			nextNode = nn
		}
		nextNode.Push(tailKeys, man)
	}
}
