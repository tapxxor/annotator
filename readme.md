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

## Receiver 

### Run with go command

```bash
cd annotator/receiver
go run receiver.go --config=../config/annotator.yml
```

## Server

### Run with go command

```bash
cd annotator/server
go run server.go --config=../config/annotator.yml
```

## Alertmanager

Configure the webhook_config receiver of alertmanager as below:

```yaml
route:
  receiver: annotator
  repeat_interval: 24h
  group_wait: 1s
  group_interval: 1s
  group_by: ['...']
```

## Sqlite

Alert status is persisted using Sqlite db. Server _init_ function creates a table with name _annotations_
and schema : 

| Column      | Type   | Description                       |
|-------------|--------|-----------------------------------|
| alert_hash  | TEXT   | sha256(StartsAt + GroupKey)       | 
| starts_id   | BIGINT | starting annotation id            |
| starts_at   | BIGINT | starting annotation timestamp     |
| ends_id     | BIGINT | ending annotation id              |
| ends_at     | BIGINT | ending annotation timestamp       |
| region_id   | BIGINT | region id of annotation (not used)|
| alertname   | TEXT   | alert name                        |
| description | TEXT   | alert description                 |
| status      | TEXT   | alert status                      |

Alert status can be one of the following:

  * init    : alert has been inserted to db from ScanF routine
  * firing  : after alert's annotation has been posted to Grafana from Post routine
  * resolved: alert resolution has been identified from ScanR routine
  * created : region annotation update time has been updated from Update routine
  * staled  : region annotation's end time is marked as older than 1 month from Delete routine



