package util

import (
	"math/rand"
	"sync"
	"time"
)

type Pool struct {
	work chan func()
	sem  chan struct{}
	Wg   sync.WaitGroup
	rng  *rand.Rand
}

func New(size int) *Pool {
	return &Pool{
		work: make(chan func()),
		sem:  make(chan struct{}, size),
		Wg:   sync.WaitGroup{},
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *Pool) NewTask(task func()) {
	select {
	case p.work <- task:
	case p.sem <- struct{}{}:
		go p.worker(task)
	}
}

func (p *Pool) worker(task func()) {
	defer func() { <-p.sem }()
	for {
		p.Wg.Add(1)
		task()
		p.Wg.Done()
		time.Sleep(time.Millisecond * time.Duration(p.rng.Intn(801)+500))
		<-p.work
	}
}
