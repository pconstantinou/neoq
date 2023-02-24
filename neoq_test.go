package neoq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"
)

var errTrigger = errors.New("triggerering a log error")
var errPeriodicTimeout = errors.New("timed out waiting for periodic job")

type testLogger struct {
	l    *log.Logger
	done chan bool
}

func (h testLogger) Info(m string, args ...any) {
	h.l.Println(m)
	h.done <- true
}
func (h testLogger) Debug(m string, args ...any) {
	h.l.Println(m)
	h.done <- true
}
func (h testLogger) Error(m string, err error, args ...any) {
	h.l.Println(m, err)
	h.done <- true
}

func TestWorkerListenConn(t *testing.T) {
	const queue = "testing"
	timeout := false
	numJobs := 1
	doneCnt := 0
	var done = make(chan bool, numJobs)

	ctx := context.TODO()
	backend, err := NewMemBackend()
	if err != nil {
		t.Fatal(err)
	}

	nq, err := New(ctx, WithBackend(backend))
	if err != nil {
		t.Fatal(err)
	}
	defer nq.Shutdown(ctx)

	handler := NewHandler(func(ctx context.Context) (err error) {
		done <- true
		return
	})
	handler.WithOptions(
		HandlerDeadline(500*time.Millisecond),
		HandlerConcurrency(1),
	)

	if err != nil {
		t.Error(err)
	}

	// Listen for jobs on the queue
	nq.Listen(ctx, queue, handler)

	// allow time for listener to start
	time.Sleep(5 * time.Millisecond)

	for i := 0; i < numJobs; i++ {
		jid, err := nq.Enqueue(ctx, Job{
			Queue: queue,
			Payload: map[string]interface{}{
				"message": fmt.Sprintf("hello world: %d", i),
			},
		})
		if err != nil || jid == -1 {
			t.Fatal("job was not enqueued. either it was duplicate or this error caused it:", err)
		}
	}

	for {
		select {
		case <-time.After(5 * time.Second):
			timeout = true
			err = errors.New("timed out waiting for job")
		case <-done:
			doneCnt++
		}

		if doneCnt >= numJobs {
			break
		}

		if timeout {
			break
		}
	}

	if timeout {
		t.Error(err)
	}
}

func TestWorkerListenCron(t *testing.T) {
	const cron = "* * * * * *"
	ctx := context.TODO()
	nq, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer nq.Shutdown(ctx)

	var done = make(chan bool)
	handler := NewHandler(func(ctx context.Context) (err error) {
		done <- true
		return
	})

	handler.WithOptions(
		HandlerDeadline(500*time.Millisecond),
		HandlerConcurrency(1),
	)

	if err != nil {
		t.Error(err)
	}

	nq.ListenCron(ctx, cron, handler)

	// allow time for listener to start
	time.Sleep(5 * time.Millisecond)

	select {
	case <-time.After(1 * time.Second):
		err = errPeriodicTimeout
	case <-done:
	}

	if err != nil {
		t.Error(err)
	}
}

func TestNeoqAddLogger(t *testing.T) {
	const queue = "testing"
	var done = make(chan bool)

	ctx := context.TODO()

	buf := &strings.Builder{}
	nq, err := New(ctx, WithLogger(testLogger{l: log.New(buf, "", 0), done: done}))
	if err != nil {
		t.Fatal(err)
	}
	defer nq.Shutdown(ctx)

	handler := NewHandler(func(ctx context.Context) (err error) {
		err = errTrigger
		return
	})
	if err != nil {
		t.Error(err)
	}

	// Listen for jobs on the queue
	err = nq.Listen(ctx, queue, handler)
	if err != nil {
		t.Error(err)
	}

	_, err = nq.Enqueue(ctx, Job{Queue: queue})
	if err != nil {
		t.Error(err)
	}

	<-done
	expectedLogMsg := "job failed job failed to process: triggerering a log error"
	actualLogMsg := strings.Trim(buf.String(), "\n")
	if actualLogMsg != expectedLogMsg {
		t.Error(fmt.Errorf("%s != %s", actualLogMsg, expectedLogMsg)) //nolint:all
	}
}
