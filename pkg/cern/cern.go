package cern

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gmgigi96/CERNJobs/pkg/poster"
	"github.com/gmgigi96/CERNJobs/pkg/registry"
)

func init() {
	registry.Register("cern", New)
}

type cern struct {
	client *http.Client
}

const jobsURL = "https://careers.smartrecruiters.com/CERN/staff"

func New(_ map[string]any) (poster.JobPoster, error) {
	return &cern{
		client: &http.Client{Timeout: 60},
	}, nil
}

func (c *cern) GetJobPosting(ctx context.Context) ([]*poster.JobPosting, error) {
	res, err := c.client.Get(jobsURL)
	if err != nil {
		return nil, fmt.Errorf("error doing GET request to %s: %w", jobsURL, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %s", res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing document: %w", err)
	}

	var jobs []*poster.JobPosting
	// <li class="opening-job job">
	// 	<a href="https://jobs.smartrecruiters.com/CERN/743999918559253-applied-physicist-ep-nu-2023-99-ld-?trid=08c65188-8d9e-4a09-bff1-3337c1661b63" class="link--block details">
	// 		<h4 class="details-title job-title link--block-target">Applied Physicist (EP-NU-2023-99-LD)</h4>
	// 		<ul class="job-list list--dotted">
	// 			<li class="job-desc">Geneva, Switzerland</li>
	// 			<li class="job-desc">EP</li>
	// 		</ul>
	// 	</a>
	// </li>
	doc.Find(".opening-job.job").Each(func(_ int, s *goquery.Selection) {
		title := s.Find(".job-title").Text()
		url, _ := s.Find("a").Attr("href")
		department := s.Find(".job-desc").Last().Text()
		jobs = append(jobs, &poster.JobPosting{
			Title:      strings.TrimSpace(title),
			DetailsURL: url,
			Department: department,
		})
	})

	return jobs, nil
}
