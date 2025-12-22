# =========================
# 基本配置
# =========================
APP_NAME := access_device
CMD_DIR  := ./cmd/access_device
GO       := go
TAGS     := pap

# =========================
# 默认目标
# =========================
.PHONY: all
all: tidy clean build

# =========================
# Go Modules 整理
# =========================
.PHONY: tidy
tidy:
	@echo "==> go mod tidy"
	$(GO) mod tidy

# =========================
# 清理 Go 编译缓存
# =========================
.PHONY: clean
clean:
	@echo "==> go clean -cache"
	$(GO) clean -cache

# =========================
# 编译（PAP）
# =========================
.PHONY: build
build:
	@echo "==> go build -tags $(TAGS) -o $(APP_NAME) $(CMD_DIR)"
	$(GO) build -tags "$(TAGS)" -o $(APP_NAME) $(CMD_DIR)

# =========================
# 运行（本地调试）
# =========================
.PHONY: run
run: build
	@echo "==> run $(APP_NAME)"
	./$(APP_NAME)

# =========================
# 全量编译（pap + token + sms）
# =========================
.PHONY: build-all
build-all:
	@echo "==> go build -tags pap,token,sms -o $(APP_NAME) $(CMD_DIR)"
	$(GO) build -tags "pap token sms" -o $(APP_NAME) $(CMD_DIR)

# =========================
# 清理产物
# =========================
.PHONY: distclean
distclean: clean
	@echo "==> remove binary"
	rm -f $(APP_NAME)

