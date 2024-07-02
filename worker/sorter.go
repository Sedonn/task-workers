package worker

import (
	"fmt"

	"sedonn/task-workers/task"
	"sedonn/task-workers/workerpool"
)

type taskUnknownError task.Task

func (e *taskUnknownError) Error() string {
	return fmt.Sprintf("task id = %v unknown error: %v", e.ID, e.Err)
}

type taskHandleError task.Task

func (e *taskHandleError) Error() string {
	return fmt.Sprintf("task id = %v handle error", e.ID)
}

// The TaskSorter is the worker that sorts a tasks between channels.
type TaskSorter struct {
	inCh    <-chan *task.Task
	outCh   chan *task.Task
	errorCh chan error
}

var _ workerpool.Worker = (*TaskSorter)(nil)

// NewTaskSorter create a new task sort worker.
func NewTaskSorter(inCh <-chan *task.Task) *TaskSorter {
	return &TaskSorter{
		inCh:    inCh,
		outCh:   make(chan *task.Task, 1),
		errorCh: make(chan error, 1),
	}
}

// OutCh returns channel of successfully completed tasks.
func (ts *TaskSorter) OutCh() <-chan *task.Task { return ts.outCh }

// ErrorCh returns channel of tasks with errors.
func (ts *TaskSorter) ErrorCh() <-chan error { return ts.errorCh }

// Do starts worker job.
func (ts *TaskSorter) Do() {
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

// Finish made cleanup after competition of a job.
func (ts *TaskSorter) Finish() {
	close(ts.outCh)
	close(ts.errorCh)
}
