.PHONY: up down down-all build logs links restart \
       up-infra up-services up-app \
       down-infra down-services down-app \
       build-infra build-services build-app \
       logs-infra logs-services logs-app logs-test \
       browser self-test test-pester test-pester-outside test-msal test-browser-playwright build-test \
       generate-certs clean clean-certs rotate-cert \
       setup-localhost clean-localhost install-box install-wiremock

export BOX_ROOT:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
DOCKER_ARCH_RAW:=$(shell docker version --format '{{.Server.Arch}}' 2>/dev/null || uname -m)
DOCKER_ARCH_NORMALIZED:=$(if $(filter x86_64 amd64,$(DOCKER_ARCH_RAW)),amd64,$(if $(filter aarch64 arm64,$(DOCKER_ARCH_RAW)),arm64,amd64))
PLATFORM?=$(if $(filter arm64,$(DOCKER_ARCH_NORMALIZED)),linux/amd64,linux/$(DOCKER_ARCH_NORMALIZED))
DOCKER_COMPOSE:=PLATFORM=$(PLATFORM) docker compose --project-directory $(BOX_ROOT) -f $(BOX_ROOT)/compose.yaml

# --- All ---

up:
	@if [ ! -f $(BOX_ROOT)/certs/tls-cert.pem ] || \
		[ ! -f $(BOX_ROOT)/certs/tls-key.pem ] || \
		[ ! -f $(BOX_ROOT)/certs/dev-root-ca.crt ] || \
		[ ! -f $(BOX_ROOT)/certs/identity-signing.key ]; then \
		$(MAKE) -C $(BOX_ROOT) build; \
	fi
	$(DOCKER_COMPOSE) up -d --remove-orphans --no-recreate
	@$(MAKE) -C $(BOX_ROOT) links

down:
	$(DOCKER_COMPOSE) down --remove-orphans

down-all:
	$(BOX_ROOT)/tests/wiremock/wiremock.ps1 stop
	@echo "Stopping all containers on infra-shared-net..."
	@containers=$$(docker ps -aq --filter network=infra-shared-net); \
	if [ -n "$$containers" ]; then \
		docker rm -f $$containers; \
	fi
	$(DOCKER_COMPOSE) --profile "*" down -v --remove-orphans
	@echo "Removing infra-shared-net network..."
	@for net in $$(docker network ls --filter name=infra-shared-net -q) $$(docker network ls --filter name=hostingcompose -q); do \
		docker network rm $$net 2>/dev/null || true; \
	done

build: generate-certs
	$(DOCKER_COMPOSE) build

logs:
	$(DOCKER_COMPOSE) logs -f

restart: down up

links:
	@echo "https://browser.web.internal"

# --- Infra (dns, entry, portal) ---

up-infra:
	@if [ ! -f $(BOX_ROOT)/certs/tls-cert.pem ] || \
		[ ! -f $(BOX_ROOT)/certs/tls-key.pem ] || \
		[ ! -f $(BOX_ROOT)/certs/dev-root-ca.crt ] || \
		[ ! -f $(BOX_ROOT)/certs/identity-signing.key ]; then \
		$(MAKE) -C $(BOX_ROOT) generate-certs; \
	fi
	$(DOCKER_COMPOSE) up -d --remove-orphans --no-recreate dns entry portal

down-infra:
	$(DOCKER_COMPOSE) stop dns entry portal
	$(DOCKER_COMPOSE) rm -f dns entry portal

build-infra:
	$(DOCKER_COMPOSE) build portal

logs-infra:
	$(DOCKER_COMPOSE) logs -f dns entry portal

# --- Services (entra, postgres, servicebus, sqledge, browser) ---

up-services:
	$(MAKE) -C $(BOX_ROOT) up-infra
	$(DOCKER_COMPOSE) up -d --no-recreate entra postgres servicebus sqledge browser

down-services:
	$(DOCKER_COMPOSE) stop entra postgres servicebus sqledge browser
	$(DOCKER_COMPOSE) rm -f entra postgres servicebus sqledge browser

build-services:
	$(DOCKER_COMPOSE) build entra browser

logs-services:
	$(DOCKER_COMPOSE) logs -f entra postgres servicebus sqledge browser

# --- App ---

up-app:
	$(MAKE) -C $(BOX_ROOT) up-services
	$(DOCKER_COMPOSE) up -d --no-recreate app

down-app:
	$(DOCKER_COMPOSE) stop app
	$(DOCKER_COMPOSE) rm -f app

build-app:
	$(DOCKER_COMPOSE) build app

logs-app:
	$(DOCKER_COMPOSE) logs -f app

# --- Browser ---

browser:
	$(MAKE) -C $(BOX_ROOT) up
	$(DOCKER_COMPOSE) up -d --no-recreate
	open -a "Google Chrome" "https://browser.web.internal"

# --- Test ---

self-test:
	$(MAKE) -C $(BOX_ROOT) generate-certs
	$(DOCKER_COMPOSE) --profile self-test build test msal-client browser-test
	$(DOCKER_COMPOSE) --profile self-test up -d --remove-orphans --no-recreate dns entry entra postgres servicebus browser portal
	$(DOCKER_COMPOSE) --profile self-test up -d --force-recreate --no-deps browser
	$(MAKE) -C $(BOX_ROOT) test-pester test-pester-outside test-browser-playwright test-msal

test-pester:
	trap '$(BOX_ROOT)/tests/wiremock/wiremock.ps1 stop' EXIT; \
	$(BOX_ROOT)/tests/wiremock/wiremock.ps1 start; \
	sleep 2; \
	$(DOCKER_COMPOSE) --profile self-test run --rm --no-deps test

test-pester-outside:
	pwsh -NoProfile -NonInteractive -Command '$$ErrorActionPreference = "Stop"; if (-not (Get-Command Invoke-Pester -ErrorAction SilentlyContinue)) { Set-PSRepository -Name PSGallery -InstallationPolicy Trusted; Install-Module -Name Pester -Force -AllowClobber -Scope CurrentUser }; Import-Module Pester -ErrorAction Stop; $$res = Invoke-Pester "$(BOX_ROOT)/tests/*Outside.Tests.ps1" -Output Detailed -PassThru; if ($$res.FailedCount -gt 0) { exit 1 }'

test-msal:
	sleep 2
	$(DOCKER_COMPOSE) --profile self-test run --rm --no-deps msal-client

test-browser-playwright:
	$(DOCKER_COMPOSE) --profile self-test run --rm --no-deps browser-test

build-test: install-wiremock
	$(DOCKER_COMPOSE) build test msal-client

logs-test:
	$(DOCKER_COMPOSE) --profile self-test logs -f test browser-test msal-client

# --- Certs ---

generate-certs:
	$(DOCKER_COMPOSE) --profile setup run --rm --build gencert

clean-certs:
	@echo "Cleaning certificates..."
	rm -rf $(BOX_ROOT)/certs/_ca
	rm -f $(BOX_ROOT)/certs/tls-cert.pem \
		$(BOX_ROOT)/certs/tls-key.pem \
		$(BOX_ROOT)/certs/dev-root-ca.crt \
		$(BOX_ROOT)/certs/identity-signing.key

rotate-cert: clean clean-certs generate-certs
	@if [ -f /etc/resolver/internal ] || [ -f "$$HOME/Library/Keychains/infra-localhost.keychain-db" ]; then \
		echo "Detected localhost setup. Cleaning and refreshing localhost certificate trust..."; \
		$(BOX_ROOT)/localhost/clean-dns-and-cert.sh; \
		$(BOX_ROOT)/localhost/setup-dns-and-cert.sh; \
	fi
	@echo "Certificates rotated successfully."

# --- Cleanup ---

clean:
	$(MAKE) -C $(BOX_ROOT) down-all
	@echo "Cleaning up infrastructure..."
	docker network prune -f
	docker volume prune -f

# --- Localhost ---

setup-localhost:
	$(DOCKER_COMPOSE) up -d --no-recreate dns
	$(BOX_ROOT)/localhost/setup-dns-and-cert.sh

clean-localhost:
	$(BOX_ROOT)/localhost/clean-dns-and-cert.sh

install-box:
	$(BOX_ROOT)/localhost/install-box.sh "$(BOX_ROOT)"

install-wiremock:
	$(BOX_ROOT)/tests/wiremock/wiremock.ps1 install
