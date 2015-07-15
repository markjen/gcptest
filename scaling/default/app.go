package scaling

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"

	"github.com/markjen/gcptest/scaling/model"
)

func init() {
	http.HandleFunc("/clear", handlerWrapper(clear))
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

	delayStr := r.URL.Query().Get("delay")
	if delayStr == "" {
		delayStr = "10"
	}

	tasks := make([]*taskqueue.Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = taskqueue.NewPOSTTask(
			"/work",
			url.Values{
				"delay": []string{delayStr},
			})
	}
	_, err := taskqueue.AddMulti(c, tasks, "auto-worker-push")
	if err != nil {
		c.Errorf("Error loading tasks: %s", err)
		output.WriteLine("Error loading tasks")
		return http.StatusInternalServerError
	}

	output.WriteLine("Loaded %d tasks into queue with delay %s", count, delayStr)
	return http.StatusOK
}

func clear(w http.ResponseWriter, r *http.Request, c appengine.Context, output *crBuffer) int {
	q := datastore.NewQuery(model.WorkerExecKind).KeysOnly()
	iter := q.Run(c)

	count := 0
	success := 0
	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		output.WriteLine("Successfully deleted %d of %d entities", success, count)
	}()

	for {
		key, err := iter.Next(nil)
		if err == datastore.Done {
			return http.StatusOK
		}
		if err != nil {
			c.Errorf("Error clearing WorkerExec entities: %s", err)
			output.WriteLine("Error clearing WorkerExec entities")
			return http.StatusInternalServerError
		}

		count++
		wg.Add(1)
		go func(k *datastore.Key) {
			datastore.Delete(c, k)
			if err != nil {
				c.Errorf("Error deleteing key %s", k)
			} else {
				success++
			}
			wg.Done()
		}(key)
	}
}
