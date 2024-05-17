package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"sedonn/task-workers/workerpool"
)

const (
	TaskCount         uint = 10
	TaskChannelBuffer uint = 10
	MaxTaskHandlers   uint = 4
	MaxTaskSorters    uint = 4
)

var ErrUnknown = errors.New("some error occurred")

type Task struct {
	id                     uint
	done                   bool
	createTime, finishTime time.Time
	err                    error
}

func (t *Task) String() string {
	return fmt.Sprintf("Task\nid: %v\ncreated: %s\nfinished: %s\ndone: %v\n",
		t.id,
		t.createTime.Format(time.RFC3339Nano), t.finishTime.Format(time.RFC3339Nano),
		t.done)
}

type TaskUnknownError Task

func (e *TaskUnknownError) Error() string {
	return fmt.Sprintf("task id = %v unknown error: %v", e.id, e.err)
}

type TaskHandleError Task

func (e *TaskHandleError) Error() string {
	return fmt.Sprintf("task id = %v handle error", e.id)
}

func makeTasks() <-chan *Task {
	tasksCh := make(chan *Task, TaskChannelBuffer)

	go func() {
		for id := range TaskCount {
			var err error
			// Condition for generating a tasks with error
			if id%2 == 0 {
				err = ErrUnknown
			}

			tasksCh <- &Task{
				id:         uint(time.Now().Unix()) + id,
				createTime: time.Now(),
				err:        err,
			}
		}
		close(tasksCh)
	}()

	return tasksCh
}

func handleTask(inCh <-chan *Task, outCh chan<- *Task) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for task := range inCh {
		// Condition for the task successful completion
		if rnd.Int()%2 != 0 && task.err == nil {
			task.done = true
		}
		task.finishTime = time.Now()

		time.Sleep(time.Millisecond * 150)

		outCh <- task
	}
}

func sortTask(inCh <-chan *Task, doneOutCh chan<- *Task, undoneOutCh chan<- error) {
	for task := range inCh {
		switch {
		case task.err != nil:
			undoneOutCh <- (*TaskUnknownError)(task)
		case task.done:
			doneOutCh <- task
		case !task.done:
			undoneOutCh <- (*TaskHandleError)(task)
		}
	}
}

func main() {
	tasksCh := makeTasks()
	handledTasksCh := make(chan *Task, TaskChannelBuffer)
	doneTasksCh := make(chan *Task, TaskChannelBuffer)
	undoneTasksCh := make(chan error, TaskChannelBuffer)

	taskHandlers := workerpool.New(MaxTaskHandlers)
	taskHandlers.SetWork(func() { handleTask(tasksCh, handledTasksCh) })
	taskHandlers.SetFinishCallback(func() { close(handledTasksCh) })

	if err := taskHandlers.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	taskSorters := workerpool.New(MaxTaskSorters)
	taskSorters.SetWork(func() { sortTask(handledTasksCh, doneTasksCh, undoneTasksCh) })
	taskSorters.SetFinishCallback(func() {
		close(doneTasksCh)
		close(undoneTasksCh)
	})

	if err := taskSorters.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var doneTasks []*Task
	var undoneTasks []error
	taskReceiversWg := &sync.WaitGroup{}

	taskReceiversWg.Add(1)
	go func() {
		defer taskReceiversWg.Done()

		for t := range doneTasksCh {
			doneTasks = append(doneTasks, t)
		}
	}()

	taskReceiversWg.Add(1)
	go func() {
		defer taskReceiversWg.Done()

		for err := range undoneTasksCh {
			undoneTasks = append(undoneTasks, err)
		}
	}()

	taskReceiversWg.Wait()

	println("Done tasks:")
	for _, task := range doneTasks {
		fmt.Println(task)
	}

	println("Errors:")
	for _, err := range undoneTasks {
		fmt.Println(err)
	}
}
