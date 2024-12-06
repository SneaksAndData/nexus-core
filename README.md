![coverage](https://raw.githubusercontent.com/SneaksAndData/nexus-configuration-controller/badges/.badges/main/coverage.svg)

# Introduction
Nexus Core is a core library used by Nexus (Go) application stack:
- [Nexus](https://github.com/SneaksAndData/nexus)
- [Nexus Configuration Controller](https://github.com/SneaksAndData/nexus-configuration-controller)
- [Nexus SDK - Go](https://github.com/SneaksAndData/nexus-sdk-go)
- [Nexus CRD](https://github.com/SneaksAndData/nexus-crd)

Nexus SDKs for other languages (Python) also utilize Nexus Core indirectly. Nexus Core provides:
- CRD API types and generated clients
- Service API used by Nexus: Cassandra, Blob Storage, Kubernetes helper methods
