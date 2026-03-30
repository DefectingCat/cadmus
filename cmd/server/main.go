package main

import (
	"log"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"rua.plus/cadmus/web/templates/pages"
)

func main() {
	// 获取端口配置
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 创建基础 HTTP server
	mux := http.NewServeMux()

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 首页
	mux.Handle("/", templ.Handler(pages.HomePage("Cadmus - 博客平台")))

	// 静态文件服务
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// 启动服务器
	addr := ":" + port
	log.Printf("Cadmus server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}