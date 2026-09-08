package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/La-Perla-App/backend-core-laperla/pkg/server/middlewares"
	"github.com/La-Perla-App/backend-core-laperla/pkg/server/swagger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Server struct {
	mux                *chi.Mux
	chiMiddlewares     []func(http.Handler) http.Handler
	routerGroup        func(r chi.Router)
	registerServiceFns []RegisterServiceFn
	gatewayPrefixes    []string
	openAPISchema      []byte
	debugRoute         string
	protosZipRoute     string
	protosZip          []byte
	swaggerRoute       string
	prodMode           bool
	httpServer         *http.Server
}

func NewServer(opts ...ServerOption) Server {
	server := Server{
		mux: chi.NewMux(),
	}
	for _, opt := range opts {
		opt(&server)
	}
	var mwrs []func(http.Handler) http.Handler
	if !server.prodMode {
		mwrs = append(mwrs, middleware.Logger)
	}
	mwrs = append(mwrs, middleware.Recoverer)
	mwrs = append(mwrs, server.chiMiddlewares...)
	for _, m := range mwrs {
		server.mux.Use(m)
	}
	return server
}

func (server *Server) RegisterHandler(path string, handler http.HandlerFunc, methods ...string) {
	server.mux.Group(func(r chi.Router) {
		m, _ := middlewares.ApplyTranscoderMiddlewares(server.gatewayPrefixes)

		r.Handle(path, m(handler))
	})
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (server Server) Listen(addr string) error {
	defer CloseEnvironment()

	if !server.prodMode {
		// Go profiling routes
		if server.debugRoute != "" {
			server.mux.Mount(server.debugRoute, middleware.Profiler())
		}

		// Register Swagger files
		if server.swaggerRoute != "" && server.openAPISchema != nil {
			swagger.RegisterSwaggerAssets(server.swaggerRoute, server.openAPISchema, server.mux)
		}

		if server.protosZipRoute != "" && server.protosZip != nil {
			server.mux.Get(server.protosZipRoute, func(w http.ResponseWriter, r *http.Request) {
				timestamp := time.Now().Format("20060102_150405")
				w.Header().Set("Content-Type", "application/zip")
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"proto_sources_%s.zip\"", timestamp))
				w.Write(server.protosZip)
			})
		}
	}

	server.mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ok"))
	})

	server.mux.Group(func(r chi.Router) {
		RegisterServicesHandlers(r, server.registerServiceFns, server.gatewayPrefixes)
		if server.routerGroup != nil {
			r.Group(server.routerGroup)
		}
	})

	c := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
		},
		AllowedHeaders: []string{
			"*",
		},
		ExposedHeaders:       []string{"Link", "Set-Cookie"},
		OptionsSuccessStatus: http.StatusNoContent,
		AllowCredentials:     true,
		MaxAge:               300,
	})

	server.httpServer = &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(c.Handler(server.mux), &http2.Server{}),
	}

	errChan := make(chan error, 1)

	go func() {
		errChan <- server.httpServer.ListenAndServe()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err

	case <-sigChan:
		fmt.Println("Gracefully exiting...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.httpServer.Shutdown(shutdownCtx); err != nil {
			return nil
		}
	}

	return nil
}
