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

### Run locally

```bash
cd annotator/receiver
go run receiver.go --config=../config/annotator.yml
```

or 

```bash
TODO
```

## Server

```bash
cd annotator/server
go run server.go --config=../config/annotator.yml
```

or

```bash
TODO
```