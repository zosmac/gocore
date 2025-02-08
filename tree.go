// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"cmp"
	"iter"
	"slices"
)

type (
	// Node is the comparable type for each node of Tree.
	Node = cmp.Ordered

	// Tree defines a hierarchy of Nodes of comparable type.
	Tree[N Node] map[N]Tree[N]

	// Table provides an optional dictionary of values for the Nodes. If used, insert new Nodes here first to ensure uniqueness in Tree.
	Table[N Node, V any] map[N]V
)

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

// SortedFunc walks the tree, returning an ordered sequence of each node's value and depth in the tree and descending through its subnodes, ordered with a comparison function.
func (tr Tree[N]) SortedFunc(cmp func(a, b N) int) iter.Seq2[int, N] {
	return func(yield func(int, N) bool) {
		tr.sort(0, cmp, yield)
	}
}

// sort walks the tree and orders each node's subtree.
func (tr Tree[N]) sort(depth int, cmp func(a, b N) int, yield func(int, N) bool) bool {
	nodes := []N{}
	for node := range tr {
		nodes = append(nodes, node)
	}
	slices.SortFunc(nodes, cmp)
	for _, node := range nodes {
		if !yield(depth, node) {
			return false
		}
		if !tr[node].sort(depth+1, cmp, yield) {
			return false
		}
	}
	return true
}

// DepthTree enables sort of deepest process trees first.
func (tr Tree[N]) DepthTree() int {
	depth := 1
	for _, tr := range tr {
		depth = max(depth, tr.DepthTree()+1)
	}
	return depth
}

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

func (tr Tree[N]) Family(node N) Tree[N] {
	if _, ok := tr[node]; ok {
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
