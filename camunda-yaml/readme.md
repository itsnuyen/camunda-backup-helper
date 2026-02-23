# Quick Install

```shell
helm install minio oci://registry-1.docker.io/bitnamicharts/minio -f ./minio-local.yaml
```


```shell

helm repo add camunda https://helm.camunda.io

helm repo update

helm upgrade --install camunda camunda/camunda-platform --version 13.4.2 --values=./camunda-yaml/values-local.yaml
```