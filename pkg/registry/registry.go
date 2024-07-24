package registry

import "github.com/gmgigi96/CERNJobs/pkg/poster"

// NewFunc is the function used to create a new JobPoster implementation,
// given a configuration.
type NewFunc func(map[string]any) (poster.JobPoster, error)

// registry holds the implementations of the JobPoster interface.
var registry = make(map[string]NewFunc)

// Register registers the functions to create JobPoster implementations.
func Register(name string, f NewFunc) {
	registry[name] = f
}
