# ============================================================
# Control Plane Makefile
# ============================================================

COMPOSE := docker compose

.DEFAULT_GOAL := help

# ------------------------------------------------------------
# Help
# ------------------------------------------------------------
help:
	@echo ""
	@echo "Control Plane Makefile"
	@echo ""
	@echo "Common targets:"
	@echo "  make up                     Start all services"
	@echo "  make down                   Stop and remove all services"
	@echo "  make build                  Build all services"
	@echo "  make build-nc               Build all services (no cache)"
	@echo ""
	@echo "Service lifecycle:"
	@echo "  make restart-<svc>          Stop/rm/build/up service"
	@echo "  make restart-nc-<svc>       Same as above, no cache"
	@echo ""
	@echo "Logs & exec:"
	@echo "  make logs-<svc>             Tail logs"
	@echo "  make sh-<svc>               Exec shell"
	@echo ""
	@echo "Examples:"
	@echo "  make restart-freeradius"
	@echo "  make restart-nc-freeradius"
	@echo "  make logs-freeradius"
	@echo ""

# ------------------------------------------------------------
# Global lifecycle
# ------------------------------------------------------------
up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

build:
	$(COMPOSE) build

build-nc:
	$(COMPOSE) build --no-cache

ps:
	$(COMPOSE) ps

# ------------------------------------------------------------
# Per-service lifecycle
# ------------------------------------------------------------
restart-%:
	$(COMPOSE) stop $* || true
	$(COMPOSE) rm -f $* || true
	$(COMPOSE) build $*
	$(COMPOSE) up -d $*

restart-nc-%:
	$(COMPOSE) stop $* || true
	$(COMPOSE) rm -f $* || true
	$(COMPOSE) build --no-cache $*
	$(COMPOSE) up -d $*

# ------------------------------------------------------------
# Logs / shell
# ------------------------------------------------------------
logs-%:
	$(COMPOSE) logs -f $*

sh-%:
	$(COMPOSE) exec $* sh

bash-%:
	$(COMPOSE) exec $* bash

# ------------------------------------------------------------
# FreeRADIUS helpers
# ------------------------------------------------------------
radius-test:
	radtest testuser testpass 127.0.0.1 0 testing123

radius-debug:
	$(COMPOSE) exec freeradius freeradius -X
