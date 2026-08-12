# Sidecar Pattern

Hands-on implementation of the **Sidecar Pattern** from *Designing Distributed Systems*.

## Architecture

```text
┌─────────────────────────┐
│     Inventory           │
│        │                │
│        ▼                │
│ inventory-sidecar       │
└─────────────────────────┘

┌─────────────────────────┐
│      Order              │
│        │                │
│        ▼                │
│ order-sidecar           │
└─────────────────────────┘

┌─────────────────────────┐
│      Payment            │
│        │                │
│        ▼                │
│ payment-sidecar         │
└─────────────────────────┘
```

Each service runs independently and sends logging events to a dedicated Sidecar process.

## What I Learned

* Sidecar separation from core business logic
* Shared Sidecar client
* Async logging
* Worker pool
* Buffered channel / backpressure
* Graceful shutdown
* Worker draining with WaitGroup
* Dockerized services
* Container networking
* Per-service Sidecar instances
* Configuration through environment variables
* Persistent log storage

## Run

```bash
docker compose up --build
```

Services:

* Inventory: `localhost:8082`
* Order: `localhost:8080`
* Payment: `localhost:8081`
* Sidecar: `localhost:8888`

Logs are persisted to:

```text
./data/logs/
```

## Key Experiment

Inside a container:

```text
127.0.0.1 → current container (container itself :) )
sidecar    → Sidecar container
```

Docker Compose provides service-name-based networking between containers.
