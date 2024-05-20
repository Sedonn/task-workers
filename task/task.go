package task

import (
	"errors"
	"fmt"
	"time"
)

var ErrUnknown = errors.New("some error occurred")

type Task struct {
	ID                     int
	Done                   bool
	CreateTime, FinishTime time.Time
	Err                    error
}

func (t *Task) String() string {
	return fmt.Sprintf("Task\nid: %v\ncreated: %s\nfinished: %s\ndone: %v\n",
		t.ID,
		t.CreateTime.Format(time.RFC3339Nano), t.FinishTime.Format(time.RFC3339Nano),
		t.Done)
}

func Make(count int) <-chan *Task {
	tasksCh := make(chan *Task, 10)

	go func() {
		for id := range count {
			var err error
			// Condition for generating a tasks with error
			if id%2 == 0 {
				err = ErrUnknown
			}

			tasksCh <- &Task{
				ID:         int(time.Now().Unix()) + id,
				CreateTime: time.Now(),
				Err:        err,
			}
		}
		close(tasksCh)
	}()

	return tasksCh
}
