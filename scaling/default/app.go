package scaling

import (
	"bytes"
	"fmt"
	"net/http"

	"appengine"
	"appengine/datastore"

	"strconv"

	"github.com/markjen/gcptest/scaling/model"

	"net/url"

	"appengine/taskqueue"
)

func init() {
	http.HandleFunc("/load", handlerWrapper(load))
	http.HandleFunc("/", handlerWrapper(index))
}

type crBuffer struct {
	bytes.Buffer
}

func (c *crBuffer) WriteLine(s string, a ...interface{}) {
	if len(a) > 0 {
		c.WriteString(fmt.Sprintf(s, a...))
	} else {
		c.WriteString(s)
	}
	c.WriteByte('\n')
}

func handlerWrapper(f func(w http.ResponseWriter, r *http.Request, c appengine.Context, output *crBuffer) int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		var output crBuffer
		status := f(w, r, c, &output)
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		w.Write(output.Bytes())
	}
}

func index(w http.ResponseWriter, r *http.Request, c appengine.Context, output *crBuffer) int {
	q := datastore.NewQuery(model.WorkerExecKind)
	iter := q.Run(c)
	for {
		var we model.WorkerExec
		key, err := iter.Next(&we)
		if err == datastore.Done {
			return http.StatusOK
		}
		if err != nil {
			output.WriteLine("Error retrieving WorkerExec: %s", err)
			return http.StatusInternalServerError
		}
		output.WriteLine("%s: %s", key.StringID(), we.String())
	}
}

func load(w http.ResponseWriter, r *http.Request, c appengine.Context, output *crBuffer) int {
	// Determine how many tasks to queue
	count := 10
	countStr := r.URL.Query().Get("count")
	if countStr != "" {
		parsedCount, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			output.WriteLine("Could not parse count: %s", countStr)
			return http.StatusBadRequest
		}
		if parsedCount <= 0 {
			output.WriteLine("Invalid count specified: %d", parsedCount)
			return http.StatusBadRequest
		}
		count = int(parsedCount)
	}

	//	moduleHostname, err := appengine.ModuleHostname(c, "auto-worker", "1", "")
	//	if err != nil {
	//		c.Errorf("Error getting worker module hostname: %s", err)
	//		output.WriteLine("Could not find worker module")
	//		return http.StatusInternalServerError
	//	}

	for i := 0; i < count; i++ {
		t := taskqueue.NewPOSTTask(
			"/work",
			url.Values{
				"delay": []string{"10"},
			})
		//		t.Header = http.Header{
		//			"Host": []string{moduleHostname},
		//		}
		taskqueue.Add(c, t, "auto-worker-push")
	}

	output.WriteLine("Loaded %d tasks into queue", count)
	return http.StatusOK
}
