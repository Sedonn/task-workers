package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	TaskCount         = 20
	TaskChannelBuffer = 10
	MaxTaskHandlers   = 4
	MaxTaskSorters    = 4
)

type Task struct {
	id         int
	createTime time.Time
	finishTime time.Time
	err        string
	result     *TaskResult
}

func (t *Task) String() string {
	return fmt.Sprintf("Task\nid: %v\ncreateTime: %s\nfinishTime: %s\ndone: %v\nmessage: %s\n",
		t.id, t.createTime.Format(time.RFC3339), t.finishTime.Format(time.RFC3339Nano), t.result.done, t.result.message)
}

type TaskResult struct {
	done    bool
	message []byte
}

type TaskUnknownError Task

func (e TaskUnknownError) Error() string {
	task := Task(e)
	return fmt.Sprintf("Task unknown error!\nid: %v\nerror: %v\n", task.id, task.err)
}

type TaskHandleError Task

func (e TaskHandleError) Error() string {
	task := Task(e)
	return fmt.Sprintf("Task process error!\nid: %d\ntime: %s\nerror %s\n", task.id, task.createTime.Format(time.RFC3339), task.err)
}

func initTaskGenerator() <-chan *Task {
	tasksCh := make(chan *Task, TaskChannelBuffer)

	go func() {
		for count := 0; count < TaskCount; count++ {
			err := ""
			if time.Now().Nanosecond()%2 > 0 {
				err = "Some error occurred"
			}
			tasksCh <- &Task{id: int(time.Now().Unix()) + count, createTime: time.Now(), err: err}
		}
		close(tasksCh)
	}()

	return tasksCh
}

func handleTask(wg *sync.WaitGroup, inCh <-chan *Task, outCh chan<- *Task) {
	defer wg.Done()

	for task := range inCh {
		if task.createTime.After(time.Now().Add(-20*time.Second)) || task.err != "" {
			task.result = &TaskResult{true, []byte("Task has been completed!")}
		} else {
			task.result = &TaskResult{false, []byte("Something went wrong!")}
		}
		task.finishTime = time.Now()

		time.Sleep(time.Millisecond * 150)

		outCh <- task
	}
}

func sortTask(wg *sync.WaitGroup, inCh <-chan *Task, doneOutCh chan<- *Task, undoneOutCh chan<- error) {
	defer wg.Done()

	for task := range inCh {
		switch {
		case task.err != "":
			undoneOutCh <- TaskUnknownError(*task)
		case task.result.done:
			doneOutCh <- task
		case !task.result.done:
			undoneOutCh <- TaskHandleError(*task)
		}
	}
}

func main() {
	tasksCh := initTaskGenerator()
	processedTasksCh := make(chan *Task, TaskChannelBuffer)
	doneTasksCh := make(chan *Task)
	undoneTasksCh := make(chan error)

	taskHandlersWg := &sync.WaitGroup{}
	taskSortersWg := &sync.WaitGroup{}

	go func() {
		for workerCount := 0; workerCount < MaxTaskHandlers; workerCount++ {
			taskHandlersWg.Add(1)
			go handleTask(taskHandlersWg, tasksCh, processedTasksCh)
		}
		taskHandlersWg.Wait()
		close(processedTasksCh)
	}()

	go func() {
		for workerCount := 0; workerCount < MaxTaskSorters; workerCount++ {
			taskSortersWg.Add(1)
			go sortTask(taskSortersWg, processedTasksCh, doneTasksCh, undoneTasksCh)
		}
		taskSortersWg.Wait()
		close(doneTasksCh)
		close(undoneTasksCh)
	}()

	doneTasks := map[int]*Task{}
	undoneTasks := []error{}
	taskReceiversWg := &sync.WaitGroup{}

	taskReceiversWg.Add(1)
	go func() {
		defer taskReceiversWg.Done()

		for task := range doneTasksCh {
			doneTasks[task.id] = task
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
