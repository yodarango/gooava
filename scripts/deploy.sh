#!/bin/zsh

source ~/.zshrc

# Check if a commit message was provided
if [ "$#" -ne 1 ]; then
    echo "Please provide a commit message"
    exit 1
fi

# The commit message is the first argument to the script
COMMIT_MESSAGE="$1"

# Add changes to the staging area
# You can adjust this to add specific files or use other git add options
git add .

# Commit the changes with the provided commit message
git commit -m "$COMMIT_MESSAGE"

# Push changes to the Git repository
git push

# Check if the push was successful
if [ $? -eq 0 ]; then
    echo "🐈 Done pushing changes to git. Now pulling changes to VPS."
else
    echo "Git push failed"
    exit 1
fi

# Copy the files to the VPS
# Cleanup note: only dangling images are pruned per deploy. The builder cache is
# capped with --keep-storage instead of a time-based prune — the old
# `--filter until=24h` wiped the Go build/module caches after every deploy and
# forced a multi-minute plaid-go recompile every time.
ssh_main "\
cd /var/www/repos/gooava/app; \
git reset --hard origin/main; \
git pull; \
echo '👍 pulled changes from git and reset to origin'; \
echo 'Current directory: '; pwd; \
echo '🏗️ Building docker now...';\
docker compose down; \
export DOCKER_BUILDKIT=1; \
export COMPOSE_DOCKER_CLI_BUILD=1; \
docker compose build --parallel; \
docker compose up -d; \
echo '🧹 Cleaning dangling images...'; \
docker image prune -f; \
docker builder prune -f --keep-storage 5GB; \
echo '🚀🚀🚀 Deployment successful'"


echo "⭐️🚀✅ Deployment successful"
