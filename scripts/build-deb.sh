#!/usr/bin/env bash

version=$(git describe --tags)
vers_nov=${version#*v}

readarray -d "." -t splitted < <(printf "%s" "$vers_nov")
vers_majmin="${splitted[0]}"."${splitted[1]}"
vers_minor="${splitted[2]}"

archs=(386 arm64 amd64)

function generate() {
  rootname="harlock_${vers_majmin}-${vers_minor}_$1"
  desc="Package: harlock
Version: $vers_majmin
Architecture: $1
Maintainer: Gianmarco Marcello <g.marcello@antima.it>
Description:  A small language with first-class support for embedding data within binaries"

    mkdir "$rootname"
    mkdir -p "$rootname"/usr/local/bin

    GOARCH="$1" make build
    mv harlock "$rootname"/usr/local/bin
    mkdir "$rootname"/DEBIAN
    echo "$desc" >> "$rootname"/DEBIAN/control

    dpkg-deb --build --root-owner-group "$rootname"
    rm -r "$rootname"

    mv "$rootname".deb dist/
}

mkdir -p dist

for arch in "${archs[@]}"; do
  generate "$arch"
done
