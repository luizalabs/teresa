package healthcheck

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/jinzhu/gorm"
)

type K8sOperations interface {
	HealthCheck() error
}

type Server struct {
	k8s K8sOperations
	DB  *gorm.DB
}

type healthCheckResponse struct {
	K8sError string `json:"k8s_error"`
	DBError  string `json:"db_error"`
}

func (s *Server) healthCheck(w http.ResponseWriter, _ *http.Request) {
	k8sError := s.k8s.HealthCheck()
	dbError := s.DB.DB().Ping()
	if k8sError != nil || dbError != nil {
		k8sErrorMsg := ""
		if k8sError != nil {
			k8sErrorMsg = k8sError.Error()
		}
		dbErrorMsg := ""
		if dbError != nil {
			dbErrorMsg = dbError.Error()
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(healthCheckResponse{
			K8sError: k8sErrorMsg,
			DBError:  dbErrorMsg,
		})
		return
	}
	w.Write([]byte("OK"))
}

func (s *Server) Run(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck/", s.healthCheck)
	server := &http.Server{Handler: mux}

	return server.Serve(l)
}

func New(k K8sOperations, db *gorm.DB) *Server {
	return &Server{k8s: k, DB: db}
}
