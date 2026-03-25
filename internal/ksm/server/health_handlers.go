package server

import (
	"net/http"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	dbHealth := DatabaseHealth{
		Status: "connected",
	}

	// Check database health
	if err := s.db.HealthCheck(); err != nil {
		dbHealth.Status = "error"
		dbHealth.Error = err.Error()
	}

	// Create overall health response
	health := HealthResponse{
		Status:   "healthy",
		Database: dbHealth,
	}

	// If database is unhealthy, mark overall status as unhealthy
	if dbHealth.Status == "error" {
		health.Status = "unhealthy"
		s.sendJSONResponse(w, http.StatusServiceUnavailable, health)
		return
	}

	s.sendJSONResponse(w, http.StatusOK, health)
}
