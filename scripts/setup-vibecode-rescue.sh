#!/bin/bash
# Setup script for VibeCode Rescue blog site
# This script creates the site and all 50 blog posts about vibecoding and automaintainer

set -e

# Configuration
SITE_NAME="VibeCode Rescue"
SITE_SLUG="vibecode-rescue"
SITE_VERSION="v1"

echo "🚀 Setting up VibeCode Rescue blog site..."

# Check if sl-cli is built
if [ ! -f "./sl-cli" ]; then
    echo "📦 Building sl-cli..."
    go build -o sl-cli ./cmd/sl-cli
fi

# Create the site
echo "📝 Creating site: $SITE_NAME ($SITE_SLUG)"
./sl-cli site create --name "$SITE_NAME" --slug "$SITE_SLUG"

# Create version
echo "🔖 Creating version: $SITE_VERSION"
./sl-cli site version create "$SITE_SLUG" --version "$SITE_VERSION"

# Copy all site files from the repository
SITE_DIR="$HOME/.superlandings/sites/$SITE_SLUG/$SITE_VERSION"
mkdir -p "$SITE_DIR"

echo "📚 Copying site files..."
cp sites/vibecode-rescue/v1/layout.html "$SITE_DIR/"
cp sites/vibecode-rescue/v1/index.html "$SITE_DIR/"
cp -r sites/vibecode-rescue/v1/pages "$SITE_DIR/"
cp -r sites/vibecode-rescue/v1/blog "$SITE_DIR/"
echo "  ✅ Copied layout, index, pages, and 50 blog posts"

echo "🎉 Setup complete! The VibeCode Rescue blog site is ready."
echo "📍 Site: $SITE_SLUG"
echo "📂 Path: $HOME/.superlandings/sites/$SITE_SLUG/$SITE_VERSION"
echo "🌐 Start the daemon to view: ./sl-cli backend start --daemon --port 3099"