# GitHub Rate Limit Prometheus exporter


[![Build Status](https://travis-ci.org/marceloalmeida/github-ratelimit-exporter.svg?branch=master)](https://travis-ci.org/marceloalmeida/github-ratelimit-exporter)
[![Maintainability](https://api.codeclimate.com/v1/badges/cdd1c4ee47b13b902f92/maintainability)](https://codeclimate.com/github/marcelosousaalmeida/github-ratelimit-exporter/maintainability)
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

