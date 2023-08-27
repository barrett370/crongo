package crongo_test

import (
	"context"
	"sync"
	"testing"
	"time"

	undertest "github.com/barrett370/crongo"
	"github.com/stretchr/testify/require"
)

type mockTask struct {
	sync.Mutex
	called   int
	duration time.Duration
}

func (m *mockTask) Run(ctx context.Context) error {
	select {
	case <-time.After(m.duration):

		m.Lock()
		m.called++
		m.Unlock()

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func Test_Scheduler(t *testing.T) {

	testcases := []struct {
		name         string
		interval     time.Duration
		taskDuration time.Duration
		waitTime     time.Duration
		wantCalls    int
		wantErrs     int
	}{
		{
			name:      "scheduler starts + stops as expected",
			interval:  time.Millisecond,
			wantCalls: 1,
		},
		{
			name:         "context timeout is sent properly",
			interval:     time.Nanosecond,
			taskDuration: time.Second,
			waitTime:     time.Millisecond,
			wantCalls:    0,
			wantErrs:     1,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			errs := make(chan error, tt.wantErrs)
			task := &mockTask{
				duration: tt.taskDuration,
			}
			c := make(chan time.Time)
			s := undertest.New(tt.name, task, tt.interval, undertest.WithDefaultLogger, undertest.WithErrorsOut(errs), undertest.WithMockTicker(c))

			s.Start()
			c <- time.Now()
			time.Sleep(tt.waitTime)
			s.Stop()
			close(errs)
			var es []error
			for err := range errs {
				es = append(es, err)
			}
			require.Len(t, es, tt.wantErrs)
			task.Lock()
			require.Equal(t, tt.wantCalls, task.called)
			task.Unlock()
		})
	}

}
