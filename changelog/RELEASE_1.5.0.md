# Release 1.5.0

## Key Highlights

### 1. Unified & Generalized Checkpointing (Cassandra & S3)
* **Generalization of Checkpoint Store:** Abstracted the checkpoint persistence (`pkg/checkpoint/store`) into a generic interface decoupled from payload persistence. It will be refered to as `Checkpointing API` in future releases.
* **Cassandra Checkpointing:** Cassandra code moved to a separate implementation of a `Checkpointing API`, `cassandra_store.go`. 
* **Cassandra Vendor Refactor:** Following the refactor of a `Checkpointing API`, several implementations of `cassandra_store.go` where added.
  - **ScyllaDB** ported to `cassandra_store`.
  - **AstraDB** ported to `cassandra_store`.
  - **AWS Keyspaces (NEW VENDOR)** added as a `cassandra_store` implementation.
  - **Indexed Store**: Added a store that utilizes SAI via `cassandra_store_indexed.go`, backed by `cassandra_store.go`. This was previosuly the only supported implementation.
  - **Bare Store (NEW):** Added `cassandra_store_bare.go` that supports non-indexed Cassandra deployments, for example AWS Keyspaces.
  - Added required schema migrations/setups in `test-resources/scylla-config` (e.g., `checkpoints_by_host.cql`, `checkpoints_by_tag.cql`, `checkpoints_indexed.cql`), to ensure code coverage of all new functionality.
* **Lightweight Payload Storage:** Default checkpoint buffer implementation now supports `SERIALIZE_TO_BACKEND` payload configuration. When enabled, payload will be compressed and saved directly in a corresponding Cassandra table. 
* **BUGFIX: JSON data in Cassandra values** Base64 encoding is now used for serializing JSON fields to prevent data corruption when saving to Cassandra engines that treat JSON as a non-text type.

### 2. Payload Storage & Security Utilities
* **Abstract Payload Store:** Decoupled payload persistence from blob storage and S3 in particular, enabling usage of non-S3 backends for payload persistence.
  - `MemoryPassThroughBuffer` can now be used for end-to-end local testing.
  - Payload publishing is not longer exclusively handled by S3 presign. See below for more info.
* **Payload Proxy Configuration:** Custom payload proxy URL can now be provided, instead of S3 pre-signed URLs used before. This allows developers to deploy Nexus payload persistence on a non-blob, non-S3 backends.
  - **Secure URL Signing:** Added cryptographic URL signing utilities under a new `pkg/urlsign` package (supporting key management, signing, and verification).
  - **BREAKING:** `PayloadUri` will now always refer to Nexus Scheduler address rather than a S3 endpoint. This allows `SERIALIZE_TO_BACKEND` and `SERIALIZE_TO_S3` to work transiently for the client. Note this requires a scheduler to provide correct public base address and serve path through configuration. Any CORS or trust policies targeting 1.4.x S3 addresses will not work with this version!

### 3. Additional features for algorithm templates
* **Config, Secret & Volume Mapping:** Expanded the `NexusAlgorithm` Kubernetes spec to support mounting of ConfigMaps, Secrets, and persistent storage volumes via persistent volume claims. Algorithm developers can now mount additional storage to their containers if needed.
* **Self-Healing Pipeline Buffers:** Automatically resolve pipeline buffer processing failures directly to `LifecycleStageFailed`.
* **BREAKING: Transient Backend Errors:** "Not Found" responses from a checkpoint backend will no longer raise an exception. Clients will receive NIL result and must handle it as a non-fatal error.

### 4. Shard Management & Networking
* **In-Cluster Shard Client:** Added support for in-cluster shard client, intended to be used for testing and local development.

### 5. Utilities & DevX Upgrades
* **Transition to Just:** Retired dependency on Makefiles/bash scripts in several workflows by adopting a unified `justfile` for automating codebase tasks (lint, generate, run-scylla, test, etc.).
* **Improved Testing Infrastructure:** Used context poll loops in buffer tests to eliminate flaky/timed-out behaviors, and dramatically increased Cassandra/Scylla test coverage.
