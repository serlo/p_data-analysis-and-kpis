#
# Makefile to automate the upload of current grafana dashboard
# 

grafana_host ?= https://stats.dev.serlo.local
grafana_user ?= admin
grafana_password ?= admin

.PHONY: dashb-backup
dashb-backup:
	curl -X GET -u $(grafana_user):$(grafana_password) -k "${grafana_host}/api/dashboards/db/author-activity" > dashboards/author-activity.json
	curl -X GET -u $(grafana_user):$(grafana_password) -k "${grafana_host}/api/dashboards/db/registrations" > dashboards/registrations.json


.PHONY: dashb-upload
dashb-upload:
	curl -X POST -u $(grafana_user):$(grafana_password) -k -H "Content-Type: application/json" --data-binary @./dashboards/author-activity.json "${grafana_host}/api/dashboards/db"
	curl -X POST -u $(grafana_user):$(grafana_password) -k -H "Content-Type: application/json" --data-binary @./dashboards/registrations.json "${grafana_host}/api/dashboards/db"

