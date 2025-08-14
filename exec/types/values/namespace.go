package values

import (
	"errors"
	"fmt"
	"github.com/anhcraft/rice/exec/types"
	"iter"
	"strings"
)

var _ = types.Namespace.DefineType((*Namespace)(nil))
var _ IndexedCollection = (*Namespace)(nil)
var _ NamespacedLayer = (*namespaceNodes)(nil)

// NamespacedLayer represents a layer having namespace
type NamespacedLayer interface {
	IndexedCollection
}

// namespaceNodes a NamespacedLayer of namespace children
type namespaceNodes map[types.Value]types.Value

func (n namespaceNodes) Keys() []types.Value {
	k := make([]types.Value, len(n))
	j := 0
	for i := range n {
		k[j] = i
		j++
	}
	return k
}

func (n namespaceNodes) PutElement(id types.Value, item types.Value) error {
	n[id] = item
	return nil
}

func (n namespaceNodes) Element(id types.Value) (types.Value, error) {
	v, ok := n[id]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (n namespaceNodes) Iterate() iter.Seq[types.Value] {
	return func(yield func(types.Value) bool) {
		for _, v := range n {
			if !yield(v) {
				return
			}
		}
	}
}

func (n namespaceNodes) Size() Int {
	return Int(len(n))
}

// Namespace the namespace tree
type Namespace struct {
	final NamespacedLayer
}

func (ns *Namespace) PutElement(id types.Value, item types.Value) error {
	return errors.New("unsupported operation")
}

func (ns *Namespace) String() string {
	return "Namespace"
}

func (ns *Namespace) Type() types.Type {
	return types.Namespace
}

func NewNamespace() *Namespace {
	return &Namespace{
		final: make(namespaceNodes),
	}
}

// Merge merges a layer
func (ns *Namespace) Merge(layer NamespacedLayer) {
	for _, k := range layer.Keys() {
		v, err := layer.Element(k)
		if err != nil {
			panic(err)
		}
		err = ns.final.PutElement(k, v)
		if err != nil {
			panic(err)
		}
	}
}

// Element gets a child
func (ns *Namespace) Element(id types.Value) (types.Value, error) {
	if key, ok := id.(Identifier); ok {
		return ns.final.Element(key)
	} else {
		return nil, elementNotIdErr
	}
}

func (ns *Namespace) Size() Int {
	return ns.final.Size()
}

func (ns *Namespace) Keys() []types.Value {
	return ns.final.Keys()
}

func (ns *Namespace) Iterate() iter.Seq[types.Value] {
	return ns.final.Iterate()
}

// Path recursively traverse the given identifier path with respect to the precedence
func (ns *Namespace) Path(ids ...Identifier) (types.Value, bool) {
	node := ns
	for _, id := range ids {
		child, err := node.Element(id)
		if child == nil || err != nil {
			return nil, false
		}
		if v, ok := child.(*Namespace); !ok {
			return child, true
		} else {
			node = v
		}
	}
	return node, true
}

// RequirePath recursively traverse the given identifier path; create descendant namespaces
// if they do not exist yet
func (ns *Namespace) RequirePath(ids ...Identifier) *Namespace {
	node := ns
	for _, id := range ids {
		child, err := node.final.Element(id)
		if err != nil {
			panic(err)
		}

		if child == nil {
			child = NewNamespace()
			err = node.final.PutElement(id, child)
			if err != nil {
				panic(err)
			}
		} else if _, ok := child.(*Namespace); !ok {
			panic(fmt.Sprintf("expected namespace but got %T at id %q", child, id))
		}

		node = child.(*Namespace)
	}
	return node
}

////////////////////////////

func (ns *Namespace) Debug() string {
	var sb strings.Builder
	debugNS(&sb, 0, ns)
	return sb.String()
}

func debugNS(sb *strings.Builder, indent int, ns *Namespace) {
	sb.WriteString(strings.Repeat("  ", indent))
	sb.WriteString("└─")
	sb.WriteString("|Namespace|\n")
	debugLayer(sb, indent+1, ns.final)
}

func debugLayer(sb *strings.Builder, indent int, layer NamespacedLayer) {
	for _, k := range layer.Keys() {
		sb.WriteString(strings.Repeat("  ", indent))
		sb.WriteString("└─")
		v, err := layer.Element(k)
		if err != nil {
			sb.WriteString(fmt.Sprintf("%v: (error: %v)", k, err))
		} else if nsl, ok := v.(*Namespace); ok {
			sb.WriteString(fmt.Sprintf("%v: Namespace\n", k))
			debugNS(sb, indent+2, nsl)
		} else if nsl, ok := v.(NamespacedLayer); ok {
			sb.WriteString(fmt.Sprintf("%v: NamespacedLayer\n", k))
			debugLayer(sb, indent+1, nsl)
		} else {
			sb.WriteString(fmt.Sprintf("%v: %v", k, v))
		}
		sb.WriteString("\n")
	}
}
