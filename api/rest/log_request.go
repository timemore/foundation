package rest

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/jmoiron/sqlx"
	"github.com/nyaruka/phonenumbers"
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
	inBody, err := io.ReadAll(req.Request.Body)
	if err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			panic(err)
		}
	}
	req.Request.Body = io.NopCloser(bytes.NewReader(inBody))
	c := NewResponseCapture(resp.ResponseWriter)
	resp.ResponseWriter = c
	chain.ProcessFilter(req, resp)
	latency := lf.clock.Since(startTime)
	if lf.db != nil {
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
			var mapPayload propertyMap
			if err := json.Unmarshal(inBody, &mapPayload); err == nil {
				lf.mapFilter(mapPayload)
				bodyLog.Request = mapPayload
			}

			var mapResponse propertyMap
			if err := json.Unmarshal(c.Bytes(), &mapResponse); err == nil {
				lf.mapFilter(mapResponse)
				bodyLog.Response = mapResponse
			}

			reqHeader := req.Request.Header
			if bHeader, err := json.Marshal(reqHeader); err == nil {
				var mapHeader propertyMap
				_ = json.Unmarshal(bHeader, &mapHeader)
				lf.mapFilter(mapHeader)
				bodyLog.Header = mapHeader
			}

			inBodyLog, err := json.Marshal(&bodyLog)
			if err != nil {
				panic(err)
			}
			var requestLogBodyRequest propertyMap
			_ = json.Unmarshal(inBodyLog, &requestLogBodyRequest)
			tNow := time.Now().UTC()
			_ = lf.db.MustExec(`INSERT INTO logs `+
				`(method, status_code, uri, referer, user_agent, ip_address, latency, body, created_at, updated_at) VALUES `+
				`($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)`,
				req.Request.Method, resp.StatusCode(), req.Request.RequestURI, req.Request.Referer(), ipAddr, req.Request.UserAgent(), latency.Nanoseconds(), requestLogBodyRequest, tNow)
		}
	}
}

type propertyMap map[string]interface{}

func (p propertyMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (p *propertyMap) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	var i interface{}
	err := json.Unmarshal(source, &i)
	if err != nil {
		return err
	}

	*p, ok = i.(map[string]interface{})
	if !ok {
		return errors.New("type assertion .(map[string]interface{}) failed")
	}

	return nil
}

type LoggerBody struct {
	Header   propertyMap `json:"header,omitempty"`
	Request  propertyMap `json:"request,omitempty"`
	Response propertyMap `json:"response,omitempty"`
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

const (
	Base64         string = "^(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})$"
	PrintableASCII string = "^[\x20-\x7E]+$"
	DataURI        string = "^data:.+\\/(.+);base64$"
	Email          string = "^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
	CreditCard     string = "^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\\d{3})\\d{11})$"
	PhoneNumber    string = "^(?:(?:\\(?(?:00|\\+)([1-4]\\d\\d|[1-9]\\d?)\\)?)?[\\-\\.\\ \\\\/]?)?((?:\\(?\\d{1,}\\)?[\\-\\.\\ \\\\/]?){0,})(?:[\\-\\.\\ \\\\/]?(?:#|ext\\.?|extension|x)[\\-\\.\\ \\\\/]?(\\d+))?$"
)

var (
	rxBase64      = regexp.MustCompile(Base64)
	rxDataURI     = regexp.MustCompile(DataURI)
	rxEmail       = regexp.MustCompile(Email)
	rxCreditCard  = regexp.MustCompile(CreditCard)
	rxPhoneNumber = regexp.MustCompile(PhoneNumber)
)

var SafeFields = []string{"authorization", "api-key", "api", "apikey", "merchant-key", "enterprise-token", "token", "user-token"}

func (lf *LogFilter) mapFilter(aMap map[string]interface{}) {
	for key, val := range aMap {
		if inArray(strings.ToLower(key), SafeFields) {
			delete(aMap, key)
		}
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			lf.mapFilter(val.(map[string]interface{}))

		case []interface{}:
			lf.parseArray(val.([]interface{}))

		default:
			// todo: sanitize data
			// https://gist.github.com/xigang/98504205d86803baedc5
			aMap[key] = concreteVal
			if str, ok := concreteVal.(string); ok {
				strCount := len(str)
				dataURI := strings.Split(str, ",")
				switch {
				case rxDataURI.MatchString(dataURI[0]):
					if rxBase64.MatchString(dataURI[1]) {
						aMap[key] = dataURI[0] + ",****"
					}
				case rxEmail.MatchString(str):
					emailAddr := strings.SplitN(str, "@", 2)
					addr := emailAddr[0]
					for _, c := range []string{"a", "i", "e", "u", "o", "0"} {
						addr = strings.ReplaceAll(strings.ToLower(addr), c, "*")
					}
					aMap[key] = addr + "@" + emailAddr[1]
				case rxCreditCard.MatchString(str):
					aMap[key] = str[:3] + strings.Repeat("*", len(str[3:]))
				case rxPhoneNumber.MatchString(str):
					phoneNumberStr := strings.TrimPrefix(str, "+")
					reLocalPhoneNumber := regexp.MustCompile("^(.*?)0(.*)$")
					repStr := "${1}62$2"
					phoneNumberStr = reLocalPhoneNumber.ReplaceAllString(phoneNumberStr, repStr)
					if _, err := phonenumbers.Parse(phoneNumberStr, ""); err == nil {
						aMap[key] = strings.Repeat("*", len(str[:strCount-4])) + str[strCount-4:]
					}
				case inArray(strings.ToLower(key), SafeFields):
					aMap[key] = "***"
				}
			}

		}
	}
}

func inArray(str string, haystacks []string) bool {
	for _, h := range haystacks {
		if h == str {
			return true
		}
	}
	return false
}

func (lf *LogFilter) parseArray(anArray []interface{}) {
	for _, val := range anArray {
		switch t := val.(type) {
		case map[string]interface{}:
			lf.mapFilter(val.(map[string]interface{}))
		case []interface{}:
			lf.parseArray(val.([]interface{}))
		default:
			println(t)
		}
	}
}
