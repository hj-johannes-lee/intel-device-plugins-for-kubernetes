#!/bin/sh -eu
#
# Copyright 2019-2021 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#
# Invoke this script with a version as parameter
# and it will update all hard-coded image versions
# in the source code.
#
# Adapted from https://github.com/intel/pmem-csi/

if [ $# != 1 ] || [ "$1" = "?" ] || [ "$1" = "--help" ]; then
    echo "Usage: $0 <image version>" >&2
    exit 1
fi

# change all devel in related files
# :devel =devel :-devel 'devel' => <image version>
files=$(git grep -l "\(:\|=\|:-\|'\)devel\(.*\)" build/docker Jenkinsfile demo/build-image.sh pkg/controllers/*/*_test.go pkg/controllers/reconciler_test.go test/e2e/*/*.go)
for file in $files; do
    sed -i -e "s;\(:\|=\|:-\|'\)devel\(.*\);\1$1\2;g" $file
done

# change yaml files' image version
# devel or 0.xy.z => <image version>
files=$(git grep -l '' *.yaml)
for file in $files; do
   sed -i -e "s;\(image\|initImage\|containerImage\|InitImage\)\(:.*intel/.*:\).*;\1\2$1;g" $file
done

# change ImageMinVersion to <image version>
files=$(git grep -l '\(ImageMinVersion.*"\).*\("\)' pkg/controllers/reconciler.go)
for file in $files; do
   sed -i -e "s;\(ImageMinVersion.*\"\).*\(\"\);\1$1\2;g" $file
done
