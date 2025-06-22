.DEFAULT_GOAL = build

.PHONY: build
build:
	mkdocs build

.PHONY: serve
serve:
	mkdocs serve

.PHONY: publish
publish:
	mkdocs gh-deploy
