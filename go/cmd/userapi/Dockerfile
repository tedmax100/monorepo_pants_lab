# Build context root = repository root (Pants sets this automatically).
# Pants compiles the Go binary for the current platform (linux/amd64 in CI)
# and places it at the address path (/ → . in directory name): go.cmd.userapi/bin
#
# Uses distroless for a minimal, secure production image (no shell, no package manager).
# For debugging, swap to gcr.io/distroless/base-debian12:debug which includes busybox.
FROM gcr.io/distroless/base-debian12

COPY go.cmd.userapi/bin /usr/local/bin/userapi

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/userapi"]
