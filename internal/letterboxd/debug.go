package letterboxd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"sort"
	"strings"
)

const maxDebugBody = 8192

type debugError struct {
	err   error
	stack string
}

func (e debugError) Error() string {
	if e.stack == "" {
		return e.err.Error()
	}
	return e.err.Error() + "\n" + e.stack
}

func (e debugError) Unwrap() error {
	return e.err
}

func (c *Client) wrapDebug(err error) error {
	if err == nil || c == nil || !c.Debug {
		return err
	}
	var existing debugError
	if errors.As(err, &existing) {
		return err
	}
	return debugError{err: err, stack: string(debug.Stack())}
}

func (c *Client) httpStatusError(req *http.Request, resp *http.Response) error {
	return c.httpStatusErrorWithBody(req, resp, readBodySnippet(resp.Body))
}

func (c *Client) httpStatusErrorWithBody(req *http.Request, resp *http.Response, body string) error {
	if resp == nil {
		return c.wrapDebug(errors.New("unexpected nil response"))
	}
	if !c.Debug {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, req.URL.String())
	}
	lines := []string{fmt.Sprintf("unexpected status %d for %s", resp.StatusCode, req.URL.String())}
	if body != "" {
		lines = append(lines, "response body: "+body)
	}
	if headerLine := formatHeaders(resp.Header); headerLine != "" {
		lines = append(lines, "response headers: "+headerLine)
	}
	if headerLine := formatHeaders(req.Header); headerLine != "" {
		lines = append(lines, "request headers: "+headerLine)
	}
	return c.wrapDebug(errors.New(strings.Join(lines, "\n")))
}

func readBodySnippet(r io.Reader) string {
	data, _ := io.ReadAll(io.LimitReader(r, maxDebugBody+1))
	if len(data) == 0 {
		return ""
	}
	truncated := len(data) > maxDebugBody
	if truncated {
		data = data[:maxDebugBody]
	}
	body := strings.TrimSpace(string(data))
	if body == "" {
		return ""
	}
	if truncated {
		body += "...(truncated)"
	}
	return body
}

func formatHeaders(headers http.Header) string {
	if len(headers) == 0 {
		return ""
	}
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, key := range keys {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(formatHeaderValue(key, headers.Values(key)))
	}
	return b.String()
}

func formatHeaderValue(key string, values []string) string {
	switch strings.ToLower(key) {
	case "cookie", "set-cookie", "authorization":
		return "<redacted>"
	}
	return strings.Join(values, ",")
}
