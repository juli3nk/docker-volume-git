TARGETS := $(shell ls scripts | grep -vE 'clean|dev|help|release')

TMUX := $(shell command -v tmux 2> /dev/null)

.dapper:
	@echo Downloading dapper
	@curl -sL https://releases.rancher.com/dapper/latest/dapper-`uname -s`-`uname -m|sed 's/v7l//'` > .dapper.tmp
	@@chmod +x .dapper.tmp
	@./.dapper.tmp -v
	@mv .dapper.tmp .dapper

.tmass:
	@echo Downloading tmass
	@curl -sL https://github.com/juliengk/tmass/releases/download/0.3.0/tmass -o .tmass.tmp
	@@chmod +x .tmass.tmp
	@./.tmass.tmp version
	@mv .tmass.tmp .tmass

$(TARGETS): .dapper
	./.dapper $@

clean:
	@./scripts/clean

dev: .dapper .tmass
ifndef TMUX
	$(error "tmux is not available, please install it")
endif

	./.tmass load -l scripts/dev/tmux/ gitvol
	tmux a -d -t gitvol

help:
	@./scripts/help

release:
	./scripts/release

.DEFAULT_GOAL := ci

.PHONY: .dapper .tmass $(TARGETS) clean dev help release
