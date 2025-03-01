package greedy

import (
	"fmt"
	"testing"
)

//  每个boss 有金币奖励，  打每个boss 要入门票 , 最多能打 k 个boss ，给你初始资金， 赚最多的金币

type Profit struct {
	Ticket int
	Get    int
}

type BossHeap struct {
	Profit []Profit
	Size   int
	Cmp    func(i int, j int) bool
}

func NewBossHeap(profit []Profit, cmp func(i int, j int) bool) *BossHeap {
	heap := &BossHeap{
		Profit: profit,
		Size:   len(profit),
		Cmp:    cmp,
	}
	heap.heapify()
	return heap
}

func exchange(arr []Profit, i, j int) {
	tmp := arr[j]
	arr[j] = arr[i]
	arr[i] = tmp
}

func (h *BossHeap) heapify() {
	size := len(h.Profit)
	for i := (size - 1) / 2; i >= 0; i-- {
		h.heapDown(i)
	}
}

func (h *BossHeap) heapDown(cur int) {
	if cur >= h.Size {
		return
	}
	parent := cur
	left := cur*2 + 1
	right := cur*2 + 2
	if left < h.Size && h.Cmp(left, parent) {
		exchange(h.Profit, left, parent)
		h.heapDown(left)
		return
	}
	if right < h.Size && h.Cmp(right, parent) {
		exchange(h.Profit, right, parent)
		h.heapDown(right)
		return
	}
}

func (h *BossHeap) heapUp(cur int) {
	if cur == 0 {
		return
	}
	par := (cur - 1) / 2
	if h.Cmp(cur, par) {
		exchange(h.Profit, cur, par)
		h.heapUp(par)
	}
}

func (h *BossHeap) CmpTicket(i, j int) bool {
	return h.Profit[i].Ticket < h.Profit[j].Ticket
}

func (h *BossHeap) Pop() Profit {
	h.Size--
	exchange(h.Profit, 0, h.Size)
	h.heapify()
	return h.Profit[h.Size]
}

// 不适用指针, 逃逸到堆，增加gc量
func (h *BossHeap) Peek() Profit {
	return h.Profit[h.Size-1]
}
func (h *BossHeap) IsEmpty() bool {
	return h.Size == 0
}
func (h *BossHeap) Push(profit Profit) {
	if h.Size == len(h.Profit) {
		return
	}
	h.Profit[h.Size] = profit
	h.heapUp(h.Size)
	h.Size++
	return
}
func TestBoss(t *testing.T) {
	arr := []Profit{
		{
			Ticket: 4,
			Get:    2,
		},
		{
			Ticket: 3,
			Get:    5,
		},
		{
			Ticket: 2,
			Get:    6,
		},
		{
			Ticket: 1,
			Get:    5,
		},
	}

	// 对原数组有影响
	cmpTicketAsc := func(i, j int) bool {
		return arr[i].Ticket < arr[j].Ticket
	}

	cmpGetDesc := func(i, j int) bool {
		return arr[i].Get > arr[j].Get
	}

	ticketAscHeap := NewBossHeap(arr, cmpTicketAsc)

	get := make([]Profit, len(arr))

	copy(get, arr)

	getDescHeap := NewBossHeap(get, cmpGetDesc)

	//fmt.Println(ticketAscHeap.Profit[0])
	//fmt.Println(GetDescHeap.Profit[0])
	//fmt.Println(ticketAscHeap.Pop())
	//fmt.Println(GetDescHeap.Pop())
	//fmt.Println(ticketAscHeap.Profit[0])
	//fmt.Println(GetDescHeap.Profit[0])

	have := 1
	K := 3 // 最多能玩3个项目
	for i := 0; i < K; i++ {
		// 把所有能玩的项目导入到利润倒叙堆
		for !ticketAscHeap.IsEmpty() && ticketAscHeap.Peek().Ticket <= have {
			getDescHeap.Push(ticketAscHeap.Pop())
		}
		maxProfit := getDescHeap.Pop()
		have += maxProfit.Get
	}
	fmt.Println(have)
}
