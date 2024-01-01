package backends

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/acaloiaro/neoq"
	"github.com/acaloiaro/neoq/handler"
	"github.com/acaloiaro/neoq/jobs"
)

const proceedKey = "proceed"
const foundKey = "found"
const queue1 = "queue1"

// TestOverrideFingerprint provides a test case that works with all the backends that
// verifies that overriding jobs with a new fingerprint works
func TestOverrideFingerprint(t *testing.T, ctx context.Context, nq neoq.Neoq) error {
	fingerprint1 := "fingerprint1" + time.Now().String()
	fingerprint2 := "fingerprint2" + time.Now().String()

	var err error
	jobsToDo := 3
	jobsProcessed := 0

	fingerPrints := make(map[string]int)

	done1 := make(chan bool)
	proceed := make(chan bool)

	h1 := handler.New(queue1, func(ctx context.Context) (err error) {
		var j *jobs.Job
		j, err = jobs.FromContext(ctx)
		if err != nil || j == nil {
			t.Errorf("Unexpected error: %s", err)
			return err
		}
		message := j.Payload["message"].(string)

		if j.Payload[foundKey] == false {
			t.Errorf("Found job that should not be found: %s", message)
		}
		if _, ok := j.Payload[proceedKey]; ok {
			proceed <- true
		}

		fingerPrints[j.Fingerprint]++
		jobsProcessed++
		t.Logf("Handled job %d with fingerprint %s and ID %d Payload: %s", jobsProcessed, j.Fingerprint, j.ID, message)
		if jobsToDo == jobsProcessed {
			done1 <- true
		}
		return
	})

	if err := nq.Start(ctx, h1); err != nil {
		t.Fatal(err)
	}

	go func() {
		_, err = nq.Enqueue(ctx, &jobs.Job{
			Queue: queue1,
			Payload: map[string]any{"message": "(1) first queued item we'll wait until this is processed",
				proceedKey: true, foundKey: true},
			RunAfter:    time.Now().Add(time.Millisecond),
			Fingerprint: fingerprint1,
		})
		if err != nil {
			err = fmt.Errorf("job was not enqueued.%w", err)
			t.Error(err)
		}

		<-proceed

		_, err = nq.Enqueue(ctx, &jobs.Job{
			Queue: queue1,
			Payload: map[string]any{
				"message": "first queued item - should be overwritten", foundKey: false},
			RunAfter:    time.Now().Add(5 * time.Second),
			Fingerprint: fingerprint1,
		})
		if err != nil {
			err = fmt.Errorf("job was not enqueued.%w", err)
			t.Error(err)
		}
		_, err = nq.Enqueue(ctx, &jobs.Job{
			Queue:       queue1,
			Payload:     map[string]any{"message": "should not be queued", foundKey: false},
			RunAfter:    time.Now().Add(time.Second),
			Fingerprint: fingerprint1,
		})
		if !errors.Is(err, jobs.ErrJobFingerprintConflict) {
			t.Errorf("Should have returned a [%v] but returned [%v]", jobs.ErrJobFingerprintConflict, err)
		}
		_, err = nq.Enqueue(ctx, &jobs.Job{
			Queue: queue1,
			Payload: map[string]any{"message": "(2) the item that overwrites may be found",
				proceedKey: true, foundKey: true},
			RunAfter:    time.Now().Add(time.Second),
			Fingerprint: fingerprint1,
		}, neoq.WithOverrideMatchingFingerprint())
		if err != nil {
			t.Errorf("Should have returned nil but returned %v", err)
		}
		<-proceed
		_, err = nq.Enqueue(ctx, &jobs.Job{
			Queue:       queue1,
			Payload:     map[string]interface{}{"message": "(3) the new item", proceedKey: true, foundKey: true},
			RunAfter:    time.Now().Add(time.Second),
			Fingerprint: fingerprint2,
		})
		if err != nil {
			err = fmt.Errorf("job was not enqueued.%w", err)
			t.Error(err)
		}
		<-proceed
	}()

	timeoutTimer := time.After(50 * time.Second)
results_loop:
	for {
		select {
		case <-timeoutTimer:
			err = jobs.ErrJobTimeout
			break results_loop
		case <-done1:
			break results_loop
		}
	}
	if err != nil {
		t.Error(err)
	}
	if jobsProcessed != jobsToDo {
		t.Errorf("handler should have handled %d jobs, but handled %d. %v", jobsToDo, jobsProcessed, fingerPrints)
	}
	return err
}
