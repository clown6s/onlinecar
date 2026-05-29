#!/bin/sh
set -euo pipefail

ENDPOINTS="http://etcd-1:2379,http://etcd-2:2379,http://etcd-3:2379"
ETCDCTL="etcdctl --endpoints=$ENDPOINTS"

echo ">>> Checking if auth is already enabled..."
if $ETCDCTL --user=root:"${ETCD_ROOT_PASSWORD}" endpoint health 2>/dev/null; then
    echo ">>> Auth already enabled, skipping setup."
    exit 0
fi

echo ">>> Setting root password..."
$ETCDCTL user passwd root --new-user-password="${ETCD_ROOT_PASSWORD}"

echo ">>> Checking if onlinecar user exists..."
if $ETCDCTL --user=root:"${ETCD_ROOT_PASSWORD}" user get onlinecar 2>/dev/null; then
    echo ">>> User onlinecar already exists, updating password..."
    $ETCDCTL --user=root:"${ETCD_ROOT_PASSWORD}" user passwd onlinecar --new-user-password="${ETCD_APP_PASSWORD}"
else
    echo ">>> Creating onlinecar user..."
    $ETCDCTL --user=root:"${ETCD_ROOT_PASSWORD}" user add onlinecar --new-user-password="${ETCD_APP_PASSWORD}"
fi

echo ">>> Creating role and granting permissions..."
$ETCDCTL --user=root:"${ETCD_ROOT_PASSWORD}" role add onlinecar-role 2>/dev/null || true
$ETCDCTL --user=root:"${ETCD_ROOT_PASSWORD}" role grant-permission onlinecar-role readwrite ''
$ETCDCTL --user=root:"${ETCD_ROOT_PASSWORD}" user grant onlinecar onlinecar-role

echo ">>> Enabling auth..."
$ETCDCTL --user=root:"${ETCD_ROOT_PASSWORD}" auth enable

echo ">>> etcd auth setup completed!"
