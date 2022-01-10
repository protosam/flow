package flow

import "fmt"

const (
	// task is not queued or ran
	STATE_NOTSET = iota

	// task is in a queue but has not been ran
	STATE_PENDING

	// task has been canceled before running
	STATE_CANCELED

	// task is currently running
	STATE_RUNNING

	// task ran is completed with failure
	STATE_FAILED

	// task ran is completed with success
	STATE_SUCCESS

	// task ran is completed without explicit success or failure
	STATE_COMPLETE
)

var ErrRequirementOnQueuedTask = fmt.Errorf("cannot add requirements to a queued task")

type TaskState int

type TaskFunc func(Task)

type TaskList []Task

// workflow queue implementation
type Queue interface {
	GetTaskList() TaskList
	GetQueueSize() int
	GetRunning() int
	GetRunCapacity() int
	SetRunCapacity(int)
	AddTask(Task) error
	Start()
	Stop()
	Waiter()
}

// task implementation for workflow queues
type Task interface {
	// will set the task ID if it is not set and returns ID value
	Id() string

	// provides slice of task errors
	GetErrors() []error

	// adds error to task errors
	Errorf(string, ...interface{}) error

	// provides tasks that must be completed as run requirements
	GetRequiredTasks() TaskList

	// adds a task to be completed to run requirements
	RequiresTask(Task) error

	// adds a conditional check for run requirements
	RequiresCondition(func() bool) error

	// provides current task state
	GetState() TaskState

	// updates task state to pending, should only be set by queue
	Pending() error

	// updates task state to canceled
	Cancel() error
	// updates task state to failed
	Failed() error
	// updates task state to success
	Success() error

	// validates if run requirements are met
	CanRun() bool

	// runs task
	Run()
}
