package proxy

import (
	"sync"
	"testing"
)

func Test_gcd(t *testing.T) {
	a := 10
	b := 12
	res := gcd(a, b)
	if res != 2 {
		t.Error("res not equal 2")
		t.FailNow()
	}
}

func Test_nGCD(t *testing.T) {
	type args struct {
		nums []int
	}
	testCases := []struct {
		desc string
		args args
		want int
	}{
		{
			desc: "case 0",
			args: args{
				nums: []int{4, 8, 16},
			},
			want: 4,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			if got := nGCD(tC.args.nums, len(tC.args.nums)); got != tC.want {
				t.Errorf("want %d, got: %d", tC.want, got)
				t.FailNow()
			}
		})
	}
}

func Test_balancer(t *testing.T) {
	ws := []W{Weight(20), Weight(30), Weight(50)}
	bla := NewBalancer(ws)

	count := make(map[int]int)
	for i := 0; i < 10; i++ {
		idx := bla.Distribute()
		count[idx]++
	}

	t.Log(count)
}

func Test_balancer_concurrent(t *testing.T) {
	ws := []W{Weight(20), Weight(30), Weight(50)}
	bla := NewBalancer(ws)

	count := make(map[int]int)
	m := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(1000)

	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			idx := bla.Distribute()
			m.Lock()
			count[idx]++
			m.Unlock()
		}()
	}

	wg.Wait()
	t.Log(count)
}
