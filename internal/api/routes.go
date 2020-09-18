package api

import "github.com/gorilla/mux"

// SetupRoutes creates a router for app endpoints supported by API
func (s *Server) SetupRoutes() {
	r := mux.NewRouter()
	// Public
	r.Methods("POST").Path("/").HandlerFunc(s.createOrUpdateAWSResource)
	r.Methods("DELETE").Path("/{resourceId}").HandlerFunc(s.deleteAWSResource)

	// Internal
	r.Methods("GET").Path("/alive").HandlerFunc(s.isAlive)
	r.Methods("GET").Path("/health").HandlerFunc(s.isReady)
	s.Router = r
}
