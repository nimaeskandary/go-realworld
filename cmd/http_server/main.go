package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/nimaeskandary/go-realworld/cmd/http_server/app"
	http_handler_types "github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler/types"
	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/middleware"
	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	"github.com/nimaeskandary/go-realworld/pkg/util"
)

func main() {
	ctx := context.Background()
	cleanupManager := util.NewCleanupManager(ctx, true)
	defer cleanupManager.Cleanup()

	args := app.ParseArgs()
	configData, err := os.ReadFile(args.ConfigPath)
	if err != nil {
		log.Fatalf("failed to read config file %v: %v", args.ConfigPath, err)
	}

	// setup dependency injection system
	var httpHandler http_handler_types.HttpHandler
	var config app.HttpServerConfig
	fxApp := util.CreateFxAppAndExtract(app.ModuleList(configData), &httpHandler, &config)

	log.Println("starting dependency injection system...")
	if err := fxApp.Start(ctx); err != nil {
		log.Fatalf("dependency injection system failed to start: %v", err)
	}

	cleanupManager.RegisterCleanupFunc(func() {
		log.Printf("stopping dependency injection system...")
		if err := fxApp.Stop(ctx); err != nil {
			log.Printf("dependency injection system failed to stop gracefully: %v", err)
		}
	})

	// setup http server
	mux := http.NewServeMux()
	mux.Handle("/", httpHandler.GetHandler())

	if config.IsSwaggerEnabled {
		addSwaggerHandler(mux)
	}

	muxWithMiddleware := middleware.WithCorsMiddleware(config.AllowedOrigins, mux)

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%v", config.Port),
		Handler: muxWithMiddleware,
	}

	cleanupManager.RegisterCleanupFunc(func() {
		log.Printf("stopping http server...")
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("http server failed to stop gracefully: %v", err)
		}
	})

	// start http server
	log.Printf("http server listening on port %v", config.Port)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server encountered an error: %v", err)
	}
}

func addSwaggerHandler(mux *http.ServeMux) {
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		swagger, err := api_gen.GetSwagger()
		if err != nil {
			http.Error(w, "failed to get swagger spec", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(swagger)
		if err != nil {
			log.Printf("error encoding swagger spec: %v", err)
		}
	})
}
