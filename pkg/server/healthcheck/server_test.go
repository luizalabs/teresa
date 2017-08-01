package healthcheck

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type fakeK8s struct {
	err error
}

func (f *fakeK8s) HealthCheck() error {
	return f.err
}

func TestHealthCheckHandler(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database", err)
	}
	defer db.Close()
	s := New(&fakeK8s{}, db)

	req, err := http.NewRequest("GET", "/healthcheck/", nil)
	if err != nil {
		t.Fatal("error creating http request", err)
	}
	res := httptest.NewRecorder()
	s.healthCheck(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, res.Code)
	}
}

func TestHealthCheckHandlerError(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database", err)
	}
	defer db.Close()

	expectedK8sErrorMsg := "boom"
	s := New(&fakeK8s{err: errors.New(expectedK8sErrorMsg)}, db)

	req, err := http.NewRequest("GET", "/healthcheck/", nil)
	if err != nil {
		t.Fatal("error creating http request", err)
	}
	res := httptest.NewRecorder()
	s.healthCheck(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("expected %d, got %d", http.StatusInternalServerError, res.Code)
	}

	body, _ := ioutil.ReadAll(res.Body)
	hcRes := new(healthCheckResponse)
	json.Unmarshal(body, hcRes)

	if hcRes.K8sError != expectedK8sErrorMsg {
		t.Errorf("expected %s, got %s", expectedK8sErrorMsg, hcRes.K8sError)
	}
}
