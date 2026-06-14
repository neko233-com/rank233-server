package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yourname/rank233-server/ranker"
)

type Server struct {
	ranker     *ranker.Ranker
	persister  *ranker.Persister
	token      string
}

func New(r *ranker.Ranker, p *ranker.Persister, token string) *Server {
	if token == "" {
		token = "neko233"
	}
	return &Server{ranker: r, persister: p, token: token}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/version", s.handleVersion)

	mux.HandleFunc("/api/ranklist/create", s.auth(s.handleCreate))
	mux.HandleFunc("/api/ranklist/delete", s.auth(s.handleDelete))
	mux.HandleFunc("/api/ranklist/put", s.auth(s.handlePut))
	mux.HandleFunc("/api/ranklist/remove", s.auth(s.handleRemove))
	mux.HandleFunc("/api/ranklist/clean", s.auth(s.handleClean))

	mux.HandleFunc("/api/ranklist/rank", s.auth(s.handleGetRank))
	mux.HandleFunc("/api/ranklist/top", s.auth(s.handleTopN))
	mux.HandleFunc("/api/ranklist/range", s.auth(s.handleRange))
	mux.HandleFunc("/api/ranklist/snapshot", s.auth(s.handleSnapshot))
	mux.HandleFunc("/api/ranklist/version", s.auth(s.handleVersionRead))
	mux.HandleFunc("/api/ranklist/list", s.auth(s.handleList))

	return mux
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Token")
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		if token != s.token {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}
		next(w, r)
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": "rank233-server v1.0.0"})
}

type createReq struct {
	ServerID string `json:"server_id"`
	Capacity int32  `json:"capacity"`
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.ServerID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "server_id required"})
		return
	}
	if err := s.ranker.Create(req.ServerID, req.Capacity); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	if s.persister != nil {
		s.persister.Save(req.ServerID)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "created", "server_id": req.ServerID})
}

type deleteReq struct {
	ServerID string `json:"server_id"`
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req deleteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if !s.ranker.Delete(req.ServerID) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if s.persister != nil {
		s.persister.CleanFile(req.ServerID)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

type putReq struct {
	ServerID  string `json:"server_id"`
	PlayerID  int64  `json:"player_id"`
	Primary   int64  `json:"primary"`
	Secondary int64  `json:"secondary"`
	Arrival   int64  `json:"arrival"`
}

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req putReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	sc := ranker.Score{Primary: req.Primary, Secondary: req.Secondary, Arrival: req.Arrival}
	updated, accepted, err := s.ranker.Put(req.ServerID, req.PlayerID, sc)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if s.persister != nil && updated {
		s.persister.Save(req.ServerID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"updated": updated, "accepted": accepted})
}

type removeReq struct {
	ServerID string `json:"server_id"`
	PlayerID int64  `json:"player_id"`
}

func (s *Server) handleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req removeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	removed, err := s.ranker.Remove(req.ServerID, req.PlayerID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if s.persister != nil && removed {
		s.persister.Save(req.ServerID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"removed": removed})
}

type cleanReq struct {
	ServerID    string `json:"server_id"`
	BeforeTime  int64  `json:"before_time"`
}

func (s *Server) handleClean(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req cleanReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	removed, err := s.persister.CleanBefore(req.ServerID, req.BeforeTime)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if s.persister != nil {
		s.persister.Save(req.ServerID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"removed": removed})
}

func (s *Server) handleGetRank(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	playerStr := r.URL.Query().Get("player_id")
	if serverID == "" || playerStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "server_id and player_id required"})
		return
	}
	playerID, err := strconv.ParseInt(playerStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid player_id"})
		return
	}
	rl, ok := s.ranker.Get(serverID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ranklist not found"})
		return
	}
	snap := rl.Snapshot()
	rank, exists := snap.GetRank(playerID)
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "player not found"})
		return
	}
	sc, _ := snap.GetScore(playerID)
	writeJSON(w, http.StatusOK, map[string]any{
		"player_id": playerID,
		"rank":      rank,
		"primary":   sc.Primary,
		"secondary": sc.Secondary,
		"arrival":   sc.Arrival,
		"version":   snap.Version(),
	})
}

func (s *Server) handleTopN(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	limitStr := r.URL.Query().Get("limit")
	if serverID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "server_id required"})
		return
	}
	limit := int32(10)
	if limitStr != "" {
		v, err := strconv.ParseInt(limitStr, 10, 32)
		if err == nil && v > 0 {
			limit = int32(v)
		}
	}
	rl, ok := s.ranker.Get(serverID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ranklist not found"})
		return
	}
	snap := rl.Snapshot()
	entries := snap.GetTopN(limit)
	result := make([]map[string]any, len(entries))
	for i, e := range entries {
		result[i] = map[string]any{
			"rank":       e.Rank,
			"player_id":  e.PlayerID,
			"primary":    e.Score.Primary,
			"secondary":  e.Score.Secondary,
			"arrival":    e.Score.Arrival,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version": snap.Version(),
		"entries": result,
	})
}

func (s *Server) handleRange(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if serverID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "server_id required"})
		return
	}
	start, _ := strconv.ParseInt(startStr, 10, 32)
	end, _ := strconv.ParseInt(endStr, 10, 32)
	if start <= 0 {
		start = 1
	}
	if end <= 0 {
		end = 10
	}
	rl, ok := s.ranker.Get(serverID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ranklist not found"})
		return
	}
	snap := rl.Snapshot()
	entries := snap.GetRange(int32(start), int32(end))
	result := make([]map[string]any, len(entries))
	for i, e := range entries {
		result[i] = map[string]any{
			"rank":       e.Rank,
			"player_id":  e.PlayerID,
			"primary":    e.Score.Primary,
			"secondary":  e.Score.Secondary,
			"arrival":    e.Score.Arrival,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version": snap.Version(),
		"entries": result,
	})
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	if serverID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "server_id required"})
		return
	}
	rl, ok := s.ranker.Get(serverID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ranklist not found"})
		return
	}
	snap := rl.Snapshot()
	writeJSON(w, http.StatusOK, map[string]any{
		"server_id": snap.Name(),
		"capacity":  snap.Capacity(),
		"len":       snap.Len(),
		"version":   snap.Version(),
	})
}

func (s *Server) handleVersionRead(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	verStr := r.URL.Query().Get("version")
	if serverID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "server_id required"})
		return
	}
	rl, ok := s.ranker.Get(serverID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ranklist not found"})
		return
	}
	currentVer := rl.Version()
	if verStr != "" {
		reqVer, err := strconv.ParseInt(verStr, 10, 64)
		if err == nil && reqVer == currentVer {
			writeJSON(w, http.StatusOK, map[string]any{
				"server_id":    serverID,
				"version":      currentVer,
				"version_match": true,
			})
			return
		}
	}
	snap := rl.Snapshot()
	entries := snap.GetTopN(snap.Len())
	result := make([]map[string]any, len(entries))
	for i, e := range entries {
		result[i] = map[string]any{
			"rank":       e.Rank,
			"player_id":  e.PlayerID,
			"primary":    e.Score.Primary,
			"secondary":  e.Score.Secondary,
			"arrival":    e.Score.Arrival,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"server_id":     serverID,
		"version":       snap.Version(),
		"version_match": false,
		"entries":       result,
	})
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	names := s.ranker.ListNames()
	writeJSON(w, http.StatusOK, map[string]any{"ranklists": names})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if !strings.HasPrefix(r.URL.Path, "/healthz") {
			log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
		}
	})
}
