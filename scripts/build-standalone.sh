#!/usr/bin/env bash

version=$(git describe --tags)
vers_nov=${version#*v}

target_archs=(386 amd64 arm arm64)
target_oss=(linux darwin windows)

mkdir -p dist

for os in "${target_oss[@]}"; do
  for arch in "${target_archs[@]}"; do
    if [ "$os" = "darwin" ]; then
     if [ "$arch" = "386" ] || [ "$arch" = "arm" ]; then
      # darwin/386 and darwin/arm support has been dropped since go1.15
      continue
      fi
    fi
    GOARCH="$arch" GOOS="$os" make build || exit 1
    arname="harlock_${vers_nov}_${os}_${arch}"
    if [ "$os" = "windows" ]; then
      zip "$arname".zip README.md LICENSE harlock.exe || exit 1
      mv "$arname.zip" dist/
    else
      tar -czvf "$arname".tar.gz README.md LICENSE harlock || exit 1
      mv "$arname.tar.gz" dist/
    fi
  done
done

rm harlock harlock.exe