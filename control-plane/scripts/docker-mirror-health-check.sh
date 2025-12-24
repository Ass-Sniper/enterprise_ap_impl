#!/bin/bash

# ==============================================================================
# 脚本名称: docker-mirror-health-check.sh
# 脚本功能: 自动检测当前 Docker 配置的镜像加速器 (Registry Mirrors) 是否可用
# 适用环境: Ubuntu/CentOS/Debian 等安装了 Docker 的 Linux 系统
# 使用方法: 
#   1. chmod +x docker-mirror-health-check.sh
#   2. ./docker-mirror-health-check.sh
# 备注说明: 
#   - 脚本会读取 /etc/docker/daemon.json 生效后的配置（通过 docker info）
#   - 成功的定义：能够联通 Docker V2 API 接口并返回 200 或 401 状态码
# ==============================================================================

# 获取当前 Docker 生效的镜像源列表
# 逻辑：从 docker info 中提取 Registry Mirrors 下方的 https 链接，并去除行首空格和末尾逗号
MIRRORS=$(docker info 2>/dev/null | grep -A 15 "Registry Mirrors" | grep "https" | sed 's/^[[:space:]]*//; s/,$//')

echo "================================================================"
echo "          Docker 镜像源可用性检测 (Mirror Health Check)         "
echo "================================================================"

# 检查是否配置了镜像源
if [ -z "$MIRRORS" ]; then
    echo -e "\033[33m提示:\033[0m 未发现已配置的 Docker 镜像加速器。"
    echo "请先在 /etc/docker/daemon.json 中配置镜像源并重启 Docker。"
    exit 1
fi

echo "检测到以下镜像源，正在开始测速..."
echo "----------------------------------------------------------------"

# 遍历每个镜像源进行测试
for url in $MIRRORS; do
    # 打印正在检测的域名
    printf "连接测试: %-40s " "$url"
    
    # 记录起始毫秒时间
    start=$(date +%s%3N)
    
    # 执行测试请求
    # -o /dev/null: 不保存输出内容
    # -s: 静默模式，不显示进度条
    # -L: 允许重定向
    # -w: 格式化输出 HTTP 状态码
    # --connect-timeout: 连接超时时间设为 5 秒
    code=$(curl -o /dev/null -s -L -w "%{http_code}" --connect-timeout 5 "$url/v2/")
    
    # 记录结束毫秒时间
    end=$(date +%s%3N)
    duration=$((end - start))

    # 判断逻辑：200 OK 或 401 Unauthorized (API 握手成功但未登录) 均视为接口可用
    if [ "$code" -eq 200 ] || [ "$code" -eq 401 ]; then
        echo -e "\033[32m[ 成功 ]\033[0m 响应: $code | 延迟: ${duration}ms"
    else
        echo -e "\033[31m[ 失败 ]\033[0m 响应: $code | 延迟: ${duration}ms"
    fi
done

echo "----------------------------------------------------------------"
echo "建议: 如果结果全部 [失败]，请考虑更换最新的加速地址或使用代理。"
echo "================================================================"
