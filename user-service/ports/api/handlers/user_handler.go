package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/devbenho/bazar-user-service/internal/utils"

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
// @Summary Register a new user
// @Description Register a new user with the provided details
// @Tags users
// @Accept json
// @Produce json
// @Param user body dtos.CreateUserRequest true "User details"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /user [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var createUserRequest dtos.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&createUserRequest); err != nil {
		return err
	}

	result, err := h.service.Register(r.Context(), &createUserRequest)
	if err != nil {
		return err
	}

	response := utils.NewSuccessResponse(http.StatusCreated, "User registered successfully", result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(response)
}

// Login handles user login requests
// @Summary Login a user
// @Description Login a user with the provided credentials
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body dtos.AuthDTO true "User credentials"
// @Success 200 {object} dtos.AuthResponseDTO
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /auth/login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var authDTO dtos.AuthDTO

	if err := json.NewDecoder(r.Body).Decode(&authDTO); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return err
	}

	response, err := h.service.Login(r.Context(), &authDTO)
	if err != nil {
		return err
	}

	result := utils.NewSuccessResponse(http.StatusOK, "User logged in successfully", response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
	return nil
}

// GetUserByID handles requests to fetch user details by ID
// @Summary Get user by ID
// @Description Get user details by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dtos.UserResponseDTO
// @Failure 404 {object} utils.ErrorResponse
// @Router /user/{id} [get]
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"] // Get the user ID from the request path

	user, err := h.service.GetUserByID(id)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
	return nil
}

// UpdateUser handles user update requests
// @Summary Update user details
// @Description Update user details by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dtos.UpdateUserRequest true "Updated user details"
// @Success 200 {object} dtos.UserResponseDTO
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /user/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]

	var updateUserRequest dtos.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateUserRequest); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return err
	}

	updatedUser, err := h.service.UpdateUser(id, &updateUserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
	return nil
}

// DeleteUser handles user deletion requests
// @Summary Delete user
// @Description Delete user by ID
// @Tags users
// @Param id path string true "User ID"
// @Success 204
// @Failure 500 {object} utils.ErrorResponse
// @Router /user/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.service.DeleteUser(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
