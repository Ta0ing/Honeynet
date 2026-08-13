package pots

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func webRequestValues(r *http.Request, raw []byte) map[string][]string {
	values := make(map[string][]string)
	for key, entries := range r.URL.Query() {
		values[key] = append([]string(nil), entries...)
	}
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		if parsed, err := url.ParseQuery(string(raw)); err == nil {
			for key, entries := range parsed {
				values[key] = append(values[key], entries...)
			}
		}
	}
	if strings.Contains(contentType, "application/json") && len(raw) > 0 {
		var document any
		if json.Unmarshal(raw, &document) == nil {
			collectWebJSONValues(values, document)
		}
	}
	return values
}

func collectWebJSONValues(values map[string][]string, value any) {
	switch item := value.(type) {
	case map[string]any:
		for key, child := range item {
			switch scalar := child.(type) {
			case string:
				values[key] = append(values[key], scalar)
			case float64, bool:
				values[key] = append(values[key], fmt.Sprint(scalar))
			default:
				collectWebJSONValues(values, child)
			}
		}
	case []any:
		for _, child := range item {
			collectWebJSONValues(values, child)
		}
	}
}

func webCredentials(r *http.Request, values map[string][]string) (string, string, string) {
	if username, password, ok := r.BasicAuth(); ok {
		return username, password, "basic"
	}
	password := first(values, "password", "pass", "passwd", "pwd", "token", "SecretID", "passWord", "userPassword", "userpassword", "login_password", "secretKey", "admin_password", "svpn_password", "j_password", "txtPassword", "pma_password", "user[password]", "user_password", "PASSWORD", "pw", "UNAME_PASSWORD")
	username := first(values, "username", "user", "login", "email", "account", "userid", "user_id", "userName", "loginName", "loginid", "login_username", "user_name", "principal", "accessKey", "admin_name", "svpn_name", "j_username", "txtUserName", "pma_username", "user[login]", "user_login", "UNAME", "un")
	if username == "" && password != "" {
		username = first(values, "name")
	}
	return username, password, "form"
}
