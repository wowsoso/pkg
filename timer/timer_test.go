package timer

import (
	"testing"
	"time"
)

func isChannelOpening(c <-chan struct{}) bool {
	select {
	case <-c:
		return false
	default:
		return true
	}
}

func TestTaskMethodDoneShouldBeOpening(t *testing.T) {
	if isChannelOpening(newTask(1).Done()) == false {
		t.Fatalf("channel should be opening")
	}
}

func TestTickMethodRegisterAppendTaskSucceed(t *testing.T) {
	tick := &Tick{tasks: make([]*Task, 0)}

	tick.register(newTask(1))

	if len(tick.tasks) != 1 {
		t.Fatalf("tick method register not append task")
	}
}

func TestTimerMethodShiftBehaviorCorrect(t *testing.T) {
	timer := NewTimer(1000, 59, 1*time.Millisecond)

	for i := 0; i < 1000; i++ {
		timer.t2[0].tasks = append(timer.t2[0].tasks, newTask(uint(i)))
	}

	for i := 0; i < 1000; i++ {
		timer.t2[58].tasks = append(timer.t2[58].tasks, newTask(uint(i)))
	}

	timer.shift()

	if len(timer.t2[0].tasks) != 0 {
		t.Fatalf("timer method shift t2[0].task should empty")
	}

	taskCount := 0
	for _, tick := range timer.t1 {
		taskCount = taskCount + len(tick.tasks)
	}

	if taskCount != 1000 {
		t.Fatalf("timer method shift t1 should 1000 tasks, got: %d", taskCount)
	}

	if len(timer.t2[57].tasks) != 1000 {
		t.Fatalf("timer method shift t2[57] should 1000 tasks, got: %d", len(timer.t2[57].tasks))
	}

	if len(timer.t2[58].tasks) != 0 {
		t.Fatalf("timer method shift t2[58] should empty, got: %d", len(timer.t2[58].tasks))
	}
}

func TestTimerMethodNextBehaviorCorrect(t *testing.T) {
	timer := NewTimer(1000, 59, 1*time.Millisecond)
	timer.pos = 999

	timer.next()

	if timer.pos != 0 {
		t.Fatalf("timer method next pos should be 0, got: %d", timer.pos)
	}

	timer.pos = 0
	timer.next()
	if timer.pos != 1 {
		t.Fatalf("timer method next pos should be 1, got: %d", timer.pos)
	}
}

func TestTimerMethodOnTickerBehaviorCorrect(t *testing.T) {
	timer := NewTimer(1000, 59, 1*time.Millisecond)

	tasks := make([]*Task, 0)

	for i := 0; i < 60; i++ {
		task := newTask(uint(i))
		tasks = append(tasks, task)
		timer.t1[0].tasks = append(timer.t1[0].tasks, task)
	}

	timer.onTicker()

	for _, task := range tasks {
		if isChannelOpening(task.Done()) == true {
			t.Fatalf("timer method onTicker should closed all last tick's tasks")
			break
		}
	}

	if len(timer.t1[0].tasks) != 0 {
		t.Fatalf("timer method onTicker t1[0].tasks should be empty, got: %d", len(timer.t1[0].tasks))
	}
}

func TestTimerMethodAddTask(t *testing.T) {
	timer := NewTimer(1000, 59, 1*time.Millisecond)

	task1 := newTask(uint(1))
	task2 := newTask(uint(999))
	task3 := newTask(uint(1000*60 - 1))
	task4 := newTask(uint(1000 * 60))

	timer.addTask(task1)
	timer.addTask(task2)
	timer.addTask(task3)
	timer.addTask(task4)

	if len(timer.t1[1].tasks) != 1 {
		t.Fatalf("timer method addTask t1[0].tasks size should be 1, got: %d", len(timer.t1[0].tasks))
	}

	if len(timer.t1[999].tasks) != 1 {
		t.Fatalf("timer method addTask t1[999].tasks size should be 1, got: %d", len(timer.t1[999].tasks))
	}

	if len(timer.t2[58].tasks) != 1 {
		t.Fatalf("timer method addTask t2[58].tasks size should be 1, got: %d", len(timer.t2[58].tasks))
	}

	if isChannelOpening(task4.done) == true {
		t.Fatalf("timer method addTask task4 channel should be closed")
	}
}

func TestTimerStartBehaviorCorrect(t *testing.T) {
	timer := NewTimer(1000, 59, 1*time.Millisecond)
	go timer.Start()

	task := timer.Task(uint(1))
	<-task.Done()

	timer.Stop()

	if isChannelOpening(timer.stopChan) == true {
		t.Fatalf("stopChan should be closed")
	}

}

func TestTimerMethodTaskBehaviorCorrect(t *testing.T) {
	timer := NewTimer(1000, 59, 1*time.Millisecond)

	timer.taskChan = make(chan *Task, 1)
	task := timer.Task(1000 * 60)
	if task != nil {
		t.Fatalf("timer method Task task should be nil")
	}

	timer.taskChan = make(chan *Task, 1)
	task = timer.Task(1000*60 - 1)
	if task != <-timer.taskChan {
		t.Fatalf("timer method Task task should be task")
	}
}
