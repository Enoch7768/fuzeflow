# Architecture

## Phase 1 boundary

```
                         FuzeFlow
                            |
             +--------------+--------------+
             |              |              |
             v              v              v
           Server         Worker           CLI
             |
             v
          HTTP API
             |
       +-----+------+
       |            |
       v            v
 PostgreSQL       Redis
```

The API owns HTTP concerns. Future workflow and execution packages will own domain behavior rather than being embedded in handlers.

PostgreSQL is the durable source of record. Redis will be used for transient coordination and queues once the execution subsystem is introduced.

## Phase boundaries

Phase 1 does not implement workflow execution. This keeps the foundation testable and prevents premature abstractions around queues, workers, and integrations.
