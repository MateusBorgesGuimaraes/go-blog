package handlers

import (
	customMiddleware "blog-api/internal/middleware"
	"blog-api/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
)

var validate = validator.New()

// PostHandler agrupa as dependências que os handlers de posts precisam.
type PostHandler struct {
	Queries *repository.Queries
}

func NewPostHandler(queries *repository.Queries) *PostHandler {
	return &PostHandler{Queries: queries}
}

type CreatePostRequest struct {
	Title         string `json:"title" validate:"required,min=3,max=255"`
	Slug          string `json:"slug" validate:"required,min=3,max=255"`
	Content       string `json:"content" validate:"required"`
	Excerpt       string `json:"excerpt" validate:"max=500"`
	CoverImageURL string `json:"cover_image_url" validate:"omitempty,url"`
}

type UpdatePostRequest struct {
	Title         string `json:"title" validate:"required,min=3,max=255"`
	Content       string `json:"content" validate:"required"`
	Excerpt       string `json:"excerpt" validate:"max=500"`
	CoverImageURL string `json:"cover_image_url" validate:"omitempty,url"`
}

// ListPosts responde GET /api/posts (público, só published)
func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	posts, err := h.Queries.ListPublishedPosts(r.Context(), repository.ListPublishedPostsParams{
		Limit: limit, Offset: offset,
	})
	if err != nil {
		http.Error(w, "erro ao buscar posts", http.StatusInternalServerError)
		return
	}

	// Coleta os IDs pra buscar as tags de todos de uma vez
	ids := make([]int32, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
	}

	tagRows, err := h.Queries.ListTagsByPostIDs(r.Context(), ids)
	if err != nil {
		http.Error(w, "erro ao buscar tags dos posts", http.StatusInternalServerError)
		return
	}

	tagsByPost := make(map[int32][]repository.Tag)
	for _, row := range tagRows {
		tagsByPost[row.PostID] = append(tagsByPost[row.PostID], repository.Tag{
			ID: row.ID, Name: row.Name, Slug: row.Slug,
		})
	}

	// Monta a resposta final combinando post + suas tags
	response := make([]PostWithTags, len(posts))
	for i, p := range posts {
		response[i] = PostWithTags{
			ListPublishedPostsRow: p,
			Tags:                  tagsByPost[p.ID], // nil se não tiver tags — vira [] no JSON
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type PostWithTags struct {
	repository.ListPublishedPostsRow
	Tags []repository.Tag `json:"tags"`
}

// ListAllPosts responde GET /api/admin/posts (admin, todos os status)
func (h *PostHandler) ListAllPosts(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	posts, err := h.Queries.ListAllPosts(r.Context(), repository.ListAllPostsParams{
		Limit: limit, Offset: offset,
	})
	if err != nil {
		http.Error(w, "erro ao buscar posts", http.StatusInternalServerError)
		return
	}

	ids := make([]int32, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
	}

	tagRows, err := h.Queries.ListTagsByPostIDs(r.Context(), ids)
	if err != nil {
		http.Error(w, "erro ao buscar tags dos posts", http.StatusInternalServerError)
		return
	}

	tagsByPost := make(map[int32][]repository.Tag)
	for _, row := range tagRows {
		tagsByPost[row.PostID] = append(tagsByPost[row.PostID], repository.Tag{
			ID: row.ID, Name: row.Name, Slug: row.Slug,
		})
	}

	response := make([]PostWithAllTags, len(posts))
	for i, p := range posts {
		response[i] = PostWithAllTags{
			ListAllPostsRow: p,
			Tags:            tagsByPost[p.ID],
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type PostWithAllTags struct {
	repository.ListAllPostsRow
	Tags []repository.Tag `json:"tags"`
}

// ListAllAuthorPosts responde GET /api/admin/posts/author (admin, todos os status)
func (h *PostHandler) ListAllAuthorPosts(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	userID := r.Context().Value(customMiddleware.UserIDKey).(int32)

	posts, err := h.Queries.ListAllPostsByAuthor(r.Context(), repository.ListAllPostsByAuthorParams{
		AuthorID: userID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		http.Error(w, "erro ao buscar posts", http.StatusInternalServerError)
		return
	}

	ids := make([]int32, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
	}

	tagRows, err := h.Queries.ListTagsByPostIDs(r.Context(), ids)
	if err != nil {
		http.Error(w, "erro ao buscar tags dos posts", http.StatusInternalServerError)
		return
	}

	tagsByPost := make(map[int32][]repository.Tag)
	for _, row := range tagRows {
		tagsByPost[row.PostID] = append(tagsByPost[row.PostID], repository.Tag{
			ID: row.ID, Name: row.Name, Slug: row.Slug,
		})
	}

	response := make([]PostWithAuthorTags, len(posts))
	for i, p := range posts {
		response[i] = PostWithAuthorTags{
			ListAllPostsByAuthorRow: p,
			Tags:                    tagsByPost[p.ID],
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type PostWithAuthorTags struct {
	repository.ListAllPostsByAuthorRow
	Tags []repository.Tag `json:"tags"`
}

// PublishPost responde PATCH /api/admin/posts/:id/publish
func (h *PostHandler) PublishPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	_, err = h.Queries.GetPostByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	}
	publishedPost, err := h.Queries.PublishPost(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "erro ao publicar post", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(publishedPost)
}

func parsePagination(r *http.Request) (limit, offset int32) {
	limit = 10
	offset = 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	return limit, offset
}

// SearchPostBySlug responde GET /api/posts/:slug
func (h *PostHandler) SearchPostBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	post, err := h.Queries.GetPostBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	}

	tags, err := h.Queries.ListTagsByPostID(r.Context(), post.ID)
	if err != nil {
		http.Error(w, "erro ao buscar tags do post", http.StatusInternalServerError)
		return
	}

	response := struct {
		repository.GetPostBySlugRow
		Tags []repository.Tag `json:"tags"`
	}{
		GetPostBySlugRow: post,
		Tags:              tags,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SearchPostById responde GET /api/posts/id/:id
func (h *PostHandler) SearchPostById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	post, err := h.Queries.GetPostByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "erro ao buscar post", http.StatusNotFound)
		return
	}

	tags, err := h.Queries.ListTagsByPostID(r.Context(), post.ID)
	if err != nil {
		http.Error(w, "erro ao buscar tags do post", http.StatusInternalServerError)
		return
	}

	response := struct {
		repository.GetPostByIDRow
		Tags []repository.Tag `json:"tags"`
	}{
		GetPostByIDRow: post,
		Tags:              tags,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreatePost responde POST /api/admin/posts
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req CreatePostRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		http.Error(w, formatValidationError(err), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(customMiddleware.UserIDKey).(int32)

	post, err := h.Queries.CreatePost(r.Context(), repository.CreatePostParams{
		Title:         req.Title,
		Slug:          req.Slug,
		Content:       req.Content,
		Excerpt:       pgtype.Text{String: req.Excerpt, Valid: req.Excerpt != ""},
		CoverImageUrl: pgtype.Text{String: req.CoverImageURL, Valid: req.CoverImageURL != ""},
		Status:        "draft",
		AuthorID:      userID,
	})

	if err != nil {
		http.Error(w, "erro ao criar post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// EditPost responde PUT /api/admin/posts/:id
func (h *PostHandler) EditPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	post, err := h.Queries.GetPostByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	}

	userID := r.Context().Value(customMiddleware.UserIDKey).(int32)
	if userID != post.AuthorID {
		http.Error(w, "voce nao pode editar post que nao te pertencem", http.StatusUnauthorized)
		return
	}

	var req UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		http.Error(w, formatValidationError(err), http.StatusBadRequest)
		return
	}

	updatedPost, err := h.Queries.UpdatePost(r.Context(), repository.UpdatePostParams{
		ID:            int32(id),
		Title:         req.Title,
		Content:       req.Content,
		Excerpt:       pgtype.Text{String: req.Excerpt, Valid: req.Excerpt != ""},
		CoverImageUrl: pgtype.Text{String: req.CoverImageURL, Valid: req.CoverImageURL != ""},
	})
	if err != nil {
		http.Error(w, "erro ao atualizar post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedPost)
}

// DeletePost responde DELETE /api/admin/posts/:id
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	post, err := h.Queries.GetPostByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	}

	userID := r.Context().Value(customMiddleware.UserIDKey).(int32)
	if userID != post.AuthorID {
		http.Error(w, "voce nao pode excluir posts que nao te pertencem", http.StatusUnauthorized)
		return
	}

	err = h.Queries.DeletePost(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Erro ao deletar post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func formatValidationError(err error) string {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		fieldErr := validationErrors[0]
		return fmt.Sprintf("campo '%s' é inválido: falhou na regra '%s'", fieldErr.Field(), fieldErr.Tag())
	}
	return "dados inválidos"
}
