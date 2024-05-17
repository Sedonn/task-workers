package workerpool

import (
	"errors"
	"sync"
)

var ErrWorkEmpty = errors.New("worker pool function empty")

// A Pool creates worker pool and distributes single work-function to workers.
type Pool struct {
	wg       *sync.WaitGroup
	capacity uint
	work     func() // will be executed by all of workers.
	onFinish func() // will be executed when all of the workers end their work.
}

// New creates pool with certain count of workers.
func New(capacity uint) *Pool {
	return &Pool{
		wg:       &sync.WaitGroup{},
		capacity: capacity,
	}
}

// SetWork register the main work function.
//
// Pool implies that work function will be use only channels in his parameters.
func (p *Pool) SetWork(f func()) {
	p.work = f
}

// SetFinishCallback sets the finish callback function.
//
// Passed function will be executed when all of the workers finish their work.
func (p *Pool) SetFinishCallback(f func()) {
	p.onFinish = f
}

// Run tells all of workers to do function passed to SetWork method.
func (p *Pool) Run() error {
	if p.work == nil {
		return ErrWorkEmpty
	}

	go func() {
		for range p.capacity {
			p.wg.Add(1)
			go func() {
				defer p.wg.Done()
				p.work()
			}()
		}

		p.wg.Wait()

		if p.onFinish != nil {
			p.onFinish()
		}
	}()

	return nil
}
