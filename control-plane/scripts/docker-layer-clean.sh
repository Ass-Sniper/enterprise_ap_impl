#!/bin/sh

# =========================
# Docker image layer cleaner (POSIX sh)
# =========================

WARN_LAYERS=40
DANGER_LAYERS=60

echo "Scanning Docker images for excessive layers..."
echo

# -----------------------------------
# 1. Scan non-dangling images
# -----------------------------------
docker images --format '{{.Repository}}:{{.Tag}} {{.ID}}' | while read IMAGE ID
do
    if [ "$IMAGE" = "<none>:<none>" ]; then
        continue
    fi

    LAYERS=$(docker history "$ID" 2>/dev/null | wc -l | tr -d ' ')

    if [ "$LAYERS" -lt "$WARN_LAYERS" ]; then
        LEVEL="OK"
    elif [ "$LAYERS" -lt "$DANGER_LAYERS" ]; then
        LEVEL="WARN"
    else
        LEVEL="DANGER"
    fi

    printf "Image: %-50s Layers: %-4s [%s]\n" "$IMAGE" "$LAYERS" "$LEVEL"

    if [ "$LAYERS" -ge "$WARN_LAYERS" ]; then
        printf "  Delete this image? [y/N] "
        read ans
        case "$ans" in
            y|Y)
                docker rmi "$ID"
                ;;
            *)
                echo "  Skip"
                ;;
        esac
    fi

    echo
done

# -----------------------------------
# 2. Dangling images summary
# -----------------------------------
DANGLING_COUNT=$(docker images -f dangling=true -q | wc -l | tr -d ' ')

if [ "$DANGLING_COUNT" -gt 0 ]; then
    echo "Found $DANGLING_COUNT dangling images (<none>:<none>)"

    # Extract dangling UNIQUE SIZE from docker system df -v
    DANGLING_SIZE=$(docker system df -v 2>/dev/null \
        | awk '
            BEGIN { sum=0 }
            $1=="<none>" && $2=="<none>" {
                size=$7
                gsub(/MB/,"",size)
                gsub(/GB/,"",size)
                if ($7 ~ /GB/) size=size*1024
                sum+=size
            }
            END {
                if (sum>=1024)
                    printf "%.2f GB", sum/1024
                else
                    printf "%.2f MB", sum
            }'
    )

    echo "Dangling images unique disk usage: $DANGLING_SIZE"

    printf "Delete ALL dangling images? [y/N] "
    read ans
    case "$ans" in
        y|Y)
            docker image prune -f
            echo "Dangling images deleted."
            ;;
        *)
            echo "Dangling images kept."
            ;;
    esac
else
    echo "No dangling images found."
fi

echo
echo "Done."
