package rest

import (
  "bytes"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "time"

  "github.com/emicklei/go-restful/v3"
  "github.com/jinzhu/gorm/dialects/postgres"
  "github.com/jmoiron/sqlx"
  "github.com/tomasen/realip"
)

type LogFilter struct {
	db    *sqlx.DB
	clock timer
}

type timer interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

type realClock struct{}

func (rc *realClock) Now() time.Time {
	return time.Now()
}

func (rc *realClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

func NewRequestLoggingFilter(db *sqlx.DB) *LogFilter {
	return &LogFilter{
		db:    db,
		clock: &realClock{},
	}
}

func (lf *LogFilter) Filter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	startTime := lf.clock.Now().UTC()
	inBody, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		panic(err)
	}
	req.Request.Body = ioutil.NopCloser(bytes.NewReader(inBody))
	c := NewResponseCapture(resp.ResponseWriter)
	resp.ResponseWriter = c
	chain.ProcessFilter(req, resp)
	latency := lf.clock.Since(startTime)
	if err := lf.db.Ping(); err != nil {
		panic(err)
	}
	// ignore error, just log
	dbDriverName := lf.db.DriverName()
	switch dbDriverName {
	case "postgres", "postgresql", "pg":
		ipAddr := realip.FromRequest(req.Request)
		if ipAddr == "" {
			ipAddr = req.Request.RemoteAddr
		}
		bodyLog := LoggerBody{}
		var jsonbRequestBody postgres.Jsonb
		var jsonbResponseBody postgres.Jsonb
		if err := json.Unmarshal(inBody, &jsonbRequestBody); err == nil {
			bodyLog.Request = jsonbRequestBody
		}
		if err := json.Unmarshal(c.Bytes(), &jsonbResponseBody); err == nil {
			bodyLog.Response = jsonbResponseBody
		}
		if bHeader, err := json.Marshal(req.Request.Header); err == nil {
			var jsonbRequestHeader postgres.Jsonb
			_ = json.Unmarshal(bHeader, &jsonbRequestHeader)
			bodyLog.Header = jsonbRequestHeader
		}

		inBodyLog, err := json.Marshal(&bodyLog)
		if err != nil {
			panic(err)
		}
		var requestLogBodyRequest postgres.Jsonb
		_ = json.Unmarshal(inBodyLog, &requestLogBodyRequest)
		tNow := time.Now().UTC()
		_ = lf.db.MustExec(`INSERT INTO logs `+
			`(method, status_code, uri, referer, user_agent, ip_address, latency, body, created_at, updated_at) VALUES `+
			`($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)`,
			req.Request.Method, resp.StatusCode(), req.Request.RequestURI, req.Request.Referer(), ipAddr, req.Request.UserAgent(), latency.Nanoseconds(), requestLogBodyRequest, tNow)
	}
}

type LoggerBody struct {
	Header   postgres.Jsonb `json:"header,omitempty"`
	Request  postgres.Jsonb `json:"request,omitempty"`
	Response postgres.Jsonb `json:"response,omitempty"`
}

type HTTPReqInfo struct {
	Method    string
	Uri       string
	Referer   string
	IPAddr    string
	UserAgent string
	Duration  time.Duration
	Request   *http.Request
}

type ResponseCapture struct {
	http.ResponseWriter
	wroteHeader bool
	status      int
	body        *bytes.Buffer
}

func NewResponseCapture(w http.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter: w,
		wroteHeader:    false,
		body:           new(bytes.Buffer),
	}
}

func (c ResponseCapture) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c ResponseCapture) Write(data []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}
	c.body.Write(data)
	return c.ResponseWriter.Write(data)
}

func (c *ResponseCapture) WriteHeader(statusCode int) {
	c.status = statusCode
	c.wroteHeader = true
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c ResponseCapture) Bytes() []byte {
	return c.body.Bytes()
}

func (c ResponseCapture) StatusCode() int {
	return c.status
}
