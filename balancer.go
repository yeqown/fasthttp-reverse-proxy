package proxy

import (
	"sync"
)

// IBalancer .
type IBalancer interface {
	Distribute() int
}

// W
type W interface {
	Weight() int
}

// Weight .
type Weight uint

func (w Weight) Weight() int {
	return int(w)
}

// NewBalancer .
// TODO:
func NewBalancer(ws []W) IBalancer {
	rrb := roundrobinBalancer{
		// choices:      ws,
		mutex:        sync.Mutex{},
		weights:      make([]int, len(ws)),
		maxWeight:    0,
		maxGCD:       1,
		lenOfWeights: len(ws),
		i:            -1,
		cw:           0,
	}

	tmpGCD := make([]int, 0, rrb.lenOfWeights)

	for idx, w := range ws {
		rrb.weights[idx] = w.Weight()
		if w.Weight() > rrb.maxWeight {
			rrb.maxWeight = w.Weight()
		}
		tmpGCD = append(tmpGCD, w.Weight())
	}

	rrb.maxGCD = nGCD(tmpGCD, rrb.lenOfWeights)

	return &rrb
}

type roundrobinBalancer struct {
	// choices      []W
	mutex        sync.Mutex
	weights      []int // weight
	maxWeight    int   // 0
	maxGCD       int   // 1
	lenOfWeights int   // 0
	i            int   // last choice, -1
	cw           int   // current weight, 0
}

// Distribute to implement round robin algorithm
// it returns the idx of the choosing in ws ([]W)
// TODO: changing to no lock call
func (rrb *roundrobinBalancer) Distribute() int {
	rrb.mutex.Lock()
	defer rrb.mutex.Unlock()

	for {
		rrb.i = (rrb.i + 1) % rrb.lenOfWeights

		if rrb.i == 0 {
			rrb.cw = rrb.cw - rrb.maxGCD
			if rrb.cw <= 0 {
				rrb.cw = rrb.maxWeight
				if rrb.cw == 0 {
					return 0
				}
			}
		}

		if rrb.weights[rrb.i] >= rrb.cw {
			return rrb.i
		}
	}
}

// gcd
func gcd(a, b int) int {
	if a < b {
		a, b = b, a // swap a & b
	}

	if b == 0 {
		return a
	}

	return gcd(b, a%b)
}

// nGCD gcd in N nums
func nGCD(data []int, n int) int {
	if n == 1 {
		return data[0]
	}
	return gcd(data[n-1], nGCD(data, n-1))
}
