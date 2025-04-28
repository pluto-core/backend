# scripts/build.sh
#!/usr/bin/env bash
set -e
make fmt
make test
make build