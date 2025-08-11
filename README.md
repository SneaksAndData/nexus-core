![coverage](https://raw.githubusercontent.com/SneaksAndData/nexus-core/badges/.badges/main/coverage.svg)

# Introduction
Nexus Core is a core library used by Nexus (Go) application stack:
- [Nexus](https://github.com/SneaksAndData/nexus)
- [Nexus Receiver](https://github.com/SneaksAndData/nexus-receiver)
- [Nexus Supervisor](https://github.com/SneaksAndData/nexus-supervisor)
- [Nexus Configuration Controller](https://github.com/SneaksAndData/nexus-configuration-controller)
- [Nexus SDK - Go](https://github.com/SneaksAndData/nexus-sdk-go)
- [Nexus CRD](https://github.com/SneaksAndData/nexus-crd)

Nexus SDKs for other languages like [Python](https://github.com/SneaksAndData/nexus-sdk-py) also utilize Nexus Core indirectly. Nexus Core provides:
- CRD API types and generated clients
- Service API used by Nexus: Cassandra, Blob Storage, Kubernetes helper methods
- Vendor-specific implementations for service API
- Utility methods for Kubernetes interactions
- Simple actor framework which uses Kubernetes `workqueue` + channels as an actor's mailbox implementations and primary communication mechanism.


## Development Notes

Since Nexus Core provides backbone functionality to other Nexus applications, issues found in those will most likely require a patch in Core. If you wish to use Nexus and help develop the Core, please open an issue and tag project maintainers.
