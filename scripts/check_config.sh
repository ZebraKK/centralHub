#!/usr/bin/env bash
# 检查 config.json 是否为合法 JSON 并可选校验字段

# 颜色定义
RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
NC='\033[0m' # No Color

# 检查 jq 是否安装
if ! command -v jq >/dev/null 2>&1; then
  echo -e "${YELLOW}未检测到 jq, 请先安装 jq 后再运行本脚本.${NC}"
  exit 1
fi

# 支持传入文件参数，未指定时默认 config.json
CONFIG_FILE="${1:-config.json}"
# 支持自定义必需字段，未指定时默认
REQUIRED_KEYS=(${2:-server database logger external})

if [ ! -f "$CONFIG_FILE" ]; then
  echo -e "${RED}未找到 $CONFIG_FILE${NC}"
  exit 1
fi

# 检查 JSON 格式
if ! jq empty "$CONFIG_FILE" 2>/dev/null; then
  echo -e "${RED}$CONFIG_FILE 格式错误，请修正后再提交。${NC}"
  exit 1
fi

echo -e "${GREEN}$CONFIG_FILE 格式合法。${NC}"

# 检查必需字段存在且不为null
for key in "${REQUIRED_KEYS[@]}"; do
  if ! jq -e "has(\"$key\") and .${key} != null" "$CONFIG_FILE" >/dev/null; then
    echo -e "${RED}缺少必需字段或字段为 null: $key${NC}"
    exit 1
  fi
  # 可选：类型校验示例（如需）
  # if ! jq -e ".${key} | type == \"object\"" "$CONFIG_FILE" >/dev/null; then
  #   echo -e "${RED}字段 $key 类型错误，应为 object${NC}"
  #   exit 1
  # fi

done

echo -e "${GREEN}$CONFIG_FILE 字段校验通过。${NC}"
exit 0
