package handlers

import (
	"encoding/xml"
	"net/http"
	"time"

	"blog-api/internal/repository"
)

type FeedHandler struct {
	Queries *repository.Queries
	BaseURL string
}

func NewFeedHandler(queries *repository.Queries, baseURL string) *FeedHandler {
	return &FeedHandler{Queries: queries, BaseURL: baseURL}
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// ServeFeed responde GET /api/feed.xml
func (h *FeedHandler) ServeFeed(w http.ResponseWriter, r *http.Request) {
	posts, err := h.Queries.ListPublishedPosts(r.Context(), repository.ListPublishedPostsParams{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		http.Error(w, "erro ao gerar feed", http.StatusInternalServerError)
		return
	}

	items := make([]Item, 0, len(posts))
	for _, post := range posts {
		postURL := h.BaseURL + "/posts/" + post.Slug

		pubDate := post.CreatedAt.Time.Format(time.RFC1123Z)
		if post.PublishedAt.Valid {
			pubDate = post.PublishedAt.Time.Format(time.RFC1123Z)
		}

		items = append(items, Item{
			Title:       post.Title,
			Link:        postURL,
			Description: post.Excerpt.String,
			PubDate:     pubDate,
			GUID:        postURL,
		})
	}

	feed := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:       "Dev notes",
			Link:        h.BaseURL,
			Description: "Artigos sobre backend, Go e arquitetura",
			Items:       items,
		},
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(xml.Header))

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(feed); err != nil {
		http.Error(w, "erro ao gerar feed", http.StatusInternalServerError)
		return
	}
}
