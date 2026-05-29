#!/bin/sh
set -e

# etcd-setup.sh — 初始化 etcd 用户和认证
# 由 etcd-setup 容器一次性执行

ETCD_ROOT_PASSWORD="${ETCD_ROOT_PASSWORD:-OnlineCar@2026#Etcd}"
ETCD_APP_PASSWORD="${ETCD_APP_PASSWORD:-OnlineCar@2026#EtcdApp}"
ETCD_ENDPOINTS="http://etcd-1:2379,http://etcd-2:2379,http://etcd-3:2379"

echo "[etcd-setup] 等待 etcd 集群就绪..."

# 重试等待 etcd 可用（最多 60 秒）
RETRIES=0
MAX_RETRIES=30
until etcdctl --endpoints=$ETCD_ENDPOINTS endpoint health 2>/dev/null; do
    RETRIES=$((RETRIES + 1))
    if [ $RETRIES -ge $MAX_RETRIES ]; then
        echo "[etcd-setup] 错误: etcd 集群在 ${MAX_RETRIES} 次重试后仍不可用"
        exit 1
    fi
    echo "[etcd-setup] 等待 etcd 就绪... ($RETRIES/$MAX_RETRIES)"
    sleep 2
done

echo "[etcd-setup] etcd 集群已就绪"

# 检查是否已初始化（避免重复执行）
if etcdctl --endpoints=$ETCD_ENDPOINTS --user=root:"$ETCD_ROOT_PASSWORD" user list 2>/dev/null; then
    echo "[etcd-setup] etcd 认证已存在，跳过初始化"
    exit 0
fi

# 创建 root 用户密码
echo "[etcd-setup] 创建 root 用户..."
etcdctl --endpoints=$ETCD_ENDPOINTS user add root --new-user-password="$ETCD_ROOT_PASSWORD" --no-password-confirm 2>/dev/null || true

# 创建应用用户
echo "[etcd-setup] 创建 onlinecar 应用用户..."
etcdctl --endpoints=$ETCD_ENDPOINTS user add onlinecar --new-user-password="$ETCD_APP_PASSWORD" --no-password-confirm 2>/dev/null || true

# 创建角色
echo "[etcd-setup] 创建 onlinecar 角色并赋权..."
etcdctl --endpoints=$ETCD_ENDPOINTS role add onlinecar 2>/dev/null || true
etcdctl --endpoints=$ETCD_ENDPOINTS role grant-permission onlinecar --from-key "" --to-key "" readwrite 2>/dev/null || true

# 绑定用户到角色
etcdctl --endpoints=$ETCD_ENDPOINTS user grant-role onlinecar onlinecar 2>/dev/null || true

# 启用认证
echo "[etcd-setup] 启用 etcd 认证..."
etcdctl --endpoints=$ETCD_ENDPOINTS auth enable 2>/dev/null || true

echo "[etcd-setup] etcd 初始化完成"
