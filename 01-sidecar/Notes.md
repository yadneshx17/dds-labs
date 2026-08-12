# Pattern: Sidecar

## Problem

Some functionality is useful to multiple services but isn't part of their core business logic.

**Examples**:
- Logging
- Metrics
- Tracing
- Security
- Configuration

## Idea

Run the supporting functionality as a companion process alongside the
application.

The application owns the core business logic.
The Sidecar owns the supporting/cross-cutting functionality.


## Architecture

Inventory ──► Inventory Sidecar
Order     ──► Order Sidecar
Payment   ──► Payment Sidecar

*The Sidecar has the same deployment/lifecycle unit as the application
it supports.*

## Why?

The application doesn't need to own logging infrastructure.

## Implementation (What I Built)

```text
Application
    ↓
LogEvent
    ↓
buffered channel
    ↓
worker pool
    ↓
Sidecar HTTP API
    ↓
log file
```

## Problems I Encountered

1. **One goroutine per request** — Initially every request created a goroutine to send logs.
   - Prob: background work could be lost during process shutdown
   - Fix: worker pool + `WaitGroup`

2. **Producer faster than logging** 
   - Prob: Producer can generate logs faster than workers can process them.
   - Fix: buffered channel → bounded queue → backpressure

3. **Server killed/Terminated during active requests** 
   - Prob: Immediate termination can interrupt in-flight requests.
   - Fix: graceful shutdown → `srv.Shutdown(context)`

4. **Background workers still running after shutdown**
   - Prob: `http.Server.Shutdown()` doesn't automatically wait for
   application-created goroutines.
   - Fix: `close(channel)` → workers drain remaining events → `wg.Wait()` → process exits

5. **`localhost` inside Docker**
   - Fix: `localhost` means the current container → use Docker service name

6. **Centralized logger wasn't actually Sidecar**
   - Fix: each service needs its own companion Sidecar
   - Initially:
        ```
        Order ───────┐
        Payment ─────┼──► One Logger
        Inventory ───┘
        ```
        This is centralized logging.
        
    - Correct Sidecar topology:
        ```
        Order     ──► Order Sidecar
        Payment   ──► Payment Sidecar
        Inventory ──► Inventory Sidecar
        ```

## Trade-offs

Pros:

- isolates cross-cutting functionality
- application code stays simpler
- reusable Sidecar implementation
- Sidecar can evolve independently

Cons:

- extra process/container
- communication overhead
- more deployment complexity
- failure handling between app and Sidecar

## 2-Minute Explanation (not ai generated apparently no one cares)

The Sidecar pattern is used when an application needs supporting
functionality that isn't part of its core business logic.

Instead of implementing that functionality directly inside the
application, we deploy it as a companion process alongside the
application.

In my implementation, Order, Payment and Inventory each have their
own logging Sidecar. The services create structured log events and
send them asynchronously through a buffered channel to a worker pool.
Workers forward those events to the Sidecar over HTTP.

I also implemented graceful shutdown so active requests can finish
and queued logging work can be drained before the process exits.

Docker Compose provides the deployment and service-to-sidecar
networking.

## Recall Questions

- What makes something a Sidecar?
- Why isn't a centralized logging service a Sidecar?
- Why use a worker pool?
- Why buffer the channel?
- What happens when the buffer is full?
- Why do we need graceful shutdown?
- Why doesn't `srv.Shutdown()` wait for my worker goroutines?
- Why is `127.0.0.1` wrong for container-to-container communication?
- What are the trade-offs of Sidecar?
