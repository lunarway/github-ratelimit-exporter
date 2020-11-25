package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	graceful "gopkg.in/tylerb/graceful.v1"
)

func main() {
	flags := pflag.NewFlagSet("github-ratelimit-exporter", pflag.ExitOnError)

	address := flags.String("web.listen-address", "0.0.0.0:9756", "HTTP server address exposing Prometheus metrics")
	shutdownTimeout := flags.Duration("web.shutdown-timeout", 10*time.Second, "HTTP server graceful shutdown timeout. Set to 0 to disable shutdown timeout")
	readTimeout := flags.Duration("web.request-read-timeout", 5*time.Second, "HTTP server read request timeout")
	githubAddr := flags.String("github.url", "https://api.github.com/rate_limit", "URL for GitHub rate limit API")
	var githubUsers githubUserValues
	flags.Var(&githubUsers, "github.user", "GitHub users to get rate limits for. Repeat flag for multiple users.")

	developmentLog := flags.Bool("log.development", false, "Log in human readable format")
	var logLevel zapcore.Level
	flags.AddGoFlag(&flag.Flag{
		DefValue: `"info"`,
		Name:     "log.level",
		Usage:    "Logging level. Available values are 'debug', 'info', 'error'",
		Value:    &logLevel,
	})

	flags.Parse(os.Args[1:])

	log := newLogger(logLevel, *developmentLog)
	defer log.Sync()

	log.Info("Starting GitHub ratelimit exporter")
	log.Infof("Listening on: '%s'", *address)
	log.Infof("Scrapping: '%s' for user names %v", *githubAddr, githubUsers.String())

	var (
		rateLimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "github_ratelimit_limit_info",
			Help: "Maximum number of requests permitted in a single rate limit window",
		}, []string{"username", "resource"})
		rateRemaining = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "github_ratelimit_remaining_info",
			Help: "Number of requests remaining in the current rate limit window",
		}, []string{"username", "resource"})
		rateReset = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "github_ratelimit_reset_epoch_seconds_info",
			Help: "Time at which the current rate limit window resets in UTC epoch seconds",
		}, []string{"username", "resource"})
		rateErrors = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "github_ratelimit_errors_total",
			Help: "Total number of errors collecting rate limit values from GitHub",
		})
	)
	prometheus.MustRegister(rateLimit, rateRemaining, rateReset, rateErrors)

	observe := func(userName, resource string, v values) {
		log.With("values", v).
			With("username", userName).
			With("resource", resource).
			Infof("Observing rate limit values: user=%s resource=%s remaining=%d", userName, resource, v.Remaining)

		rateLimit.WithLabelValues(userName, resource).Set(float64(v.Limit))
		rateRemaining.WithLabelValues(userName, resource).Set(float64(v.Remaining))
		rateReset.WithLabelValues(userName, resource).Set(float64(v.Reset))
	}

	server := &graceful.Server{
		Timeout: *shutdownTimeout,
		LogFunc: log.Infof,
		Server: &http.Server{
			Addr:        *address,
			ReadTimeout: *readTimeout,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var scrapeError bool
				for _, githubUser := range githubUsers.values {
					logger := log.With("username", githubUser.userName)
					logger.Infof("Getting latest rate limit values for '%s'", githubUser.userName)
					res, err := getCurrentLimits(*githubAddr, githubUser.userName, githubUser.accessToken, log)
					if err != nil {
						rateErrors.Inc()
						logger.Errorf("Failed to get latest values: %v", err)
						scrapeError = true
						continue
					}

					observe(githubUser.userName, "core", res.Resources.Core)
					observe(githubUser.userName, "search", res.Resources.Search)
					observe(githubUser.userName, "graphql", res.Resources.GraphQL)
					observe(githubUser.userName, "integration_manifest", res.Resources.IntegrationManifest)
				}
				if scrapeError {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				promhttp.Handler().ServeHTTP(w, r)
			}),
		},
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func getCurrentLimits(addr, userName, accessToken string, log *zap.SugaredLogger) (gitHubRateLimit, error) {
	req, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		return gitHubRateLimit{}, fmt.Errorf("create http request: %w", err)
	}

	if (userName != "") && (accessToken != "") {
		req.SetBasicAuth(userName, accessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return gitHubRateLimit{}, fmt.Errorf("execute http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Failed to read response body: %s", body)
		} else {
			log.Errorf("Response body: %s", body)
		}
		return gitHubRateLimit{}, fmt.Errorf("http response status code %s", resp.Status)
	}

	var res gitHubRateLimit
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return gitHubRateLimit{}, fmt.Errorf("json unmarshal response: %w", err)
	}
	return res, nil
}

type gitHubRateLimit struct {
	Resources struct {
		Core                values `json:"core"`
		Search              values `json:"search"`
		GraphQL             values `json:"graphql"`
		IntegrationManifest values `json:"integration_manifest"`
	} `json:"resources"`
}

type values struct {
	Limit     int `json:"limit"`
	Remaining int `json:"remaining"`
	Reset     int `json:"reset"`
}

func newLogger(level zapcore.Level, development bool) *zap.SugaredLogger {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: development,
		Sampling:    nil,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "@timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	if development {
		cfg.Encoding = "console"
	}
	rawLog, err := cfg.Build()
	if err != nil {
		fmt.Printf("Failed to instantiate logger: %v", err)
		os.Exit(1)
	}
	return rawLog.Sugar()
}

// githubUserValues is a flag that collects user names and access tokens for a
// GitHub user.
type githubUserValues struct {
	values []githubUser
}

type githubUser struct {
	userName    string
	accessToken string
}

var _ pflag.Value = &githubUserValues{}

func (g *githubUserValues) String() string {
	var userNames []string
	for _, user := range g.values {
		userNames = append(userNames, user.userName)
	}
	return fmt.Sprintf("[ %s ]", strings.Join(userNames, ","))
}

func (g *githubUserValues) Set(s string) error {
	sections := strings.SplitN(s, "=", 2)
	if len(sections) != 2 {
		return fmt.Errorf("github user must be formatted as username=access-token: was %s", s)
	}
	userName := sections[0]
	accessToken := sections[1]
	if len(userName) == 0 || len(accessToken) == 0 {
		return fmt.Errorf("github user must be formatted as username=access-token: was %s", s)
	}
	g.values = append(g.values, githubUser{
		userName:    userName,
		accessToken: accessToken,
	})
	return nil
}

func (g *githubUserValues) Type() string {
	return "username=access-token"
}
