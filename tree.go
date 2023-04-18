// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"sort"
)

type (
	// Node is the comparable type for each node of Tree.
	Node interface{ ~int | ~string }

	// Comparator is a value returned by the Order function of a Table for sorting.
	Comparator interface{ ~int | ~string }

	// Table refers to the data corresponding the nodes of a tree.
	Table[N Node, V any] map[N]V

	// Tree defines a hierarchy of Nodes of comparable type.
	Tree[N Node, C Comparator, V any] map[N]Tree[N, C, V]

	// Order distinguishes similar Node values for sorting.
	Order[N Node, C Comparator, V any] func(N, Table[N, V]) C

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

// Traverse walks the tree and invokes function fn for each node.
func (tr Tree[N, C, V]) Traverse(depth int, tbl Table[N, V], order Order[N, C, V], display Display[N, V]) {
	for _, node := range tr.order(tbl, order) {
		display(depth, node, tbl)
		tr[node].Traverse(depth+1, tbl, order, display)
	}
}

// FlatTree walks the tree to build an ordered slice of the nodes.
func (tr Tree[N, C, V]) Flatten(tbl Table[N, V], order Order[N, C, V]) []N {
	var flat []N
	for _, node := range tr.order(tbl, order) {
		flat = append(append(flat, node), tr[node].Flatten(tbl, order)...)
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

// order sorts the top nodes of a tree and returns as an ordered slice.
func (tr Tree[N, C, V]) order(tbl Table[N, V], fn Order[N, C, V]) []N {
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
