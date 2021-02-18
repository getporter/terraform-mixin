#!/usr/bin/env bash

set -euo pipefail
export REGISTRY=${REGISTRY:-$USER}
export REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )"
export PORTER_HOME=${PORTER_HOME:-$REPO_DIR/bin}
# Run tests in a temp directory
export TEST_DIR=/tmp/porter/terraform
mkdir -p ${TEST_DIR}
pushd ${TEST_DIR}
trap popd EXIT

function verify-output() {
  # Verify the output matches the expected value
  output=`${PORTER_HOME}/porter installation output show $1`
  if [[ "${output}" != "$2" ]]; then
    echo "Output '$1' value: '${output}' does not match expected"
    return 1
  fi

  # Verify the output has no extra newline (mixin should trim newline added by terraform cli)
  if [[ "$(${PORTER_HOME}/porter installation output show $1 | wc -l)" > 1 ]]; then
    echo "Output '$1' has an extra newline character"
    return 1
  fi
}


# Copy terraform assets
cp -r ${REPO_DIR}/build/testdata/bundles/terraform/terraform .

# Copy in the terraform porter manifest
cp ${REPO_DIR}/build/testdata/bundles/terraform/porter.yaml .

${PORTER_HOME}/porter build

${PORTER_HOME}/porter install --debug --param file_contents='foo!'

echo "Verifying installation output after install"
verify-output "file_contents" 'foo!'

${PORTER_HOME}/porter invoke --action=plan --debug

${PORTER_HOME}/porter upgrade --debug --param file_contents='bar!'

echo "Verifying installation output after upgrade"
verify-output "file_contents" 'bar!'

${PORTER_HOME}/porter uninstall --debug
