package httpapi

import (
	"errors"
	"testing"

	"github.com/honeynet/honeynet/internal/geoip"
)

type stubGeoResolver struct {
	result geoip.Result
	err    error
}

func (s stubGeoResolver) Lookup(string) (geoip.Result, error) { return s.result, s.err }

func TestEventLocationPrefersServerIPIPResult(t *testing.T) {
	api := &API{geoIP: stubGeoResolver{result: geoip.Result{Geo: "中国 / 山东 / 济南", ASN: "AS4837"}}}
	location, asn := api.eventLocation("202.194.98.103", "节点提供的位置", "AS1")
	if location != "中国 / 山东 / 济南" || asn != "AS4837" {
		t.Fatalf("eventLocation() = %q, %q", location, asn)
	}
}

func TestEventLocationFallsBackWhenLookupFails(t *testing.T) {
	api := &API{geoIP: stubGeoResolver{err: errors.New("lookup failed")}}
	location, asn := api.eventLocation("203.0.113.1", "原位置", "AS64500")
	if location != "原位置" || asn != "AS64500" {
		t.Fatalf("eventLocation() = %q, %q", location, asn)
	}
}

func TestEventLocationWithoutResolverUsesAgentFallback(t *testing.T) {
	api := &API{}
	location, asn := api.eventLocation("203.0.113.1", "节点位置", "AS64500")
	if location != "节点位置" || asn != "AS64500" {
		t.Fatalf("eventLocation() = %q, %q", location, asn)
	}
}

func TestEventLocationClearsASNForInternalAddress(t *testing.T) {
	api := &API{geoIP: stubGeoResolver{result: geoip.Result{Geo: "内网地址", Internal: true}}}
	location, asn := api.eventLocation("127.0.0.1", "伪造位置", "AS64500")
	if location != "内网地址" || asn != "" {
		t.Fatalf("eventLocation() = %q, %q", location, asn)
	}
}
