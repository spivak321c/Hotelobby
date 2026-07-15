#!/usr/bin/env bash
set -euo pipefail

# Install bun (not pre-installed on Render)
curl -fsSL https://bun.sh/install | bash
export BUN_INSTALL="$HOME/.bun"
export PATH="$BUN_INSTALL/bin:$PATH"

# Build frontend
cd frontend
bun install
bun run build

# Copy to Go embed directory
mkdir -p ../backend/cmd/server/static
cp -r build/* ../backend/cmd/server/static/

# Debug: show what was copied
echo "=== Static files after copy ==="
find ../backend/cmd/server/static -type f | head -20
echo "=== JS files ==="
find ../backend/cmd/server/static -name "*.js" | head -10
echo "=== End debug ==="

# Build Go binary (output to repo root as "app")
cd ../backend
go build -tags netgo -ldflags '-s -w' -o ../app ./cmd/server
