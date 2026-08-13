package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/honeynet/honeynet/internal/threatintel"
)

func (a *API) threatIntelStatus(c *gin.Context) {
	if a.threatIntel == nil {
		ok(c, threatintel.Status{Enabled: false, Source: "免费社区威胁情报库", IPv4IPv6: true})
		return
	}
	ok(c, a.threatIntel.Status())
}

func (a *API) queryThreatIntel(c *gin.Context) {
	address := strings.TrimSpace(c.Query("ip"))
	parsed, err := netip.ParseAddr(address)
	if err != nil {
		fail(c, http.StatusBadRequest, "INVALID_IP", "请输入有效的 IPv4 或 IPv6 地址")
		return
	}
	if parsed.Is4In6() {
		parsed = parsed.Unmap()
	}
	if a.threatIntel == nil || !a.threatIntel.Status().Loaded {
		fail(c, http.StatusServiceUnavailable, "INTELLIGENCE_DATABASE_UNAVAILABLE", "威胁情报数据库尚未就绪")
		return
	}
	result, found := a.threatIntel.Lookup(parsed.String())
	tags := []string{}
	if found {
		tags = threatintel.EventTags(result)
	}
	ok(c, gin.H{"ip": parsed.String(), "matched": found, "intelligence": result, "tags": tags})
}

func (a *API) updateThreatIntel(c *gin.Context) {
	if a.threatIntel == nil || !a.threatIntel.Status().Enabled {
		fail(c, http.StatusConflict, "INTELLIGENCE_DATABASE_DISABLED", "威胁情报数据库未启用")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 55*time.Second)
	defer cancel()
	if err := a.threatIntel.Update(ctx); err != nil {
		if errors.Is(err, threatintel.ErrUpdateInProgress) {
			fail(c, http.StatusConflict, "INTELLIGENCE_UPDATE_RUNNING", "威胁情报数据库正在更新")
			return
		}
		fail(c, http.StatusBadGateway, "INTELLIGENCE_UPDATE_FAILED", "威胁情报数据库更新失败，请查看状态详情")
		return
	}
	ok(c, a.threatIntel.Status())
}
