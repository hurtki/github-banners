package github

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v81/github"
	"github.com/hurtki/github-banners/api/internal/config"
	"github.com/hurtki/github-banners/api/internal/domain"
	log "github.com/hurtki/github-banners/api/internal/logger"
	"github.com/hurtki/github-banners/api/internal/service"
)

// Handler manages HTTP request handling
type Handler struct {
	config        *config.Config
	logger        log.Logger
	statsService  *service.StatsService
}

//constructor to create http handler
func NewHandler(cfg *config.Config, logger log.Logger, statsService *service.StatsService) *Handler {
	return &Handler{
		config:       cfg,
		logger:       logger,
		statsService: statsService,
	}
}

//chi routers
func (h *Handler) RegisterRoutes(router interface {
	Get(pattern string, handlerFn http.HandlerFunc)
	Post(pattern string, handlerFn http.HandlerFunc)
}) {
	router.Get("/api/stats/{username}", h.GetUserStats)
	router.Post("/api/batch", h.GetBatchStats)
	router.Get("/health", h.HealthCheck)
}

//GET /api/stats/{username}
func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.PathValue("username"))
	if username == "" {
		h.respondError(w, "Username is required", http.StatusBadRequest)
		return
	}

	//username format validation
	if !isValidGitHubUsername(username) {
		h.respondError(w, "Invalid GitHub username format", http.StatusBadRequest)
		return
	}

	requestLogger := h.logger.With(
		"username", username,
		"method", r.Method,
		"path", r.URL.Path,
		"ip", r.RemoteAddr,
		"user_agent", r.UserAgent(),
	)
	requestLogger.Info("Processing GitHub stats request")

	ctx, cancel := context.WithTimeout(r.Context(), h.config.RequestTimeout)
	defer cancel()

	startTime := time.Now()
	userStats, err := h.statsService.GetUserStats(ctx, username)
	elapsed := time.Since(startTime)

	if err != nil {
		requestLogger.Error("Failed to fetch GitHub stats", 
			"error", err.Error(), 
			"duration", elapsed.String())
		h.handleGitHubError(w, err)
		return
	}

	requestLogger.Info("Successfully fetched GitHub stats",
		"duration", elapsed.String(),
		"cached", userStats.Cached,
		"total_repos", userStats.Stats.TotalRepos,
		"total_stars", userStats.Stats.TotalStars)

	h.respondJSON(w, userStats, http.StatusOK)
}

//POST /api/batch for multiple users
func (h *Handler) GetBatchStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Usernames []string `json:"usernames"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if len(request.Usernames) == 0 {
		h.respondError(w, "No usernames provided", http.StatusBadRequest)
		return
	}

	if len(request.Usernames) > 10 {
		h.respondError(w, "Maximum 10 usernames per batch", http.StatusBadRequest)
		return
	}

	for _, username := range request.Usernames {
		if !isValidGitHubUsername(username) {
			h.respondError(w, "Invalid GitHub username: "+username, http.StatusBadRequest)
			return
		}
	}

	requestLogger := h.logger.With(
		"usernames", strings.Join(request.Usernames, ","),
		"count", len(request.Usernames),
		"method", r.Method,
		"ip", r.RemoteAddr,
	)
	requestLogger.Info("Processing batch GitHub stats request")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second) // Longer timeout for batch
	defer cancel()

	startTime := time.Now()
	results, err := h.statsService.GetMultipleUsers(ctx, request.Usernames)
	elapsed := time.Since(startTime)

	if err != nil {
		requestLogger.Error("Failed to fetch batch GitHub stats", 
			"error", err.Error(), 
			"duration", elapsed.String())
		h.respondError(w, "Failed to fetch some user stats", http.StatusPartialContent)
		return
	}

	response := map[string]interface{}{
		"results": results,
		"count":   len(results),
		"fetched_at": time.Now().UTC(),
	}

	requestLogger.Info("Successfully fetched batch GitHub stats",
		"duration", elapsed.String(),
		"success_count", len(results))

	h.respondJSON(w, response, http.StatusOK)
}

//GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	healthResponse := map[string]interface{}{
		"status":    "healthy",
		"service":   "github-stats-api",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
	}

	h.respondJSON(w, healthResponse, http.StatusOK)
}

func (h *Handler) handleGitHubError(w http.ResponseWriter, err error) {
	if ghErr, ok := err.(*github.ErrorResponse); ok {
		switch ghErr.Response.StatusCode {
		case http.StatusNotFound:
			h.respondError(w, "GitHub user not found", http.StatusNotFound)
		case http.StatusForbidden:
			// Rate limit exceeded
			h.respondError(w, "GitHub API rate limit exceeded", http.StatusTooManyRequests)
		case http.StatusUnauthorized:
			h.respondError(w, "GitHub authentication failed", http.StatusInternalServerError)
		default:
			h.respondError(w, "GitHub API error: "+ghErr.Message, ghErr.Response.StatusCode)
		}
		return
	}

	//context timeout
	if err == context.DeadlineExceeded {
		h.respondError(w, "Request timeout", http.StatusGatewayTimeout)
		return
	}

	// Network or other errors
	h.respondError(w, "Failed to fetch GitHub data: "+err.Error(), http.StatusInternalServerError)
}

func (h *Handler) respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes
	
	//CORS headers
	if len(h.config.CORSOrigins) > 0 {
		w.Header().Set("Access-Control-Allow-Origin", h.config.CORSOrigins[0])
	}
	
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

func (h *Handler) respondError(w http.ResponseWriter, message string, statusCode int) {
	errorResponse := domain.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Code:    statusCode,
		Message: message,
	}

	h.respondJSON(w, errorResponse, statusCode)
}

func isValidGitHubUsername(username string) bool {
	if len(username) < 1 || len(username) > 39 {
		return false
	}

	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return false
		}
	}

	if strings.HasPrefix(username, "-") ||
		strings.HasSuffix(username, "-") ||
		strings.Contains(username, "--") {
		return false
	}

	return true
}

var startTime = time.Now()
