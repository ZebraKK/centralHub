#!/bin/bash
# 检查 config.json 是否为合法 JSON 并可选校验字段


# 支持传入文件参数，未指定时默认 config.json
CONFIG_FILE="${1:-config.json}"

if [ ! -f "$CONFIG_FILE" ]; then
  echo "未找到 $CONFIG_FILE"
  exit 1
fi

# 检查 JSON 格式
if ! jq empty "$CONFIG_FILE" 2>/dev/null; then
  echo "$CONFIG_FILE 格式错误，请修正后再提交。"
  exit 1
fi

echo "$CONFIG_FILE 格式合法。"

# 可选：检查必需字段
REQUIRED_KEYS=(server database logger external)
for key in "${REQUIRED_KEYS[@]}"; do
  if ! jq -e ".${key}" "$CONFIG_FILE" >/dev/null; then
    echo "缺少必需字段: $key"
    exit 1
  fi
done

echo "$CONFIG_FILE 字段校验通过。"
exit 0
