# Camunda Backup Helper

## Create a camunda cluster

```shell
helm upgrade --install camunda camunda/camunda-platform --version 13.4.2 --values=./camunda-yaml/values-local.yaml
```

How to build

```
docker build --platform linux/amd64 -t camunda-backup-helper:0.0.2 .
docker tag camunda-backup-helper:0.0.2 itsnuyen/camunda-backup-helper:0.0.2
docker push itsnuyen/camunda-backup-helper:0.0.2
```

## Rest Interface 

### Create backup

### List backups

### Search Backup

### Perform Backup

### Cleanup Backup

### Snapshot Registration

## Online Tool