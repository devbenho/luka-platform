package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/devbenho/bazar-user-service/internal/dtos"
	"github.com/devbenho/bazar-user-service/internal/services"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	service services.IUserService
}

func NewUserHandler(service services.IUserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// Register handles user registration requests
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var createUserRequest dtos.CreateUserRequest

	// Parse the request body into the DTO
	if err := json.NewDecoder(r.Body).Decode(&createUserRequest); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Call the service to register the user
	response, err := h.service.Register(r.Context(), &createUserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the response as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login handles user login requests
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var authDTO dtos.AuthDTO

	// Parse the request body into the AuthDTO
	if err := json.NewDecoder(r.Body).Decode(&authDTO); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Call the service to log in the user
	response, err := h.service.Login(r.Context(), &authDTO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Return the response as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetUserByID handles requests to fetch user details by ID
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Call the service to get the user details
	user, err := h.service.GetUserByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Return the user details as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// UpdateUser handles user update requests
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var updateUserRequest dtos.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateUserRequest); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Call the service to update the user details
	updatedUser, err := h.service.UpdateUser(id, &updateUserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated user details as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}

// DeleteUser handles user deletion requests
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Call the service to delete the user
	if err := h.service.DeleteUser(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return a no-content response
	w.WriteHeader(http.StatusNoContent)
}
