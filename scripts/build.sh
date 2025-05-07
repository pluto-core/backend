#!/usr/bin/env bash
set -e

# переходим в корень проекта
cd "$(dirname "$0")/.."

TAG="k3d-registry.local:5000/pluto-backend:0.0.1-latest"

# соберём auth-service (раскомментируйте, если нужно)
# docker build -f deploy/dockerfiles/auth.Dockerfile    -t ${TAG}-auth    .

# соберём manifest-service
docker build -f deploy/dockerfiles/manifest.Dockerfile -t ${TAG}-manifest .

# пушим
# docker push ${TAG}-auth
docker push ${TAG}-manifest