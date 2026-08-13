package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/honeynet/honeynet/internal/store"
	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxOEMLogoUploadBytes = 2 << 20
	maxOEMLogoPixels      = 4_000_000
)

type platformOEMView struct {
	SystemName              string `json:"system_name"`
	SystemVersion           string `json:"system_version"`
	Copyright               string `json:"copyright"`
	SystemLogoURL           string `json:"system_logo_url"`
	CompanyLogoURL          string `json:"company_logo_url"`
	CustomerServicePhone    string `json:"customer_service_phone"`
	CustomerServiceEmail    string `json:"customer_service_email"`
	OfficialWebsiteURL      string `json:"official_website_url"`
	ProductDocumentationURL string `json:"product_documentation_url"`
	Revision                int64  `json:"revision"`
}

type platformOEMUpdate struct {
	SystemName              string `json:"system_name"`
	SystemVersion           string `json:"system_version"`
	Copyright               string `json:"copyright"`
	CustomerServicePhone    string `json:"customer_service_phone"`
	CustomerServiceEmail    string `json:"customer_service_email"`
	OfficialWebsiteURL      string `json:"official_website_url"`
	ProductDocumentationURL string `json:"product_documentation_url"`
	Revision                int64  `json:"revision"`
}

func (a *API) platformOEM(c *gin.Context) {
	item, err := a.loadPlatformOEM(c.Request.Context())
	if err != nil {
		fail(c, http.StatusInternalServerError, "PLATFORM_CONFIG_READ_FAILED", "读取平台配置失败")
		return
	}
	ok(c, platformOEMPublicView(item))
}

func (a *API) updatePlatformOEM(c *gin.Context) {
	var request platformOEMUpdate
	if c.ShouldBindJSON(&request) != nil {
		fail(c, http.StatusBadRequest, "INVALID_ARGUMENT", "平台配置格式错误")
		return
	}
	normalized, err := normalizePlatformOEMUpdate(request)
	if err != nil {
		fail(c, http.StatusBadRequest, "INVALID_PLATFORM_CONFIG", err.Error())
		return
	}
	before, err := a.loadPlatformOEM(c.Request.Context())
	if err != nil {
		fail(c, http.StatusInternalServerError, "PLATFORM_CONFIG_READ_FAILED", "读取平台配置失败")
		return
	}
	if request.Revision < 1 || request.Revision != before.Revision {
		fail(c, http.StatusConflict, "PLATFORM_CONFIG_CHANGED", "平台配置已被其他管理员修改，请刷新后重试")
		return
	}
	updates := map[string]any{
		"system_name": normalized.SystemName, "system_version": normalized.SystemVersion,
		"copyright": normalized.Copyright, "customer_service_phone": normalized.CustomerServicePhone,
		"customer_service_email": normalized.CustomerServiceEmail, "official_website_url": normalized.OfficialWebsiteURL,
		"product_documentation_url": normalized.ProductDocumentationURL,
		"revision":                  gorm.Expr("revision + 1"), "updated_at": time.Now(),
	}
	result := a.db.WithContext(c.Request.Context()).Model(&store.PlatformOEMSetting{}).
		Where("id = ? AND revision = ?", store.PlatformOEMSettingID, request.Revision).Updates(updates)
	if result.Error != nil {
		fail(c, http.StatusInternalServerError, "PLATFORM_CONFIG_SAVE_FAILED", "保存平台配置失败")
		return
	}
	if result.RowsAffected != 1 {
		fail(c, http.StatusConflict, "PLATFORM_CONFIG_CHANGED", "平台配置已被其他管理员修改，请刷新后重试")
		return
	}
	// Construct the committed revision from the CAS input instead of reading it
	// back. A second administrator may commit immediately after this request;
	// a post-write SELECT could otherwise return and audit their revision.
	after := before
	after.SystemName = normalized.SystemName
	after.SystemVersion = normalized.SystemVersion
	after.Copyright = normalized.Copyright
	after.CustomerServicePhone = normalized.CustomerServicePhone
	after.CustomerServiceEmail = normalized.CustomerServiceEmail
	after.OfficialWebsiteURL = normalized.OfficialWebsiteURL
	after.ProductDocumentationURL = normalized.ProductDocumentationURL
	after.Revision = request.Revision + 1
	setAuditChange(c, "platform_oem", store.PlatformOEMSettingID, platformOEMPublicView(before), platformOEMPublicView(after))
	ok(c, platformOEMPublicView(after))
}

func (a *API) uploadPlatformOEMLogo(c *gin.Context) {
	kind, okKind := platformLogoKind(c.Param("kind"))
	if !okKind {
		fail(c, http.StatusNotFound, "PLATFORM_ASSET_NOT_FOUND", "平台图片类型不存在")
		return
	}
	revision, err := strconv.ParseInt(strings.TrimSpace(c.PostForm("revision")), 10, 64)
	if err != nil || revision < 1 {
		fail(c, http.StatusBadRequest, "INVALID_REVISION", "请提交当前平台配置版本")
		return
	}
	header, err := c.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "INVALID_LOGO", "请选择 PNG、JPEG 或 WebP 图片")
		return
	}
	content, mimeType, err := normalizeOEMLogo(header)
	if err != nil {
		fail(c, http.StatusBadRequest, "INVALID_LOGO", err.Error())
		return
	}
	digest := sha256.Sum256(content)
	digestText := hex.EncodeToString(digest[:])
	before, err := a.loadPlatformOEM(c.Request.Context())
	if err != nil {
		fail(c, http.StatusInternalServerError, "PLATFORM_CONFIG_READ_FAILED", "读取平台配置失败")
		return
	}
	if before.Revision != revision {
		fail(c, http.StatusConflict, "PLATFORM_CONFIG_CHANGED", "平台配置已被其他管理员修改，请刷新后重试")
		return
	}
	updates := map[string]any{"revision": gorm.Expr("revision + 1"), "updated_at": time.Now()}
	if kind == "system" {
		updates["system_logo"] = content
		updates["system_logo_mime"] = mimeType
		updates["system_logo_sha256"] = digestText
		updates["system_logo_size"] = len(content)
	} else {
		updates["company_logo"] = content
		updates["company_logo_mime"] = mimeType
		updates["company_logo_sha256"] = digestText
		updates["company_logo_size"] = len(content)
	}
	result := a.db.WithContext(c.Request.Context()).Model(&store.PlatformOEMSetting{}).
		Where("id = ? AND revision = ?", store.PlatformOEMSettingID, revision).Updates(updates)
	if result.Error != nil {
		fail(c, http.StatusInternalServerError, "PLATFORM_ASSET_SAVE_FAILED", "保存平台图片失败")
		return
	}
	if result.RowsAffected != 1 {
		fail(c, http.StatusConflict, "PLATFORM_CONFIG_CHANGED", "平台配置已被其他管理员修改，请刷新后重试")
		return
	}
	item := before
	item.Revision = revision + 1
	if kind == "system" {
		item.SystemLogoMIME = mimeType
		item.SystemLogoSHA256 = digestText
		item.SystemLogoSize = int64(len(content))
	} else {
		item.CompanyLogoMIME = mimeType
		item.CompanyLogoSHA256 = digestText
		item.CompanyLogoSize = int64(len(content))
	}
	setAuditChange(c, "platform_logo", kind, platformLogoAuditView(before, kind), platformLogoAuditView(item, kind))
	ok(c, platformOEMPublicView(item))
}

func (a *API) deletePlatformOEMLogo(c *gin.Context) {
	kind, okKind := platformLogoKind(c.Param("kind"))
	if !okKind {
		fail(c, http.StatusNotFound, "PLATFORM_ASSET_NOT_FOUND", "平台图片类型不存在")
		return
	}
	revision, err := strconv.ParseInt(strings.TrimSpace(c.Query("revision")), 10, 64)
	if err != nil || revision < 1 {
		fail(c, http.StatusBadRequest, "INVALID_REVISION", "请提交当前平台配置版本")
		return
	}
	before, err := a.loadPlatformOEM(c.Request.Context())
	if err != nil {
		fail(c, http.StatusInternalServerError, "PLATFORM_CONFIG_READ_FAILED", "读取平台配置失败")
		return
	}
	if before.Revision != revision {
		fail(c, http.StatusConflict, "PLATFORM_CONFIG_CHANGED", "平台配置已被其他管理员修改，请刷新后重试")
		return
	}
	updates := map[string]any{"revision": gorm.Expr("revision + 1"), "updated_at": time.Now()}
	if kind == "system" {
		updates["system_logo"] = nil
		updates["system_logo_mime"] = ""
		updates["system_logo_sha256"] = ""
		updates["system_logo_size"] = 0
	} else {
		updates["company_logo"] = nil
		updates["company_logo_mime"] = ""
		updates["company_logo_sha256"] = ""
		updates["company_logo_size"] = 0
	}
	result := a.db.WithContext(c.Request.Context()).Model(&store.PlatformOEMSetting{}).
		Where("id = ? AND revision = ?", store.PlatformOEMSettingID, revision).Updates(updates)
	if result.Error != nil {
		fail(c, http.StatusInternalServerError, "PLATFORM_ASSET_DELETE_FAILED", "清除平台图片失败")
		return
	}
	if result.RowsAffected != 1 {
		fail(c, http.StatusConflict, "PLATFORM_CONFIG_CHANGED", "平台配置已被其他管理员修改，请刷新后重试")
		return
	}
	item := before
	item.Revision = revision + 1
	if kind == "system" {
		item.SystemLogoMIME = ""
		item.SystemLogoSHA256 = ""
		item.SystemLogoSize = 0
	} else {
		item.CompanyLogoMIME = ""
		item.CompanyLogoSHA256 = ""
		item.CompanyLogoSize = 0
	}
	setAuditChange(c, "platform_logo", kind, platformLogoAuditView(before, kind), platformLogoAuditView(item, kind))
	ok(c, platformOEMPublicView(item))
}

func (a *API) servePlatformOEMLogo(c *gin.Context) {
	kind, okKind := platformLogoKind(c.Param("kind"))
	if !okKind {
		c.Status(http.StatusNotFound)
		return
	}
	for attempt := 0; attempt < 2; attempt++ {
		var metadata store.PlatformOEMSetting
		metadataFields := "system_logo_mime, system_logo_sha256"
		if kind == "company" {
			metadataFields = "company_logo_mime, company_logo_sha256"
		}
		if err := a.db.WithContext(c.Request.Context()).Select(metadataFields).First(&metadata, "id = ?", store.PlatformOEMSettingID).Error; err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		mimeType, digest := metadata.SystemLogoMIME, metadata.SystemLogoSHA256
		if kind == "company" {
			mimeType, digest = metadata.CompanyLogoMIME, metadata.CompanyLogoSHA256
		}
		custom := mimeType != "" && digest != ""
		var content []byte
		if !custom {
			content, mimeType, digest = defaultPlatformLogo()
		}
		etag := `"` + digest + `"`
		setPlatformLogoHeaders(c, kind, mimeType, etag)
		if c.GetHeader("If-None-Match") == etag {
			c.Status(http.StatusNotModified)
			return
		}
		if custom {
			var asset store.PlatformOEMSetting
			blobField, digestField := "system_logo", "system_logo_sha256"
			if kind == "company" {
				blobField, digestField = "company_logo", "company_logo_sha256"
			}
			err := a.db.WithContext(c.Request.Context()).Select(blobField).
				Where("id = ? AND "+digestField+" = ?", store.PlatformOEMSettingID, digest).First(&asset).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue // the logo changed between metadata and BLOB reads
			}
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			content = asset.SystemLogo
			if kind == "company" {
				content = asset.CompanyLogo
			}
			if len(content) == 0 {
				continue
			}
		}
		c.Data(http.StatusOK, mimeType, content)
		return
	}
	c.Status(http.StatusServiceUnavailable)
}

func setPlatformLogoHeaders(c *gin.Context, kind, mimeType, etag string) {
	c.Header("ETag", etag)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	c.Header("Content-Disposition", `inline; filename="`+kind+platformLogoExtension(mimeType)+`"`)
}

func (a *API) loadPlatformOEM(ctx context.Context) (store.PlatformOEMSetting, error) {
	var item store.PlatformOEMSetting
	db := a.db.WithContext(ctx)
	readMetadata := func() error {
		item = store.PlatformOEMSetting{}
		return db.Session(&gorm.Session{}).Select(platformOEMMetadataColumns).First(&item, "id = ?", store.PlatformOEMSettingID).Error
	}
	if err := readMetadata(); err == nil {
		return item, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return item, err
	}
	// The row is normally created by MigrateAndSeed. OnConflict also makes
	// direct router tests and rolling upgrades safe before their first read.
	version := strings.TrimSpace(a.cfg.Version)
	if version == "" {
		version = "dev"
	}
	seed := store.PlatformOEMSetting{Base: store.Base{ID: store.PlatformOEMSettingID}, SystemName: "Honeynet", SystemVersion: version, Copyright: "© Honeynet", Revision: 1}
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(&seed).Error; err != nil {
		return item, err
	}
	err := readMetadata()
	return item, err
}

const platformOEMMetadataColumns = "id, created_at, updated_at, system_name, system_version, copyright, customer_service_phone, customer_service_email, official_website_url, product_documentation_url, system_logo_mime, system_logo_sha256, system_logo_size, company_logo_mime, company_logo_sha256, company_logo_size, revision"

func platformOEMPublicView(item store.PlatformOEMSetting) platformOEMView {
	view := platformOEMView{
		SystemName: item.SystemName, SystemVersion: item.SystemVersion, Copyright: item.Copyright,
		CustomerServicePhone: item.CustomerServicePhone, CustomerServiceEmail: item.CustomerServiceEmail,
		OfficialWebsiteURL: item.OfficialWebsiteURL, ProductDocumentationURL: item.ProductDocumentationURL,
		Revision: item.Revision,
	}
	systemVersion, companyVersion := item.SystemLogoSHA256, item.CompanyLogoSHA256
	if systemVersion == "" {
		_, _, systemVersion = defaultPlatformLogo()
	}
	if companyVersion == "" {
		_, _, companyVersion = defaultPlatformLogo()
	}
	view.SystemLogoURL = "/api/v1/platform/assets/system?v=" + systemVersion
	view.CompanyLogoURL = "/api/v1/platform/assets/company?v=" + companyVersion
	return view
}

func platformLogoAuditView(item store.PlatformOEMSetting, kind string) map[string]any {
	if kind == "company" {
		return map[string]any{"kind": kind, "configured": item.CompanyLogoSHA256 != "", "mime": item.CompanyLogoMIME, "bytes": item.CompanyLogoSize, "sha256": item.CompanyLogoSHA256}
	}
	return map[string]any{"kind": kind, "configured": item.SystemLogoSHA256 != "", "mime": item.SystemLogoMIME, "bytes": item.SystemLogoSize, "sha256": item.SystemLogoSHA256}
}

func normalizePlatformOEMUpdate(request platformOEMUpdate) (platformOEMUpdate, error) {
	request.SystemName = strings.TrimSpace(request.SystemName)
	request.SystemVersion = strings.TrimSpace(request.SystemVersion)
	request.Copyright = strings.TrimSpace(request.Copyright)
	request.CustomerServicePhone = strings.TrimSpace(request.CustomerServicePhone)
	request.CustomerServiceEmail = strings.TrimSpace(request.CustomerServiceEmail)
	request.OfficialWebsiteURL = strings.TrimSpace(request.OfficialWebsiteURL)
	request.ProductDocumentationURL = strings.TrimSpace(request.ProductDocumentationURL)
	checks := []struct {
		label string
		value string
		max   int
		req   bool
	}{
		{"系统名称", request.SystemName, 128, true}, {"系统版本", request.SystemVersion, 64, true},
		{"版权信息", request.Copyright, 512, false}, {"客服热线", request.CustomerServicePhone, 64, false},
	}
	for _, check := range checks {
		if check.req && check.value == "" {
			return request, errors.New(check.label + "不能为空")
		}
		if utf8.RuneCountInString(check.value) > check.max || hasUnsafeTextControl(check.value) || strings.ContainsAny(check.value, "<>") {
			return request, errors.New(check.label + "格式错误或内容过长")
		}
	}
	if request.CustomerServiceEmail != "" {
		parsed, err := mail.ParseAddress(request.CustomerServiceEmail)
		if err != nil || parsed.Address != request.CustomerServiceEmail || utf8.RuneCountInString(request.CustomerServiceEmail) > 254 {
			return request, errors.New("客服邮箱格式错误")
		}
	}
	if err := validatePlatformURL("官网地址", request.OfficialWebsiteURL); err != nil {
		return request, err
	}
	if err := validatePlatformURL("产品文档地址", request.ProductDocumentationURL); err != nil {
		return request, err
	}
	return request, nil
}

func validatePlatformURL(label, value string) error {
	if value == "" {
		return nil
	}
	if len(value) > 1024 || hasUnsafeTextControl(value) {
		return errors.New(label + "格式错误或内容过长")
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil {
		return errors.New(label + "必须是有效的 HTTP 或 HTTPS 地址")
	}
	return nil
}

func hasUnsafeTextControl(value string) bool {
	return strings.IndexFunc(value, func(r rune) bool { return unicode.IsControl(r) || unicode.In(r, unicode.Cf) }) >= 0
}

func platformLogoKind(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "system":
		return "system", true
	case "company":
		return "company", true
	default:
		return "", false
	}
}

func normalizeOEMLogo(header *multipart.FileHeader) ([]byte, string, error) {
	if header == nil || header.Size < 1 || header.Size > maxOEMLogoUploadBytes {
		return nil, "", errors.New("图片大小必须在 1 B 到 2 MiB 之间")
	}
	file, err := header.Open()
	if err != nil {
		return nil, "", errors.New("无法读取图片")
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, maxOEMLogoUploadBytes+1))
	if err != nil || len(raw) == 0 || len(raw) > maxOEMLogoUploadBytes {
		return nil, "", errors.New("图片大小必须在 1 B 到 2 MiB 之间")
	}
	detected := http.DetectContentType(raw)
	if detected != "image/png" && detected != "image/jpeg" && detected != "image/webp" {
		return nil, "", errors.New("仅支持 PNG、JPEG 或 WebP 图片，不支持 SVG")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 4096 || config.Height > 4096 || int64(config.Width)*int64(config.Height) > maxOEMLogoPixels {
		return nil, "", errors.New("图片无法解码或像素尺寸超过限制")
	}
	decoded, decodedFormat, err := image.Decode(bytes.NewReader(raw))
	if err != nil || decodedFormat != format {
		return nil, "", errors.New("图片内容不完整或无法解码")
	}
	var output bytes.Buffer
	if format == "png" {
		err = png.Encode(&output, decoded)
		detected = "image/png"
	} else if format == "jpeg" {
		err = jpeg.Encode(&output, decoded, &jpeg.Options{Quality: 90})
		detected = "image/jpeg"
	} else if format == "webp" {
		err = png.Encode(&output, decoded)
		detected = "image/png"
	} else {
		return nil, "", errors.New("图片格式与文件内容不匹配")
	}
	if err != nil || output.Len() == 0 || output.Len() > 4<<20 {
		return nil, "", errors.New("图片标准化处理失败或内容过大")
	}
	return output.Bytes(), detected, nil
}

func platformLogoExtension(mimeType string) string {
	if mimeType == "image/jpeg" {
		return ".jpg"
	}
	return ".png"
}

var (
	defaultPlatformLogoOnce sync.Once
	defaultPlatformLogoData []byte
	defaultPlatformLogoHash string
)

// defaultPlatformLogo returns a code-owned raster fallback. User-controlled
// SVG is never accepted, and resetting either logo always leaves a usable URL.
func defaultPlatformLogo() ([]byte, string, string) {
	defaultPlatformLogoOnce.Do(func() {
		canvas := image.NewRGBA(image.Rect(0, 0, 96, 96))
		background := color.RGBA{R: 201, G: 24, B: 43, A: 255}
		foreground := color.RGBA{R: 255, G: 255, B: 255, A: 255}
		for y := 0; y < 96; y++ {
			for x := 0; x < 96; x++ {
				canvas.SetRGBA(x, y, background)
			}
		}
		for y := 22; y < 74; y++ {
			for x := 25; x < 36; x++ {
				canvas.SetRGBA(x, y, foreground)
			}
			for x := 60; x < 71; x++ {
				canvas.SetRGBA(x, y, foreground)
			}
		}
		for y := 43; y < 54; y++ {
			for x := 36; x < 60; x++ {
				canvas.SetRGBA(x, y, foreground)
			}
		}
		var encoded bytes.Buffer
		_ = png.Encode(&encoded, canvas)
		defaultPlatformLogoData = encoded.Bytes()
		digest := sha256.Sum256(defaultPlatformLogoData)
		defaultPlatformLogoHash = hex.EncodeToString(digest[:])
	})
	return defaultPlatformLogoData, "image/png", defaultPlatformLogoHash
}
