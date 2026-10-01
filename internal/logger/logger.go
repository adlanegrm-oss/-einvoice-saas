package logger

import (
	"context"
	"log/slog"
	"os"
)

var Logger *slog.Logger

// InitLogger initialise le logger au format JSON structuré
func InitLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// LogInvoiceEvent enregistre un événement lié au traitement d'une facture
func LogInvoiceEvent(ctx context.Context, level slog.Level, msg string, invoiceID string, status string, err error) {
	attrs := []slog.Attr{
		slog.String("invoice_id", invoiceID),
		slog.String("status", status),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	if Logger != nil {
		Logger.LogAttrs(ctx, level, msg, attrs...)
	} else {
		slog.LogAttrs(ctx, level, msg, attrs...)
	}
}
