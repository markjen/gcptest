package scaling

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"appengine"
)

var requestCounter int64 = 0

func init() {
	http.HandleFunc("/_ah/", ackHandler)
	http.HandleFunc("/run", handler)
}

func ackHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("Instance ID: %s", appengine.InstanceID())
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	c := appengine.NewContext(r)
	c.Infof("Instance ID: %s", appengine.InstanceID())

	requestCount := atomic.AddInt64(&requestCounter, 1)
	c.Infof("Request count: %d", requestCount)

	// How much garbage to generate per cycle in MBs. Default to 16MB.
	var size int64 = 1 << 4
	sizeStr := r.FormValue("size")
	if sizeStr != "" {
		parsedSize, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			c.Errorf("Bad size request parameter: %s", sizeStr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if parsedSize < 0 {
			c.Errorf("Negative size requested: %d", parsedSize)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		size = parsedSize
	}

	// How much garbage to generate per cycle in KBs. Default to 128MB.
	loops := 3
	loopsStr := r.FormValue("loops")
	if loopsStr != "" {
		parsedLoops, err := strconv.ParseInt(loopsStr, 10, 64)
		if err != nil {
			c.Errorf("Bad loops request parameter: %s", loopsStr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if parsedLoops <= 0 {
			c.Errorf("Negative or zero loops requested: %d", parsedLoops)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		loops = int(parsedLoops)
	}

	m := new(runtime.MemStats)
	printMemStats := func(desc string) {
		runtime.ReadMemStats(m)
		c.Infof("Alloc: %d, TotalAlloc: %d, Sys: %d (%s)", m.Alloc, m.TotalAlloc, m.Sys, desc)
	}

	printMemStats("Before test, before GC")
	runtime.GC()
	printMemStats("Before test, after GC")

	var x []byte
	for i := 0; i < loops; i++ {
		printMemStats(fmt.Sprintf("Begin loop %d", i+1))
		x = make([]byte, size*1024*1024)
		c.Infof("x size: %d", len(x))
		printMemStats("Memory eaten")
		x = nil
		runtime.GC()
		printMemStats("Memory released and after GC")
	}

	c.Infof("Finished request in %s", time.Now().Sub(start))

	w.Write([]byte("ok"))
}
