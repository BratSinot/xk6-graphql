#!/usr/bin/env sh

~/go/bin/xk6 build v0.33.0 \
  --output build/k6 \
  --with github.com/BratSinot/xk6-graphql=$PWD
