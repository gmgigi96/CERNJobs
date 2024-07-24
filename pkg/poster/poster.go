package poster

import "context"

// JobPosting holds information about a job, like the URL
// with the job description, the title, where it's located.
type JobPosting struct {
	DetailsURL string
	Title      string
	Department string
}

// JobPoster is an interface that defines a method to fetch job postings.
type JobPoster interface {
	GetJobPosting(ctx context.Context) ([]*JobPosting, error)
}
