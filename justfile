# Default recipe to start everything
all: scylla minio prepare-buckets prepare-scylla

# Start ScyllaDB container
scylla:
    docker run -d \
      --name scylla \
      -p 9042:9042 \
      scylladb/scylla \
      --listen-address 127.0.0.1 \
      --rpc-address 0.0.0.0 \
      --broadcast-rpc-address 127.0.0.1 \
      --smp 1 \
      --developer-mode 1

# Start Minio container
minio:
    docker run -d \
      --name minio \
      -p 9000:9000 \
      -p 9123:9123 \
      -p 10000:10000 \
      --restart always \
      quay.io/minio/minio \
      server /data --console-address :9123

# Create Minio buckets (depends on Minio being healthy)
prepare-buckets:
    @echo "Waiting for Minio to be healthy..."
    until curl -f http://localhost:9123; do sleep 5; done
    docker run --rm \
      --network host \
      --entrypoint "/bin/sh" \
      quay.io/minio/minio \
      -c "mc alias set e2e http://localhost:9000 minioadmin minioadmin && \
          mc admin info e2e && \
          mc mb --ignore-existing e2e/tmp && \
          mc mb --ignore-existing e2e/nexus"

# Run Scylla initialization script
prepare-scylla:
    @echo "Waiting for Scylla gossip to be running..."
    until [ "$(docker exec scylla nodetool statusgossip 2>/dev/null)" = "running" ]; do sleep 5; done
    docker run --rm \
      --network host \
      -v {{invocation_directory()}}/test-resources/storage:/opt/storage \
      --entrypoint /opt/storage/prepare-scylla.sh \
      scylladb/scylla

# Stop and remove all containers
clean:
    docker rm -f scylla minio || true
