

```
docker build . -t log-analyzer:0.0.0
kind load docker-image docker.io/library/log-analyzer:0.0.0
kubectl delete po log-analyzer -n logging
kubectl run log-analyzer --image=log-analyzer:0.0.0 -n logging

kubectl expose pod log-analyzer --port=8080 --target-port=8080 --name=log-analyzer -n logging
kubectl logs log-analyzer -n logging -f
```

```
helm upgrade --install fluent-bit fluent/fluent-bit -f fluentbit-values.yaml  -n logging
```

```
ab -c 500 -n 500 http://localhost:8080/ingest
```
