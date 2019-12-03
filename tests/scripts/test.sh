#!/bin/bash

if [ "${test}" == "all" ]; then
  go test -v ./tests/integration/... -integration-tests
  go test -v ./tests/validation/... -validation-tests
elif [ "${test}" == "integration" ] || [ "${test}" == "validation" ]; then
  go test -v ./tests/${test}/... -${test}-tests
else
  echo "Only acceptable values for environment variable 'test' are 'integration', 'validation', and 'all'."
fi

source ./tests/scripts/cleanup_cluster.sh