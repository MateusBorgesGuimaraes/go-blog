package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"blog-api/internal/config"
	"blog-api/internal/handlers"
	customMiddleware "blog-api/internal/middleware"
	"blog-api/internal/repository"
)

func main() {
	// Carrega o .env se existir (em produção, as env vars já vêm do ambiente)
	if err := godotenv.Load(); err != nil {
		slog.Info("nenhum arquivo .env encontrado, usando variáveis de ambiente do sistema")
	}

	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}
	slog.Info("conectado ao banco de dados")

	queries := repository.New(pool)
	authHandler := handlers.NewAuthHandler(queries, cfg.JWTSecret)
	userHandler := handlers.NewUserHandler(queries)
	postHandler := handlers.NewPostHandler(queries)
	tagHandler := handlers.NewTagHandler(queries)
	commentHandler := handlers.NewCommentHandler(queries)
	feedHandler := handlers.NewFeedHandler(queries, cfg.BaseURL)
	uploadHandler := handlers.NewUploadHandler(cfg.BaseURL)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Serve os arquivos estáticos de /uploads
	fileServer := http.FileServer(http.Dir("./uploads"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fileServer))

	// Rotas
	r.Get("/health", handlers.HealthCheck)

	r.Route("/api", func(r chi.Router) {
		r.Get("/feed.xml", feedHandler.ServeFeed)
		r.Post("/auth/register", userHandler.CreateUser)
		r.Post("/auth/login", authHandler.Login)
		r.Get("/posts", postHandler.ListPosts)
		r.Get("/posts/{slug}", postHandler.SearchPostBySlug)
		r.Get("/posts/id/{id}", postHandler.SearchPostById)
		r.Get("/tags", tagHandler.ListTags)
		r.Get("/tags/{slug}", tagHandler.GetTagBySlug)
		r.Get("/posts/{id}/tags", tagHandler.ListTagsByPostID)
		r.Get("/posts/{id}/comments", commentHandler.ListComments)
		r.Post("/posts/{id}/comments", commentHandler.CreateComment)

		r.Route("/admin", func(r chi.Router) {
			r.Use(customMiddleware.Auth(cfg.JWTSecret))
			r.Get("/users", userHandler.SearchUserByEmail)
			r.Get("/users/{id}", userHandler.SearchUserByID)
			r.Get("/posts", postHandler.ListAllPosts)
			r.Get("/posts/author", postHandler.ListAllAuthorPosts)
			r.Post("/posts", postHandler.CreatePost)
			r.Put("/posts/{id}", postHandler.EditPost)
			r.Patch("/posts/{id}/publish", postHandler.PublishPost)
			r.Delete("/posts/{id}", postHandler.DeletePost)
			r.Post("/tags", tagHandler.CreateTag)
			r.Delete("/tags/{id}", tagHandler.DeleteTag)
			r.Post("/posts/{id}/tags", tagHandler.AddTagToPost)
			r.Delete("/posts/{id}/tags/{tagId}", tagHandler.RemoveTagFromPost)
			r.Patch("/comments/{id}/approve", commentHandler.ApproveComment)
			r.Delete("/comments/{id}", commentHandler.DeleteComment)
			r.Post("/uploads", uploadHandler.UploadImage)
		})
	})

	slog.Info("servidor rodando", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
