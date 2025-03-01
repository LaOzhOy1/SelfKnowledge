package graph

import (
	"fmt"
	"testing"
)

// 查并集

// 一个人有多个 id 号， 只要不同的实体里 有id 号相同， 这两个实体代表的就是一个人，问一堆实体里有多少人

func TestEntityLinkCollection(t *testing.T) {

	var entities []Entity
	for i := 1; i <= 10; i++ {
		entities = append(entities, Entity{
			Bid:   i,
			Sid:   string(rune(i)),
			Alias: string(rune(i)),
		})
	}
	entities[0].Bid = 2
	entities[2].Bid = 4

	union := NewEntityLinkCollection(entities)
	map1 := map[int]Entity{}
	map2 := map[string]Entity{}
	map3 := map[string]Entity{}

	for _, entity := range entities {

		if e, ok := map1[entity.Bid]; ok {
			union.Union(e, entity)
		} else {
			map1[entity.Bid] = entity
		}
		if e, ok := map2[entity.Alias]; ok {
			union.Union(e, entity)
		} else {
			map2[entity.Alias] = entity
		}
		if e, ok := map3[entity.Sid]; ok {
			union.Union(e, entity)

		} else {
			map3[entity.Sid] = entity
		}
	}

	fmt.Println(len(union.NodesSize))
}

type EntityLinkCollection struct {
	Nodes     map[Entity]struct{}
	Parents   map[Entity]Entity
	NodesSize map[Entity]int
}

func NewEntityLinkCollection(entities []Entity) EntityLinkCollection {
	nodes := make(map[Entity]struct{})
	parents := make(map[Entity]Entity)
	nodesSize := make(map[Entity]int)

	for _, e := range entities {
		nodes[e] = struct{}{}
		parents[e] = e
		nodesSize[e] = 1
	}
	return EntityLinkCollection{
		Nodes:     nodes,
		Parents:   parents,
		NodesSize: nodesSize,
	}
}

func (e *EntityLinkCollection) findFather(e1 Entity) Entity {
	history := make([]Entity, 0)
	for e.Parents[e1] != e1 {
		history = append(history, e1)
		e1 = e.Parents[e1]
	}
	for _, h := range history {
		e.Parents[h] = e1
	}
	return e1
}

func (e *EntityLinkCollection) IsSame(e1, e2 Entity) bool {
	if _, f1 := e.Nodes[e1]; !f1 {
		return false
	}

	if _, f2 := e.Nodes[e2]; !f2 {
		return false
	}

	return e.findFather(e1) == e.findFather(e2)
}

func (e *EntityLinkCollection) Union(e1, e2 Entity) {
	if _, f1 := e.Nodes[e1]; !f1 {
		return
	}

	if _, f2 := e.Nodes[e2]; !f2 {
		return
	}
	f1 := e.findFather(e1)
	f2 := e.findFather(e2)
	if f1 != f2 {
		s1 := e.NodesSize[e1]
		s2 := e.NodesSize[e2]
		if s1 < s2 {
			e.Parents[f1] = f2
			e.NodesSize[f1] += e.NodesSize[f2]
			delete(e.NodesSize, f2)
		} else {
			e.Parents[f2] = f1
			e.NodesSize[f2] += e.NodesSize[f1]
			delete(e.NodesSize, f1)
		}
	}
}

type Entity struct {
	Bid   int
	Sid   string
	Alias string
}
