#!/bin/zsh

################################################################################
# Syncs the live (VPS) gooava MySQL database into the local docker MySQL.
# Starts the local MySQL container (and Docker Desktop on macOS) when needed.
#
# One direction only: live -> local. This REPLACES your local data.
#
# Security notes:
# - Secrets never touch a command line: mysqldump/mysql read credentials from
#   each container's own environment ($MYSQL_ROOT_PASSWORD / $MYSQL_DATABASE).
# - The dump contains plaintext Plaid access tokens and your transaction
#   history. It is written to a mktemp file and ALWAYS deleted on exit.
#
# Usage: ./scripts/sync_live_db.sh [-y|--yes]
################################################################################

set -e

source ~/.zshrc   # for ssh_main, same as deploy.sh

REMOTE_APP_DIR="/var/www/repos/gooava/app"

# always run from the repo root (this script lives in scripts/)
cd "$(dirname "$0")/.."

# --- confirmations and pre-flight checks -------------------------------------

if [ "$1" != "-y" ] && [ "$1" != "--yes" ]; then
  read "?⚠️  This will REPLACE your local gooava database with the live one. Continue? [y/N] " confirm
  if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "Aborted."
    exit 0
  fi
fi

# The local MySQL container must be healthy before we pipe a dump into it —
# start it ourselves instead of asking the user to.
if ! docker compose ps gooava_mysql 2>/dev/null | grep -q "(healthy)"; then
  # the daemon itself must be up first — on macOS, launch Docker Desktop
  if ! docker info >/dev/null 2>&1; then
    echo "🐳 Docker is not running — launching Docker Desktop..."
    if ! open -a Docker 2>/dev/null; then
      echo "❌ Could not launch Docker Desktop. Start Docker manually and re-run this script."
      exit 1
    fi

    echo "⏳ Waiting for the Docker daemon..."
    for _ in {1..30}; do
      docker info >/dev/null 2>&1 && break
      sleep 2
    done
    if ! docker info >/dev/null 2>&1; then
      echo "❌ Docker daemon did not come up in time. Start Docker manually and re-run this script."
      exit 1
    fi
  fi

  echo "🐳 Starting the local MySQL container..."
  # --wait blocks until the healthcheck (mysqladmin ping) passes
  if ! docker compose up -d --wait --wait-timeout 120 gooava_mysql; then
    echo "❌ MySQL did not become healthy in time. Check: docker compose logs gooava_mysql"
    exit 1
  fi
fi

# --- dump on the VPS ----------------------------------------------------------

# the trap guarantees the dump (sensitive!) is removed no matter how we exit
DUMP_FILE="$(mktemp -t gooava_live_dump).sql"
trap 'rm -f "$DUMP_FILE"' EXIT

echo "📥 Dumping the live database on the VPS..."
ssh_main "cd $REMOTE_APP_DIR && docker compose exec -T gooava_mysql sh -c 'mysqldump --single-transaction --no-tablespaces -u root -p\"\$MYSQL_ROOT_PASSWORD\" \"\$MYSQL_DATABASE\"'" > "$DUMP_FILE"

if [ ! -s "$DUMP_FILE" ]; then
  echo "❌ The dump came back empty. Is the stack running on the VPS?"
  exit 1
fi

if ! head -5 "$DUMP_FILE" | grep -qi "mysql dump"; then
  echo "❌ The downloaded file does not look like a MySQL dump — aborting before touching the local DB."
  exit 1
fi

echo "👍 Dump downloaded ($(du -h "$DUMP_FILE" | cut -f1) — contains sensitive data, will be deleted on exit)"

# --- restore locally ----------------------------------------------------------

# mysqldump includes DROP TABLE + SET FOREIGN_KEY_CHECKS=0, so the restore
# cleanly replaces every table (including schema_migrations)
echo "🔄 Restoring into local MySQL..."
docker compose exec -T gooava_mysql sh -c 'mysql -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE"' < "$DUMP_FILE"

echo "✅ Local database replaced with the live one."

# the app runs migrations at startup — if local code is ahead of live, a
# restart re-applies the pending migrations onto the fresh data
if docker compose ps --status running gooava_app 2>/dev/null | grep -q gooava_app; then
  echo "🔁 Restarting the app so migrations re-check against the fresh data..."
  docker compose restart gooava_app
fi

echo "⭐️ Done."
