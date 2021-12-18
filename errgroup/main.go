package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	rootCtx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(rootCtx)

	var httpServer = &http.Server{
		Addr:    ":8080",
		Handler: server{},
	}

	// 启动http server
	g.Go(func() error {
		if err := httpServer.ListenAndServe(); err != nil {
			return err
		}
		return nil
	})

	fmt.Println("启动http server")

	// 处理退出(ctrl+c)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case q := <-quit:
				switch q {
				case syscall.SIGINT, syscall.SIGTERM:
					fmt.Println("退出http server")
					fmt.Println(httpServer.Shutdown(context.Background()))
					cancel()
				}
			}
		}
	})

	err := g.Wait()
	fmt.Println(err)
	fmt.Println(ctx.Err())
}

type server struct{}

func (server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello world")
}
