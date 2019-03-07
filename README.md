Contains interesting go code examples runnable as tests. (These tests mostly
don't test anything; they are just examples. They are meant to be run individually.)

Points of interest:

- `goroutine_test.go`
  - `TestMapReducePattern` -- Example of how to manage jobs and workers using
    wait groups and range on a channel. Run with `go test -run MapReduce`.
  - Various simpler examples of goroutines.

- `marshal_test.go` -- examples of marhaling oddities and how to make marshal output nice
