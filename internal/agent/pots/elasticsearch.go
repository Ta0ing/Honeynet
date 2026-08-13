package pots

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxElasticsearchBody = 64 << 10

type ElasticsearchService struct {
	listener net.Listener
	server   *http.Server
}

func (s *ElasticsearchService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	clusterName := configString(target.Config, "cluster_name", "production-search")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(r.Body, maxElasticsearchBody))
		src, dst := endpoint(remoteAddr(r.RemoteAddr)), endpoint(localAddr(r.Context()))
		rawRequest, bodyBase64 := httpRequestSnapshot(r, body, false)
		payload := map[string]any{
			"method": r.Method, "path": r.URL.RequestURI(), "user_agent": r.UserAgent(),
			"content_type": r.Header.Get("Content-Type"), "headers": r.Header, "body": string(body), "raw_request": rawRequest,
		}
		if bodyBase64 != "" {
			payload["body"] = "[binary body]"
			payload["body_base64"] = bodyBase64
		}
		sink(protocol.NewEvent("elasticsearch.request", src, dst, payload, "database", "http"))
		if username, password, ok := r.BasicAuth(); ok {
			sink(protocol.NewEvent("elasticsearch.credential", src, dst, map[string]any{
				"username": username, "password": password, "path": r.URL.Path,
			}, "credential"))
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		response := elasticsearchResponse(clusterName, r.URL.Path)
		_ = json.NewEncoder(w).Encode(response)
	})
	s.server = &http.Server{
		Handler:           http.MaxBytesHandler(handler, maxElasticsearchBody),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ConnContext: func(ctx context.Context, conn net.Conn) context.Context {
			return context.WithValue(ctx, localAddressKey{}, conn.LocalAddr())
		},
	}
	go func() { _ = s.server.Serve(listener) }()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.server.Shutdown(shutdownCtx)
	}()
	return nil
}

func elasticsearchResponse(clusterName, path string) map[string]any {
	switch {
	case path == "/_cluster/health":
		return map[string]any{
			"cluster_name": clusterName, "status": "green", "timed_out": false,
			"number_of_nodes": 3, "number_of_data_nodes": 3, "active_primary_shards": 12,
			"active_shards": 24, "unassigned_shards": 0,
		}
	case strings.HasPrefix(path, "/_cat/indices"):
		return map[string]any{"indices": []any{}}
	case strings.HasPrefix(path, "/_nodes"):
		return map[string]any{"cluster_name": clusterName, "nodes": map[string]any{}}
	default:
		return map[string]any{
			"name": "es-data-01", "cluster_name": clusterName, "cluster_uuid": "Z4kH2f3QQYy80zR4eP2f1A",
			"version": map[string]any{"number": "8.12.2", "build_flavor": "default", "build_type": "tar", "lucene_version": "9.9.2"},
			"tagline": "You Know, for Search",
		}
	}
}

func (s *ElasticsearchService) Stop() error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
