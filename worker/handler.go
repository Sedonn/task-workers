package worker

import (
	"math/rand"
	"time"

	"sedonn/task-workers/task"
	"sedonn/task-workers/workerpool"
)

// A TaskHandler is worker which completes the tasks.
type TaskHandler struct {
	inCh  <-chan *task.Task
	outCh chan *task.Task
	rnd   *rand.Rand
}

var _ workerpool.Worker = (*TaskHandler)(nil)

// NewTaskHandler create a new task handle worker.
func NewTaskHandler(inCh <-chan *task.Task) *TaskHandler {
	return &TaskHandler{
		inCh:  inCh,
		outCh: make(chan *task.Task, 1),
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// OutCh returns channel with results of task handling.
func (th *TaskHandler) OutCh() <-chan *task.Task { return th.outCh }

// Do executes worker job in concurrent mode.
func (th *TaskHandler) Do() {
	for task := range th.inCh {
		// Condition for the task successful completion.
		done := th.rnd.Int()%2 != 0 && task.Err == nil

		task.SetDone(done)

		time.Sleep(time.Millisecond * 150)

		th.outCh <- task
	}
}

// Finish made cleanup after competition of a job.
func (th *TaskHandler) Finish() { close(th.outCh) }
