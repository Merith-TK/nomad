#!/bin/bash
set -e

echo "Installing OpenSCAD..."
sudo apt-get update
sudo apt-get install -y openscad openscad-nightly xvfb

echo "Installing Golang..."
sudo apt-get install -y golang-go

echo "Installing Python packages..."
pip install --upgrade pip
pip install -r requirements.txt

echo "Setting up git hooks..."
git config --global --add safe.directory /workspaces/nomad-core

echo "✅ Dev container setup complete!"
echo "📦 Installed: OpenSCAD, Python 3D libraries"
echo "🎨 Ready to create 3D models with code!"
