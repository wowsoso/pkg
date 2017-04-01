package timer

import (
	"time"
)

type Task struct {
	delay uint
	done  chan struct{}
}

func newTask(delay uint) *Task {
	return &Task{
		delay: delay,
		done:  make(chan struct{}),
	}
}

func (t *Task) Done() <-chan struct{} {
	return t.done
}

type Tick struct {
	tasks []*Task
}

func NewTick() *Tick {
	return &Tick{
		tasks: make([]*Task, 0),
	}
}

func (t *Tick) register(task *Task) {
	t.tasks = append(t.tasks, task)
}

type Timer struct {
	t1     []*Tick
	t2     []*Tick
	t1Size uint
	t2Size uint
	pos    uint

	ticker   *time.Ticker
	taskChan chan *Task
	stopChan chan struct{}
}

func NewTimer(t1Size, t2Size uint, tickInterval time.Duration) *Timer {
	t := &Timer{
		t1:       make([]*Tick, t1Size),
		t2:       make([]*Tick, t2Size),
		t1Size:   t1Size,
		t2Size:   t2Size,
		ticker:   time.NewTicker(tickInterval),
		taskChan: make(chan *Task),
		stopChan: make(chan struct{}),
	}

	for i := 0; i < int(t1Size); i++ {
		t.t1[i] = &Tick{
			tasks: make([]*Task, 0),
		}
	}

	for i := 0; i < int(t2Size); i++ {
		t.t2[i] = &Tick{
			tasks: make([]*Task, 0),
		}
	}

	return t
}

func (t *Timer) shift() {
	for _, task := range t.t2[0].tasks {
		t.t1[task.delay].register(task)
	}

	for i := 0; i < int(t.t2Size)-1; i++ {
		t.t2[i] = t.t2[i+1]
	}

	t.t2[t.t2Size-1] = &Tick{
		tasks: make([]*Task, 0),
	}
}

func (t *Timer) next() {
	if t.pos == t.t1Size-1 {
		t.pos = 0
		t.shift()
	} else {
		t.pos++
	}
}

func (t *Timer) onTicker() {
	pos := t.pos

	for _, item := range t.t1[pos].tasks {
		close(item.done)
	}

	t.t1[pos].tasks = make([]*Task, 0)

	t.next()
}

func (t *Timer) addTask(task *Task) {
	pos := t.pos

	if task.delay < t.t1Size {
		t.t1[(task.delay+pos)%t.t1Size].register(task)
		task.delay = (task.delay + pos) % t.t1Size
		return
	}

	if task.delay < t.t1Size*(t.t2Size+1) {
		idx := ((task.delay + pos) / (t.t1Size + 1)) - 1
		task.delay = (task.delay + pos) % t.t1Size
		t.t2[idx].tasks = append(t.t2[idx].tasks, task)
		return
	}

	// some error occur
	close(task.done)
}

func (t *Timer) Start() {
	for {
		select {
		case <-t.ticker.C:
			t.onTicker()
		case task := <-t.taskChan:
			t.addTask(task)
		case <-t.stopChan:
			return
		}
	}
}

func (t *Timer) Stop() {
	close(t.stopChan)
}

func (t *Timer) Task(delay uint) *Task {
	if delay > t.t1Size*(t.t2Size+1)-1 {
		return nil
	}

	task := newTask(delay)

	t.taskChan <- task

	return task
}
