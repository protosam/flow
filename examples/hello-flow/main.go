package main

import (
	"log"

	"github.com/protosam/flow"
)

func main() {
	task1 := flow.NewTask(func(task flow.Task) {
		log.Printf("Hello from Task 1")
	})

	task2 := flow.NewTask(func(task flow.Task) {
		log.Printf("Hello from Task 2")
	})

	// task 1 requires task 2 to run
	if err := task1.RequiresTask(task2); err != nil {
		log.Fatalf("failed to add required task to task2: %s", err)
	}

	// create a queue to add tasks to
	taskQueue := flow.NewQueue()

	// add tasks to queue
	taskQueue.AddTask(task1)
	taskQueue.AddTask(task2)

	// start running any queued tasks
	go taskQueue.Start()

	// wait until the task queue is empty
	taskQueue.Waiter()

	// stop the task queue even loop
	taskQueue.Stop()
}
