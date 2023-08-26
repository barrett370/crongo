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
	ticker   *time.Ticker
	done     chan struct{}
}

type OptFn = func(*Scheduler)

func New(name string, task Tasker, interval time.Duration, optFns ...OptFn) *Scheduler {
	s := &Scheduler{
		name:     name,
		logger:   crongolog.NoopLogger{},
		op:       task,
		interval: interval,
		ticker:   time.NewTicker(interval),
		done:     make(chan struct{}),
	}
	for _, fn := range optFns {
		fn(s)
	}
	return s
}

func WithDefaultLogger(s *Scheduler) {
	s.logger = log.New(os.Stdout, fmt.Sprintf("[CRON: %s] ", s.name), log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
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
		case <-s.ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), s.interval)
			err := s.op.Run(ctx)
			if err != nil {
				s.logger.Printf("error while running work func. err: %v\n", err)
			}
			cancel()
		case <-s.done:
			s.logger.Println("received stop signal, stopping")
			close(s.done)
			return
		}
	}
}
