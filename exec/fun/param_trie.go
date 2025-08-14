package fun

import (
	"fmt"
	"reflect"
	"rice/lib/queue"
	"rice/lib/stack"
	"slices"
	"strings"
)

/*

	Parameter Trie Registration & Lookup algorithm

	1. Basic knowledge:
	- `any` (or `interface{}`) accepts any kind of value regardless of the dimension
	- `[]any, [][]any, ...` (`any` with d>0) can only accept itself; e.g. `[]any` cannot accept `[]int`
	- Variadic: f(...T) = f() | f([]T with len > 0)

	2. Parameter Trie
	- The parameter trie (prefix trees) allows quickly search for compatible argument types
	- Each node might have a handler. If exists, prefixed arguments could be supplied to that handler
	- Each node might have checkpoint. If exists, suffixed arguments could be captured for variadic support

	3. Registration:
	3.1. Subtree creation
	- A list of params turn into a subtree - typically, represented as a linked list unless branching happens
	- Branching happens due to union types (somewhat similar to Monomorphization)
	e.g. `int|string` causes two branches `int` and `string` separately created
	- The leaves of that subtree holds the same pointer to the handler
	e.g. (int|string), bool -> HANDLER
	turn into:
		└─int
			└─bool -> HANDLER
		└─string
			└─bool -> HANDLER

	- For later lookups, we prefer specificity (a.k.a exact matching >> relative matching), therefore, we sort the
	  children array at each node (see: param_trie.go#GetOrCreateChild)
		- In essence, sort by dimension size and base type. Here is an example: `int, []int, any`
		- Array has lower precedence: used for variadic support, `int` >> `[]int`
		- `any` has the lowest precedence

	- For variadic support, splits into two cases: "at-least-one" and "none"
		- "at-least-one" attaches the handler to the last parameter
		- "none" attaches the handler to the parent of the last parameter
	e.g. (int|string), ...string -> HANDLER
	turn into:
		└─int -> HANDLER 			# none
			└─...string -> HANDLER	# at-least-one
		└─string -> HANDLER 	    # none
			└─...string -> HANDLER	# at-least-one

	3.2. Tree merging
	- That subtree is glued to the Trie (if exists) under the function name
	- We guarantee atomicity. If we cannot set the handler, an error is returned and no change is partially made
	- Some conflicts example:
	a. (int|string) -> HANDLER; int -> HANDLER
		Explain: The first registration made branch (int -> HANDLER); the second registration is duplicated
	b. -> HANDLER; ...int -> HANDLER
		Explain: The first registration attaches a handler to the root; the second registration has variadity, which
		will split into two cases with one case is "none". It fails because "none" also attempts to attach a handler
		to the root

	4. Lookup
	- (*) Rule of thumb: The provided argument list is the source of trust
	- Strategy: One-by-one matching with DFS
	4.1. Without variadic:
		+ Continue the current branch if the child accepts the argument type
		+ Return the answer when reaching the first leaf having an existing handler
		+ Precedence is implicitly guaranteed due to sorted children list
	e.g. int, any -> HANDLER_1; any, any -> HANDLER_2
	turn into:
		└─int
			└─any -> HANDLER_1
		└─any
			└─any -> HANDLER_2
	call (0, 0) -> HANDLER_1
	call ("", 0) -> HANDLER_2

	4.2. Variadic support
	- The "zero" case is designed to simplify the algorithm, which equals to "Without variadic"
	- What matters now is the "at-least-one" case

	4.2.1. Checkpoint candidate
	- A node is a checkpoint candidate when:
		+ Supports checkpoint (a.k.a it is the last parameter of a variadic function)
		+ Its supported type can hold the current argument type
	- Next goal: Ensure that candidate is valid for remaining arguments

	4.2.2. Checkpoint maintenance
	- When traversing a branch, we capture and prune locally-eligible candidates. It is possible to have more than one
	e.g. these two nodes construct the prefix of two function handlers
		└─any -> HANDLER
			└─...string -> HANDLER
	- The search on a branch ends when:
		+ Case 1: Reach the leaf, which means the branch supports enough argument types
			- Return the handler if exists
			- Otherwise, Process locally-captured checkpoints
		+ Case 2: None eligible children, which means an argument has no sub-branch to continue
			- Process locally-captured checkpoints

	4.2.3. Local checkpoint final process
	- Backwards iteration on the checkpoint list, find the first checkpoint in which the parameter at that point accepts
	all suffixed arguments
	e.g.
		└─string -> HANDLER_1
		└─...string -> HANDLER_2
		└─any -> HANDLER_3
	call ("") -> HANDLER_1
	call ("", "") -> HANDLER_2
	call (0) -> HANDLER_3

*/

// ParamTrie a trie of parameters
type ParamTrie struct {
	id   string
	root *ParamNode
}

func NewParamTrie(id string) *ParamTrie {
	return &ParamTrie{root: NewParamNode(0), id: id}
}

// Register a function
func (t *ParamTrie) Register(def *FunctionDef) error {
	q := queue.New[*ParamNode](4)
	q.Enqueue(t.root)

	zeroVar := make([]*ParamNode, 0)
	leaves := make([]*ParamNode, 0)

	for argIdx, argTypes := range def.args {
		sz := q.Size()

		if sz == 0 {
			return fmt.Errorf("function %q%s registration failed: branching error", t.id, def)
		}

		for i := 0; i < sz; i++ {
			parent, ok := q.Dequeue()
			if ok {
				// When a function is variadic, the last parameter can accept none argument
				// which means the last one does not have to exist at all
				// As such, we apply the handler to each parent of these nodes
				if def.variadic && argIdx == len(def.args)-1 {
					if parent.HasHandler() {
						return fmt.Errorf("function %q%s has duplicated handler (zero variadic)", t.id, def)
					}

					// We do not set the handler until everything is fully validated; for atomicity guarantee
					zeroVar = append(zeroVar, parent)
				}

				for _, argType := range argTypes {
					child := parent.GetOrCreateChild(argType, func() *ParamNode {
						return NewParamNode(argType)
					})
					q.Enqueue(child)
				}
			}
		}
	}

	for !q.IsEmpty() {
		leaf, ok := q.Dequeue()
		if ok {
			if leaf.HasHandler() {
				return fmt.Errorf("function %q%s has duplicated handler", t.id, def)
			}

			// We do not set the handler until everything is fully validated; for atomicity guarantee
			leaves = append(leaves, leaf)
		}
	}

	// These nodes have been validated and could have the same handler
	for _, node := range zeroVar {
		node.handler = def.handler
		node.markHasHandler()
		if def.contextual {
			node.markContextual()
		}
	}
	for _, node := range leaves {
		if def.variadic {
			node.markCheckpointSupport()
		}
		node.handler = def.handler
		node.markHasHandler()
		if def.contextual {
			node.markContextual()
		}
	}

	return nil
}

// MatchResult denotes the result of MatchHandler
type MatchResult struct {
	// Handler the executable function
	Handler reflect.Value

	// Contextual if true, the function call requires the first argument to be context.Context; then, follows
	// by the given argument list
	Contextual bool

	// VariadicIndex denotes the index of the argument matching the last parameter of a variadic function
	// Any argument starting from that point are passed the last parameter
	// This VariadicIndex is relative to the argument list and does not account for Contextual
	VariadicIndex int
}

// MatchHandler search for the most "specific" handler that accepts the given argument list
func (t *ParamTrie) MatchHandler(args []reflect.Value) (*MatchResult, error) {
	argTypes := make([]ArgType, len(args))
	for i, arg := range args {
		argType, err := getArgType(arg.Type())
		if err != nil {
			return nil, err
		}
		argTypes[i] = argType
	}

	// ArgPair: "next"-th provided arg corresponds to registered "node" children
	type ArgPair struct {
		node *ParamNode
		next int
	}

	type Checkpoint struct {
		// the last parameter of a variadic function
		node *ParamNode
		// the index of this parameter relative to the given argument list
		idx int
	}

	// try each branch, do DFS
	st := stack.New[ArgPair](4)
	st.Push(ArgPair{node: t.root, next: 0})

	// eligible checkpoints per path
	cpl := make([]Checkpoint, 0)

	for !st.IsEmpty() {
		pair, ok := st.Pop()
		if !ok {
			break
		}

		// Final process when a branch ends
		processLocalCp := func() (*ParamNode, int) {
			if len(cpl) > 0 {
				// Within the same branch, the later checkpoint gives more specificity
				for i := len(cpl) - 1; i >= 0; i-- {
					cp := cpl[i]
					check := true

					// Check for remaining arguments after cp.idx
					for j := cp.idx + 1; j < len(args); j++ {
						givenType := argTypes[j]

						if !cp.node.ArgType().CanContainMultiOf(givenType) {
							check = false
							break
						}
					}

					if check {
						return cp.node, cp.idx
					}
				}

				cpl = cpl[:0] // reset checkpoint list for the next branch
			}

			return nil, -1
		}

		// Case 1: Reach the leaf, enough arguments
		if pair.next == len(args) {
			if pair.node.HasHandler() {
				return &MatchResult{Handler: pair.node.Handler(), Contextual: pair.node.IsContextual(), VariadicIndex: -1}, nil
			}

			if cp, idx := processLocalCp(); cp != nil {
				return &MatchResult{Handler: cp.Handler(), Contextual: cp.IsContextual(), VariadicIndex: idx}, nil
			}

			continue
		}

		// Inductive steps

		givenType := argTypes[pair.next]

		if len(cpl) > 0 { // Prune local checkpoint candidates
			// If there were checkpoints, they must comply with later argument types
			cpl = slices.DeleteFunc(cpl, func(cp Checkpoint) bool {
				return !cp.node.ArgType().CanContainMultiOf(givenType)
			})
		}

		// go backward because we are doing DFS, so the next inductive step can pick
		// the first child with respect to precedence
		eligibleChildren := 0
		for i := len(pair.node.childrenNav) - 1; i >= 0; i-- {
			child := pair.node.children[pair.node.childrenNav[i]]
			supportedType := child.ArgType()

			// Normal argument type check
			canContinue := supportedType.CanAccept(givenType)

			// Variadic support
			// When a node is an eligible checkpoint, we capture them
			// It is important to retain all checkpoints because we might have yet processed all arguments
			if child.SupportCheckpoint() && supportedType.CanContainMultiOf(givenType) {
				cpl = append(cpl, Checkpoint{node: child, idx: pair.next})
				canContinue = true
			}

			if canContinue {
				st.Push(ArgPair{node: child, next: pair.next + 1})
				eligibleChildren++
			}
		}

		// Case 2: None eligible children
		if eligibleChildren == 0 {
			if cp, idx := processLocalCp(); cp != nil {
				return &MatchResult{Handler: cp.Handler(), Contextual: cp.IsContextual(), VariadicIndex: idx}, nil
			}
		}
	}

	return nil, fmt.Errorf("no matching signature for function %q with given arguments", t.id)
}

////////////////////////////

func (t *ParamTrie) debug() string {
	var sb strings.Builder
	sb.WriteString(t.id + ":\n")
	t.debugNode(t.root, 1, &sb)
	return sb.String()
}

func (t *ParamTrie) debugNode(root *ParamNode, depth int, sb *strings.Builder) {
	if root.argType.Type() != 0 {
		sb.WriteString(strings.Repeat("  ", depth-1))
		sb.WriteString("└─")
		if root.SupportCheckpoint() {
			sb.WriteString("...")
		}
		sb.WriteString(root.ArgType().String())
		if root.HasHandler() {
			sb.WriteString(" -> [HANDLER]\n")
		} else {
			sb.WriteString("\n")
		}
	}
	for _, argType := range root.childrenNav {
		t.debugNode(root.children[argType], depth+1, sb)
	}
}
