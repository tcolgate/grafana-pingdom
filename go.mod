module github.com/tcolgate/grafana-pingdom

require (
	github.com/prometheus/client_golang v0.9.2
	github.com/russellcardullo/go-pingdom v0.0.0-20181021024747-0897d314d9a6
	github.com/tcolgate/grafana-simple-json-go v0.9.7
)

replace github.com/russellcardullo/go-pingdom => github.com/tcolgate/go-pingdom v0.0.0-20181205204318-ff8ac1b65457

go 1.13
