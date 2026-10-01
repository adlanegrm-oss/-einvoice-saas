package worker

import (
	"log"
	"sync"
)

// Job représente une tâche à exécuter en arrière-plan
type Job func()

// Pool gère le traitement asynchrone des tâches lourdes
type Pool struct {
	jobQueue chan Job
	wg       sync.WaitGroup
	mu       sync.RWMutex
	closed   bool
}

// NewPool initialise le worker pool avec un nombre fixe de workers
func NewPool(workerCount int, queueSize int) *Pool {
	p := &Pool{jobQueue: make(chan Job, queueSize)}

	for i := 0; i < workerCount; i++ {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			for job := range p.jobQueue {
				run(workerID, job)
			}
		}(i + 1)
	}
	return p
}

// run exécute un job en isolant les paniques : le worker reste vivant.
func run(workerID int, job Job) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Worker %d] panique dans un job : %v", workerID, r)
		}
	}()
	job()
}

// Submit ajoute un job dans la file sans bloquer l'appelant.
// Renvoie false si la file est pleine ou si le pool est arrêté.
func (p *Pool) Submit(job Job) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		return false
	}
	select {
	case p.jobQueue <- job:
		return true
	default:
		log.Println("[Worker Pool] File d'attente pleine, tâche rejetée.")
		return false
	}
}

// Stop arrête proprement le pool (idempotent) et attend la fin des jobs en cours.
func (p *Pool) Stop() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	close(p.jobQueue)
	p.mu.Unlock()
	p.wg.Wait()
}
