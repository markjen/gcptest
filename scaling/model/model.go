package model

import (
	"fmt"
	"time"

	"appengine"
	"appengine/datastore"
)

const WorkerExecKind = "WorkerExec"

type WorkerExec struct {
	Started       time.Time
	Finished      time.Time
	InstanceID    string
	RequestNumber int64
}

func (we *WorkerExec) String() string {
	return fmt.Sprintf("%40s %40s", we.Started, we.Finished)
}

func SaveWorkerExec(c appengine.Context, w *WorkerExec) (*datastore.Key, error) {
	stringID := fmt.Sprintf("%s:%010d", w.InstanceID[:10], w.RequestNumber)
	key := datastore.NewKey(c, WorkerExecKind, stringID, 0, nil)
	return datastore.Put(c, key, w)
}
