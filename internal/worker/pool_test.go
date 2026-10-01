package worker

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestPoolRunsJobs(t *testing.T) {
	p := NewPool(2, 8)
	var n int32
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		if !p.Submit(func() { atomic.AddInt32(&n, 1); wg.Done() }) {
			t.Fatal("soumission refusée")
		}
	}
	wg.Wait()
	p.Stop()
	if n != 5 {
		t.Errorf("5 jobs attendus, %d exécutés", n)
	}
}

func TestSubmitAfterStopDoesNotPanic(t *testing.T) {
	p := NewPool(1, 1)
	p.Stop()
	p.Stop() // idempotent
	if p.Submit(func() {}) {
		t.Error("un pool arrêté ne doit rien accepter")
	}
}

func TestPanickingJobDoesNotKillWorker(t *testing.T) {
	p := NewPool(1, 4)
	p.Submit(func() { panic("boom") })
	done := make(chan struct{})
	p.Submit(func() { close(done) })
	<-done // le job suivant s'exécute : le worker a survécu
	p.Stop()
}

func TestFullQueueRejects(t *testing.T) {
	p := NewPool(1, 1)
	block := make(chan struct{})
	started := make(chan struct{})
	p.Submit(func() { close(started); <-block })
	<-started
	p.Submit(func() {}) // remplit la file (taille 1)
	if p.Submit(func() {}) {
		t.Error("la file pleine doit refuser")
	}
	close(block)
	p.Stop()
}
