package graph

import (
	"fmt"
	"testing"
)

// 通过遍历每个点， 选出边最小的点进行相连， 知道所有点遍历完， 每次都是 集合 与 点进行相结合， 故不需要查并集

type PrimeNode struct {
	Id   int
	Next []*PrimeEdge
}

type PrimeEdge struct {
	From   *PrimeNode
	To     *PrimeNode
	Weight int
}

type PrimeGraph struct {
	Nodes map[int]*PrimeNode
	Edges map[*PrimeEdge]struct{}
}

func TestPrime(t *testing.T) {
	arr := [][]int{
		{1, 2, 1},
		{2, 3, 2},
		{3, 4, 3},
		{4, 1, 4},
		{1, 3, 5},
		{2, 4, 6},
	}
	g := NewPrimeGraph(arr, 4)
	g.Prime(3)
}

func NewPrimeGraph(arr [][]int, n int) *PrimeGraph {
	nodes := make(map[int]*PrimeNode, n)
	edges := make(map[*PrimeEdge]struct{}, len(arr))

	primeGraph := &PrimeGraph{
		Nodes: nodes,
		Edges: edges,
	}

	for i := 1; i <= n; i++ {
		node := &PrimeNode{
			Id:   i,
			Next: nil,
		}
		primeGraph.Nodes[i] = node
	}
	for i := 0; i < len(arr); i++ {
		from := primeGraph.Nodes[arr[i][0]]
		to := primeGraph.Nodes[arr[i][1]]
		edge := &PrimeEdge{
			From:   from,
			To:     to,
			Weight: arr[i][2],
		}
		// 假设 arr[i][0] arr [i][1] 必定存在
		from.Next = append(from.Next, edge)
		primeGraph.Edges[edge] = struct{}{}
	}

	return primeGraph
}

func (p *PrimeGraph) Prime(n int) {
	distances := make(map[*PrimeNode]int)
	selected := make(map[*PrimeNode]struct{})
	distances[p.Nodes[1]] = 0

	min := p.selectNextMinNode(selected, distances)
	for min != nil {
		distance := distances[min]
		edges := min.Next
		for _, edge := range edges {
			if d, ok := distances[edge.To]; ok {
				if distance+edge.Weight < d {
					distances[edge.To] = distance + edge.Weight
				}
			} else {
				distances[edge.To] = distance + edge.Weight
			}
		}
		selected[min] = struct{}{}
		min = p.selectNextMinNode(selected, distances)
	}

	fmt.Println(distances[p.Nodes[n]])
}

func (p *PrimeGraph) selectNextMinNode(selected map[*PrimeNode]struct{}, distances map[*PrimeNode]int) *PrimeNode {
	min := -1
	var minNode *PrimeNode
	for node, d := range distances {
		if _, ok := selected[node]; !ok {
			if min == -1 {
				min = d
				minNode = node
				continue
			}
			if d < min {
				min = d
				minNode = node
			}
		}
	}
	return minNode
}
