package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	publicBase  = "http://127.0.0.1:8181"
	privateBase = "http://127.0.0.1:8182"
	tokenPath   = "/data/admin-token.json"
	database    = "r2db"
)

type tokenDocument struct {
	Token string `json:"token"`
}

type result struct {
	Mode          string         `json:"mode"`
	Checks        map[string]any `json:"checks"`
	CatalogSHA256 string         `json:"catalog_sha256,omitempty"`
	QuerySHA256   string         `json:"query_sha256,omitempty"`
	TokenSHA256   string         `json:"token_sha256"`
}

func main() {
	if len(os.Args) != 2 {
		fatal("usage: r2-client <matrix|write-a|verify-a|write-b|verify-rollback|leak-scan>")
	}
	raw, err := os.ReadFile(tokenPath)
	must(err)
	var doc tokenDocument
	must(json.Unmarshal(raw, &doc))
	if doc.Token == "" {
		fatal("empty token")
	}
	if os.Args[1] == "exec-equal-supervisor" {
		must(os.Setenv("INFLUXDB3_EXTERNAL_BEARER_TOKEN", doc.Token))
		must(syscall.Exec("/usr/bin/setpriv", []string{
			"setpriv", "--reuid=1500", "--regid=1500", "--init-groups",
			"/usr/local/bin/supervisor.sh",
		}, os.Environ()))
	}
	externalToken := os.Getenv("INFLUXDB3_EXTERNAL_BEARER_TOKEN")
	if len(externalToken) < 32 {
		fatal("external bearer token contract unavailable")
	}
	st, err := os.Stat(tokenPath)
	must(err)
	if st.Mode().Perm() != 0600 {
		fatal(fmt.Sprintf("token mode is %04o, expected 0600", st.Mode().Perm()))
	}
	out := result{
		Mode:        os.Args[1],
		Checks:      map[string]any{},
		TokenSHA256: digest([]byte(doc.Token)),
	}
	switch os.Args[1] {
	case "matrix":
		runMatrix(doc.Token, externalToken, &out)
	case "write-a":
		createDatabase(doc.Token, externalToken)
		write(externalToken, "r2metric,kind=baseline value=11i 1700000000000000000")
		verify(doc.Token, externalToken, true, false, &out)
	case "verify-a":
		verify(doc.Token, externalToken, true, false, &out)
	case "write-b":
		write(externalToken, "r2metric,kind=postchange value=22i 1700000001000000000")
		verify(doc.Token, externalToken, true, true, &out)
	case "verify-rollback":
		verify(doc.Token, externalToken, true, false, &out)
	case "leak-scan":
		runLeakScan(doc.Token, externalToken, &out)
	default:
		fatal("unknown mode")
	}
	encoded, err := json.Marshal(out)
	must(err)
	fmt.Println(string(encoded))
}

func runMatrix(internalToken, externalToken string, out *result) {
	matrix := map[string]int{
		"healthz_no_auth":       request("GET", publicBase+"/healthz", "", nil),
		"health_no_auth":        request("GET", publicBase+"/health", "", nil),
		"health_malformed_auth": request("GET", publicBase+"/health", "Basic invalid", nil),
		"health_wrong_bearer":   request("GET", publicBase+"/health", "Bearer wrong", nil),
		"health_internal_bearer": request("GET", publicBase+"/health", "Bearer "+internalToken, nil),
		"health_external_bearer": request("GET", publicBase+"/health", "Bearer "+externalToken, nil),
		"query_no_auth":         request("POST", publicBase+"/api/v3/query_sql", "", []byte(`{"db":"r2db","q":"SELECT 1","format":"json"}`)),
		"query_wrong_bearer":    request("POST", publicBase+"/api/v3/query_sql", "Bearer wrong", []byte(`{"db":"r2db","q":"SELECT 1","format":"json"}`)),
		"query_internal_bearer": request("POST", publicBase+"/api/v3/query_sql", "Bearer "+internalToken, []byte(`{"db":"r2db","q":"SELECT 1","format":"json"}`)),
	}
	expected := map[string]int{
		"healthz_no_auth": 200, "health_no_auth": 401, "health_malformed_auth": 401,
		"health_wrong_bearer": 401, "health_internal_bearer": 401, "health_external_bearer": 200,
		"query_no_auth": 401, "query_wrong_bearer": 401, "query_internal_bearer": 401,
	}
	for key, want := range expected {
		if matrix[key] != want {
			fatal(fmt.Sprintf("%s=%d want %d", key, matrix[key], want))
		}
	}
	out.Checks["http_negative_matrix"] = matrix

	pubStatus, pubHeaders, pubBody := fetch("GET", publicBase+"/health", "Bearer "+externalToken, nil)
	privStatus, privHeaders, privBody := fetch("GET", privateBase+"/health", "Bearer "+internalToken, nil)
	if pubStatus != privStatus || !bytes.Equal(pubBody, privBody) {
		fatal("adapter health status/body parity failed")
	}
	out.Checks["health_status_body_parity"] = true
	out.Checks["health_content_type_parity"] = pubHeaders.Get("Content-Type") == privHeaders.Get("Content-Type")
}

func createDatabase(_ string, externalToken string) {
	status := request("POST", publicBase+"/api/v3/configure/database", "Bearer "+externalToken, []byte(`{"db":"r2db"}`))
	if status != 200 && status != 201 && status != 204 && status != 409 {
		fatal(fmt.Sprintf("create database status=%d", status))
	}
}

func write(externalToken, line string) {
	status := request("POST", publicBase+"/api/v3/write_lp?db="+database+"&precision=nanosecond", "Bearer "+externalToken, []byte(line))
	if status != 200 && status != 204 {
		fatal(fmt.Sprintf("write status=%d", status))
	}
}

func verify(internalToken, externalToken string, expectA, expectB bool, out *result) {
	body := query(internalToken, externalToken, "SELECT kind, value FROM r2metric ORDER BY time")
	text := string(body)
	hasA := strings.Contains(text, "baseline") && strings.Contains(text, "11")
	hasB := strings.Contains(text, "postchange") && strings.Contains(text, "22")
	if hasA != expectA || hasB != expectB {
		fatal(fmt.Sprintf("sentinel mismatch has_a=%v has_b=%v", hasA, hasB))
	}
	out.Checks["baseline_sentinel_present"] = hasA
	out.Checks["postchange_sentinel_present"] = hasB
	out.QuerySHA256 = digest(canonicalJSON(body))

	status, _, catalog := fetch("GET", publicBase+"/api/v3/configure/database?format=json", "Bearer "+externalToken, nil)
	if status != 200 {
		fatal(fmt.Sprintf("catalog status=%d", status))
	}
	if !strings.Contains(string(catalog), database) {
		fatal("database missing from catalog")
	}
	out.CatalogSHA256 = digest(canonicalJSON(catalog))
}

func query(internalToken, externalToken, sql string) []byte {
	payload, err := json.Marshal(map[string]any{"db": database, "q": sql, "format": "json"})
	must(err)
	pubStatus, pubHeaders, pubBody := fetch("POST", publicBase+"/api/v3/query_sql", "Bearer "+externalToken, payload)
	privStatus, privHeaders, privBody := fetch("POST", privateBase+"/api/v3/query_sql", "Bearer "+internalToken, payload)
	if pubStatus != 200 || privStatus != 200 {
		fatal(fmt.Sprintf("query status public=%d private=%d", pubStatus, privStatus))
	}
	if !bytes.Equal(pubBody, privBody) {
		fatal("adapter query body parity failed")
	}
	if pubHeaders.Get("Content-Type") != privHeaders.Get("Content-Type") {
		fatal("adapter query content-type parity failed")
	}
	return pubBody
}

func runLeakScan(token, externalToken string, out *result) {
	roots := []string{"/data", "/evidence", "/etc", "/home", "/tmp", "/usr/local"}
	filesScanned := 0
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() || path == tokenPath {
				return nil
			}
			info, err := entry.Info()
			if err != nil || !info.Mode().IsRegular() || info.Size() > 32<<20 {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			filesScanned++
			if bytes.Contains(data, []byte(token)) {
				fatal("protected token found outside allowed file: " + path)
			}
			if bytes.Contains(data, []byte(externalToken)) {
				fatal("external bearer token found in regular file: " + path)
			}
			return nil
		})
	}
	processFiles := 0
	procEntries, _ := os.ReadDir("/proc")
	for _, entry := range procEntries {
		if !entry.IsDir() {
			continue
		}
		for _, leaf := range []string{"cmdline", "environ"} {
			data, err := os.ReadFile(filepath.Join("/proc", entry.Name(), leaf))
			if err != nil {
				continue
			}
			processFiles++
			if bytes.Contains(data, []byte(token)) {
				fatal("protected token found in process " + leaf)
			}
		}
	}
	out.Checks["allowed_token_file"] = tokenPath
	out.Checks["negative_regular_files_scanned"] = filesScanned
	out.Checks["negative_process_surfaces_scanned"] = processFiles
	out.Checks["token_hits_outside_allowed_file"] = 0
	out.Checks["external_token_regular_file_hits"] = 0
}

func request(method, url, auth string, body []byte) int {
	status, _, _ := fetch(method, url, auth, body)
	return status
}

func fetch(method, url, auth string, body []byte) (int, http.Header, []byte) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	must(err)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	must(err)
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	must(err)
	return resp.StatusCode, resp.Header.Clone(), data
}

func canonicalJSON(raw []byte) []byte {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return raw
	}
	normalized, err := json.Marshal(value)
	must(err)
	return normalized
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func must(err error) {
	if err != nil {
		fatal(err.Error())
	}
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "fatal:", message)
	os.Exit(1)
}
