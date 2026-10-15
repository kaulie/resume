package main

import (
	"context"
	"log"
	"net/http"
	"os"
)

func main() {
	ctx := context.Background()
	pool, err := connectDB(ctx)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer pool.Close()

	if err := migrate(ctx, pool); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	a := &app{pool: pool}
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	log.Printf("✅ resume backend listening on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, a.routes()); err != nil {
		log.Fatal(err)
	}
}
