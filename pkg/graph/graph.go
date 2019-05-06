package graph

import (
	"context"
	"errors"
	"log"
	"sync"
)

var (
	ErrUnsolvable = errors.New("graph: unsolvable graph")
)

type done struct {
	id  int
	err error
}

type work struct {
	id   int
	ctx  context.Context
	done chan done
}

type ProcessFunc func(ctx context.Context, id int) error

type Graph struct {
	Concurrency int
	Nodes       map[int][]int
	Process     ProcessFunc

	wg        *sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	inFlight  map[int]bool
	completed map[int]bool
	work      chan work
	err       error
	done      chan done
}

func (g *Graph) init() {
	g.completed = map[int]bool{}
	g.inFlight = map[int]bool{}
	g.work = make(chan work, g.Concurrency)
	g.done = make(chan done)
	g.wg = &sync.WaitGroup{}
	g.ctx, g.cancel = context.WithCancel(context.Background())
}

func (g *Graph) Solve(ctx context.Context) error {
	g.init()

	g.wg.Add(g.Concurrency)
	for i := 0; i < g.Concurrency; i++ {
		go worker(i, g.Process, g.wg, g.work)
	}
	err := g.pump(ctx)
	g.wg.Wait()
	close(g.done)
	return err
}

// Worker processes individual items from the work queue.
func worker(i int, process ProcessFunc, wg *sync.WaitGroup, work chan work) {
	log.Printf("worker[%d]: starting", i)
	defer log.Printf("worker[%d]: stopping", i)
	defer wg.Done()

	for work := range work {
		log.Printf("worker[%d]: starting work=%d", i, work.id)
		err := process(work.ctx, work.id)
		work.done <- struct {
			id  int
			err error
		}{id: work.id, err: err}
		log.Printf("worker[%d]: finished work=%d", i, work.id)
	}
}

// Reads from done channel and pumps work into the work channel. This function
// sets state on the graph object.
func (g *Graph) pump(ctx context.Context) error {
	log.Printf("pump: starting")
	defer close(g.work)

	// Prime the initial channel with work to be done. Block when sending work
	// here as we need to get something into the initial channels or else we'll
	// hit deadlock.
	if !g.sendWork(true) {
		return ErrUnsolvable
	}

	// Wait for a worker to be freed to send more work down the work channel.
	// If jobs are still in flight, we continue to read from the done channel.
	for !g.finished() || g.working() {
		select {
		case done := <-g.done:
			log.Printf("pump: work done work=%d", done.id)
			g.complete(done.id)

			// If there's an error, mark this globally.
			if done.err != nil {
				log.Printf("pump: receive error: %v", done.err)
				g.errored(done.err)
			}

			if !g.finished() {
				sent := g.sendWork(false)
				// Unsolvable case: no in flight processing and no new work
				// sent to the queue. A circular dependency must exist.
				if !sent && !g.working() {
					return ErrUnsolvable
				}
			}
		case <-ctx.Done():
			log.Printf("pump: context cancelled - waiting for workers to exit")
			g.errored(ctx.Err())
		}
	}

	log.Println("pump: finished")
	return g.err
}

// Errorred sets the error on the graph and also should cancel all work.
func (g *Graph) errored(err error) {
	g.err = err
	g.cancel()
}

func (g *Graph) working() bool {
	return len(g.inFlight) > 0
}

// Finished
func (g *Graph) finished() bool {
	return g.err != nil || len(g.completed) >= len(g.Nodes)
}

// Complete a set of work marking it done and not in flight.
func (g *Graph) complete(id int) {
	g.completed[id] = true
	delete(g.inFlight, id)
}

// SendWork pushes work into the work channel. If block is set to true it will
// block and push all ready work into the channel. If block is false it returns
// as soon as the channel blocks.
func (g *Graph) sendWork(block bool) (sent bool) {
	for id := range g.Nodes {
		if g.ready(id) {
			if block {
				g.work <- work{id: id, ctx: g.ctx, done: g.done}
			} else {
				select {
				case g.work <- work{id: id, ctx: g.ctx, done: g.done}:
				default:
					return
				}
			}

			g.inFlight[id] = true
			log.Printf("pump: send work work=%d inFlight=%+v", id, g.inFlight)
			sent = true
		}
	}
	return
}

// Ready returns whether work can be started on.
func (g *Graph) ready(id int) bool {
	if g.inFlight[id] {
		return false
	}
	if g.completed[id] {
		return false
	}
	for _, dep := range g.Nodes[id] {
		if !g.completed[dep] {
			return false
		}
	}
	return true // All dependencies are completed.
}
