package workerpool

import (
	"sync"
)

// Worker describes the worker behavior.
//
// Do is the main work function which executes in concurrent mode. Realization should be concurrent-safe.
//
// Finish is the callback-method and executes when all of the workers end their work.
type Worker interface {
	Do()
	Finish()
}

// A Pool run certain worker by multiple goroutines, which are adjust by capacity property.
type Pool struct {
	wg       *sync.WaitGroup
	worker   Worker
	capacity int
}

// New creates pool with certain worker.
func New(w Worker, capacity int) *Pool {
	return &Pool{
		wg:       &sync.WaitGroup{},
		worker:   w,
		capacity: capacity,
	}
}

// Run starts worker in concurrent mode.
func (p *Pool) Run() {
	go func() {
		for range p.capacity {
			p.wg.Add(1)
			go func() {
				defer p.wg.Done()
				p.worker.Do()
			}()
		}

		p.wg.Wait()

		p.worker.Finish()
	}()
}
