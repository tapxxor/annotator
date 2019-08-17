# Annotator

Annotator consists of two modules the receiver and the server.

- A webhook_config in configured in alertmanager in order to send every alert to receiver.
  Receiver saves the alert payload on disk and acknoledges to the alertmanager. Alertmanager
  retransmissions ensure that even if the receiver is unavailable, when it will become 
  active it will get all the missing alerts.
- Annotator server scans periodically the disk for saved alerts and through API calls to 
  Grafana creates or updates annotations.

```yaml
    ┌──────────────┐     ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
    │ Alertmanager │     │   Receiver  │    │    Server   │    │   Grafana   │
    └──────────────┘     └─────────────┘    └─────────────┘    └─────────────┘
           ╲                  ╱ |                ╱ ╲                   ╱ 
            ╲_POST_alert ____╱  |               ╱   ╲_POST_annotation_╱
                                | Write        ╱ Read
                                |             ╱
        ┌──────┬─────────┬─────────┬─────┬─────────┐
        │ Disk │ alert 1 │ alert 2 │ ... │ alert x │
        └──────┴─────────┴─────────┴─────┴─────────┘
```

## Run prometheus

```bash
docker volume create prometheus_data

docker run --name prometheus -d -p 9090:9090 \
--restart=always \
-v prometheus_data:/prometheus \
-v "$(pwd)"/config:/prometheus_config \
-v "$(pwd)"/rules:/etc/prometheus/rules \
prom/prometheus \
--web.enable-lifecycle \
--config.file=/prometheus_config/prometheus.yml \
--storage.tsdb.retention=7d
```

## Run alertmanager

```bash
docker run --name alertmanager -d -p 9093:9093 \
--restart=always \
-v $(pwd)/config/:/alertmanager_config \
prom/alertmanager:master \
--config.file=/alertmanager_config/alertmanager.yml 
```

## Run grafana

```bash
docker volume create grafana-data
docker run -d --name=grafana -p 3000:3000 \
--restart=always \
-v grafana-data:/var/lib/grafana \
grafana/grafana
```

## Run sqlite
docker run -d --name sqlite3 --restart=always \
-v `pwd`/db:/root/db \
nouchka/sqlite3 annotations.db


```bash
curl -H "Authorization: Bearer eyJrIjoiYVZGZkJ4VktxRjRzcDBDbkZNY0JtYW5RWW4zZ3JrTEgiLCJuIjoiYW5ub3RhdG9yIiwiaWQiOjF9" http://localhost:3000/api/dashboards/home
```

curl -X POST \
-H "Authorization: Bearer eyJrIjoiYVZGZkJ4VktxRjRzcDBDbkZNY0JtYW5RWW4zZ3JrTEgiLCJuIjoiYW5ub3RhdG9yIiwiaWQiOjF9" \
-H "Content-Type: application/json" --data @test/ann1.json \
http://localhost:3000/api/annotations


curl -X POST \
-H "Authorization: Bearer eyJrIjoiYVZGZkJ4VktxRjRzcDBDbkZNY0JtYW5RWW4zZ3JrTEgiLCJuIjoiYW5ub3RhdG9yIiwiaWQiOjF9" \
-H "Content-Type: application/json" --data @test/ann2.json \
http://localhost:3000/api/annotations


curl -X POST \
-H "Authorization: Bearer eyJrIjoiYVZGZkJ4VktxRjRzcDBDbkZNY0JtYW5RWW4zZ3JrTEgiLCJuIjoiYW5ub3RhdG9yIiwiaWQiOjF9" \
-H "Content-Type: application/json" --data @test/ann3.json \
http://localhost:3000/api/annotations

curl -X POST \
-H "Authorization: Bearer eyJrIjoiYVZGZkJ4VktxRjRzcDBDbkZNY0JtYW5RWW4zZ3JrTEgiLCJuIjoiYW5ub3RhdG9yIiwiaWQiOjF9" \
-H "Content-Type: application/json" --data @test/ann4.json \
http://localhost:3000/api/annotations

# Request
POST /api/annotations HTTP/1.1
Accept: application/json
Content-Type: application/json
{
  "time":1565446815000,
  "isRegion":true,
  "timeEnd":1565447070000,
  "tags":["tag1","tag2"],
  "text":"Annotation Description"
}

Response
HTTP/1.1 200
Content-Type: application/json

{
    "message":"Annotation added",
    "id": 1,
    "endId": 2
}



curl -X GET \
-H "Authorization: Bearer eyJrIjoiYVZGZkJ4VktxRjRzcDBDbkZNY0JtYW5RWW4zZ3JrTEgiLCJuIjoiYW5ub3RhdG9yIiwiaWQiOjF9" \
http://localhost:3000/api/annotations