#!/bin/bash

result=0
if [ "${test}" == "all" ]; then
  go test -v ./tests/integration/... -integration-tests
  integrationResult=$?
  go test -v ./tests/validation/... -validation-tests
  validationResult=$?
  result=$(( integrationResult > validationResult ? integrationResult : validationResult ))
elif [ "${test}" == "integration" ] || [ "${test}" == "validation" ]; then
  go test -v ./tests/${test}/... -${test}-tests
  result=$?
else
  echo "Only acceptable values for environment variable 'test' are 'integration', 'validation', and 'all'."
fi

source ./tests/scripts/cleanup_cluster.sh
exit $result