package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var config struct {
	JobsFile         string
	TelegramAPIToken string
	TelegramChatID   int64
}

func init() {
	flag.StringVar(&config.JobsFile, "jobs-file", "/tmp/jobs-cern", "specify where found jobs are cached")
	flag.StringVar(&config.TelegramAPIToken, "tg-token", "", "telegram API token")
	flag.Int64Var(&config.TelegramChatID, "tg-chat-id", 0, "telegram chat ID")

	flag.Parse()

	if config.TelegramAPIToken == "" {
		log.Fatal("telegram api token not provided")
	}
	if config.TelegramChatID == 0 {
		log.Fatal("telegram chat id not provided")
	}
}

type JobList struct {
	file string
	jobs []JobPosting
}

func (j *JobList) init(src string) error {
	if j == nil {
		*j = JobList{}
	}
	j.file = src

	f, err := os.Open(j.file)
	if err != nil {
		if os.IsNotExist(err) {
			j.jobs = make([]JobPosting, 0)
			return j.save()
		}
		return fmt.Errorf("error opening file %s: %w", j.file, err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&j.jobs); err != nil {
		if errors.Is(err, io.EOF) {
			// file is emtpy
			j.jobs = make([]JobPosting, 0)
			return j.save()
		}
		return fmt.Errorf("error decoding file %s: %w", j.file, err)
	}
	return nil
}

const jobsURL = "https://careers.smartrecruiters.com/CERN/staff"

type JobPosting struct {
	DetailsURL string
	Title      string
	Department string
}

func getJobPosting() ([]JobPosting, error) {
	res, err := http.Get(jobsURL)
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

	var jobs []JobPosting
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
		jobs = append(jobs, JobPosting{
			Title:      strings.TrimSpace(title),
			DetailsURL: url,
			Department: department,
		})
	})

	return jobs, nil
}

func (j *JobList) has(job *JobPosting) bool {
	for _, jj := range j.jobs {
		if jj.Department == job.Department && jj.Title == job.Title {
			return true
		}
	}
	return false
}

func (j *JobList) hasNewJobs(got []JobPosting) ([]JobPosting, error) {
	var found []JobPosting
	for _, job := range got {
		if !j.has(&job) {
			found = append(found, job)
		}
	}
	if len(found) != 0 {
		j.jobs = got
		if err := j.save(); err != nil {
			return nil, err
		}
	}
	return found, nil
}

func (j *JobList) save() error {
	f, err := os.OpenFile(j.file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", j.file, err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(j.jobs); err != nil {
		return fmt.Errorf("error encoding job list into json: %w", err)
	}
	return nil
}

func main() {
	var list JobList
	if err := list.init(config.JobsFile); err != nil {
		log.Fatal(err)
	}

	got, err := getJobPosting()
	if err != nil {
		log.Fatal(err)
	}

	found, err := list.hasNewJobs(got)
	if err != nil {
		log.Fatal(err)
	}

	if len(found) == 0 {
		return
	}

	if err := sendOnTelegram(found); err != nil {
		log.Fatal(err)
	}
}

func (j JobPosting) MessageString() string {
	f := `💼 %s

👉 [More details](%s)`
	return fmt.Sprintf(f, j.Title, j.DetailsURL)
}

func sendOnTelegram(l []JobPosting) error {
	bot, err := tgbotapi.NewBotAPI(config.TelegramAPIToken)
	if err != nil {
		return fmt.Errorf("error creating telegram bot: %w", err)
	}

	for _, job := range l {
		text := job.MessageString()
		log.Println("sending", text, "to telegram")
		msg := tgbotapi.NewMessage(config.TelegramChatID, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err := bot.Send(msg); err != nil {
			return fmt.Errorf("error sending message to telegram: %w", err)
		}
	}
	return nil
}
