// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"iter"
	"slices"
)

type (
	// Node is the comparable type for each node of Tree.
	Node = comparable

	// Tree defines a hierarchy of Nodes of comparable type.
	Tree[N Node] map[N]Tree[N]

	// Table provides an optional dictionary of values for the Nodes. If used, insert new Nodes here first to ensure uniqueness in Tree.
	Table[N Node, V Value[N]] map[N]V

	// Value defines the interface for a node's value to specify its parent node for building a tree.
	Value[N Node] interface {
		HasParent() bool
		Parent() N
	}
)

// Parent returns the node and value for a parent node as determined by the current node's value.
func (tb Table[N, V]) Parent(node N) (N, V) {
	if ok := tb[node].HasParent(); ok {
		return tb[node].Parent(), tb[tb[node].Parent()]
	}
	var n N
	var v V
	return n, v // no parent, return empty node and value
}

// BuildTree builds the tree.
func (tb Table[N, V]) BuildTree() Tree[N] {
	tr := Tree[N]{}
	for node, val := range tb {
		nodes := []N{node}
		for ; val.HasParent(); val = tb[val.Parent()] {
			nodes = append([]N{val.Parent()}, nodes...)
		}
		tr.Add(nodes...)
	}
	return tr
}

// Ordered walks a map, returning an ordered sequence of each map element's key and value by key, ordered with a comparison function.
func Ordered[K comparable, V any](m map[K]V, cmp func(a, b K) int) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		keys := []K{}
		for key := range m {
			keys = append(keys, key)
		}
		slices.SortFunc(keys, cmp)
		for _, key := range keys {
			if !yield(key, m[key]) {
				return
			}
		}
	}
}

// Add adds new Nodes as a branch to the Tree.
func (tr Tree[N]) Add(nodes ...N) {
	if len(nodes) > 0 {
		if _, ok := tr[nodes[0]]; !ok {
			tr[nodes[0]] = Tree[N]{}
		}
		tr[nodes[0]].Add(nodes[1:]...)
	}
}

// All walks the tree, returning a sequence of each node and its depth in the tree and descending through its subnodes.
func (tr Tree[N]) All() iter.Seq2[int, N] {
	return func(yield func(int, N) bool) {
		tr.push(0, yield)
	}
}

// push pushes all elements to the yield function.
func (tr Tree[N]) push(depth int, yield func(int, N) bool) bool {
	for node, tr := range tr {
		if !yield(depth, node) {
			return false
		}
		if !tr.push(depth+1, yield) {
			return false
		}
	}
	return true
}

// Ordered walks the tree, ordering node values with a comparison function at each depth as it descends each node's subtree, reporting each node's depth and value in depth first order.
func (tr Tree[N]) Ordered(cmp func(a, b N) int) iter.Seq2[int, N] {
	return func(yield func(int, N) bool) {
		tr.order(0, cmp, yield)
	}
}

// order walks the tree and orders each node's subtree.
func (tr Tree[N]) order(depth int, cmp func(a, b N) int, yield func(int, N) bool) bool {
	for node := range Ordered(tr, cmp) {
		if !yield(depth, node) {
			return false
		}
		if !tr[node].order(depth+1, cmp, yield) {
			return false
		}
	}
	return true
}

// DepthTree enables ordering of tree nodes depth first.
func (tr Tree[N]) DepthTree() int {
	depth := 1
	for _, tr := range tr {
		depth = max(depth, tr.DepthTree()+1)
	}
	return depth
}

// Ancestors creates a slice of the ancestor nodes of the node.
func (tr Tree[N]) Ancestors(node N) []N {
	nodes := make([]N, 10)
	for depth, n := range tr.All() {
		if depth >= cap(nodes) {
			ns := make([]N, 2*cap(nodes))
			copy(ns, nodes)
			nodes = ns
		}
		nodes[depth] = n
		if node == n {
			return nodes[:depth]
		}
	}
	return nil
}

// FindTree finds the subtree anchored by a specific node.
func (tr Tree[N]) FindTree(node N) Tree[N] {
	if _, ok := tr[node]; ok {
		return tr
	}
	for _, tr := range tr {
		if tr = tr.FindTree(node); tr != nil {
			return tr
		}
	}
	return nil
}

// Family creates a tree of the selected node with its ancestors.
func (tr Tree[N]) Family(node N) Tree[N] {
	if _, ok := tr[node]; ok { // no ancestors
		return tr
	}
	anc := tr.Ancestors(node) // ancestors
	if len(anc) == 0 {
		return Tree[N]{} // not found
	}
	tr = tr.FindTree(node) // self and descendants
	parent := anc[len(anc)-1]
	fam := Tree[N]{}
	fam.Add(anc...)
	tb := fam.FindTree(parent)
	tb[parent][node] = tr[node]
	return fam
}
