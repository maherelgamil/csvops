#!/bin/bash

VERSION=$1
NAME=csvops

PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "windows/amd64")
mkdir -p dist

for PLATFORM in "${PLATFORMS[@]}"
do
  OS=$(echo $PLATFORM | cut -d'/' -f1)
  ARCH=$(echo $PLATFORM | cut -d'/' -f2)
  EXT=""
  if [ "$OS" == "windows" ]; then
    EXT=".exe"
  fi

  OUTPUT="dist/${NAME}_${VERSION}_${OS}_${ARCH}${EXT}"
  echo "🔧 Building $OUTPUT"
  env GOOS=$OS GOARCH=$ARCH go build -ldflags "-X main.version=$VERSION" -o $OUTPUT

  # Optional: zip or tar the output
  # shellcheck disable=SC2164
  cd dist
  if [ "$OS" == "windows" ]; then
    zip "${NAME}_${VERSION}_${OS}_${ARCH}.zip" "$(basename $OUTPUT)"
  else
    tar -czf "${NAME}_${VERSION}_${OS}_${ARCH}.tar.gz" "$(basename $OUTPUT)"
  fi
  rm "$(basename $OUTPUT)"
  cd ..
done