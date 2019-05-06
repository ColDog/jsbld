package graph

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type errMap map[int]bool

func runGraph(t *testing.T, g *Graph, errs errMap) (map[int]int, error) {
	l := sync.Mutex{}
	completed := map[int]int{}
	g.Process = func(ctx context.Context, id int) error {
		time.Sleep(100 * time.Millisecond)
		l.Lock()
		completed[id] += 1
		l.Unlock()
		if errs[id] {
			return fmt.Errorf("err: %d", id)
		}
		return nil
	}
	err := g.Solve(context.Background())
	for i, c := range completed {
		if c != 1 {
			t.Fatalf("completed more than once (%d)", i)
		}
	}
	return completed, err
}

func TestGraphN(t *testing.T) {
	g := &Graph{
		Concurrency: 2,
		Nodes: map[int][]int{
			1: []int{},
			2: []int{1, 3},
			3: []int{},
			4: []int{3},
			5: []int{4},
			6: []int{4},
			7: []int{4},
			8: []int{4},
		},
	}

	completed, err := runGraph(t, g, errMap{})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if len(completed) != len(g.Nodes) {
		t.Fatalf("invalid completion %+v", completed)
	}
	fmt.Printf("c: %v\n", completed)
}

func TestGraphErr(t *testing.T) {
	g := &Graph{
		Concurrency: 1,
		Nodes: map[int][]int{
			1: []int{},
			2: []int{1, 3},
			3: []int{},
		},
	}

	completed, err := runGraph(t, g, errMap{3: true})
	if err == nil {
		t.Fatalf("requires error")
	}
	fmt.Printf("c: %v, err: %v\n", completed, err)
}

func TestGraphCircular(t *testing.T) {
	g := &Graph{
		Concurrency: 1,
		Nodes: map[int][]int{
			3: []int{},
			1: []int{2, 3},
			2: []int{1},
		},
	}

	completed, err := runGraph(t, g, errMap{})
	if err == nil {
		t.Fatalf("requires error")
	}
	fmt.Printf("c: %v, err: %v\n", completed, err)
}
