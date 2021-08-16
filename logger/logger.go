package logger

import (
	"io"
	"net/http"
	"os"
	"path"
	"time"

	zlogsentry "github.com/archdx/zerolog-sentry"
	"github.com/rez-go/stev"
	"github.com/rs/zerolog"
	"github.com/tomasen/realip"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	Logger = zerolog.Logger
)

type PkgLogger struct {
	Logger
}

const EnvPrefixDefault = "LOG_"

func newRollingFile(config Config) io.Writer {
	filena := config.Filename
	if filena == "" {
		filena = "service.log"
	}
	if err := os.MkdirAll(config.Directory, 0744); err != nil {
		return nil
	}
	logFilename := path.Join(config.Directory, filena)

	return &lumberjack.Logger{
		Filename:   logFilename,
		MaxBackups: config.MaxBackups, // files
		MaxSize:    config.MaxSize,    // megabytes
		MaxAge:     config.MaxAge,     // days
	}
}

func newLogger() Logger {
	if logPretty := os.Getenv(EnvPrefixDefault + "PRETTY"); logPretty == "true" {
		return zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	return zerolog.New(os.Stderr)
}

func NewPkgLogger() PkgLogger {
	logger := newLogger()
	logLevelStr := os.Getenv(EnvPrefixDefault + "LEVEL")
	if logLevelStr != "" {
		logLevel, err := zerolog.ParseLevel(logLevelStr)
		if err == nil {
			logger = logger.Level(logLevel)
		}
	}
	cfg := configSkeletonPtr()
	err := stev.LoadEnv(EnvPrefixDefault, &cfg)
	if err == nil {
		writers := []io.Writer{
			zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
			newRollingFile(cfg),
		}
		zlogLevels := []zerolog.Level{
			zerolog.WarnLevel,
			zerolog.FatalLevel,
			zerolog.DebugLevel,
			zerolog.PanicLevel,
		}
		w, err := zlogsentry.New("https://0b1da1c9d24049e2ae313a216e578263@o415299.ingest.sentry.io/5908432", zlogsentry.WithLevels(zlogLevels...))
		if err != nil {
			panic(err)
		}

		defer w.Close()
		mw := zerolog.MultiLevelWriter(writers...)
		logger = logger.Output(mw)
	}
	logCtx := logger.With().Timestamp().CallerWithSkipFrameCount(2)
	return PkgLogger{logCtx.Logger()}
}

func (logger PkgLogger) WithRequest(req *http.Request) *Logger {
	if req == nil {
		return &logger.Logger
	}

	var urlStr string
	if req.URL != nil {
		urlStr = req.URL.String()
	}
	remoteAddr := realip.FromRequest(req)
	if remoteAddr == "" {
		remoteAddr = req.RemoteAddr
	}
	l := logger.With().
		Str("method", req.Method).
		Str("url", urlStr).
		Str("remote_ip", remoteAddr).
		Str("user_agent", req.UserAgent()).
		Logger()

	return &l
}
