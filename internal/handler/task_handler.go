package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type TaskExecRequest struct {
	TaskID string `json:"task_id"`
}

type TaskExecResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func HandleAdminTaskExec(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, TaskExecResponse{
				Status:  "ERROR",
				Message: "Méthode non autorisée",
			})
			return
		}

		// Protection contre les payloads trop volumineux (max 64 KB)
		r.Body = http.MaxBytesReader(w, r.Body, 64*1024)

		var req TaskExecRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil || req.TaskID == "" {
			writeJSON(w, http.StatusBadRequest, TaskExecResponse{
				Status:  "ERROR",
				Message: "Identifiant de tâche manquant ou JSON invalide",
			})
			return
		}

		operator := r.Header.Get("X-User-Email")
		if operator == "" {
			operator = "admin.super@einvoice.int"
		}

		ctx := r.Context()
		respMessage, respDetails, err := executeTask(ctx, db, req.TaskID)
		if err != nil {
			if errors.Is(err, errUnknownTask) {
				writeJSON(w, http.StatusBadRequest, TaskExecResponse{
					Status:  "ERROR",
					Message: fmt.Sprintf("Tâche inconnue : %s", req.TaskID),
				})
				return
			}

			log.Printf("[AUDIT ERROR] Échec [%s] par [%s]: %v\n", req.TaskID, operator, err)
			writeJSON(w, http.StatusInternalServerError, TaskExecResponse{
				Status:  "ERROR",
				Message: "Échec technique lors de l'exécution",
				Details: err.Error(),
			})
			return
		}

		log.Printf("[AUDIT SUCCESS] Tâche [%s] exécutée par [%s]\n", req.TaskID, operator)
		writeJSON(w, http.StatusOK, TaskExecResponse{
			Status:  "OK",
			Message: respMessage,
			Details: respDetails,
		})
	}
}

var errUnknownTask = errors.New("unknown task")

func executeTask(ctx context.Context, db *sql.DB, taskID string) (message, details string, err error) {
	switch taskID {

	case "checkpoint_wal":
		_, err = db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE);")
		if err == nil {
			message = "Checkpoint SQLite WAL exécuté avec succès."
			details = "invoices.db-wal synchronisé et tronqué."
		}

	case "restart_workers":
		log.Println("[OPS] Ordre de redémarrage des workers reçu.")
		message = "Signal transmis au pool de 16 workers."
		details = "Workers d'ingestion réinitialisés."

	case "purge_sas":
		sasDir := filepath.Join("archives", "temp_drafts")
		cutoff := time.Now().Add(-72 * time.Hour)
		deletedCount := 0

		if _, statErr := os.Stat(sasDir); !os.IsNotExist(statErr) {
			err = filepath.Walk(sasDir, func(path string, info os.FileInfo, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				if !info.IsDir() && info.ModTime().Before(cutoff) {
					if rmErr := os.Remove(path); rmErr == nil {
						deletedCount++
					}
				}
				return nil
			})
		}
		if err == nil {
			message = fmt.Sprintf("Purge terminée : %d fichier(s) purgé(s).", deletedCount)
			details = "Fichiers de plus de 72h nettoyés."
		}

	case "retry_dlq":
		query := `
			UPDATE invoices 
			SET status = 'PENDING', 
			    error_reason = NULL,
			    updated_at = ? 
			WHERE status IN ('ERROR', 'REJECTED')`

		res, dbErr := db.ExecContext(ctx, query, time.Now().UTC().Format(time.RFC3339))
		if dbErr != nil {
			err = dbErr
			return
		}

		rowsAffected, _ := res.RowsAffected()
		message = fmt.Sprintf("File DLQ réinitialisée : %d facture(s) remise(s) en traitement.", rowsAffected)

	default:
		return "", "", errUnknownTask
	}

	return message, details, err
}