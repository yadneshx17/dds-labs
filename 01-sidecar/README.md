# Sidecar Pattern

Hands-on implementation of the **Sidecar Pattern** from *Designing Distributed Systems*.

## Architecture

```text
┌─────────────────────────┐
│ Inventory               │
│        │                │
│        ▼                │
│ inventory-sidecar       │
└─────────────────────────┘

┌─────────────────────────┐
│ Order                   │
│        │                │
│        ▼                │
│ order-sidecar           │
└─────────────────────────┘

┌─────────────────────────┐
│ Payment                 │
│        │                │
│        ▼                │
│ payment-sidecar         │
└─────────────────────────┘
```

Each service runs independently and sends logging events to a dedicated Sidecar process.

## What I Learned

* Sidecar pattern and separation of core/non-core functionality
* Shared Sidecar client
* Asynchronous logging
* Worker pool with buffered channels
* Backpressure and bounded concurrency
* Graceful shutdown
* `WaitGroup` for draining background workers
* Docker container networking
* Docker Compose service discovery
* `localhost` vs container-to-container networking
* Bind mounts for persistent logs

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
127.0.0.1 → current container
sidecar    → Sidecar container
```

Docker Compose provides service-name-based networking between containers.
