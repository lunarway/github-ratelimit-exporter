# GitHub Rate Limit Prometheus exporter
===
[![Build Status](https://travis-ci.org/marceloalmeida/github-ratelimit-exporter.svg?branch=master)](https://travis-ci.org/marceloalmeida/github-ratelimit-exporter)
[![Code Climate](https://codeclimate.com/github/marceloalmeida/github-ratelimit-exporter/badges/gpa.svg)](https://codeclimate.com/github/marceloalmeida/github-ratelimit-exporter)
[![Issue Count](https://codeclimate.com/github/marceloalmeida/github-ratelimit-exporter/badges/issue_count.svg)](https://codeclimate.com/github/marceloalmeida/github-ratelimit-exporter)
---

Simple GitHub Rate Limit Prometheus exporter, useful to keep on track your GitHub API requests.

# Usage:
```sh
docker run marceloalmeida/github-ratelimit-exporter:latest
```

# Usage (parsing parameters):
```sh
docker run marceloalmeida/github-ratelimit-exporter:latest --help
docker run marceloalmeida/github-ratelimit-exporter:latest -addr 0.0.0.0:8080
```

