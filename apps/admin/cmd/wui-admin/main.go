package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	db     = make(map[string]interface{})
	dbMu   sync.RWMutex
	adminPassword = "admin123"
	jwtSecret = "wui-admin-secret"
)

func main() {
	if p := os.Getenv("ADMIN_PASSWORD"); p != "" {
		adminPassword = p
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/api/admin/login", handleLogin)
	http.HandleFunc("/api/admin/licenses", withAuth(handleLicenses))
	http.HandleFunc("/api/admin/licenses/create", withAuth(handleCreateLicense))
	http.HandleFunc("/api/admin/stats", withAuth(handleStats))

	log.Printf("WUI Admin started on :8081")
	log.Printf("Password: %s", adminPassword)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", 401)
			return
		}
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		if token != jwtSecret {
			http.Error(w, "Invalid token", 401)
			return
		}
		next(w, r)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.Password != adminPassword {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid password"})
		return
	}
	
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": jwtSecret,
		"admin": map[string]interface{}{
			"id":       1,
			"username": "admin",
			"role":     "superadmin",
		},
	})
}

func handleLicenses(w http.ResponseWriter, r *http.Request) {
	dbMu.RLock()
	var licenses []map[string]interface{}
	for _, v := range db {
		if l, ok := v.(map[string]interface{}); ok {
			if l["type"] == "license" {
				licenses = append(licenses, l)
			}
		}
	}
	dbMu.RUnlock()
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  licenses,
		"total": len(licenses),
	})
}

func handleCreateLicense(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type       string `json:"type"`
		Plan       string `json:"plan"`
		MaxTunnels int    `json:"maxTunnels"`
		MaxUsers   int    `json:"maxUsers"`
		Days       int    `json:"days"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	key := fmt.Sprintf("WUI-%d-%d", time.Now().Unix(), req.MaxTunnels)
	expiresAt := time.Now().AddDate(0, 0, req.Days)
	
	license := map[string]interface{}{
		"type":        "license",
		"key":         key,
		"plan":        req.Plan,
		"status":      "inactive",
		"maxTunnels":  req.MaxTunnels,
		"maxUsers":    req.MaxUsers,
		"expiresAt":   expiresAt.Format(time.RFC3339),
	}
	
	dbMu.Lock()
	db[key] = license
	dbMu.Unlock()
	
	json.NewEncoder(w).Encode(license)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	dbMu.RLock()
	var total, active int
	for _, v := range db {
		if l, ok := v.(map[string]interface{}); ok {
			if l["type"] == "license" {
				total++
				if l["status"] == "active" {
					active++
				}
			}
		}
	}
	dbMu.RUnlock()
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"totalLicenses":  total,
		"activeLicenses": active,
	})
}
