#!/bin/sh

exxport EMAIL=infra@qubit.com
exxport PASSWORD={{ lookup "password" }}
exxport APIKEY={{ lookup "apiKey" }}

exec /bin/grafana-pingdom
