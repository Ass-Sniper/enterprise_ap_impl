package accessdevice

import (
	"access_device/app"
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "cp-redis:6379",
		DB:   0,
	})

	ps := app.BuildPortalServer(rdb)

	http.HandleFunc("/portal_auth", ps.PortalAuthHandler)

	fmt.Printf("接入设备 (NAS) 模拟器运行在 :8080...\n")
	log.Printf("认证重定向功能已开启 -> 目标: http://127.0.0.1:8088/login")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
