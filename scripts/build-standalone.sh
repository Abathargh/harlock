#!/usr/bin/env bash

version=$(git describe --tags)
vers_nov=${version#*v}

target_archs=(386 amd64 arm64)
target_oss=(linux darwin windows)

mkdir -p dist

for os in "${target_oss[@]}"; do
  for arch in "${target_archs[@]}"; do
    GOARCH="$arch" GOOS="$os" make build
    arname="harlock_${vers_nov}_${os}_${arch}"
    if [ "$os" = "windows" ]; then
      zip "$arname".zip README.md LICENSE harlock.exe
      mv "$arname.zip" dist/
    else
      tar -czvf "$arname".tar.gz README.md LICENSE harlock
      mv "$arname.tar.gz" dist/
    fi
  done
done

rm harlock harlock.exe