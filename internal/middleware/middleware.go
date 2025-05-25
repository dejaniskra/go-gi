package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dejaniskra/go-gi/logger"
	"github.com/google/uuid"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		transactionID := uuid.New().String()

		ctx := logger.WithFields(r.Context(), logger.Field{
			Key:   "transaction_id",
			Value: transactionID,
		})

		w.Header().Set("Transaction-ID", transactionID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorCtx(r.Context(), "panic recovered", logger.Field{
					Key:   "error",
					Value: fmt.Sprintf("%v", err),
				})

				transactionID := ""
				for _, f := range logger.FieldsFromContext(r.Context()) {
					if f.Key == "transaction_id" {
						transactionID = fmt.Sprintf("%v", f.Value)
						break
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Transaction-ID", transactionID)
				w.WriteHeader(http.StatusInternalServerError)

				// Write JSON error response
				json.NewEncoder(w).Encode(map[string]string{
					"error":          "Internal Server Error",
					"transaction_id": transactionID,
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
