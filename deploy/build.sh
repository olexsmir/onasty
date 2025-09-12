#!/usr/bin/env bash
set -e

echo "📦 Building frontend..."
cd ../web
task install
task build

echo "📂 Copying frontend files..."
rm -rf ../deploy/frontend/*
cp -r dist/* ../deploy/frontend/
