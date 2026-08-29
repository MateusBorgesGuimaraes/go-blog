package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"blog-api/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Queries *repository.Queries
}

func NewUserHandler(queries *repository.Queries) *UserHandler {
	return &UserHandler{Queries: queries}
}

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserResponse struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// CreateUser responde POST /api/auth/register
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		http.Error(w, formatValidationError(err), http.StatusBadRequest)
		return
	}

	_, err := h.Queries.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		http.Error(w, "já existe uma conta com esse email", http.StatusConflict)
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "erro ao verificar email", http.StatusInternalServerError)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "erro ao processar senha", http.StatusInternalServerError)
		return
	}

	user, err := h.Queries.CreateUser(r.Context(), repository.CreateUserParams{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		Role:         "editor",
	})
	if err != nil {
		http.Error(w, "erro ao criar usuário", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	})
}

// SearchUserByID responde GET /api/admin/users/:id
func (h *UserHandler) SearchUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	user, err := h.Queries.GetUserByID(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "usuário não encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "erro ao buscar usuário", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UserResponse{
		ID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role,
	})
}

// SearchUserByEmail responde GET /api/admin/users?email=...
func (h *UserHandler) SearchUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "parâmetro 'email' é obrigatório", http.StatusBadRequest)
		return
	}

	user, err := h.Queries.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "usuário não encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "erro ao buscar usuário", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UserResponse{
		ID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role,
	})
}
