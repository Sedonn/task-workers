package main

import (
	"fmt"
	"sync"

	"sedonn/task-workers/task"
	"sedonn/task-workers/worker"
	"sedonn/task-workers/workerpool"
)

const (
	TaskCount       = 50
	MaxTaskHandlers = 10
	MaxTaskSorters  = 10
)

func main() {
	tasksCh := task.Make(TaskCount)

	taskHandler := worker.NewTaskHandler(tasksCh)
	workerpool.New(taskHandler, MaxTaskHandlers).Run()

	taskSorter := worker.NewTaskSorter(taskHandler.OutCh())
	workerpool.New(taskSorter, MaxTaskSorters).Run()

	var doneTasks []*task.Task
	var errorTasks []error
	taskReceiversWg := &sync.WaitGroup{}

	taskReceiversWg.Add(2)
	go func() {
		defer taskReceiversWg.Done()

		for t := range taskSorter.OutCh() {
			doneTasks = append(doneTasks, t)
		}
	}()
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
