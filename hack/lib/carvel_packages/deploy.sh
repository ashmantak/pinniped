#!/usr/bin/env bash

# Copyright 2020-2023 the Pinniped contributors. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

#
# This script can be used in conjunction with prepare-for-integration-tests.sh.
# When invoked with the PINNIPED_USE_LOCAL_KIND_REGISTRY environment variable set to a non-empty value,
# the integration tests script will create a local docker registry and configure kind to use the registry
# and will build the Pinniped binary and container image.
# This script will then create Carvel Packages for supervisor,concierge and local-user-authenticator.
# It will also create a Carvel PackageRepository.
# The PackageRepository will be installed on the kind cluster, then PackageInstall resources
# will be created to deploy an instance of each of the packages on the cluster.
# Once this script has completed, Pinniped can be interacted with as if it had been deployed in the usual way,
# for example by running tests or by preparing supervisor for manual interactions:
#  source /tmp/integration-test-env && go test -v -race -count 1 -timeout 0 ./test/integration -run  /TestE2EFullIntegration_Browser
#  hack/prepare-supervisor-on-kind.sh --oidc
#
# Example usage:
#   PINNIPED_USE_LOCAL_KIND_REGISTRY=1 ./hack/prepare-for-integration-tests.sh --clean --pre-install ./hack/lib/carvel_packages/build.sh --alternate-deploy ./hack/lib/carvel_packages/deploy.sh
#
set -euo pipefail

#
# Helper functions
#
function log_note() {
  GREEN='\033[0;32m'
  NC='\033[0m'
  if [[ ${COLORTERM:-unknown} =~ ^(truecolor|24bit)$ ]]; then
    echo -e "${GREEN}$*${NC}"
  else
    echo "$*"
  fi
}

function log_error() {
  RED='\033[0;31m'
  NC='\033[0m'
  if [[ ${COLORTERM:-unknown} =~ ^(truecolor|24bit)$ ]]; then
    echo -e "🙁${RED} Error: $* ${NC}"
  else
    echo ":( Error: $*"
  fi
}

function check_dependency() {
  if ! command -v "$1" >/dev/null; then
    log_error "Missing dependency..."
    log_error "$2"
    exit 1
  fi
}

# this script is best invoked from the root directory
# it is designed to be passed as --pre-install flag to hack/prepare-for-integration-tests.sh
hack_lib_path="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$hack_lib_path/../../" || exit 1

# arguments provided to scripts called by hack/prepare-for-integration-tests.sh
# - app: unimportant, but always first
# - tag: uuidgen in hack/prepare-for-integration-tests.sh
#        if this script is run standalone, then auto-fill with a unique value
app=${1:-"app-argument-not-provided"}
tag=${2:-"tag-argument-not-provided"}
registry=${3:-"registry-argument-not-provided"}
repo=${4:-"repo-argument-not-provided"}
data_values_file=${5:-"ytt-data-values-file-argument-not-provided"}

log_note "deploy.sh called with app: ${app} tag: ${tag} registry: ${registry} repo: ${repo} data_values_file: ${data_values_file}"

log_note "Begin deploy of carvel package for ${app}..."

if [[ "${PINNIPED_USE_LOCAL_KIND_REGISTRY:-}" == "" ]]; then
  log_error "Building the Carvel package requires configuring kind with a local registry."
  log_error "Please set the environment variable PINNIPED_USE_LOCAL_KIND_REGISTRY."
  log_error "For example:"
  log_error "    PINNIPED_USE_LOCAL_KIND_REGISTRY=1 ./hack/prepare-for-integration-tests.sh --clean --pre-install ./hack/lib/carvel_packages/build.sh --alternate-deploy ./hack/lib/carvel_packages/deploy.sh"
  exit 1
fi


pinniped_package_version="${tag}" # ie, "0.25.0"

registry_repo="$registry/$repo"
registry_repo_tag="${registry_repo}:${tag}"

api_group_suffix="pinniped.dev"

# Package prefix for concierge, supervisor, local-user-authenticator
package_repo_prefix="${registry_repo}/package" # + $resource_name + ":" + $tag

# Pinniped Package repository
package_repository_repo="pinniped-package-repository"
package_repository_repo_tag="${registry_repo}/${package_repository_repo}:${tag}"


# deploy kapp-controller onto kind cluster
log_note "Installing kapp-controller on cluster..."
KAPP_CONTROLLER_GLOBAL_NAMESPACE="kapp-controller-packaging-global"
kapp deploy --app kapp-controller --file "https://github.com/vmware-tanzu/carvel-kapp-controller/releases/latest/download/release.yml" -y

# ensure this directory exists though this script will run several times
mkdir -p "deploy_carvel/install"

log_note "Deploying Pinniped PackageRepository..."
pinniped_package_repository_name="pinniped-package-repository"
pinniped_package_repository_file="deploy_carvel/install/packagerepository.${pinniped_package_version}.yml"
echo -n "" > "${pinniped_package_repository_file}"
cat <<EOT >> "${pinniped_package_repository_file}"
---
apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageRepository
metadata:
  name: "${pinniped_package_repository_name}"
  namespace: "${KAPP_CONTROLLER_GLOBAL_NAMESPACE}"
spec:
  fetch:
    imgpkgBundle:
      image: "${package_repository_repo_tag}"
EOT

kapp deploy --app "${pinniped_package_repository_name}" --file "${pinniped_package_repository_file}" -y
kapp inspect --app "${pinniped_package_repository_name}" --tree


resource_name="${app}"

log_note "Creating RBAC for ${resource_name} PackageInstall..."

namespace="${resource_name}-install-ns"
pinniped_package_rbac_prefix="pinniped-package-rbac-${resource_name}"
pinniped_package_rbac_file="deploy_carvel/install/${pinniped_package_rbac_prefix}-${resource_name}-rbac.yml"
echo -n "" > "${pinniped_package_rbac_file}"
# NOTE: this script is for development purposes running on a local kind cluster.
# For any other use case, the generated artifacts should be properly reviewed.
# For example, the RBAC generated here should be adjusted to conform to the
# principle of LEAST privilege.
cat <<EOF >> "${pinniped_package_rbac_file}"
---
apiVersion: v1
kind: Namespace
metadata:
  name: "${namespace}"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: "${pinniped_package_rbac_prefix}-sa-superadmin-dangerous"
  namespace: "${namespace}"
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: "${pinniped_package_rbac_prefix}-role-superadmin-dangerous"
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: "${pinniped_package_rbac_prefix}-role-binding-superadmin-dangerous"
subjects:
- kind: ServiceAccount
  name: "${pinniped_package_rbac_prefix}-sa-superadmin-dangerous"
  namespace: "${namespace}"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: "${pinniped_package_rbac_prefix}-role-superadmin-dangerous"

EOF

kapp deploy --app "${pinniped_package_rbac_prefix}" --file "${pinniped_package_rbac_file}" -y


log_note "Creating ${resource_name} PackageInstall..."
NAMESPACE="${resource_name}-install-ns"
PINNIPED_PACKAGE_RBAC_PREFIX="pinniped-package-rbac-${resource_name}"
RESOURCE_PACKAGE_VERSION="${resource_name}.pinniped.dev"
PACKAGE_INSTALL_FILE_NAME="deploy_carvel/install/${resource_name}-pkginstall.yml"
SECRET_NAME="${resource_name}-package-install-secret"

log_note "Generating ${PACKAGE_INSTALL_FILE_NAME}..."
cat > "${PACKAGE_INSTALL_FILE_NAME}" << EOF
---
apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
metadata:
    # name, does not have to be versioned, versionSelection.constraints below will handle
    name: "${resource_name}-package-install"
    namespace: "${NAMESPACE}"
spec:
  serviceAccountName: "${PINNIPED_PACKAGE_RBAC_PREFIX}-sa-superadmin-dangerous"
  packageRef:
    refName: "${RESOURCE_PACKAGE_VERSION}"
    versionSelection:
      constraints: "${pinniped_package_version}"
  values:
  - secretRef:
      name: "${SECRET_NAME}"
EOF

log_note "Creating secret ${SECRET_NAME} with ${data_values_file}..."
kubectl create secret generic "${SECRET_NAME}" --namespace "${NAMESPACE}" --from-file "${data_values_file}" --dry-run=client -o yaml | kubectl apply -f-

KAPP_CONTROLLER_APP_NAME="${resource_name}-pkginstall"
log_note "Deploying ${KAPP_CONTROLLER_APP_NAME}..."
kapp deploy --app "${KAPP_CONTROLLER_APP_NAME}" --file "${PACKAGE_INSTALL_FILE_NAME}" -y


log_note "Verifying PackageInstall resources..."
kubectl get PackageInstall -A | grep pinniped
kubectl get secret -A | grep pinniped

log_note "Listing all package resources (PackageRepository, Package, PackageInstall)..."
kubectl get pkgi && kubectl get pkgr && kubectl get pkg

log_note "Listing all kapp cli apps..."
kapp ls --all-namespaces

log_note "Listing all kapp-controller apps..."
kubectl get app --all-namespaces

log_note "Complete deploy of carvel package for ${app}..."
