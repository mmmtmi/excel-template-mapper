package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/mmmtmi/excel-template-mapper/graph"
	"github.com/mmmtmi/excel-template-mapper/internal/dbconn"
	"github.com/mmmtmi/excel-template-mapper/internal/service"
	"github.com/mmmtmi/excel-template-mapper/internal/store/mysql"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	cfg, err := dbconn.LoadMySQLConfigFromEnv(".env")
	if err != nil {
		log.Fatal(err)
	}

	db, err := dbconn.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Edit by CODEX: wire Processor for Excel conversion mutations.
	tplRepo := &mysql.TemplateRepository{DB: db}
	ruleRepo := &mysql.RuleRepository{DB: db}
	proc := service.NewProcessor(tplRepo, ruleRepo)

	resolvers := &graph.Resolver{
		DB:        db,
		Processor: proc,
	}
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolvers}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	// Edit by CODEX: required for `Upload` scalar (multipart/form-data).
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	// Edit by CODEX: apply request body limit (includes multipart metadata).
	const maxUpload = 7 << 20 // 7MiB
	http.Handle("/query", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Edit by CODEX: CORS for local React dev servers.
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:5173" || origin == "http://127.0.0.1:5173" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
		}
		srv.ServeHTTP(w, r)
	}))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	// Edit by CODEX: use http.Server to enable timeouts.
	httpSrv := &http.Server{
		Addr:              ":" + port,
		Handler:           nil, // use http.DefaultServeMux
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(httpSrv.ListenAndServe())
}
