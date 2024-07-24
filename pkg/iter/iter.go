package iter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/gmgigi96/CERNJobs/pkg/poster"
	"github.com/gmgigi96/CERNJobs/pkg/registry"
)

func init() {
	registry.Register("iter", New)
}

type iter struct {
	client *http.Client
}

func New(_ map[string]any) (poster.JobPoster, error) {
	return &iter{
		client: &http.Client{Timeout: 60},
	}, nil
}

const jobsURL = "https://www.iter.org/jobs"

func (i *iter) GetJobPosting(ctx context.Context) ([]*poster.JobPosting, error) {
	res, err := i.client.Get(jobsURL)
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
	// <tr class="job 7501">
	//   <td data-value="24008004023059">04 Aug,  2024</td>
	//   <td>
	//     <b>
	//       <a class="job-link" data-desc="Engineering Data Management Coordinator IO0527" data-job="io" href="https://career5.successfactors.eu/career?career_ns=job_listing&amp;company=ITER&amp;navBarLevel=JOB_SEARCH&amp;rcm_site_locale=en_GB&amp;career_job_req_id=7501" target="_blank">Engineering Data Management Coordinator IO0527</a>
	//     </b>
	//   </td>
	//   <td>
	//     <span class="badge new posting">New Posting</span>
	//   </td>
	//   <td>
	//     <a data-job="io" data-desc="Engineering Data Management Coordinator IO0527" href="https://career5.successfactors.eu/career?career_ns=job_listing&amp;company=ITER&amp;navBarLevel=JOB_SEARCH&amp;rcm_site_locale=en_GB&amp;career_job_req_id=7501" target="_blank" class="btn btn-primary btn-large job-link">
	//       <i class="fa fa-eye" aria-hidden="true">
	//       </i>
	//     </a>
	//   </td>
	// </tr>
	doc.Find(".job").Each(func(_ int, s *goquery.Selection) {
		title := s.Find("a.job-link").Text()
		u, _ := s.Find("a.job-link").Attr("href")
		u, _ = url.QueryUnescape(u)

		jobs = append(jobs, &poster.JobPosting{
			DetailsURL: u,
			Title:      title,
		})
	})

	return jobs, nil
}
