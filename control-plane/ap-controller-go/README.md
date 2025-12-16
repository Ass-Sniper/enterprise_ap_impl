# ap-controller-go

Go-based implementation of the AP Controller.

## Design principles

- controller.yaml driven (no hard-coded logic)
- Redis-backed Session Schema v2
- role_rules with priority + wildcard support
- policy_version aware (AP-safe updates)
- HMAC-signed audit logs
- batch_status API for portal-sync.sh

## Build

```bash
docker build -t ap-controller-go .
```

## Run

```bash
docker run \
  -p 8443:8443 \
  -e AUDIT_SECRET=change_me \
  ap-controller-go
```

## Notes

- Python implementation remains untouched

- Data-plane scripts (portal-fw.sh / portal-sync.sh) unchanged

- docker-compose.yml can switch implementations safely
