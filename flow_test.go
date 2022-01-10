package flow

import (
	"testing"
)

// General usage testing
func TestFlow(t *testing.T) {
	t.Logf("performing test flow")

	task1 := NewTask(func(task Task) {
		t.Logf("task 1 ran")
	})

	t.Logf("task1 with id %s", task1.Id())

	task2 := NewTask(func(task Task) {
		t.Logf("task 2 ran")
	})
	task3 := NewTask(func(task Task) {
		t.Logf("task 3 ran")
	})
	task4 := NewTask(func(task Task) {
		t.Logf("task 4 ran")
	})
	task5 := NewTask(func(task Task) {
		t.Logf("task 5 ran")
	})
	task6 := NewTask(func(task Task) {
		t.Logf("task 6 ran")
	})

	t.Logf("staging required tasks")
	// setup some required tasks
	if err := task1.RequiresTask(task2); err != nil {
		t.Fatalf("failed to add required task to task2: %s", err)
	}
	if err := task2.RequiresTask(task4); err != nil {
		t.Fatalf("failed to add required task to task4: %s", err)
	}
	if err := task2.RequiresTask(task3); err != nil {
		t.Fatalf("failed to add required task to task3: %s", err)
	}
	if err := task3.RequiresTask(task5); err != nil {
		t.Fatalf("failed to add required task to task5: %s", err)
	}
	if err := task4.RequiresTask(task6); err != nil {
		t.Fatalf("failed to add required task to task6: %s", err)
	}
	if err := task5.RequiresTask(task6); err != nil {
		t.Fatalf("failed to add required task to task6: %s", err)
	}

	t.Logf("creating new queue")
	taskQueue := NewQueue()

	t.Logf("adding tasks to queue")
	taskQueue.AddTask(task1)
	taskQueue.AddTask(task2)
	taskQueue.AddTask(task3)
	taskQueue.AddTask(task4)
	taskQueue.AddTask(task5)
	taskQueue.AddTask(task6)

	t.Logf("starting task queue and waiting 5 seconds")
	go taskQueue.Start()

	t.Logf("waiting for queue to empty")
	taskQueue.Waiter()

	t.Logf("stopping task queue")
	taskQueue.Stop()
}
