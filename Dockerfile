FROM alpine
COPY grafana-pingdom /bin/grafana-pingdom 

ENTRYPOINT ["/bin/grafana-pingdom "]
