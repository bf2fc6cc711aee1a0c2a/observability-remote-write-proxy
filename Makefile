prom1:
	docker run --network=host -v $(shell pwd)/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus --web.enable-remote-write-receiver --config.file=/etc/prometheus/prometheus.yml

prom2:
	docker run --network=host -v $(shell pwd)/prometheus2.yml:/etc/prometheus/prometheus.yml prom/prometheus --web.listen-address=:9091 --web.enable-remote-write-receiver --config.file=/etc/prometheus/prometheus.yml
