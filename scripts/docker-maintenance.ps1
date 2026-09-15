$ErrorActionPreference = 'Stop'

# Keep a small reusable cache so normal rebuilds remain fast without allowing
# BuildKit data to grow indefinitely. Named volumes are deliberately untouched.
docker builder prune --all --force --keep-storage 1GB
docker image prune --force
docker system df
