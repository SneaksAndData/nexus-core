set shell := ["bash", "-c"]

# images
SCYLLA_IMAGE := "scylladb/scylla"
MINIO_IMAGE  := "mojatter/s2-server"

# configurations
SCYLLA_CONFIG := invocation_directory() / "test-resources/scylla-config"

# Default recipe
fresh: stop up

# Start CI environment
up: scylla minio prepare-buckets prepare-scylla start-kind-cluster shards-kubeconfig

# Start ScyllaDB with health checks
scylla:
    @echo "🚀 Starting Scylla..."
    docker run -d \
      --name scylla \
      -p 9042:9042 \
      -p 10000:10000 \
      --health-cmd "nodetool statusgossip | grep -q 'running'" \
      --health-interval 5s \
      --health-retries 10 \
      {{SCYLLA_IMAGE}} \
      --listen-address 127.0.0.1 \
      --rpc-address 0.0.0.0 \
      --broadcast-rpc-address 0.0.0.0 \
      --smp 1 \
      --developer-mode 1

# Start Minio (s2-server image has no shell/curl, so it can't run container-internal health checks)
minio:
    @echo "📦 Starting Minio..."
    docker run -d \
      --name minio \
      -p 9000:9000 \
      -p 9123:9123 \
      --restart always \
      -e S2_SERVER_CONSOLE_LISTEN=:9123 \
      -e S2_SERVER_USER=minioadmin \
      -e S2_SERVER_PASSWORD=minioadmin \
      -e S2_SERVER_BUCKETS=tmp,nexus \
      {{MINIO_IMAGE}}

# Wait for Minio to become reachable (buckets are created at startup via S2_SERVER_BUCKETS)
prepare-buckets:
    @echo "⏳ Waiting for Minio..."
    until curl -sf http://localhost:9000/healthz > /dev/null; do sleep 2; done

# Run Scylla initialization
prepare-scylla:
    @echo "⏳ Waiting for Scylla..."
    until [ "$(docker inspect -f '{{ "{{" }}.State.Health.Status{{ "}}" }}' scylla)" == "healthy" ]; do sleep 2; done
    docker run --rm \
      --network host \
      -v {{SCYLLA_CONFIG}}:/opt/storage \
      --entrypoint /opt/storage/prepare-scylla.sh \
      {{SCYLLA_IMAGE}}

start-kind-cluster:
    kind create cluster --config=test-resources/kind.yaml --name nexus-shard-0

shards-kubeconfig:
    mkdir -p ./test-resources/kind && \
    kind export kubeconfig --name nexus-shard-0 --kubeconfig ./test-resources/kind/kind-nexus-shard-0.kubeconfig

# Cleanup CI environment
stop:
    @echo "🧹 Cleaning up..."
    docker rm -f scylla minio 2>/dev/null || true
    kind delete cluster --name nexus-shard-0

# View logs
logs name="":
    docker logs -f {{if name == "" { "scylla" } else { name }}}
