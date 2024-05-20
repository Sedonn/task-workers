package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"sedonn/task-workers/task"
	"sedonn/task-workers/workerpool"
)

const (
	TaskCount         = 50
	TaskChannelBuffer = 10
	MaxTaskHandlers   = 10
	MaxTaskSorters    = 10
)

type taskUnknownError task.Task

func (e *taskUnknownError) Error() string {
	return fmt.Sprintf("task id = %v unknown error: %v", e.ID, e.Err)
}

type taskHandleError task.Task

func (e *taskHandleError) Error() string {
	return fmt.Sprintf("task id = %v handle error", e.ID)
}

type taskHandler struct {
	inCh  <-chan *task.Task
	outCh chan *task.Task
}

func newTaskHandler(inCh <-chan *task.Task) *taskHandler {
	return &taskHandler{
		inCh:  inCh,
		outCh: make(chan *task.Task, TaskChannelBuffer),
	}
}

func (th *taskHandler) OutCh() <-chan *task.Task { return th.outCh }

func (th *taskHandler) Do() {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for task := range th.inCh {
		// Condition for the task successful completion
		if rnd.Int()%2 != 0 && task.Err == nil {
			task.Done = true
		}
		task.FinishTime = time.Now()

		time.Sleep(time.Millisecond * 150)

		th.outCh <- task
	}
}

func (th *taskHandler) Finish() { close(th.outCh) }

var _ workerpool.Worker = (*taskHandler)(nil)

type taskSorter struct {
	inCh    <-chan *task.Task
	outCh   chan *task.Task
	errorCh chan error
}

func newTaskSorter(inCh <-chan *task.Task) *taskSorter {
	return &taskSorter{
		inCh:    inCh,
		outCh:   make(chan *task.Task, TaskChannelBuffer),
		errorCh: make(chan error, TaskChannelBuffer),
	}
}

func (ts *taskSorter) OutCh() <-chan *task.Task { return ts.outCh }

func (ts *taskSorter) ErrorCh() <-chan error { return ts.errorCh }

func (ts *taskSorter) Do() {
	for task := range ts.inCh {
		switch {
		case task.Err != nil:
			ts.errorCh <- (*taskUnknownError)(task)
		case task.Done:
			ts.outCh <- task
		case !task.Done:
			ts.errorCh <- (*taskHandleError)(task)
		}
	}
}

func (ts *taskSorter) Finish() {
	close(ts.outCh)
	close(ts.errorCh)
}

var _ workerpool.Worker = (*taskSorter)(nil)

func main() {
	tasksCh := task.Make(TaskCount)

	taskHandler := newTaskHandler(tasksCh)
	workerpool.New(taskHandler, MaxTaskHandlers).Run()

	taskSorter := newTaskSorter(taskHandler.OutCh())
	workerpool.New(taskSorter, MaxTaskSorters).Run()

	var doneTasks []*task.Task
	var errorTasks []error
	taskReceiversWg := &sync.WaitGroup{}

	taskReceiversWg.Add(1)
	go func() {
		defer taskReceiversWg.Done()

		for t := range taskSorter.OutCh() {
			doneTasks = append(doneTasks, t)
		}
	}()

	taskReceiversWg.Add(1)
	go func() {
		defer taskReceiversWg.Done()

		for err := range taskSorter.ErrorCh() {
			errorTasks = append(errorTasks, err)
		}
	}()

	taskReceiversWg.Wait()

	fmt.Println("Done tasks:")
	for _, task := range doneTasks {
		fmt.Println(task)
	}

	fmt.Println("Errors:")
	for _, err := range errorTasks {
		fmt.Println(err)
	}
}
