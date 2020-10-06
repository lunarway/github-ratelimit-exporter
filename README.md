# Prometheus metrics exporter for the GitHub API rate limits


[![Build Status](https://travis-ci.com/lunarway/github-ratelimit-exporter.svg?branch=master)](https://travis-ci.com/lunarway/github-ratelimit-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/lunarway/github-ratelimit-exporter)](https://goreportcard.com/report/github.com/lunarway/github-ratelimit-exporter)
[![Docker Repository on Quay](https://quay.io/repository/lunarway/github-ratelimit-exporter/status "Docker Repository on Quay")](https://quay.io/repository/lunarway/github-ratelimit-exporter)

Prometheus exporter for [GitHub API rate limits](https://docs.github.com/en/free-pro-team@latest/developers/apps/rate-limits-for-github-apps) written in Go.
Allows for exporting scanning data into Prometheus by scraping the GitHub API for rate limit information.
Note that the endpoint used does not count in the rate limiter, ie. running the exporter does put additional load on the API usage.


# Installation

Several pre-compiled binaries are available from the [releases page](https://github.com/lunarway/github-ratelimit-exporter/releases).

A docker image is also available on our Quay.io registry.

```
docker run quay.io/lunarway/github-ratelimit-exporter --github.user octocat --github.access-token <access token for user>
```

# Usage

Provide a GitHub user name and access token to exporter limits based on that user.
If you omit these fields the limits are based on unauthenticated requests.
These are limited by the calling IP address.

To create a personal access token follow official documentation on [Creating a personal access token
](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token).

It exposes prometheus metrics on `/` on port `9532` (can be configured).

```
github-ratelimit-exporter --github.user octocat --github.access-token <access token for user>
```

See all configuration options with the `--help` flag

```
$ github-ratelimit-exporter --help
Usage of github-ratelimit-exporter:
      --github.access-token string          Access token for GitHub user defined in flag github.user
      --github.url string                   URL for GitHub rate limit API (default "https://api.github.com/rate_limit")
      --github.user string                  GitHub user to get rate limits for
      --log.development                     Log in human readable format
      --log.level Level                     Logging level. Available values are 'debug', 'info', 'error' (default info)
      --web.listen-address string           HTTP server address exposing Prometheus metrics (default "0.0.0.0:9756")
      --web.request-read-timeout duration   HTTP server read request timeout (default 5s)
      --web.shutdown-timeout duration       HTTP server graceful shutdown timeout. Set to 0 to disable shutdown timeout (default 10s)
```

# Build

The exporter can be build using the standard Go tool chain if you have it available.

```
go build
```

You can build inside a Docker image as well.
This produces a `github-ratelimit-exporter` image that can run with the binary as entry point.

```
docker build -t github-ratelimit-exporter .
```

This is useful if the exporter is to be depoyled in Kubernetes or other dockerized environments.

Here is an example of running the exporter locally.

```
$ docker run -p 9756:9756 github-ratelimit-exporter:latest --github.user octocat --github.access-token <access token> --log.development=true
2020-10-06T20:13:10.714Z	info	src/main.go:45	Starting GitHub ratelimit exporter
2020-10-06T20:13:10.715Z	info	src/main.go:46	Listening on: '0.0.0.0:9756'
2020-10-06T20:13:10.715Z	info	src/main.go:47	Scrapping: 'https://api.github.com/rate_limit' with user name 'astrochimp' and access token '****************************************'
2020-10-06T20:13:32.174Z	info	src/main.go:86	Getting latest rate limit values
2020-10-06T20:13:35.516Z	info	src/main.go:72	Observing rate limit values: resource=core remaining=4839	{"values": {"limit":5000,"remaining":4839,"reset":1602017397}, "resource": "core"}
2020-10-06T20:13:35.517Z	info	src/main.go:72	Observing rate limit values: resource=search remaining=30	{"values": {"limit":30,"remaining":30,"reset":1602015275}, "resource": "search"}
2020-10-06T20:13:35.517Z	info	src/main.go:72	Observing rate limit values: resource=graphql remaining=5000	{"values": {"limit":5000,"remaining":5000,"reset":1602018815}, "resource": "graphql"}
2020-10-06T20:13:35.517Z	info	src/main.go:72	Observing rate limit values: resource=integration_manifest remaining=5000	{"values": {"limit":5000,"remaining":5000,"reset":1602018815}, "resource": "integration_manifest"}
```

# Deployment

To deploy the exporter in Kubernetes, you can find a simple Kubernetes deployment and secret yaml in the `examples` folder.
You have to add your GitHub user access token in the `secrets.yaml` and the GitHub user that you want to get metrics from in the args section of the `deployment.yaml`.
The examples assumes that you have a namespace in kubernetes named: `monitoring`.

It further assumes that you have [kubernetes service discovery](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config) configured for you Prometheus instance and a target that will gather metrics from pods, similar to this:

```yaml
- job_name: 'kubernetes-pods'
  kubernetes_sd_configs:
  - role: pod

  relabel_configs:
  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
    action: keep
    regex: true
  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
    action: replace
    target_label: __metrics_path__
    regex: (.+)
  - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
    action: replace
    regex: (.+):(?:\d+);(\d+)
    replacement: ${1}:${2}
    target_label: __address__
  - action: labelmap
    regex: __meta_kubernetes_pod_label_(.+)
```

To deploy it to your kubernetes cluster run the following commands:

```
kubectl apply -f examples/kubernetes.yaml
```

# Development

The project uses Go modules so you need Go version >=1.15 to run it.
Run builds and tests with the standard Go tool chain.

```
go build
go test
```

# Credits

This exporter is a fork of [marceloalmeida/github-ratelimit-exporter](https://github.com/marceloalmeida/github-ratelimit-exporter).
