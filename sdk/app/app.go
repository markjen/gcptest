package app

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func init() {
	http.HandleFunc("/_ah/", ackHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/stop", stopHandler)
	http.HandleFunc("/", statusHandler)
}

func logInstanceID(r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, "Instance ID: %s", appengine.InstanceID())
}

func ackHandler(w http.ResponseWriter, r *http.Request) {
	logInstanceID(r)
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	logInstanceID(r)

}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	logInstanceID(r)

}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	logInstanceID(r)
	w.Write([]byte("ok"))
}
