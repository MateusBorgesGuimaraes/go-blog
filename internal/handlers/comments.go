package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	customMiddleware "blog-api/internal/middleware"
	"blog-api/internal/repository"

	"github.com/go-chi/chi/v5"
)

type CommentHandler struct {
	Queries *repository.Queries
}

func NewCommentHandler(queries *repository.Queries) *CommentHandler {
	return &CommentHandler{Queries: queries}
}

type CreateCommentRequest struct {
	AuthorName string `json:"author_name" validate:"required,min=2,max=255"`
	Content    string `json:"content" validate:"required,min=3"`
}

// ListComments responde GET /api/posts/:id/comments (público, só aprovados)
func (h *CommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	comments, err := h.Queries.ListApprovedCommentsByPostID(r.Context(), int32(postID))
	if err != nil {
		http.Error(w, "erro ao buscar comentários", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comments)
}

// CreateComment responde POST /api/posts/:id/comments (público)
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		http.Error(w, formatValidationError(err), http.StatusBadRequest)
		return
	}

	comment, err := h.Queries.CreateComment(r.Context(), repository.CreateCommentParams{
		PostID:     int32(postID),
		AuthorName: req.AuthorName,
		Content:    req.Content,
	})
	if err != nil {
		http.Error(w, "erro ao criar comentário", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// ApproveComment responde PATCH /api/admin/comments/:id/approve
func (h *CommentHandler) ApproveComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	commentInfo, err := h.Queries.GetCommentWithPostAuthor(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "comentário não encontrado", http.StatusNotFound)
		return
	}

	userID := r.Context().Value(customMiddleware.UserIDKey).(int32)
	if userID != commentInfo.PostAuthorID {
		http.Error(w, "você não pode moderar comentários de posts que não são seus", http.StatusForbidden)
		return
	}

	comment, err := h.Queries.ApproveComment(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "erro ao aprovar comentário", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comment)
}

// DeleteComment responde DELETE /api/admin/comments/:id
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	commentInfo, err := h.Queries.GetCommentWithPostAuthor(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "comentário não encontrado", http.StatusNotFound)
		return
	}

	userID := r.Context().Value(customMiddleware.UserIDKey).(int32)
	if userID != commentInfo.PostAuthorID {
		http.Error(w, "você não pode excluir comentários de posts que não são seus", http.StatusForbidden)
		return
	}

	if err := h.Queries.DeleteComment(r.Context(), int32(id)); err != nil {
		http.Error(w, "erro ao deletar comentário", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
