package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/honeynet/honeynet/internal/config"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPlatformOEMPublicDefaultsAndRasterAsset(t *testing.T) {
	router, _, _, _ := platformOEMTestRouter(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/platform/branding", nil)
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("branding status=%d body=%s", response.Code, response.Body.String())
	}
	view := decodePlatformOEMView(t, response.Body.Bytes())
	if view.SystemName != "Honeynet" || view.SystemVersion != "0.24.0-test" || view.Revision != 1 {
		t.Fatalf("unexpected default branding: %#v", view)
	}
	if view.SystemLogoURL == "" || view.CompanyLogoURL == "" {
		t.Fatalf("default logo URLs are missing: %#v", view)
	}

	asset := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, view.SystemLogoURL, nil)
	router.ServeHTTP(asset, request)
	if asset.Code != http.StatusOK || asset.Header().Get("Content-Type") != "image/png" || len(asset.Body.Bytes()) == 0 {
		t.Fatalf("default asset status=%d type=%q bytes=%d", asset.Code, asset.Header().Get("Content-Type"), asset.Body.Len())
	}
	if !strings.Contains(asset.Header().Get("Content-Security-Policy"), "sandbox") || asset.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("unsafe asset headers: %#v", asset.Header())
	}
	etag := asset.Header().Get("ETag")
	cached := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, view.SystemLogoURL, nil)
	request.Header.Set("If-None-Match", etag)
	router.ServeHTTP(cached, request)
	if cached.Code != http.StatusNotModified {
		t.Fatalf("conditional asset status=%d", cached.Code)
	}
}

func TestPlatformOEMMetadataReadDoesNotLoadLogoBlobs(t *testing.T) {
	router, db, _, _ := platformOEMTestRouter(t)
	large := bytes.Repeat([]byte{0x7f}, 1<<20)
	if err := db.Model(&store.PlatformOEMSetting{}).Where("id = ?", store.PlatformOEMSettingID).Updates(map[string]any{
		"system_logo": large, "system_logo_mime": "image/png", "system_logo_sha256": strings.Repeat("a", 64), "system_logo_size": len(large),
		"company_logo": large, "company_logo_mime": "image/png", "company_logo_sha256": strings.Repeat("b", 64), "company_logo_size": len(large),
	}).Error; err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/platform/branding", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("branding status=%d body=%s", response.Code, response.Body.String())
	}
	if response.Body.Len() > 4096 {
		t.Fatalf("metadata response unexpectedly contains logo data: %d bytes", response.Body.Len())
	}
	var metadata store.PlatformOEMSetting
	if err := db.Select(platformOEMMetadataColumns).First(&metadata, "id = ?", store.PlatformOEMSettingID).Error; err != nil {
		t.Fatal(err)
	}
	if len(metadata.SystemLogo) != 0 || len(metadata.CompanyLogo) != 0 {
		t.Fatal("metadata projection loaded logo blobs")
	}
}

func TestPlatformOEMConditionalAssetDoesNotLoadLogoBlob(t *testing.T) {
	router, db, token, _ := platformOEMTestRouter(t)
	uploaded := performPlatformOEMUpload(router, "/api/v1/platform/assets/system", token, 1, "logo.png", testPlatformPNG(t))
	if uploaded.Code != http.StatusOK {
		t.Fatalf("upload status=%d body=%s", uploaded.Code, uploaded.Body.String())
	}
	view := decodePlatformOEMView(t, uploaded.Body.Bytes())
	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, view.SystemLogoURL, nil))
	if first.Code != http.StatusOK || first.Header().Get("ETag") == "" {
		t.Fatalf("asset status=%d etag=%q", first.Code, first.Header().Get("ETag"))
	}

	var blobReads atomic.Int32
	callbackName := "test:count_platform_logo_blob_reads"
	if err := db.Callback().Query().Before("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		for _, selected := range tx.Statement.Selects {
			if strings.TrimSpace(selected) == "system_logo" || strings.Contains(selected, "system_logo,") {
				blobReads.Add(1)
			}
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Callback().Query().Remove(callbackName) })

	cached := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, view.SystemLogoURL, nil)
	request.Header.Set("If-None-Match", first.Header().Get("ETag"))
	router.ServeHTTP(cached, request)
	if cached.Code != http.StatusNotModified {
		t.Fatalf("conditional asset status=%d", cached.Code)
	}
	if count := blobReads.Load(); count != 0 {
		t.Fatalf("conditional request loaded logo BLOB %d time(s)", count)
	}
}

func TestPlatformOEMAdminUpdateValidationConcurrencyAndAudit(t *testing.T) {
	router, db, adminToken, operatorToken := platformOEMTestRouter(t)
	body := `{"system_name":"企业欺骗防御平台","system_version":"2026.1","copyright":"版权所有","customer_service_phone":"400-000-0000","customer_service_email":"soc@example.com","official_website_url":"https://example.com","product_documentation_url":"https://docs.example.com","revision":1}`

	denied := performPlatformOEMJSON(router, http.MethodPut, "/api/v1/platform/settings", body, operatorToken)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("operator update status=%d body=%s", denied.Code, denied.Body.String())
	}
	updated := performPlatformOEMJSON(router, http.MethodPut, "/api/v1/platform/settings", body, adminToken)
	if updated.Code != http.StatusOK {
		t.Fatalf("admin update status=%d body=%s", updated.Code, updated.Body.String())
	}
	view := decodePlatformOEMView(t, updated.Body.Bytes())
	if view.SystemName != "企业欺骗防御平台" || view.Revision != 2 {
		t.Fatalf("unexpected updated view: %#v", view)
	}
	stale := performPlatformOEMJSON(router, http.MethodPut, "/api/v1/platform/settings", body, adminToken)
	if stale.Code != http.StatusConflict || !strings.Contains(stale.Body.String(), "PLATFORM_CONFIG_CHANGED") {
		t.Fatalf("stale update status=%d body=%s", stale.Code, stale.Body.String())
	}
	invalid := strings.Replace(body, "https://example.com", "javascript:alert(1)", 1)
	invalid = strings.Replace(invalid, `"revision":1`, `"revision":2`, 1)
	rejected := performPlatformOEMJSON(router, http.MethodPut, "/api/v1/platform/settings", invalid, adminToken)
	if rejected.Code != http.StatusBadRequest {
		t.Fatalf("unsafe URL status=%d body=%s", rejected.Code, rejected.Body.String())
	}

	var audits []store.AuditLog
	if err := db.Find(&audits).Error; err != nil || len(audits) != 1 {
		t.Fatalf("audits=%d err=%v", len(audits), err)
	}
	if strings.Contains(strings.ToLower(string(audits[0].Detail)), "blob") || strings.Contains(string(audits[0].Detail), "system_logo\"") {
		t.Fatalf("audit unexpectedly includes binary fields: %s", audits[0].Detail)
	}
}

func TestPlatformOEMLogoUploadRejectsSVGNormalizesPNGAndCanReset(t *testing.T) {
	router, db, adminToken, _ := platformOEMTestRouter(t)

	rejected := performPlatformOEMUpload(router, "/api/v1/platform/assets/system", adminToken, 1, "logo.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`))
	if rejected.Code != http.StatusBadRequest || !strings.Contains(rejected.Body.String(), "INVALID_LOGO") {
		t.Fatalf("SVG upload status=%d body=%s", rejected.Code, rejected.Body.String())
	}

	imageData := testPlatformPNG(t)
	uploaded := performPlatformOEMUpload(router, "/api/v1/platform/assets/system", adminToken, 1, "anything.bin", imageData)
	if uploaded.Code != http.StatusOK {
		t.Fatalf("PNG upload status=%d body=%s", uploaded.Code, uploaded.Body.String())
	}
	view := decodePlatformOEMView(t, uploaded.Body.Bytes())
	if view.Revision != 2 || !strings.Contains(view.SystemLogoURL, "v=") || strings.Contains(view.SystemLogoURL, "default") {
		t.Fatalf("unexpected upload response: %#v", view)
	}

	asset := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, view.SystemLogoURL, nil)
	router.ServeHTTP(asset, request)
	if asset.Code != http.StatusOK || asset.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("uploaded asset status=%d type=%q", asset.Code, asset.Header().Get("Content-Type"))
	}
	if _, _, err := image.Decode(bytes.NewReader(asset.Body.Bytes())); err != nil {
		t.Fatalf("stored logo is not a normalized image: %v", err)
	}

	reset := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodDelete, "/api/v1/platform/assets/system?revision=2", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(reset, request)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status=%d body=%s", reset.Code, reset.Body.String())
	}
	view = decodePlatformOEMView(t, reset.Body.Bytes())
	if view.Revision != 3 || view.SystemLogoURL == "" || strings.Contains(view.SystemLogoURL, "v=default") {
		t.Fatalf("unexpected reset response: %#v", view)
	}

	var audits []store.AuditLog
	if err := db.Order("created_at ASC").Find(&audits).Error; err != nil || len(audits) != 2 {
		t.Fatalf("logo audits=%d err=%v", len(audits), err)
	}
	for _, audit := range audits {
		text := string(audit.Detail)
		if strings.Contains(text, "iVBOR") || strings.Contains(text, "system_logo\"") {
			t.Fatalf("logo audit contains binary data: %s", text)
		}
		if !strings.Contains(text, "sha256") || !strings.Contains(text, "bytes") {
			t.Fatalf("logo audit misses safe metadata: %s", text)
		}
	}
}

func TestPlatformOEMLogoAcceptsJPEGAndWebPAndRejectsUnsafeImages(t *testing.T) {
	jpegData := testPlatformJPEG(t)
	webpData, err := base64.StdEncoding.DecodeString("UklGRqgBAABXRUJQVlA4WAoAAAAQAAAADwAADwAAQUxQSMMAAAABJ6KokSTleucYX+ffKpmImP90cY3gJjDi4Yt3MsjBEVyDKzDosHgVjnhRNcEIDAJPkqBqsFUZHNa2bUYvTsZ2PLbtd/uvKa4hov9J0f2PkPe6REkkGzolkTTzFG0Ox9PlFiD0CxS+kOGDtxoynjaCfx0pfk52CPuInrOR75lzRugygtv4zEiy90UwfSD9NheMITJWLaXWayO8XeOlWRXVnIGk2W6WdYoYMQ+KqixQNPowgt+6a1BSKbUtz+lUFAoBAAAAVlA4IL4AAACQAgCdASoQABAAAwA0JbACdDBPCIUMfAMdCCz96AD+/XSg/QKbH4r3Q3ycN/bSDK/T/zVo4u6nvclvG/SqxWOuup+XhN9BojvaW+Tv+MvxvX/hr/o/5Qns9LtmX/+qKdl/yWznhuasl7nkxvSTI4xf3Y85VSB/lU/8Ofj/b9JrA+ifvIOYZm2x1RP/dhfmsf5diuSfR7+z+r/+HR3zEo/+XM/B+vkYw73Pzx+ROaAB/ZoBSzEs3rzZe6qsAAAA")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name     string
		contents []byte
		mime     string
	}{
		{name: "jpeg", contents: jpegData, mime: "image/jpeg"},
		{name: "webp normalized to png", contents: webpData, mime: "image/png"},
	} {
		t.Run(test.name, func(t *testing.T) {
			router, _, token, _ := platformOEMTestRouter(t)
			response := performPlatformOEMUpload(router, "/api/v1/platform/assets/company", token, 1, "logo.dat", test.contents)
			if response.Code != http.StatusOK {
				t.Fatalf("upload status=%d body=%s", response.Code, response.Body.String())
			}
			view := decodePlatformOEMView(t, response.Body.Bytes())
			asset := httptest.NewRecorder()
			router.ServeHTTP(asset, httptest.NewRequest(http.MethodGet, view.CompanyLogoURL, nil))
			if asset.Code != http.StatusOK || asset.Header().Get("Content-Type") != test.mime {
				t.Fatalf("asset status=%d mime=%q", asset.Code, asset.Header().Get("Content-Type"))
			}
		})
	}

	unsafeCases := []struct {
		name     string
		contents []byte
	}{
		{name: "fake extension", contents: []byte("not an image")},
		{name: "oversize", contents: bytes.Repeat([]byte{'A'}, maxOEMLogoUploadBytes+1)},
		{name: "pixel bomb", contents: testPlatformLargePNG(t)},
	}
	for _, test := range unsafeCases {
		t.Run(test.name, func(t *testing.T) {
			router, _, token, _ := platformOEMTestRouter(t)
			response := performPlatformOEMUpload(router, "/api/v1/platform/assets/system", token, 1, "logo.png", test.contents)
			if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "INVALID_LOGO") {
				t.Fatalf("unsafe image status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestPlatformOEMRejectsTextXSSAndUnicodeControls(t *testing.T) {
	base := `{"system_name":"Honeynet","system_version":"2026.1","copyright":"copyright","customer_service_phone":"400","customer_service_email":"soc@example.com","official_website_url":"https://example.com","product_documentation_url":"https://docs.example.com","revision":1}`
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "html", body: strings.Replace(base, `"copyright":"copyright"`, `"copyright":"<img src=x onerror=alert(1)>"`, 1)},
		{name: "bidi control", body: strings.Replace(base, `"copyright":"copyright"`, `"copyright":"safe\u202Eexe"`, 1)},
		{name: "control", body: strings.Replace(base, `"copyright":"copyright"`, `"copyright":"safe\u0001text"`, 1)},
		{name: "email", body: strings.Replace(base, "soc@example.com", "invalid@@example.com", 1)},
		{name: "URL credentials", body: strings.Replace(base, "https://example.com", "https://user:pass@example.com", 1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			isolatedRouter, _, isolatedToken, _ := platformOEMTestRouter(t)
			response := performPlatformOEMJSON(isolatedRouter, http.MethodPut, "/api/v1/platform/settings", test.body, isolatedToken)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestPlatformOEMConcurrentRevisionHasSingleWinner(t *testing.T) {
	router, _, token, _ := platformOEMTestRouter(t)
	const workers = 8
	start := make(chan struct{})
	statuses := make(chan int, workers)
	var wait sync.WaitGroup
	for index := 0; index < workers; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			body := `{"system_name":"并发版本` + strconv.Itoa(index) + `","system_version":"1","copyright":"","customer_service_phone":"","customer_service_email":"","official_website_url":"","product_documentation_url":"","revision":1}`
			statuses <- performPlatformOEMJSON(router, http.MethodPut, "/api/v1/platform/settings", body, token).Code
		}(index)
	}
	close(start)
	wait.Wait()
	close(statuses)
	var successes, conflicts int
	for status := range statuses {
		if status == http.StatusOK {
			successes++
		} else if status == http.StatusConflict {
			conflicts++
		}
	}
	if successes != 1 || conflicts != workers-1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
}

func platformOEMTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, string, string) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}
	if err := db.AutoMigrate(&store.User{}, &store.AuditLog{}, &store.PlatformOEMSetting{}); err != nil {
		t.Fatal(err)
	}
	admin := store.User{Base: store.NewBase(), Username: "oem-admin", PasswordHash: "unused", Role: "admin", Enabled: true, TokenVersion: 1}
	operator := store.User{Base: store.NewBase(), Username: "oem-operator", PasswordHash: "unused", Role: "operator", Enabled: true, TokenVersion: 1}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&operator).Error; err != nil {
		t.Fatal(err)
	}
	manager := NewTokenManager("platform-oem-test-secret-with-adequate-length", time.Hour).WithUserStore(db)
	adminToken, _, err := manager.Issue(AuthUser{ID: admin.ID, Username: admin.Username, Role: admin.Role, TokenVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	operatorToken, _, err := manager.Issue(AuthUser{ID: operator.ID, Username: operator.Username, Role: operator.Role, TokenVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	api := &API{cfg: config.Config{Version: "0.24.0-test"}, db: db, tokens: manager}
	router := gin.New()
	router.Use(requestID(), securityHeaders())
	v1 := router.Group("/api/v1")
	v1.GET("/platform/branding", api.platformOEM)
	v1.GET("/platform/assets/:kind", api.servePlatformOEMLogo)
	authed := v1.Group("")
	authed.Use(manager.Middleware(), api.audit())
	authed.GET("/platform/settings", requireRoles("admin"), api.platformOEM)
	authed.PUT("/platform/settings", requireRoles("admin"), api.updatePlatformOEM)
	authed.POST("/platform/assets/:kind", requireRoles("admin"), api.uploadPlatformOEMLogo)
	authed.DELETE("/platform/assets/:kind", requireRoles("admin"), api.deletePlatformOEMLogo)
	authed.DELETE("/platform/config/logos/:kind", requireRoles("admin"), api.deletePlatformOEMLogo)
	return router, db, adminToken, operatorToken
}

func performPlatformOEMJSON(router http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(response, request)
	return response
}

func performPlatformOEMUpload(router http.Handler, path, token string, revision int64, filename string, content []byte) *httptest.ResponseRecorder {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("revision", strconv.FormatInt(revision, 10))
	file, _ := writer.CreateFormFile("file", filename)
	_, _ = file.Write(content)
	_ = writer.Close()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, path, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(response, request)
	return response
}

func decodePlatformOEMView(t *testing.T, body []byte) platformOEMView {
	t.Helper()
	var response struct {
		Data platformOEMView `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("decode platform response: %v (%s)", err, body)
	}
	return response.Data
}

func testPlatformPNG(t *testing.T) []byte {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, 48, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 48; x++ {
			canvas.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 8), B: 80, A: 255})
		}
	}
	var output bytes.Buffer
	if err := png.Encode(&output, canvas); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func testPlatformJPEG(t *testing.T) []byte {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, 32, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 32; x++ {
			canvas.Set(x, y, color.RGBA{R: 20, G: uint8(x * 6), B: uint8(y * 12), A: 255})
		}
	}
	var output bytes.Buffer
	if err := jpeg.Encode(&output, canvas, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func testPlatformLargePNG(t *testing.T) []byte {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, 2001, 2000))
	var output bytes.Buffer
	if err := png.Encode(&output, canvas); err != nil {
		t.Fatal(err)
	}
	if output.Len() > maxOEMLogoUploadBytes {
		t.Fatalf("pixel test image unexpectedly exceeds byte limit: %d", output.Len())
	}
	return output.Bytes()
}
