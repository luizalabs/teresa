package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/luizalabs/teresa-api/helpers"
	"github.com/luizalabs/teresa-api/k8s"
	"github.com/luizalabs/teresa-api/restapi/operations/deployments"
)

const (
	waitConditionTickDuration = 3 * time.Second
)

type flushResponseWriter struct {
	f http.Flusher
	w io.Writer
}

func newFlushResponseWriter(w io.Writer) *flushResponseWriter {
	fw := flushResponseWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}
	return &fw
}
func (fw flushResponseWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if fw.f != nil {
		fw.f.Flush()
	}
	return
}
func (fw flushResponseWriter) Println(a ...interface{}) (n int, err error) {
	n, err = fw.Write([]byte(fmt.Sprintln(a...)))
	return
}
func (fw flushResponseWriter) Printf(format string, a ...interface{}) (n int, err error) {
	n, err = fw.Write([]byte(fmt.Sprintf(format, a...)))
	return
}

// CreateDeploymentHandler handler triggered when a deploy url is requested
var CreateDeploymentHandler deployments.CreateDeploymentHandlerFunc = func(params deployments.CreateDeploymentParams, principal interface{}) middleware.Responder {
	var r middleware.ResponderFunc = func(rw http.ResponseWriter, pr runtime.Producer) {
		tk := k8s.IToToken(principal)

		l := log.WithField("app", params.AppName).WithField("token", *tk.Email).WithField("requestId", helpers.NewShortUUID())
		r, err := k8s.Client.Deployments().Create(params.AppName, *params.Description, &params.AppTarball, helpers.FileStorage, tk)
		if err != nil {
			// FIXME: improve this... is it possible?
			var httpErr *GenericError
			if k8s.IsInputError(err) {
				l.WithError(err).Warn("error during deploy")
				httpErr = NewBadRequestError(err)
			} else if k8s.IsUnauthorizedError(err) {
				l.WithError(err).Warn("error during deploy")
				httpErr = NewUnauthorizedError(err)
			} else {
				l.WithError(err).Error("error during deploy")
				httpErr = NewInternalServerError(err)
			}
			rw.WriteHeader(int(*httpErr.Payload.Code))
			rw.Write([]byte(*httpErr.Payload.Message))
			return
		}
		l.Debug("starting the streaming of the build")
		defer r.Close()
		// creates a flushed response writer...
		w := newFlushResponseWriter(rw)
		// stream funciont output...
		io.Copy(w, r)
	}
	return r
}
