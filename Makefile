.PHONY: kind-build clean

kind-build:
	docker build . -t log-analyzer:0.0.0
	kind load docker-image docker.io/library/log-analyzer:0.0.0
	-kubectl delete po log-analyzer -n logging
	kubectl run log-analyzer --image=log-analyzer:0.0.0 -n logging

kind-run:
	kubectl run log-analyzer --image=log-analyzer:0.0.0 -n logging

kind-clean:
	kubectl delete po log-analyzer -n logging

kind-log:
	kubectl logs log-analyzer -n logging -f