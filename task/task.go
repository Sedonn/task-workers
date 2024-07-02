package task

import (
	"errors"
	"fmt"
	"time"
)

var ErrUnknown = errors.New("some error occurred")

// The Task contains data about task.
type Task struct {
	ID                    int
	Done                  bool
	CreatedAt, FinishedAt time.Time
	Err                   error
}

// SetDone completes task with certain state and sets a finish time.
func (t *Task) SetDone(done bool) {
	t.Done = done
	t.FinishedAt = time.Now()
}

func (t *Task) String() string {
	return fmt.Sprintf("Task\nid: %v\ncreated: %s\nfinished: %s\ndone: %v\n",
		t.ID,
		t.CreatedAt.Format(time.RFC3339Nano), t.FinishedAt.Format(time.RFC3339Nano),
		t.Done)
}

// Make creates channel with generated tasks.
func Make(count int) <-chan *Task {
	tasksCh := make(chan *Task, 1)

	go func() {
		for id := range count {
			var err error
			// Condition for generating a tasks with unknown error.
			if id%2 == 0 {
				err = ErrUnknown
			}

			tasksCh <- &Task{
				ID:        int(time.Now().Unix()) + id,
				CreatedAt: time.Now(),
				Err:       err,
			}
		}
		close(tasksCh)
	}()

	return tasksCh
}
