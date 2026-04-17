package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"Scrybot/internal/config"
	"Scrybot/internal/state"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var methodNotAllowedError = func(w http.ResponseWriter) { http.Error(w, "method not allowed", http.StatusMethodNotAllowed) }

type Server struct {
	hub   *Hub
	cfg   config.Config
	store state.Store
	uiFS  fs.FS
	port  string

	mu      sync.RWMutex
	running bool
}

func New(hub *Hub, cfg config.Config, store state.Store, uiFiles fs.FS, port string) *Server {
	return &Server{
		hub:     hub,
		cfg:     cfg,
		store:   store,
		uiFS:    uiFiles,
		port:    port,
		running: true,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/command", s.handleCommand)
	mux.HandleFunc("/ws/logs", s.handleWS)

	if s.uiFS != nil {
		mux.Handle("/", http.FileServer(http.FS(s.uiFS)))
	}

	addr := fmt.Sprintf(":%s", s.port)
	log.Printf("Dashboard listening on http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}

type statusResponse struct {
	LastCheck interface{} `json:"last_check"`
	NextCheck interface{} `json:"next_check"`
	SeenCount int         `json:"seen_count"`
	Running   bool        `json:"running"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowedError(w)
		return
	}

	st, err := s.store.Load()
	if err != nil {
		http.Error(w, "failed to load state", http.StatusInternalServerError)
		return
	}

	resp := statusResponse{
		SeenCount: len(st.SeenIDs),
		Running:   true,
	}

	if !st.LastCheck.IsZero() {
		resp.LastCheck = st.LastCheck
		resp.NextCheck = st.LastCheck.Add(s.cfg.PollInterval)
	}

	writeJSON(w, resp)
}

type configResponse struct {
	SearchQuery       string `json:"search_query"`
	PollInterval      string `json:"poll_interval"`
	DataDir           string `json:"data_dir"`
	WebhookConfigured bool   `json:"webhook_configured"`
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowedError(w)
		return
	}

	writeJSON(
		w, configResponse{
			SearchQuery:       s.cfg.SearchQuery,
			PollInterval:      s.cfg.PollInterval.String(),
			DataDir:           s.cfg.DataDir,
			WebhookConfigured: s.cfg.WebhookURL != "",
		},
	)
}

func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowedError(w)
		return
	}
	writeJSON(
		w, map[string]interface{}{
			"ok":     false,
			"output": "no commands implemented yet",
		},
	)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade failed: %v", err)
		return
	}
	s.hub.register(conn)
	defer func() {
		s.hub.unregister(conn)
		err := conn.Close()
		if err != nil {
			log.Printf("WS connection close failed: %v", err)
			return
		}
	}()

	// Block until client disconnects.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
