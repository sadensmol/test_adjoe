package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-test-task/test-task/src/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.GetConfig()

	//todo: init logger here

	app := App{}
	app.Start(ctx, cfg)

	// Set up signal capturing
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a termination signal
	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %s. Initiating graceful shutdown...", sig)
		cancel()
	}()

	// Wait for the context to be canceled
	<-ctx.Done()

	shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	app.Shutdown(shutdownContext)
	<-shutdownContext.Done()

}

// todo: do some final cleanup here
func (a *App) Shutdown(ctx context.Context) {
}
