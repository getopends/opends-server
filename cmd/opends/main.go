package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/getopends/opends/internal"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	/*	var (
			fHost    = flag.String("host", "", "Host")
			port     = flag.Int("port", 12345, "Port")
			tls      = flag.Bool("tls", false, "Enable TLS")
			certFile = flag.String("cert-file", "", "Cert file")
			keyfile  = flag.String("key-file", "", "Key file")
		)

		flag.Parse()
	*/

	cfg, err := internal.NewConfig()
	if err != nil {
		panic(err)
	}

	log.Printf("Config: %#+v\n", cfg)

	h := &internal.Handler{
		Service:      &internal.Service{},
		PublicRouter: mux.NewRouter(),
		Config:       &internal.Config{},
	}

	h.SetRoutes()

	host := cfg.Public.Host
	if host == "" {
		host = "0.0.0.0"
	}

	port := cfg.Public.Port
	if port == 0 {
		port = 13000
	}

	addr := fmt.Sprintf("%v:%v", host, port)

	var r http.Handler = h.PublicRouter
	if cfg.CORS.Enable {
		opts := []handlers.CORSOption{}

		if cfg.CORS.AllowCredentials {
			opts = append(opts, handlers.AllowCredentials())
		}

		if cfg.CORS.MaxAge > 0 {
			opts = append(opts, handlers.MaxAge(cfg.CORS.MaxAge))
		}

		if len(cfg.CORS.AllowedHeaders) > 0 {
			opts = append(opts, handlers.AllowedHeaders(cfg.CORS.AllowedHeaders))
		}

		if len(cfg.CORS.AllowedOrigins) > 0 {
			opts = append(opts, handlers.AllowedOrigins(cfg.CORS.AllowedOrigins))
		}

		if len(cfg.CORS.AllowedMethods) > 0 {
			opts = append(opts, handlers.AllowedMethods(cfg.CORS.AllowedMethods))
		}

		if len(cfg.CORS.ExposedHeaders) > 0 {
			opts = append(opts, handlers.ExposedHeaders(cfg.CORS.ExposedHeaders))
		}

		r = handlers.CORS(opts...)(r)
	}

	srv := http.Server{
		Addr:    addr,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Starting server at %v", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve returned err: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("got interruption signal")
	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Printf("server shutdown returned an err: %v\n", err)
	}

	log.Println("server stopped")
}

type Response struct {
	Header http.Header
	Body   io.Reader
}
