package geoip

import (
	"errors"
	"os"
	"testing"

	ipdb "github.com/ipipdotnet/ipdb-go"
)

type fakeCityDatabase struct {
	info  *ipdb.CityInfo
	err   error
	calls int
}

func (f *fakeCityDatabase) FindInfo(string, string) (*ipdb.CityInfo, error) {
	f.calls++
	return f.info, f.err
}

func TestLookupFormatsChineseCityAndASN(t *testing.T) {
	db := &fakeCityDatabase{info: &ipdb.CityInfo{CountryName: "中国", RegionName: "山东", CityName: "济南", DistrictName: "济南", ASN: "4837", IspDomain: "China Unicom"}}
	resolver := &Resolver{db: db, language: "CN"}
	got, err := resolver.Lookup("202.194.98.103")
	if err != nil {
		t.Fatal(err)
	}
	if got.Geo != "中国 / 山东 / 济南" || got.ASN != "AS4837" || got.ISP != "China Unicom" {
		t.Fatalf("Lookup() = %#v", got)
	}
	if db.calls != 1 {
		t.Fatalf("database calls = %d, want 1", db.calls)
	}
}

func TestLookupClassifiesInternalAddressWithoutDatabase(t *testing.T) {
	db := &fakeCityDatabase{err: errors.New("must not be called")}
	resolver := &Resolver{db: db, language: "CN"}
	for _, address := range []string{"127.0.0.1", "10.1.2.3", "192.168.1.1", "fc00::1"} {
		got, err := resolver.Lookup(address)
		if err != nil || got.Geo != "内网地址" || !got.Internal {
			t.Fatalf("Lookup(%q) = %#v, %v", address, got, err)
		}
	}
	if db.calls != 0 {
		t.Fatalf("database calls = %d, want 0", db.calls)
	}
}

func TestLookupRejectsInvalidAddress(t *testing.T) {
	resolver := &Resolver{db: &fakeCityDatabase{}, language: "CN"}
	if _, err := resolver.Lookup("not-an-ip"); !errors.Is(err, ErrInvalidAddress) {
		t.Fatalf("Lookup() error = %v", err)
	}
}

func TestProvidedIPIPDatabase(t *testing.T) {
	path := os.Getenv("HONEYPOT_IPIP_TEST_DB")
	if path == "" {
		t.Skip("HONEYPOT_IPIP_TEST_DB is not set")
	}
	resolver, err := Open(path, "CN")
	if err != nil {
		t.Fatal(err)
	}
	got, err := resolver.Lookup("202.194.98.103")
	if err != nil {
		t.Fatal(err)
	}
	if got.Country == "" || got.Region == "" || got.City == "" {
		t.Fatalf("database did not return city-level data: %#v", got)
	}
	t.Logf("202.194.98.103 => %#v", got)
}
