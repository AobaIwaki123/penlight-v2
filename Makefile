.PHONY: help up down restart status logs up-back down-back restart-back up-front down-front restart-front logs-back logs-front verify test gen

RUN_DIR := .run
BACK_PID := $(RUN_DIR)/backend.pid
FRONT_PID := $(RUN_DIR)/frontend.pid
BACK_LOG := $(RUN_DIR)/backend.log
FRONT_LOG := $(RUN_DIR)/frontend.log
BACK_PORT := 8080
FRONT_PORT := 3000

help: ## コマンド一覧を表示
	@echo "利用可能な Make コマンド:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

up: up-back up-front ## バックエンドとフロントエンドを両方起動
	@$(MAKE) status

down: down-front down-back ## バックエンドとフロントエンドを両方停止
	@echo "全プロセスを停止しました。"

restart: down ## バックエンドとフロントエンドを両方再起動
	@sleep 1
	@$(MAKE) up

up-back: ## バックエンドを起動 (ポート: 8080)
	@mkdir -p $(RUN_DIR)
	@if lsof -Pi :$(BACK_PORT) -sTCP:LISTEN -t >/dev/null 2>&1 ; then \
		echo "バックエンドは既に起動しています (Port: $(BACK_PORT))"; \
	else \
		echo "バックエンドを起動しています (http://localhost:$(BACK_PORT))..."; \
		nohup go run ./cmd/server < /dev/null > $(BACK_LOG) 2>&1 & echo $$! > $(BACK_PID); \
		for i in 1 2 3 4 5 6 7 8 9 10; do \
			if lsof -Pi :$(BACK_PORT) -sTCP:LISTEN -t >/dev/null 2>&1 ; then \
				echo "バックエンドが正常に起動しました (Port: $(BACK_PORT))"; \
				break; \
			fi; \
			sleep 0.5; \
		done; \
		if ! lsof -Pi :$(BACK_PORT) -sTCP:LISTEN -t >/dev/null 2>&1 ; then \
			echo "バックエンドの起動に失敗しました。ログを確認してください: $(BACK_LOG)"; \
			cat $(BACK_LOG); \
			exit 1; \
		fi; \
	fi

down-back: ## バックエンドを停止
	@echo "バックエンドを停止しています..."
	@if [ -f $(BACK_PID) ]; then \
		kill $$(cat $(BACK_PID)) 2>/dev/null || true; \
		rm -f $(BACK_PID); \
	fi
	@lsof -ti :$(BACK_PORT) | xargs kill -9 2>/dev/null || true
	@echo "バックエンドを停止しました。"

restart-back: down-back ## バックエンドを再起動
	@sleep 1
	@$(MAKE) up-back

up-front: ## フロントエンドを起動 (ポート: 3000)
	@mkdir -p $(RUN_DIR)
	@if lsof -Pi :$(FRONT_PORT) -sTCP:LISTEN -t >/dev/null 2>&1 ; then \
		echo "フロントエンドは既に起動しています (Port: $(FRONT_PORT))"; \
	else \
		echo "フロントエンドを起動しています (http://localhost:$(FRONT_PORT))..."; \
		nohup npm run --prefix frontend dev < /dev/null > $(FRONT_LOG) 2>&1 & echo $$! > $(FRONT_PID); \
		for i in 1 2 3 4 5 6 7 8 9 10; do \
			if lsof -Pi :$(FRONT_PORT) -sTCP:LISTEN -t >/dev/null 2>&1 ; then \
				echo "フロントエンドが正常に起動しました (Port: $(FRONT_PORT))"; \
				break; \
			fi; \
			sleep 0.5; \
		done; \
		if ! lsof -Pi :$(FRONT_PORT) -sTCP:LISTEN -t >/dev/null 2>&1 ; then \
			echo "フロントエンドの起動に失敗しました。ログを確認してください: $(FRONT_LOG)"; \
			cat $(FRONT_LOG); \
			exit 1; \
		fi; \
	fi

down-front: ## フロントエンドを停止
	@echo "フロントエンドを停止しています..."
	@if [ -f $(FRONT_PID) ]; then \
		kill $$(cat $(FRONT_PID)) 2>/dev/null || true; \
		rm -f $(FRONT_PID); \
	fi
	@lsof -ti :$(FRONT_PORT) | xargs kill -9 2>/dev/null || true
	@echo "フロントエンドを停止しました。"

restart-front: down-front ## フロントエンドを再起動
	@sleep 1
	@$(MAKE) up-front

status: ## 起動状態を確認
	@echo "=== プロセスステータス ==="
	@if lsof -Pi :$(BACK_PORT) -sTCP:LISTEN -t >/dev/null 2>&1 ; then \
		echo "  Backend : \033[32mRUNNING\033[0m (PID: $$(lsof -ti :$(BACK_PORT) | tr '\n' ' ')) -> http://localhost:$(BACK_PORT)"; \
	else \
		echo "  Backend : \033[31mSTOPPED\033[0m"; \
	fi
	@if lsof -Pi :$(FRONT_PORT) -sTCP:LISTEN -t >/dev/null 2>&1 ; then \
		echo "  Frontend: \033[32mRUNNING\033[0m (PID: $$(lsof -ti :$(FRONT_PORT) | tr '\n' ' ')) -> http://localhost:$(FRONT_PORT)"; \
	else \
		echo "  Frontend: \033[31mSTOPPED\033[0m"; \
	fi

logs: ## 全ログをリアルタイム表示
	@tail -f $(BACK_LOG) $(FRONT_LOG) 2>/dev/null || echo "ログファイルがまだ作成されていません。"

logs-back: ## バックエンドログを表示
	@tail -f $(BACK_LOG)

logs-front: ## フロントエンドログを表示
	@tail -f $(FRONT_LOG)

verify: ## 全自動一括検証 (scripts/verify-all.sh)
	@./scripts/verify-all.sh

test: ## Go テスト実行
	@go test ./...

gen: ## スキーマ・ER図一括再生成
	@./scripts/generate-all.sh
