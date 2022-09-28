#!/usr/bin/env bash

target_archs=(386 amd64 arm64)
target_oss=(linux darwin windows)

mkdir -p dist

for os in "${target_oss[@]}"; do
  for arch in "${target_archs[@]}"; do
    GOARCH="$arch" GOOS="$os" make build
    arname="harlock_${os}_${arch}"
    if [ "$os" = "windows" ]; then
      zip "harlock_${os}_${arch}" README.md LICENSE harlock.exe
      mv "${arname}.zip" dist/
    else
      tar -czvf "harlock_${os}_${arch}".tar.gz README.md LICENSE harlock
      mv "${arname}.tar.gz" dist/
    fi
  done
done

rm harlock harlock.exe