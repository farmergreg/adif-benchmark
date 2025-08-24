#!/usr/bin/env sh
docker build -t adif-benchmark .
docker run adif-benchmark
