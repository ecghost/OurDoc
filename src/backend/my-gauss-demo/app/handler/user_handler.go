package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"my-gauss-app/model"
)

// HandleUsers POST: 插入用户
func HandleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var users []model.User
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		log.Printf("Invalid JSON")
		return
	}

	for _, u := range users {
		if err := model.InsertUser(u); err != nil {
			log.Printf("Insert user %v failed: %v", u, err)
			http.Error(w, "Insert failed", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Users inserted successfully"))
}

// HandleQueryUsers GET: 查询所有用户
func HandleQueryUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := model.QueryAllUsers()
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
