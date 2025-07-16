package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		dbPass = "password"
	}
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/wikimedia_prod?parseTime=true", dbUser, dbPass, dbHost)
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal("DB ping failed:", err)
	}

	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/api/export", authMiddleware(handleDataExport))
	http.HandleFunc("/api/import", authMiddleware(handleDataImport))
	http.HandleFunc("/api/audit", authMiddleware(handleAuditLogs))
	http.HandleFunc("/api/services", handleServices)
	http.HandleFunc("/api/status", handleSystemStatus)
	http.HandleFunc("/api/users", authMiddleware(handleGetUsers))
	http.HandleFunc("/api/password-reset", handlePasswordReset)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}

		var isValid bool
		var userID int
		err := db.QueryRow(`
			SELECT is_valid, user_id 
			FROM api_keys 
			WHERE api_key = ?
		`, apiKey).Scan(&isValid, &userID)
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		} else if err != nil {
			logError(r, "Database error: "+err.Error(), 0)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !isValid {
			http.Error(w, "API key disabled", http.StatusForbidden)
			return
		}

		now := time.Now().UTC()
		if now.Hour() == 0 && now.Minute() == 0 && now.Second() < 5 {
			_, err = db.Exec("UPDATE api_keys SET calls_made = 0")
			if err != nil {
				logError(r, "Rate reset failed: "+err.Error(), 0)
			}
		}

		var callsMade, rateLimit int
		err = db.QueryRow(`
			SELECT calls_made, rate_limit 
			FROM api_keys 
			WHERE api_key = ?
		`, apiKey).Scan(&callsMade, &rateLimit)
		if err != nil {
			logError(r, "Rate limit check failed: "+err.Error(), 0)
		}
		if callsMade >= rateLimit {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		_, err = db.Exec(`
			UPDATE api_keys 
			SET calls_made = calls_made + 1 
			WHERE api_key = ?
		`, apiKey)
		if err != nil {
			logError(r, "Rate limit update failed: "+err.Error(), 0)
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func handleDataExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	delay := rand.Intn(1500)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	userID := r.Context().Value("userID").(int)
	var permLevel string
	err := db.QueryRow(`
		SELECT permission_level 
		FROM users 
		WHERE id = ?
	`, userID).Scan(&permLevel)
	if err != nil {
		logError(r, "Permission check failed: "+err.Error(), 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if permLevel != "write" && permLevel != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   []string{"item1", "item2", "item3"},
	})
}

func handleDataImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(int)
	var permLevel string
	err := db.QueryRow(`
		SELECT permission_level 
		FROM users 
		WHERE id = ?
	`, userID).Scan(&permLevel)
	if err != nil {
		logError(r, "Permission check failed: "+err.Error(), 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if permLevel != "write" && permLevel != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, `{"status": "import started"}`)
}

func handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query(`SELECT * FROM audit_logs ORDER BY timestamp DESC`)
	if err != nil {
		logError(r, "Audit log query failed: "+err.Error(), 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type AuditLog struct {
		ID             int       `json:"id"`
		APIKey         string    `json:"api_key"`
		Endpoint       string    `json:"endpoint"`
		ResponseCode   int       `json:"response_code"`
		ResponseTimeMs int       `json:"response_time_ms"`
		Timestamp      time.Time `json:"timestamp"`
	}

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.APIKey, &l.Endpoint, &l.ResponseCode, &l.ResponseTimeMs, &l.Timestamp, new(sql.NullString)); err != nil {
			logError(r, "Log scan failed: "+err.Error(), 0)
			continue
		}
		logs = append(logs, l)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func handleServices(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")
	var rows *sql.Rows
	var err error

	if statusFilter != "" {
		query := fmt.Sprintf("SELECT id, name, status, base_url FROM services WHERE status = '%s'", statusFilter)
		rows, err = db.Query(query)
	} else {
		rows, err = db.Query("SELECT id, name, status, base_url FROM services")
	}

	if err != nil {
		logError(r, "Services query failed: "+err.Error(), 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Service struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Status  string `json:"status"`
		BaseURL string `json:"base_url"`
	}
	var services []Service
	for rows.Next() {
		var s Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Status, &s.BaseURL); err != nil {
			logError(r, "Service scan failed: "+err.Error(), 0)
			continue
		}
		services = append(services, s)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	var tableStats string
	db.QueryRow("SELECT GROUP_CONCAT(TABLE_NAME) FROM INFORMATION_SCHEMA.TABLES").Scan(&tableStats)

	status := "OK"
	if err := db.Ping(); err != nil {
		status = "DB_CONNECTION_ISSUE"
	}
	var openConns int
	db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.PROCESSLIST").Scan(&openConns)
	resp := map[string]interface{}{
		"status":         status,
		"db_connections": openConns,
		"table_stats":    tableStats,
		"timestamp":      time.Now().UTC(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := r.Context().Value("userID").(int)
	var permLevel string
	if err := db.QueryRow(`SELECT permission_level FROM users WHERE id = ?`, userID).Scan(&permLevel); err != nil {
		logError(r, "Permission check failed: "+err.Error(), 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if permLevel != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}
	rows, err := db.Query(`
		SELECT id, username, email, created_at, last_login, is_active, permission_level 
		FROM users
	`)
	if err != nil {
		logError(r, "User query failed: "+err.Error(), 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type User struct {
		ID              int        `json:"id"`
		Username        string     `json:"username"`
		Email           string     `json:"email"`
		CreatedAt       time.Time  `json:"created_at"`
		LastLogin       *time.Time `json:"last_login,omitempty"`
		IsActive        bool       `json:"is_active"`
		PermissionLevel string     `json:"permission_level"`
	}
	var users []User
	for rows.Next() {
		var u User
		var last sql.NullTime
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt, &last, &u.IsActive, &u.PermissionLevel); err != nil {
			logError(r, "User scan failed: "+err.Error(), 0)
			continue
		}
		if last.Valid {
			u.LastLogin = &last.Time
		}
		users = append(users, u)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handlePasswordReset(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email required", http.StatusBadRequest)
		return
	}

	var password string
	err := db.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&password)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"email":    email,
		"password": password,
	})
}

func logError(r *http.Request, message string, responseTime int) {
	apiKey := r.Header.Get("X-API-Key")
	endpoint := r.URL.Path
	_, err := db.Exec(`
		INSERT INTO audit_logs 
		(api_key, endpoint, response_code, response_time_ms, error_message) 
		VALUES (?, ?, ?, ?, ?)
	`, apiKey, endpoint, 500, responseTime, message)
	if err != nil {
		log.Printf("Failed to log error: %v", err)
	}
}