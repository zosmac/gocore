// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"sort"
)

type (
	// Node is the comparable type for each node of Tree.
	Node interface{ ~int | ~string }

	// Table refers to the data corresponding the nodes of a tree.
	Table[N Node, V any] map[N]V

	// Tree defines a hierarchy of Nodes of comparable type.
	Tree[N, C Node, V any] map[N]Tree[N, C, V]

	// Order distinguishes similar Node values for sorting.
	Order[N, C Node, V any] func(N, Table[N, V]) C

	// Display shows the value of a Node at particular depth in Tree.
	Display[N Node, V any] func(int, N, Table[N, V])
)

// Add inserts a node into a tree.
func (tr Tree[N, C, V]) Add(nodes ...N) {
	if len(nodes) > 0 {
		if _, ok := tr[nodes[0]]; !ok {
			tr[nodes[0]] = Tree[N, C, V]{}
		}
		tr[nodes[0]].Add(nodes[1:]...)
	}
}

// Sortnodes sorts the top nodes of a tree and returns as an ordered slice.
func (tr Tree[N, C, V]) Sortnodes(tbl Table[N, V], fn Order[N, C, V]) []N {
	var nodes []N
	for node := range tr {
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool {
		nodei := fn(nodes[i], tbl)
		nodej := fn(nodes[j], tbl)
		return nodei < nodej ||
			nodei == nodej && nodes[i] < nodes[j]
	})
	return nodes
}

// Traverse walks the tree and invokes function fn for each node.
func (tr Tree[N, C, V]) Traverse(depth int, tbl Table[N, V], fn1 Order[N, C, V], fn2 Display[N, V]) {
	for _, node := range tr.Sortnodes(tbl, fn1) {
		fn2(depth, node, tbl)
		tr[node].Traverse(depth+1, tbl, fn1, fn2)
	}
}

// FlatTree walks the tree to build an ordered slice of the nodes.
func (tr Tree[N, C, V]) FlatTree(tbl Table[N, V], fn Order[N, C, V]) []N {
	var flat []N
	for _, node := range tr.Sortnodes(tbl, fn) {
		flat = append(append(flat, node), tr[node].FlatTree(tbl, fn)...)
	}
	return flat
}

// FindTree finds the subtree anchored by a specific node.
func (tr Tree[N, C, V]) FindTree(node N) Tree[N, C, V] {
	for n, tr := range tr {
		if n == node {
			return Tree[N, C, V]{node: tr}
		}
		if tr = tr.FindTree(node); tr != nil {
			return tr
		}
	}
	return nil
}
