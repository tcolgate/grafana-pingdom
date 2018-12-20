package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/russellcardullo/go-pingdom/pingdom"
	simplejson "github.com/tcolgate/grafana-simple-json-go"
)

type pingdomChecks interface {
	List(...map[string]string) ([]pingdom.CheckResponse, error)
	SummaryOutage(request pingdom.SummaryOutageRequest) (*pingdom.SummaryOutageResponse, error)
	Results(request pingdom.ResultsRequest) (*pingdom.ResultsResponse, error)
}

// Pingdom implements a prom collector, and grafana simple json source for
// scraping the pingdom API.
type Pingdom struct {
	client pingdomChecks
}

// GrafanaAnnotations implements the Annotator of the grafana-simple-json package.
func (p *Pingdom) GrafanaAnnotations(ctx context.Context, query string, args simplejson.AnnotationsArguments) ([]simplejson.Annotation, error) {
	chks, err := p.client.List(map[string]string{"include_tags": "true"})
	if err != nil {
		log.Printf("failed to list checks, %v", err)
		return nil, err
	}

	if query == "" {
		query = ".*"
	}

	qre, err := regexp.Compile(query)
	if err != nil {
		log.Printf("failed to compile query, %v", err)
		return nil, err
	}

	var anns []simplejson.Annotation
	for _, chk := range chks {
		if !qre.MatchString(chk.Hostname) {
			continue
		}

		outs, err := p.client.SummaryOutage(pingdom.SummaryOutageRequest{
			From: int(args.From.Unix()),
			To:   int(args.To.Unix()),
			Id:   chk.ID,
		})
		if err != nil {
			log.Printf("outage list failed (%v), %v", chk.ID, err)
			continue
		}

		j, _ := json.Marshal(outs)

		tags := []string{"down", chk.Hostname}
		for _, t := range chk.Tags {
			tags = append(tags, t.Name)
		}
		sort.Strings(tags)

		for _, out := range outs.Summary.States {
			switch out.Status {
			case "down":
				anns = append(anns,
					simplejson.Annotation{
						Time:    time.Unix(out.From, 0),
						TimeEnd: time.Unix(out.To, 0),
						Title:   chk.Name,
						Text:    chk.Hostname,
						Tags:    tags,
					},
				)
			default:
				continue
			}
		}
	}
	return anns, nil
}

var (
	respThreshDesc = prometheus.NewDesc(
		"pingdom_check_response_threshold_seconds",
		"The alert threshold for thi check",
		[]string{"name", "hostname"}, nil,
	)
	statusDesc = prometheus.NewDesc(
		"pingdom_check_status_bool",
		"",
		[]string{"name", "hostname", "status"}, nil,
	)
	lastErrorTs = prometheus.NewDesc(
		"pingdom_check_last_error_timestamp",
		"Timestamp of the last error from a check",
		[]string{"name", "hostname"}, nil,
	)
	lastTestTs = prometheus.NewDesc(
		"pingdom_check_last_test_timestamp",
		"Timestamp of the last test",
		[]string{"name", "hostname"}, nil,
	)
	lastResponse = prometheus.NewDesc(
		"pingdom_check_response_duration_seconds",
		"Time taken for the last check.",
		[]string{"name", "hostname"}, nil,
	)
)

// Describe implements prometheus.Collector
func (p *Pingdom) Describe(ch chan<- *prometheus.Desc) {
	ch <- respThreshDesc
	ch <- statusDesc
	ch <- lastErrorTs
	ch <- lastTestTs
	ch <- lastResponse
}

// Collect implements prometheus.Collector
func (p *Pingdom) Collect(ch chan<- prometheus.Metric) {
}

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	client := pingdom.NewClient(
		os.Getenv("EMAIL"),
		os.Getenv("PASSWORD"),
		os.Getenv("APIKEY"))

	p := &Pingdom{client.Checks}
	gsj := simplejson.New(
		simplejson.WithAnnotator(p),
	)

	log.Println("Start server")
	go func() {
		if err := http.ListenAndServe(":8080", gsj); err != nil {
			log.Fatalf("failed running server, %v", err)
		}
	}()
	<-stop
	log.Println("Stopped server")
}
