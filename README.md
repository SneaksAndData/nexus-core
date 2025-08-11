![coverage](https://raw.githubusercontent.com/SneaksAndData/nexus-core/badges/.badges/main/coverage.svg)

# Introduction
Nexus Core is a core library used by Nexus (Go) application stack:
- [Nexus](https://github.com/SneaksAndData/nexus)
- [Nexus Configuration Controller](https://github.com/SneaksAndData/nexus-configuration-controller)
- [Nexus SDK - Go](https://github.com/SneaksAndData/nexus-sdk-go)
- [Nexus CRD](https://github.com/SneaksAndData/nexus-crd)

Nexus SDKs for other languages (Python) also utilize Nexus Core indirectly. Nexus Core provides:
- CRD API types and generated clients
- Service API used by Nexus: Cassandra, Blob Storage, Kubernetes helper methods

## Testing

Note that some tests require a `nexus-crd` chart to be installed, that depends on `nexus-core`. When testing changing, clone `nexus-crd`, update the code and build a dev version. Then, update `Chart.yaml` with a new version and run `cd helm && helm dependency update .`.

Afterwards, run `helm upgrade nexus-test-stack --namespace nexus --kube-context kind-nexus-controller` before running tests.
