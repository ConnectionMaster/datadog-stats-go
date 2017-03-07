package datadog

import (
	"github.com/DataDog/datadog-go/statsd"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Datadog struct {
	Client *statsd.Client
}

func New(addr string) *Datadog {
	client, err := statsd.New(addr)
	if err != nil {
		log.Fatal(err)
	}

	dd := &Datadog{Client: client}

	return dd
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (dd *Datadog) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		t1 := time.Now()

		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)

		t2 := time.Now()

		dd.Client.TimeInMilliseconds("response_time", float64(t2.Sub(t1)/time.Millisecond), dd.Client.Tags, 1)
		dd.Client.Incr("status_code."+strconv.Itoa(lrw.statusCode), dd.Client.Tags, 1)
	}

	return http.HandlerFunc(fn)
}
