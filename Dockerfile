FROM alpine
COPY grafana-pingdom /bin/grafana-pingdom 
RUN apk update && apk add ca-certificates
ENTRYPOINT ["/bin/grafana-pingdom "]
