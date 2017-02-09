package main

import (
	"flag"
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"bytes"
	"io"
	"gopkg.in/tylerb/graceful.v1"
	"time"
	"strings"
)

var (
	githubUsername = ""
	githubPassword = ""
	githubAddr     = ""
	Addr           = ""
)


type GitHubRateLimit struct {
	Resources struct {
			  Core    struct {
					  Limit     int `json:"limit"`
					  Remaining int `json:"remaining"`
					  Reset     int `json:"reset"`
				  } `json:"core"`
			  Search  struct {
					  Limit     int `json:"limit"`
					  Remaining int `json:"remaining"`
					  Reset     int `json:"reset"`
				  } `json:"search"`
			  Graphql struct {
					  Limit     int `json:"limit"`
					  Remaining int `json:"remaining"`
					  Reset     int `json:"reset"`
				  } `json:"graphql"`
		  } `json:"resources"`
	Rate      struct {
			  Limit     int `json:"limit"`
			  Remaining int `json:"remaining"`
			  Reset     int `json:"reset"`
		  } `json:"rate"`
}


func (g *GitHubRateLimit) WriteTo(w io.Writer) {
	buf := &bytes.Buffer{}

	// GitHub Rate Limit: Resources
	buf.WriteString(fmt.Sprintf("# HELP %s %s\n", "github_ratelimit_resources_limit", "GitHub Rate Limit: Resources"))
	buf.WriteString(fmt.Sprintf("# TYPE %s %s\n", "github_ratelimit_resources_limit", "gauge"))

	buf.WriteString(fmt.Sprintf("%s{type=%s} %d\n", "github_ratelimit_resources_limit", "core", g.Resources.Core.Limit))
	buf.WriteString(fmt.Sprintf("%s{type=%s} %d\n", "github_ratelimit_resources_limit", "search", g.Resources.Search.Limit))
	buf.WriteString(fmt.Sprintf("%s{type=%s} %d\n", "github_ratelimit_resources_limit", "graphql", g.Resources.Graphql.Limit))


	// GitHub Rate Remaining: Resources
	buf.WriteString(fmt.Sprintf("# HELP %s %s\n", "github_ratelimit_resources_remaining", "GitHub Rate Remaining: Resources"))
	buf.WriteString(fmt.Sprintf("# TYPE %s %s\n", "github_ratelimit_resources_remaining", "gauge"))

	buf.WriteString(fmt.Sprintf("%s{type=%s} %d\n", "github_ratelimit_resources_remaining", "core", g.Resources.Core.Remaining))
	buf.WriteString(fmt.Sprintf("%s{type=%s} %d\n", "github_ratelimit_resources_remaining", "search", g.Resources.Search.Remaining))
	buf.WriteString(fmt.Sprintf("%s{type=%s} %d\n", "github_ratelimit_resources_remaining", "graphql", g.Resources.Graphql.Remaining))

	// GitHub Rate Reset: Resources
	buf.WriteString(fmt.Sprintf("# HELP %s %s\n", "github_ratelimit_resources_remaining", "GitHub Rate Reset: Resources"))
	buf.WriteString(fmt.Sprintf("# TYPE %s %s\n", "github_ratelimit_resources_remaining", "count"))

	buf.WriteString(fmt.Sprintf("%s{type=%s} %d\n", "github_ratelimit_resources_reset", "core", g.Resources.Core.Reset))
	buf.WriteString(fmt.Sprintf("%s{type=%s} %d\n", "github_ratelimit_resources_reset", "search", g.Resources.Search.Reset))
	buf.WriteString(fmt.Sprintf("%s{type=%s} %d\n", "github_ratelimit_resources_reset", "graphql", g.Resources.Graphql.Reset))


	// GitHub Rate Rate
	buf.WriteString(fmt.Sprintf("# HELP %s %s\n", "github_ratelimit_rate_limit", "GitHub Rate Limit"))
	buf.WriteString(fmt.Sprintf("# TYPE %s %s\n", "github_ratelimit_rate_limit", "gauge"))
	buf.WriteString(fmt.Sprintf("%s %d\n", "github_ratelimit_rate_limit", g.Rate.Limit))

	buf.WriteString(fmt.Sprintf("# HELP %s %s\n", "github_ratelimit_rate_remaining", "GitHub Rate Remaining"))
	buf.WriteString(fmt.Sprintf("# TYPE %s %s\n", "github_ratelimit_rate_remaining", "gauge"))
	buf.WriteString(fmt.Sprintf("%s %d\n", "github_ratelimit_rate_remaining", g.Rate.Remaining))

	buf.WriteString(fmt.Sprintf("# HELP %s %s\n", "github_ratelimit_rate_reset", "GitHub Rate Reset"))
	buf.WriteString(fmt.Sprintf("# TYPE %s %s\n", "github_ratelimit_rate_reset", "count"))
	buf.WriteString(fmt.Sprintf("%s %d\n", "github_ratelimit_rate_reset", g.Rate.Reset))

	io.Copy(w, buf)
}


func main() {
	addr := flag.String("addr", "0.0.0.0:8080", "HTTP Server address")
	url := flag.String("url", "https://api.github.com", "Github API address")
	username := flag.String("username", "", "GitHub username")
	password := flag.String("password", "", "GitHub password")
	flag.Parse()

	Addr = *addr
	githubAddr = *url + "/rate_limit"
	githubUsername = *username
	githubPassword = *password

	log.Println("Starting GitHub exporter.")
	log.Println("Listening on:", "'" + Addr + "'",)
	log.Println("Scrapping:", "'" + githubAddr + "'", "with username", "'" + githubUsername + "'", "and password", "'" + strings.Repeat("*", len(githubPassword)) + "'", ".")

	server := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:        Addr,
			ReadTimeout: time.Duration(5) * time.Second,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				req, err := http.NewRequest("GET", githubAddr, nil)
				if err != nil {
					log.Println(err)
				}

				if (githubUsername != "") && (githubPassword != "") {
					req.SetBasicAuth(githubUsername, githubPassword)
				}

				transport := http.Transport{}
				resp, err := transport.RoundTrip(req)
				if err != nil {
					log.Println(err)
				}

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
				}

				resp.Body.Close()

				res := GitHubRateLimit{}
				json.Unmarshal([]byte(body), &res)

				res.WriteTo(w)
			}),
		},
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
