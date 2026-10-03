package jobs

import (
	"sync"
	"testing"
	"time"
)

func TestTracker_LifecycleAndConcurrency(t *testing.T) {
	tracker := NewTracker()

	job1 := &Job{
		ID:        "job-1",
		Type:      "CLEARANCE_SUBMISSION",
		InvoiceID: "inv-001",
		Status:    StatusQueued,
		StartedAt: time.Now(),
	}

	job2 := &Job{
		ID:        "job-2",
		Type:      "AI_OCR",
		InvoiceID: "inv-002",
		Status:    StatusProcessing,
		StartedAt: time.Now(),
	}

	job3 := &Job{
		ID:        "job-3",
		Type:      "CLEARANCE_SUBMISSION",
		InvoiceID: "inv-003",
		Status:    StatusCompleted,
		StartedAt: time.Now(),
	}

	tracker.TrackJob(job1)
	tracker.TrackJob(job2)
	tracker.TrackJob(job3)

	t.Run("Filtrage des tâches actives", func(t *testing.T) {
		active := tracker.GetActiveJobs()
		if len(active) != 2 {
			t.Fatalf("attendu 2 jobs actifs, obtenu %d", len(active))
		}
		for _, j := range active {
			if j.Status != StatusQueued && j.Status != StatusProcessing {
				t.Errorf("job non actif trouvé: %s avec statut %s", j.ID, j.Status)
			}
		}
	})

	t.Run("Accès concurrentiel sécurisé (Data Race check)", func(t *testing.T) {
		var wg sync.WaitGroup
		iterations := 100

		wg.Add(iterations * 2)
		for i := 0; i < iterations; i++ {
			go func(id int) {
				defer wg.Done()
				tracker.TrackJob(&Job{
					ID:     "concurrent-job",
					Status: StatusProcessing,
				})
			}(i)

			go func() {
				defer wg.Done()
				_ = tracker.GetActiveJobs()
			}()
		}
		wg.Wait()
	})
}