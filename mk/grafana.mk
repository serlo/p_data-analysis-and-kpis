#
# Describes operations on the grafana instance.
#

export grafana_user ?= admin

ifeq ($(env_name),minikube)
	export grafana_host ?= https://stats.serlo.local
	export grafana_password ?= admin
	export grafana_serlo_password ?= serlo
endif
ifeq ($(env_name),dev)
	export grafana_host ?= http://stats.serlo-development.dev
endif
ifeq ($(env_name),staging)
	export grafana_host ?= https://stats.serlo-staging.dev
endif
ifeq ($(env_name),production)
	export grafana_host ?= https://stats.serlo.dev
endif

ifneq ($(env_name),minikube)
	export grafana_password ?= $(shell cat $(infrastructure_repository)/live/$(env_name)/secrets/terraform-$(env_name).tfvars | grep kpi_grafana_admin_password | awk '{ print $$3}' | sed 's/\"//g')
	export grafana_serlo_password ?= $(shell cat $(infrastructure_repository)/live/$(env_name)/secrets/terraform-$(env_name).tfvars | grep kpi_grafana_serlo_password | awk '{ print $$3}' | sed 's/\"//g')
endif


GRAFANA_CURL_AUTH=-k -u $(grafana_user):$(grafana_password) -s -S

.PHONY: grafana_setup
# setup grafana on a new cluster
grafana_setup: grafana_restore_dashboards grafana_add_default_users grafana_set_preferences

.PHONY: grafana_backup_dashboards
# download grafana dashboards to the repository
grafana_backup_dashboards: 
	bash scripts/backup-dashboard.sh

.PHONY: grafana_restore_dashboards
# load grafana dashboards to $grafana_host
grafana_restore_dashboards:
	bash scripts/restore-dashboard.sh

.PHONY: grafana_add_default_users
# add the default users to grafana
grafana_add_default_users:
	curl $(GRAFANA_CURL_AUTH) \
		-XGET $(grafana_host)/api/users \
	| grep serlo >/dev/null && echo "user serlo already created" || \
	curl $(GRAFANA_CURL_AUTH) \
		-XPOST \
		-H 'Content-Type: application/json' \
		-d "{\"name\":\"serlo\",\"email\":\"kpi-user@serlo.org\",\"login\":\"serlo\",\"password\":\"$(grafana_serlo_password)\"}" \
	$(grafana_host)/api/admin/users >/dev/null && echo "user serlo created"

.PHONY: grafana_set_preferences
# set timezone browser and home dashboard author activity
grafana_set_preferences:
	#get dashboard by uid of author activity
	curl $(GRAFANA_CURL_AUTH) \
		-XPUT  \
		-H 'Accept: application/json' \
		-H 'Content-Type: application/json' \
		-d "{\"theme\":\"\", \
			\"homeDashboardId\":$$(curl $(GRAFANA_CURL_AUTH) \
					-XGET \
					-H 'Accept: application/json' \
					-H 'Content-Type: application/json' \
					$(grafana_host)/api/dashboards/uid/yS5BVkWZk \
				| jq '.dashboard.id'),\
			\"timezone\":\"browser\" \
		}" \
		$(grafana_host)/api/org/preferences

