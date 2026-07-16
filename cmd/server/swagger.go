package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

//go:embed swagger.json
var swaggerSpec []byte

const swaggerUI = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <title>NB_Api — Swagger</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *::before, *::after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    #swagger-ui { max-width: 1460px; margin: 0 auto; padding: 20px; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "/swagger.json",
      dom_id: "#swagger-ui",
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIBundle.SwaggerUIStandalonePreset
      ],
      layout: "BaseLayout",
      deepLinking: true
    });
  </script>
</body>
</html>`

func (s *server) handleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerUI))
}

func (s *server) handleSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Parse the embedded swagger spec
	var spec map[string]any
	if err := json.Unmarshal(swaggerSpec, &spec); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"failed to parse swagger spec"}`))
		return
	}

	var servers []map[string]string

	if s.swaggerURL != "" {
		// Use the configured URL (from flag or SWAGGER_URL env var)
		servers = []map[string]string{
			{"url": s.swaggerURL, "description": "Servidor configurado"},
			{"url": "http://localhost:8081", "description": "Desenvolvimento (dev)"},
			{"url": "http://localhost:8080", "description": "Produção (Docker)"},
		}
	} else {
		// Dynamic detection from the request
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
			scheme = strings.Split(fwd, ",")[0]
		}
		host := r.Host
		if fwdHost := r.Header.Get("X-Forwarded-Host"); fwdHost != "" {
			host = strings.Split(fwdHost, ",")[0]
		}
		if host == "" {
			host = "localhost:8080"
		}
		dynamicURL := fmt.Sprintf("%s://%s", scheme, host)
		servers = []map[string]string{
			{"url": dynamicURL, "description": fmt.Sprintf("Atual (%s)", host)},
			{"url": "http://localhost:8081", "description": "Desenvolvimento (dev)"},
			{"url": "http://localhost:8080", "description": "Produção (Docker)"},
		}
	}

	spec["servers"] = servers

	// Re-encode preserving the original formatting (2-space indent)
	result, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"failed to encode swagger spec"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}
