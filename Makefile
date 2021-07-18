.PHONY: run prometheus

run:
	go run main.go

prometheus:
	curl localhost:8773/prometheus
