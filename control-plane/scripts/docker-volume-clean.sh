#!/bin/bash

# 1. 获取游离卷列表并统计数量
DANGLING_VOLUMES=$(docker volume ls -f dangling=true -q)
COUNT=$(echo "$DANGLING_VOLUMES" | sed '/^\s*$/d' | wc -l)

echo "发现 $COUNT 个未使用的 Docker 卷。"

if [ "$COUNT" -gt 0 ]; then
    echo "------------------------------------------"
    echo "这些卷目前的磁盘占用详情如下："
    
    # 构建匹配正则并显示占用详情
    REGEX=$(echo "$DANGLING_VOLUMES" | tr '\n' '|' | sed 's/|$//')
    if [ -n "$REGEX" ]; then
        docker system df -v | sed -n '/VOLUME NAME/,/^$/p' | grep -E "$REGEX"
    fi
    
    echo "------------------------------------------"
    
    # 2. 增加用户交互提示
    read -p "是否确认清理以上所有未使用的卷？(y/n): " CONFIRM
    
    if [[ "$CONFIRM" == "y" || "$CONFIRM" == "Y" ]]; then
        echo "正在执行清理..."
        docker volume prune -f
        echo "清理成功！"
    else
        echo "操作已取消。"
    fi
else
    echo "提示：目前没有发现可清理的游离卷。"
    echo "如果这些长哈希卷依然存在，说明它们正被某些（可能是已停止的）容器引用。"
fi
