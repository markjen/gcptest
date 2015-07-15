package scaling

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"appengine"

	"github.com/markjen/gcptest/scaling/model"
)

var requestCounter int64 = 0

func init() {
	http.HandleFunc("/work", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	c := appengine.NewContext(r)

	requestCount := atomic.AddInt64(&requestCounter, 1)
	c.Infof("Request count: %d", requestCount)

	// Default delay value to 60 seconds
	var delay int64 = 60
	delayStr := r.FormValue("delay")
	if delayStr != "" {
		var err error
		delay, err = strconv.ParseInt(delayStr, 10, 64)
		if err != nil {
			c.Warningf("Bad delay request parameter: %s", delayStr)
		}
	}
	if delay < 0 {
		c.Warningf("Negative delay requested: %d, clamping to 0", delay)
		delay = 0
	}

	c.Infof("Waiting %d", delay)

	time.Sleep(time.Second * time.Duration(delay))
	finish := time.Now()

	workerExec := &model.WorkerExec{
		Started:       start,
		Finished:      finish,
		InstanceID:    appengine.InstanceID(),
		RequestNumber: requestCount}
	key, err := model.SaveWorkerExec(c, workerExec)
	if err != nil {
		c.Errorf("Could not save worker exec (%+v): %+v", workerExec, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Infof("Finished, writing data: %+v (ID: %s)", workerExec, key.StringID())

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Done!"))
}
