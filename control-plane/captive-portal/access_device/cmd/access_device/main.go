package main

import (
	"fmt"
	"log"
	"net/http"

	"access_device/app"

	"github.com/redis/go-redis/v9"

	// 🔑 强制链接插件（build tag 控制是否生效）
	_ "access_device/auth/plugins/pap"
	// _ "access_device/auth/plugins/token"
	// _ "access_device/auth/plugins/sms"
)

const (
	redisServerAddress  = "172.19.0.2:6379"
	portalServerAddress = "172.19.0.1"
	portalServerPort    = 8080
	nasServerPort       = 9000
)

func main() {
	// ★ 日志全局配置：文件名 + 行号
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	rdb := redis.NewClient(&redis.Options{
		Addr: redisServerAddress,
		DB:   0,
	})

	ps := app.BuildPortalServer(rdb)

	http.HandleFunc("/portal_auth", ps.PortalAuthHandler)

	log.Printf("接入设备 (NAS) 模拟器运行在 :%d...\n", nasServerPort)
	log.Printf("认证重定向功能已开启 -> 目标: http://%s:%d/", portalServerAddress, portalServerPort)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", nasServerPort), nil))
}
