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
}

// NewPool initialise le worker pool avec un nombre fixe de workers
func NewPool(workerCount int, queueSize int) *Pool {
	p := &Pool{
		jobQueue: make(chan Job, queueSize),
	}

	for i := 0; i < workerCount; i++ {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			for job := range p.jobQueue {
				log.Printf("[Worker %d] Début de traitement asynchrone...", workerID)
				job()
				log.Printf("[Worker %d] Traitement terminé.", workerID)
			}
		}(i + 1)
	}

	return p
}

// Submit ajoute un job dans la file sans bloquer l'appelant
func (p *Pool) Submit(job Job) bool {
	select {
	case p.jobQueue <- job:
		return true
	default:
		log.Println("[Worker Pool] File d'attente pleine, tâche rejetée.")
		return false
	}
}

// Stop arrête proprement le pool de workers
func (p *Pool) Stop() {
	close(p.jobQueue)
	p.wg.Wait()
}
