package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func setupTestDB(t *testing.T) *sql.DB {
       dbUser := os.Getenv("DB_USER")
       if dbUser == "" {
	       dbUser = "root"
       }
       dbPass := os.Getenv("DB_PASS")
       if dbPass == "" {
	       dbPass = "root"
       }
       dbHost := os.Getenv("DB_HOST")
       if dbHost == "" {
	       dbHost = "localhost"
       }
       dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":3306)/wikimedia_prod?parseTime=true"
       db, err := sql.Open("mysql", dsn)
       if err != nil {
	       t.Fatalf("DB connection failed: %v", err)
       }
       return db
}

func TestRateLimitReset(t *testing.T) {
       db = setupTestDB(t)
       defer db.Close()

       // Set up a test API key with low limit
       _, err := db.Exec(`UPDATE api_keys SET calls_made = 5, rate_limit = 5, last_reset = ? WHERE api_key = 'key_ABC123'`, time.Now().Add(-25*time.Hour))
       if err != nil {
	       t.Fatalf("Failed to setup api_key: %v", err)
       }

       req := httptest.NewRequest("GET", "/api/export", nil)
       req.Header.Set("X-API-Key", "key_ABC123")
       w := httptest.NewRecorder()
       handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
	       w.WriteHeader(http.StatusOK)
       })
       handler(w, req)

       if w.Result().StatusCode == http.StatusTooManyRequests {
	       t.Errorf("Rate limit should have reset, but got 429 Too Many Requests")
       }
}

func TestAuditKeyMasking(t *testing.T) {
       db = setupTestDB(t)
       defer db.Close()

       req := httptest.NewRequest("GET", "/api/audit", nil)
       req.Header.Set("X-API-Key", "key_DEF456")
       w := httptest.NewRecorder()
       handler := authMiddleware(handleAuditLogs)
       handler(w, req)

       if w.Result().StatusCode != http.StatusOK {
	       t.Fatalf("Expected 200 OK, got %d", w.Result().StatusCode)
       }

       var logs []struct {
	       MaskedKey string `json:"api_key"`
       }
       if err := json.NewDecoder(w.Body).Decode(&logs); err != nil {
	       t.Fatalf("Failed to decode response: %v", err)
       }
       for _, l := range logs {
	       if len(l.MaskedKey) < 9 || l.MaskedKey[6:] != "..." {
		       t.Errorf("API key not masked properly: got %q", l.MaskedKey)
	       }
       }
}
