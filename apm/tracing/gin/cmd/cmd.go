package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"stdlib_learn/apm/tracing/config"
	"stdlib_learn/apm/tracing/gin/middleware"
	"stdlib_learn/apm/tracing/gin/router"
	"time"
)

func main()  {
	config.NewTracer("gin-demo")

	e := gin.New()
	e.Use(middleware.Jaeger())
	router.Register(e)

	server := &http.Server{
		Addr: ":8080",
		Handler: e,
		ReadTimeout: 15*time.Second,
		WriteTimeout: 15*time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server listen:%s\n",err)
		}
	}()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt)
	sig := <-signalChan
	log.Println("Get Signal:", sig)
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")


}
