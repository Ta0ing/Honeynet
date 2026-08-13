package pots

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/potcert"
	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func webTemplateTestRoot() string {
	return filepath.Join("..", "..", "..", "honeypot-templates-server", "services")
}

type webTemplateTestProvider struct {
	TLSProvider
	root string
}

func (p webTemplateTestProvider) TemplateRoot() string { return p.root }

func newWebTemplateTestProvider(t *testing.T) webTemplateTestProvider {
	t.Helper()
	provider, err := potcert.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return webTemplateTestProvider{TLSProvider: provider, root: webTemplateTestRoot()}
}

func webTemplateHTTPClient() *http.Client {
	return &http.Client{
		Transport:     &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, // #nosec G402 -- generated local honeypot certificate
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Timeout:       3 * time.Second,
	}
}

func TestWebTemplateManifestAvailabilityMatchesSuppliedPack(t *testing.T) {
	manifest, err := loadWebTemplateManifest(webTemplateTestRoot())
	if err != nil {
		t.Fatal(err)
	}
	configured := 0
	for code, entry := range manifest {
		if entry.Root != "" && entry.Index != "" {
			configured++
			if _, registered := webTemplateCodes[code]; !registered {
				t.Errorf("configured template %q is not registered", code)
			}
		}
	}
	if configured != 68 {
		t.Fatalf("configured static templates = %d, want 68", configured)
	}
	for code := range webTemplateCodes {
		entry, configured := manifest[code]
		if !configured || entry.Root == "" || entry.Index == "" {
			t.Errorf("registered template %q is absent from config.json", code)
		}
	}
	available := availableWebTemplates(webTemplateTestRoot())
	if len(available) != 67 {
		t.Fatalf("available static templates = %d, want 67", len(available))
	}
	if available["router-cmcc"] {
		t.Fatal("router-cmcc must not be advertised while its resource directory is absent")
	}
	if len(SupportedCodesAt(webTemplateTestRoot())) != 103 {
		t.Fatalf("runtime capabilities = %d, want 103", len(SupportedCodesAt(webTemplateTestRoot())))
	}
}

func TestEveryAvailableWebTemplateStartsNatively(t *testing.T) {
	provider := newWebTemplateTestProvider(t)
	client := webTemplateHTTPClient()
	available := availableWebTemplates(webTemplateTestRoot())
	for code := range available {
		code := code
		t.Run(code, func(t *testing.T) {
			service, err := New(code, provider)
			if err != nil {
				t.Fatal(err)
			}
			port := freePort(t)
			events := make(chan protocol.Event, 4)
			target := protocol.PotTarget{ID: code, Service: code, Port: port, DesiredStatus: "running", Config: map[string]any{"bind": "127.0.0.1"}}
			if err := service.Start(context.Background(), target, func(event protocol.Event) { events <- event }); err != nil {
				t.Fatal(err)
			}
			defer service.Stop()
			response, err := client.Get("https://127.0.0.1:" + strconv.Itoa(port) + "/")
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode == http.StatusNotFound || response.StatusCode >= 500 {
				t.Fatalf("root returned HTTP %d", response.StatusCode)
			}
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			select {
			case event := <-events:
				if event.EventType != "web.request" || event.Payload["template_source"] != "honeypot-templates-server" {
					t.Fatalf("unexpected event: %#v", event)
				}
			case <-time.After(time.Second):
				t.Fatal("request event was not captured")
			}
		})
	}
}

func TestWebTemplateTomcatServesOriginalBytes(t *testing.T) {
	provider := newWebTemplateTestProvider(t)
	service, err := New("tomcat", provider)
	if err != nil {
		t.Fatal(err)
	}
	port := freePort(t)
	if err := service.Start(context.Background(), protocol.PotTarget{ID: "tomcat", Service: "tomcat", Port: port, Config: map[string]any{"bind": "127.0.0.1"}}, func(protocol.Event) {}); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	response, err := webTemplateHTTPClient().Get("https://127.0.0.1:" + strconv.Itoa(port) + "/")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	want, err := os.ReadFile(filepath.Join(webTemplateTestRoot(), "tomcat", "root", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("Tomcat response does not equal the supplied template index")
	}
}

func TestWebTemplateQueryAssetAndResponseMetadata(t *testing.T) {
	provider := newWebTemplateTestProvider(t)
	service, err := New("phpadmin", provider)
	if err != nil {
		t.Fatal(err)
	}
	port := freePort(t)
	if err := service.Start(context.Background(), protocol.PotTarget{ID: "phpadmin", Service: "phpadmin", Port: port, Config: map[string]any{"bind": "127.0.0.1"}}, func(protocol.Event) {}); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	path := "/js/codemirror/lib/codemirror.css?v=4.6.6"
	response, err := webTemplateHTTPClient().Get("https://127.0.0.1:" + strconv.Itoa(port) + path)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	want, _ := os.ReadFile(filepath.Join(webTemplateTestRoot(), "phpadmin", "root", "js", "codemirror", "lib", "codemirror.css？v=4.6.6"))
	if string(body) != string(want) {
		t.Fatal("query-addressed asset bytes do not match the supplied resource")
	}
	if response.Header.Get("Content-Encoding") != "" {
		t.Fatalf("decoded resource retained stale Content-Encoding: %q", response.Header.Get("Content-Encoding"))
	}
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "text/css") {
		t.Fatalf("Content-Type = %q", response.Header.Get("Content-Type"))
	}
}

func TestWebTemplateTemplateCapturesPostedCredentials(t *testing.T) {
	provider := newWebTemplateTestProvider(t)
	service, err := New("oa-tongda", provider)
	if err != nil {
		t.Fatal(err)
	}
	port := freePort(t)
	events := make(chan protocol.Event, 8)
	if err := service.Start(context.Background(), protocol.PotTarget{ID: "oa", Service: "oa-tongda", Port: port, Config: map[string]any{"bind": "127.0.0.1"}}, func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	response, err := webTemplateHTTPClient().PostForm("https://127.0.0.1:"+strconv.Itoa(port)+"/login", url.Values{"username": {"alice"}, "password": {"secret"}})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	want, _ := os.ReadFile(filepath.Join(webTemplateTestRoot(), "oa-tongda", "root", "index.html"))
	if response.StatusCode != http.StatusOK || string(body) != string(want) {
		t.Fatalf("unmapped form response = HTTP %d, %d bytes; want original index", response.StatusCode, len(body))
	}
	for range 2 {
		select {
		case event := <-events:
			if event.EventType == "web.credential" {
				if event.Payload["username"] != "alice" || event.Payload["password"] != "secret" {
					t.Fatalf("credential payload = %#v", event.Payload)
				}
				return
			}
		case <-time.After(time.Second):
			t.Fatal("credential event was not captured")
		}
	}
	t.Fatal("credential event was not captured")
}

func TestWebTemplateRedirectQueryAndMethodRoutes(t *testing.T) {
	provider := newWebTemplateTestProvider(t)
	service, err := New("isport", provider)
	if err != nil {
		t.Fatal(err)
	}
	port := freePort(t)
	if err := service.Start(context.Background(), protocol.PotTarget{ID: "isport", Service: "isport", Port: port, Config: map[string]any{"bind": "127.0.0.1"}}, func(protocol.Event) {}); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	client := webTemplateHTTPClient()
	base := "https://127.0.0.1:" + strconv.Itoa(port)
	response, err := client.Get(base + "/")
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	location := response.Header.Get("Location")
	if response.StatusCode != http.StatusFound || !strings.HasPrefix(location, "/service/logon?path=") {
		t.Fatalf("root redirect = HTTP %d, Location %q", response.StatusCode, location)
	}
	response, err = client.Get(base + location)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK || len(body) == 0 {
		t.Fatalf("query route = HTTP %d, %d bytes", response.StatusCode, len(body))
	}
	response, err = client.Post(base+"/service/generateVerifyCode", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("method route = HTTP %d", response.StatusCode)
	}
}
