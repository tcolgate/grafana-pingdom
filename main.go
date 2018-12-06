package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/russellcardullo/go-pingdom/pingdom"
	simplejson "github.com/tcolgate/grafana-simple-json-go"
)

type Pingdom struct {
	client *pingdom.Client
}

func (p *Pingdom) GrafanaAnnotations(ctx context.Context, query string, args simplejson.AnnotationsArguments) ([]simplejson.Annotation, error) {
	chks, err := p.client.Checks.List()
	if err != nil {
		return nil, err
	}

	var anns []simplejson.Annotation
	for _, chk := range chks {
		outs, err := p.client.Checks.SummaryOutage(pingdom.SummaryOutageRequest{
			From: int(args.From.Unix()),
			To:   int(args.To.Unix()),
			Id:   chk.ID,
		})
		if err != nil {
			log.Printf("outage list failed (%v), %v", chk.ID, err)
			continue
		}

		tags := []string{"down"}
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
						Text:    chk.Name,
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

func main() {
	client := pingdom.NewClient(
		os.Getenv("EMAIL"),
		os.Getenv("PASSWORD"),
		os.Getenv("APIKEY"))

	p := &Pingdom{client}
	gsj := simplejson.New(
		simplejson.WithAnnotator(p),
	)

	http.ListenAndServe(":8080", gsj)
}
