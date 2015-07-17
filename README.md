# Google Cloud Platform Tests

## gc
This package contains a small app to help observe GC behavior while running
in GAE.

### findings
* runtime.GC() successfully executes the GC within the execution of a single
  request.
* if runtime.MemStats.Sys exceeds the configured amount, the instance will be
  shutdown cleanly after the request if it finishes within ~1 second (i.e.
  /_ah/stop will be called).
* if a request exceeds configured memory amount and continues to run for more
  than 1 second, the request will be killed and instance shutdown uncleanly
  (i.e. /_ah/stop will not be called).

## scaling
This package contains GAE components that test the module scaling policies,
driven by various triggers (HTTP requests and push task queues).

### findings

* Taskqueue pushes to the worker module in bursts of ~10 tasks
* Automatic scaling does not strictly respect `max_concurrent_requests`
* As worker module is hit with traffic over a period of time, automatic scaling
  increases the number of instances; but seemingly gradually?

## sdk
This package attempts to measure performance of the open source SDK in a GAE
app vs. a Managed VMs.

To get going, first get the open source SDK:

`go get google.golang.org/appengine`

### findings
