package flow

import (
	"fmt"
	"sync"
)

func NewQueue() Queue {
	q := &queue{}
	return q
}

type queue struct {
	sync.Mutex
	runCapacity int
	runCount    int
	runQueue    TaskList
	taskList    TaskList
	stopped     bool
}

func (q *queue) GetTaskList() TaskList {
	q.Lock()
	defer q.Unlock()

	// return copy of slice with new element pointers
	return append(TaskList{}, q.taskList...)
}

func (q *queue) GetQueueSize() int {
	q.Lock()
	defer q.Unlock()

	return len(q.runQueue)
}

func (q *queue) GetRunning() int {
	q.Lock()
	defer q.Unlock()

	return q.runCount
}

func (q *queue) GetRunCapacity() int {
	q.Lock()
	defer q.Unlock()

	return q.runCapacity
}

func (q *queue) SetRunCapacity(i int) {
	q.Lock()
	defer q.Unlock()

	if i >= 0 {
		q.runCapacity = i
	}
}

func (q *queue) AddTask(t Task) error {
	q.Lock()
	defer q.Unlock()

	// check if task is already in a queue
	if t.GetState() != STATE_NOTSET {
		return fmt.Errorf("can not add a task already in a queue")
	}

	// add task to runQueue
	q.runQueue = append(q.runQueue, t)

	// add task to the taskList
	q.taskList = append(q.taskList, t)

	// update task state to pending
	t.Pending()

	return nil
}

func (q *queue) Start() {
	q.runLoop()
}

func (q *queue) Stop() {
	q.Lock()
	defer q.Unlock()
	q.stopped = true
}

func (q *queue) isStop() bool {
	q.Lock()
	defer q.Unlock()
	return q.stopped
}

func (q *queue) runLoop() {
	// when q.stopped is true, this will halt runLoop
	if q.isStop() {
		return
	}

	// do next task
	q.nextTask()
	// brief yield to goroutines
	runtime_doSpin()
	// recur runLoop
	q.runLoop()
}

func (q *queue) nextTask() {
	q.Lock()
	defer q.Unlock()

	// check if there is capacity to run a task
	if q.runCapacity > 0 && q.runCount >= q.runCapacity {
		return
	}

	// find a task that can be ran
	for i := range q.runQueue {
		// can't run this task, move to next task
		if !q.runQueue[i].CanRun() {
			continue
		}

		// everything from here assumes a runnable task was found

		// increment runCount before running task in goroutine
		q.runCount++

		// run the task in a separate goroutine, where it will become detached
		// from the locking in the current scope
		go func(t Task) {
			// run the task
			t.Run()

			// decrement the runCount, locking is necessary
			q.decrementRunCount()

			// remove the task from the runQueue since it is complete
			q.removeTaskFromRunQueue(t)
		}(q.runQueue[i])

		// return because a task was found and ran
		// loop only queues 1 task per call
		return
	}
}

func (q *queue) decrementRunCount() {
	q.Lock()
	defer q.Unlock()
	q.runCount--
}

func (q *queue) removeTaskFromRunQueue(t Task) {
	q.Lock()
	defer q.Unlock()

	// find task in the queue and remove it
	for i := range q.runQueue {
		// compare task id's
		if q.runQueue[i].Id() == t.Id() {
			// rebuild runQueue without task
			q.runQueue = append(q.runQueue[:i], q.runQueue[i+1:]...)

			// end here, iterating beyond here is wasted compute
			return
		}
	}
}

func (q *queue) Waiter() {
	if q.GetQueueSize() <= 0 {
		return
	}
	runtime_doSpin()
	q.Waiter()
}
