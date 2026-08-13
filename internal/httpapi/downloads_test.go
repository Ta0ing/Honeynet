package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/honeynet/honeynet/internal/config"
)

func TestDownloadAgentChecksum(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	contents := []byte("honeynet-agent-test-binary")
	if err := os.WriteFile(filepath.Join(dir, "honeynet-agent-linux-amd64"), contents, 0755); err != nil {
		t.Fatal(err)
	}
	api := &API{cfg: config.Config{DownloadsDir: dir}}
	router := gin.New()
	router.GET("/download/agent/:os/:arch/sha256", api.downloadAgentChecksum)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/download/agent/linux/amd64/sha256", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	want := sha256.Sum256(contents)
	if strings.TrimSpace(recorder.Body.String()) != hex.EncodeToString(want[:]) {
		t.Fatalf("checksum = %q, want %q", recorder.Body.String(), hex.EncodeToString(want[:]))
	}
}

func TestDownloadAgentRejectsUnsupportedBuild(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &API{cfg: config.Config{DownloadsDir: t.TempDir()}}
	router := gin.New()
	router.GET("/download/agent/:os/:arch", api.downloadAgent)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/download/agent/darwin/amd64", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestDownloadAgentGuardChecksum(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	contents := []byte("stable-agent-guard")
	if err := os.WriteFile(filepath.Join(dir, "honeynet-agent-guard-linux-amd64"), contents, 0755); err != nil {
		t.Fatal(err)
	}
	api := &API{cfg: config.Config{DownloadsDir: dir}}
	router := gin.New()
	router.GET("/download/agent-guard/:os/:arch/sha256", api.downloadAgentGuardChecksum)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/download/agent-guard/linux/amd64/sha256", nil))
	want := sha256.Sum256(contents)
	if recorder.Code != http.StatusOK || strings.TrimSpace(recorder.Body.String()) != hex.EncodeToString(want[:]) {
		t.Fatalf("status=%d checksum=%q", recorder.Code, recorder.Body.String())
	}
	if _, _, ok := api.agentGuardBuild("windows", "amd64"); ok {
		t.Fatal("Windows must not expose the Linux process guard")
	}
}

func TestDownloadTemplateBundleChecksum(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	contents := []byte("web-template-test-bundle")
	if err := os.WriteFile(filepath.Join(dir, "honeypot-templates-server.tar.gz"), contents, 0644); err != nil {
		t.Fatal(err)
	}
	api := &API{cfg: config.Config{DownloadsDir: dir}}
	router := gin.New()
	router.GET("/download/templates/:format/sha256", api.downloadTemplateBundleChecksum)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/download/templates/tar.gz/sha256", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	want := sha256.Sum256(contents)
	if strings.TrimSpace(recorder.Body.String()) != hex.EncodeToString(want[:]) {
		t.Fatalf("checksum = %q, want %q", recorder.Body.String(), hex.EncodeToString(want[:]))
	}
}

func TestInstallersConfigureNativeTemplateRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &API{}
	for _, test := range []struct {
		name string
		path string
		call gin.HandlerFunc
		want []string
	}{
		{name: "linux", path: "/install-agent.sh", call: api.installShell, want: []string{"--console-ca", "--cacert \"$CONSOLE_CA\"", "/download/templates/tar.gz", "/download/agent-guard/$OS/$ARCH", "honeynet-agent-guard", "--template-root \"$TEMPLATE_DIR/services\"", "--enroll-only", "tar -xzf"}},
		{name: "windows", path: "/install-agent.ps1", call: api.installPowerShell, want: []string{"/download/templates/zip", "--template-root $templateRoot", "--enroll-only", "Expand-Archive"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.GET(test.path, test.call)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			for _, wanted := range test.want {
				if !strings.Contains(recorder.Body.String(), wanted) {
					t.Fatalf("installer does not contain %q", wanted)
				}
			}
		})
	}
}
