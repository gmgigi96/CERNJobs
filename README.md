# CERNJobsBot

Bot reporting new staff job positions at CERN.

## Install
```
go build
cp CERNJobs /usr/bin/CERNJobs
```

## Run
```
Usage of ./staff-bot:
  -jobs-file string
        specify where found jobs are cached (default "/tmp/jobs-cern")
  -tg-chat-id int
        telegram chat ID
  -tg-token string
        telegram API token
```

## Configuration

The bot works combined to cron. Add this line to the crontab:
```
0 0/5 0 ? * * * /usr/bin/CERNJobs -tg-chat-id <id> -tg-token <tkn>
```
