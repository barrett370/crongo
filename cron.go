package crongo

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	crongolog "github.com/barrett370/crongo/log"
)

type Scheduler struct {
	name string
	// no idea what the best approach is here, there are no good, standardised logger interfaces...
	logger   crongolog.Logger
	op       Tasker
	interval time.Duration
	c        <-chan time.Time
	done     chan struct{}
	errs     chan<- error
}

type OptFn = func(*Scheduler)

func New(name string, task Tasker, interval time.Duration, optFns ...OptFn) *Scheduler {
	s := &Scheduler{
		name:     name,
		logger:   crongolog.NoopLogger{},
		op:       task,
		interval: interval,
		// using Tick here to allow for easy mocking
		// WARNING: it leaks the underlying ticker
		// I do not see this being an issue as I intend applications of this lib to run "forever"
		c:    time.Tick(interval),
		done: make(chan struct{}),
	}
	for _, fn := range optFns {
		fn(s)
	}
	return s
}

func WithDefaultLogger(s *Scheduler) {
	s.logger = log.New(os.Stdout, fmt.Sprintf("[CRON: %s] ", s.name), log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
}

func WithErrorsOut(errs chan<- error) func(s *Scheduler) {
	return func(s *Scheduler) {
		s.errs = errs
	}
}

func WithMockTicker(c <-chan time.Time) func(*Scheduler) {
	return func(s *Scheduler) {
		s.c = c
	}
}

func (s *Scheduler) Start() {
	s.logger.Println("starting")

	go s.loop()
}

func (s *Scheduler) Stop() {
	s.logger.Println("stopping...")
	s.done <- struct{}{}
	<-s.done
}

func (s *Scheduler) loop() {
	for {
		select {
		case ts := <-s.c:
			ctx, cancel := context.WithTimeout(context.Background(), s.interval)
			s.logger.Printf("running task, ts: %v\n", ts)
			err := s.op.Run(ctx)
			if err != nil {
				s.logger.Printf("error while running work func. err: %v\n", err)
				if s.errs != nil {
					s.errs <- err
				}
			}
			cancel()
		case <-s.done:
			s.logger.Println("received stop signal, stopping")
			close(s.done)
			return
		}
	}
}
