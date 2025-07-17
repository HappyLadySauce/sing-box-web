#!/bin/bash

# 默认使用 Claude 代理
api="claude"

# 解析参数
while getopts "a:c" opt; do
  case $opt in
    a)
      api="$OPTARG"
      ;;
    c)
      run_command=true
      ;;
    \?)
      echo "无效选项: -$OPTARG" >&2
      exit 1
      ;;
  esac
done

# 设置 API 配置
case "$api" in
  "claude")
    export ANTHROPIC_BASE_URL="https://api.aicodemirror.com/api/claudecode"
    export ANTHROPIC_API_KEY="sk-ant-api03-28FTUmLVNTCLGdA2p6_G3-mwBKUZOAtgXfG41K9gFafkLZjLjsHPx54RSI_kGESMPDk-xqX9Y5ahKa3zm5ZHtg"
    ;;
  "kimi")
    export ANTHROPIC_BASE_URL="https://api.moonshot.cn/anthropic/"
    export ANTHROPIC_API_KEY="sk-z3WHOGq7mR7tHcb7u5TaJRihudUwfZINm6AimD4bQ4vOcZG2"
    ;;
  *)
    echo "未知 API: $api" >&2
    echo "可用选项: claude, kimi"
    exit 1
    ;;
esac

# 执行命令
if [ "$run_command" = true ]; then
  claude -c
fi