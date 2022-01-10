package flow

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

func NewTask(taskFn TaskFunc) Task {
	ctx, cancel := context.WithCancel(context.Background())
	t := &task{
		ctx:         ctx,
		ctxCancelFn: cancel,
		taskFn:      taskFn,
	}
	return t
}

type task struct {
	sync.Mutex

	id                 string
	idLock             sync.Mutex
	ctx                context.Context
	ctxCancelFn        context.CancelFunc
	taskFn             TaskFunc
	state              TaskState
	requiredTasks      TaskList
	requiredConditions []func() bool
	runApproved        bool
	errorsLock         sync.Mutex
	errors             []error
}

func (t *task) Id() string {
	t.idLock.Lock()
	defer t.idLock.Unlock()

	if t.id == "" {
		t.id = uuid.NewString()
	}
	return t.id
}

func (t *task) RequiresTask(x Task) error {
	t.Lock()
	defer t.Unlock()

	// adding requirements on queued tasks is not allowed
	if t.state != STATE_NOTSET {
		return ErrRequirementOnQueuedTask
	}

	// recursively path tasks of tasks.GetRequiredTasks() for inf loop
	if err := t.recursionSeeker(TaskList{}, x); err != nil {
		return err
	}

	// add required task
	t.requiredTasks = append(t.requiredTasks, x)

	return nil
}

func (t *task) RequiresCondition(fn func() bool) error {
	t.Lock()
	defer t.Unlock()

	// adding requirements on queued tasks is not allowed
	if t.state != STATE_NOTSET {
		return ErrRequirementOnQueuedTask
	}

	// add condition
	t.requiredConditions = append(t.requiredConditions, fn)

	return nil
}

func (t *task) CanRun() bool {
	t.Lock()
	defer t.Unlock()

	// task state must be pending to permit re-running
	if t.state != STATE_PENDING {
		return false
	}

	// performing conditional is not needed after run
	if t.runApproved {
		return true
	}

	// perform all conditional checks, false means can't run
	for i := range t.requiredConditions {
		if !t.requiredConditions[i]() {
			return false
		}
	}

	// perform all conditional checks, false means can't run
	for i := range t.requiredTasks {
		REQUIRED_TASK_STATE := t.requiredTasks[i].GetState()
		if REQUIRED_TASK_STATE != STATE_SUCCESS && REQUIRED_TASK_STATE != STATE_COMPLETE {
			return false
		}
	}

	// document approval, performing conditional checks is not needed anymore
	t.runApproved = true

	return true
}

//
func (t *task) Pending() error {
	t.Lock()
	defer t.Unlock()

	if t.state != STATE_NOTSET {
		return t.Errorf("tasks that are not in a queue can not have state set to pending")
	}

	t.state = STATE_PENDING
	return nil
}

//
func (t *task) Run() {
	// only continue when run conditions are met
	if !t.CanRun() {
		return
	}

	// update state to running
	t.Lock()
	t.state = STATE_RUNNING
	t.Unlock()

	// do the task
	t.taskFn(t)

	// the default state for a task that is still RUNNING is COMPLETE
	t.Lock()
	if t.state == STATE_RUNNING {
		t.state = STATE_COMPLETE
	}
	t.Unlock()
}

func (t *task) Cancel() error {
	t.Lock()
	defer t.Unlock()

	if t.state != STATE_PENDING {
		return t.Errorf("task state must be pending to be canceled")
	}

	t.state = STATE_CANCELED
	return nil
}

func (t *task) Failed() error {
	t.Lock()
	defer t.Unlock()

	if t.state != STATE_RUNNING {
		return t.Errorf("task state must be running to be set to failed")
	}

	t.state = STATE_FAILED
	return nil
}

func (t *task) Success() error {
	t.Lock()
	defer t.Unlock()

	if t.state != STATE_RUNNING {
		return t.Errorf("task state must be running to be set to success")
	}

	t.state = STATE_SUCCESS
	return nil
}

func (t *task) GetState() TaskState {
	t.Lock()
	defer t.Unlock()
	return t.state
}

func (t *task) GetRequiredTasks() TaskList {
	t.Lock()
	defer t.Unlock()

	// return copy of slice with new element pointers
	return append(TaskList{}, t.requiredTasks...)
}

func (t *task) Errorf(s string, a ...interface{}) error {
	t.errorsLock.Lock()
	defer t.errorsLock.Unlock()
	err := fmt.Errorf(s, a...)
	t.errors = append(t.errors, err)
	return err
}

func (t *task) GetErrors() []error {
	t.errorsLock.Lock()
	defer t.errorsLock.Unlock()
	// return copy of slice with new element pointers
	return append([]error{}, t.errors...)
}

func (t *task) recursionSeeker(ancestors TaskList, branch Task) error {
	if t.Id() == branch.Id() {
		return fmt.Errorf("dependency cycle would prevent task from running due to loop on self")
	}

	for i := range ancestors {
		if branch.Id() == ancestors[i].Id() {
			return fmt.Errorf("dependency cycle would prevent task from running due to branch in tree")
		}
	}

	ancestors = append(ancestors, branch)

	branches := branch.GetRequiredTasks()
	for i := range branches {
		if err := t.recursionSeeker(ancestors, branches[i]); err != nil {
			return err
		}
	}

	return nil
}
