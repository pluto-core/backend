# scripts/run_local.sh
#!/usr/bin/env bash
set -e
docker-compose -p pluto-backend -f deploy/docker-compose.local.yml up --build -d
