package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/vitalvas/oneauth/internal/ksm/config"
	"github.com/vitalvas/oneauth/internal/ksm/crypto"
	"github.com/vitalvas/oneauth/internal/ksm/database"
	"github.com/vitalvas/oneauth/internal/logger"
)

type Server struct {
	config     *config.Config
	db         database.DB
	crypto     *crypto.Engine
	httpServer *http.Server
	logger     *logrus.Logger
}

func New(configPath string) (*Server, error) {
	log := logger.New("")

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	db, err := database.New(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	cryptoEngine, err := crypto.NewEngine(cfg.Security.MasterKey)
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.WithError(closeErr).Error("Failed to close database during cleanup")
		}
		return nil, fmt.Errorf("failed to initialize crypto engine: %w", err)
	}

	return &Server{
		config: cfg,
		db:     db,
		crypto: cryptoEngine,
		logger: log,
	}, nil
}

func (s *Server) Start() error {
	router := mux.NewRouter()

	// Middleware

	// Traditional KSM Protocol
	router.HandleFunc("/wsapi/decrypt/", s.handleKSMDecrypt).Methods(http.MethodGet)

	// Modern REST API
	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/decrypt", s.handleRESTDecrypt).Methods(http.MethodPost)
	api.HandleFunc("/keys", s.handleStoreKey).Methods(http.MethodPost)
	api.HandleFunc("/keys", s.handleListKeys).Methods(http.MethodGet)
	api.HandleFunc("/keys/{key_id}", s.handleDeleteKey).Methods(http.MethodDelete)

	// Health check
	router.HandleFunc("/health", s.handleHealth).Methods(http.MethodGet)

	s.httpServer = &http.Server{
		Addr:         s.config.Server.Address,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.WithField("address", s.config.Server.Address).Info("Starting KSM server")

	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down KSM server")

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}

func (s *Server) Close() error {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close database connection")
			return fmt.Errorf("failed to close database: %w", err)
		}
	}
	return nil
}
