package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserData struct {
	Username       string      `json:"username"`
	HashedPassword string      `json:"hashed_password"`
	DashboardState interface{} `json:"dashboardState"`
	CourseData     interface{} `json:"courseData"`
}

var (
	usersData = make(map[string]*UserData)
	sessions  = make(map[string]string) // token -> username
	mu        sync.RWMutex
	dataFile  = "users.json"
	dataDir   = "data_users"
)

func initAuth() {
	os.MkdirAll(dataDir, 0755)

	// Migrate old users.json if it exists
	if _, err := os.Stat(dataFile); err == nil {
		oldData := make(map[string]*UserData)
		b, _ := os.ReadFile(dataFile)
		json.Unmarshal(b, &oldData)
		for username, user := range oldData {
			bUser, _ := json.MarshalIndent(user, "", "  ")
			os.WriteFile(dataDir+"/"+username+".json", bUser, 0644)
		}
		os.Rename(dataFile, dataFile+".bak") // Backup old file
	}

	// Load individual user files
	files, _ := os.ReadDir(dataDir)
	for _, f := range files {
		if !f.IsDir() {
			b, _ := os.ReadFile(dataDir + "/" + f.Name())
			var user UserData
			if err := json.Unmarshal(b, &user); err == nil {
				usersData[user.Username] = &user
			}
		}
	}
}

func saveUser(username string) {
	mu.RLock()
	user, exists := usersData[username]
	mu.RUnlock()

	if !exists {
		return
	}

	b, err := json.MarshalIndent(user, "", "  ")
	if err == nil {
		os.WriteFile(dataDir+"/"+username+".json", b, 0644)
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	if _, exists := usersData[req.Username]; exists {
		mu.Unlock()
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		mu.Unlock()
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	usersData[req.Username] = &UserData{
		Username:       req.Username,
		HashedPassword: string(hashed),
		DashboardState: nil,
		CourseData:     nil,
	}
	mu.Unlock()
	saveUser(req.Username)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Registered successfully"})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.RLock()
	user, exists := usersData[req.Username]
	mu.RUnlock()

	if !exists || bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := uuid.New().String()
	mu.Lock()
	sessions[token] = req.Username
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token, "username": req.Username})
}

// Middleware to get user from token
func getUserFromToken(r *http.Request) *UserData {
	token := r.Header.Get("Authorization")
	if token == "" {
		return nil
	}
	// Basic Bearer token removal
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	mu.RLock()
	username, ok := sessions[token]
	var user *UserData
	if ok {
		user = usersData[username]
	}
	mu.RUnlock()
	return user
}

func handleData(w http.ResponseWriter, r *http.Request) {
	user := getUserFromToken(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		mu.RLock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"dashboardState": user.DashboardState,
			"courseData":     user.CourseData,
		})
		mu.RUnlock()
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			DashboardState interface{} `json:"dashboardState"`
			CourseData     interface{} `json:"courseData"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		user.DashboardState = req.DashboardState
		user.CourseData = req.CourseData
		mu.Unlock()
		saveUser(user.Username)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Data saved successfully"})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
