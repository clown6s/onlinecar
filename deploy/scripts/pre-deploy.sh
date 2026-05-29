#!/usr/bin/env bash
# pre-deploy.sh — 部署前配置校验
# 用法: cd deploy && bash scripts/pre-deploy.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"

PASS=0
FAIL=0
WARN=0

pass() { PASS=$((PASS + 1)); echo "  ✓ $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  ✗ $1"; }
warn() { WARN=$((WARN + 1)); echo "  ⚠ $1"; }

echo "===== OnlineCar 部署前配置校验 ====="
echo ""

# 1. docker-compose.yaml 语法校验
echo "[1/6] docker-compose.yaml 语法校验"
if command -v docker compose &>/dev/null; then
  if docker compose -f "$DEPLOY_DIR/docker-compose.yaml" config --quiet 2>/dev/null; then
    pass "docker-compose.yaml 语法正确"
  else
    fail "docker-compose.yaml 语法错误"
  fi
elif command -v docker-compose &>/dev/null; then
  if docker-compose -f "$DEPLOY_DIR/docker-compose.yaml" config --quiet 2>/dev/null; then
    pass "docker-compose.yaml 语法正确"
  else
    fail "docker-compose.yaml 语法错误"
  fi
else
    warn "未找到 docker compose / docker-compose，跳过语法校验"
fi
echo ""

# 2. 必需配置文件存在性
echo "[2/6] 必需配置文件存在性"
REQUIRED_FILES=(
  "config/driver.yaml"
  "config/passenger.yaml"
  "config/district.yaml"
  "config/gateway.yaml"
  "config/prometheus.yaml"
  "config/promtail-config.yaml"
  "config/grafana-datasources.yaml"
  "config/grafana-dashboards.yaml"
  "config/alertmanager.yaml"
  "config/jaeger-config.yaml"
  "config/otel-collector-config.yaml"
  "config/alert_rules.yaml"
  "nginx/nginx.conf"
  "scripts/etcd-setup.sh"
)
for f in "${REQUIRED_FILES[@]}"; do
  if [ -f "$DEPLOY_DIR/$f" ]; then
    pass "$f"
  else
    fail "$f 不存在"
  fi
done
echo ""

# 3. YAML 配置关键字段校验
echo "[3/6] YAML 配置关键字段校验"
for svc in driver passenger district; do
  YAML="$DEPLOY_DIR/config/${svc}.yaml"
  if [ -f "$YAML" ]; then
    # 检查 DataSource 字段
    if grep -q "DataSource:" "$YAML"; then
      pass "${svc}.yaml DataSource 字段存在"
    else
      fail "${svc}.yaml 缺少 DataSource 字段"
    fi
    # 检查 Log 配置
    if grep -q "^Log:" "$YAML"; then
      pass "${svc}.yaml Log 配置存在"
    else
      fail "${svc}.yaml 缺少 Log 配置"
    fi
  fi
done

# Gateway 特殊校验
GW_YAML="$DEPLOY_DIR/config/gateway.yaml"
if [ -f "$GW_YAML" ]; then
  if grep -q "AuthDriver:" "$GW_YAML" && grep -q "AuthPassenger:" "$GW_YAML"; then
    pass "gateway.yaml JWT 配置存在"
  else
    fail "gateway.yaml 缺少 JWT 配置"
  fi
fi
echo ""

# 4. Dockerfile 存在性 + 多阶段构建检查
echo "[4/6] Dockerfile 校验"
for svc in driver passenger gateway district; do
  DF="$DEPLOY_DIR/${svc}/Dockerfile"
  if [ -f "$DF" ]; then
    pass "${svc}/Dockerfile 存在"
    if grep -q "FROM.*AS builder" "$DF"; then
      pass "${svc}/Dockerfile 多阶段构建"
    else
      fail "${svc}/Dockerfile 未使用多阶段构建"
    fi
    if grep -q "USER " "$DF"; then
      pass "${svc}/Dockerfile 非 root 用户"
    else
      fail "${svc}/Dockerfile 缺少 USER 指令"
    fi
  else
    fail "${svc}/Dockerfile 不存在"
  fi
done
echo ""

# 5. .dockerignore 存在性
echo "[5/6] .dockerignore 存在性"
for svc in driver passenger gateway district; do
  DI="$DEPLOY_DIR/../${svc}/.dockerignore"
  if [ -f "$DI" ]; then
    pass "${svc}/.dockerignore 存在"
  else
    warn "${svc}/.dockerignore 不存在（构建上下文较大）"
  fi
done
echo ""

# 6. 环境变量占位符检查（是否有未替换的占位符）
echo "[6/6] 环境变量占位符检查"
if grep -rE '\$\{[A-Z_]+\}' "$DEPLOY_DIR/config/" 2>/dev/null | grep -v "DB_APP_PASSWORD" | grep -qE '\$\{[A-Z_]+\}'; then
  # 只报告非 DB_APP_PASSWORD 的（DB_APP_PASSWORD 有默认值）
  UNSET_VARS=$(grep -rE '\$\{[A-Z_]+\}' "$DEPLOY_DIR/config/" 2>/dev/null | grep -v "DB_APP_PASSWORD" | head -5)
  if [ -n "$UNSET_VARS" ]; then
    warn "以下环境变量占位符需确认已设置:"
    echo "$UNSET_VARS" | sed 's/^/    /'
  fi
else
  pass "无未设置的环境变量占位符"
fi
echo ""

# 汇总
echo "===== 校验结果 ====="
echo "  通过: $PASS  失败: $FAIL  警告: $WARN"

if [ "$FAIL" -gt 0 ]; then
  echo "  ❌ 存在 $FAIL 个失败项，请修复后再部署"
  exit 1
elif [ "$WARN" -gt 0 ]; then
  echo "  ⚠️ 存在 $WARN 个警告项，建议修复"
  exit 0
else
  echo "  ✅ 所有检查通过"
  exit 0
fi
