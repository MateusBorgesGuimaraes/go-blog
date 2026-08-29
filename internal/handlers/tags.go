package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"blog-api/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type TagHandler struct {
	Queries *repository.Queries
}

func NewTagHandler(queries *repository.Queries) *TagHandler {
	return &TagHandler{Queries: queries}
}

type CreateTagRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Slug string `json:"slug" validate:"required,min=2,max=100"`
}

// AddTagToPostRequest é usado tanto pra adicionar quanto remover uma tag de um post.
type TagToPostRequest struct {
	TagID int32 `json:"tag_id" validate:"required"`
}

// ListTags responde GET /api/tags (rota pública)
func (h *TagHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.Queries.ListTags(r.Context())
	if err != nil {
		http.Error(w, "erro ao buscar tags", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tags)
}

// GetTagBySlug responde GET /api/tags/:slug (rota pública)
func (h *TagHandler) GetTagBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	tag, err := h.Queries.GetTagBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "tag não encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "erro ao buscar tag", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tag)
}

// ListTagsByPostID responde GET /api/posts/:id/tags (rota pública)
func (h *TagHandler) ListTagsByPostID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	tags, err := h.Queries.ListTagsByPostID(r.Context(), int32(postID))
	if err != nil {
		http.Error(w, "erro ao buscar tags do post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tags)
}

// CreateTag responde POST /api/admin/tags
func (h *TagHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		http.Error(w, formatValidationError(err), http.StatusBadRequest)
		return
	}

	tag, err := h.Queries.CreateTag(r.Context(), repository.CreateTagParams{
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		http.Error(w, "erro ao criar tag", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tag)
}

// DeleteTag responde DELETE /api/admin/tags/:id
func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	if err := h.Queries.DeleteTag(r.Context(), int32(id)); err != nil {
		http.Error(w, "erro ao deletar tag", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddTagToPost responde POST /api/admin/posts/:id/tags
func (h *TagHandler) AddTagToPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	var req TagToPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		http.Error(w, formatValidationError(err), http.StatusBadRequest)
		return
	}

	err = h.Queries.AddTagToPost(r.Context(), repository.AddTagToPostParams{
		PostID: int32(postID),
		TagID:  req.TagID,
	})
	if err != nil {
		http.Error(w, "erro ao adicionar tag ao post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveTagFromPost responde DELETE /api/admin/posts/:id/tags/:tagId
func (h *TagHandler) RemoveTagFromPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id de post inválido", http.StatusBadRequest)
		return
	}

	tagIDStr := chi.URLParam(r, "tagId")
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		http.Error(w, "id de tag inválido", http.StatusBadRequest)
		return
	}

	err = h.Queries.RemoveTagFromPost(r.Context(), repository.RemoveTagFromPostParams{
		PostID: int32(postID),
		TagID:  int32(tagID),
	})
	if err != nil {
		http.Error(w, "erro ao remover tag do post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
