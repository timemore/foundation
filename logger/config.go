package logger

type Config struct {
	// Enable console logging
	ConsoleLog bool `split_words:"true" default:"true"`

	// EncodeLogsAsJson makes the log framework log JSON
	EncodeLogsAsJson bool `split_words:"true"`
	// FileLoggingEnabled makes the framework log to a file
	// the fields below can be skipped if this value is false!
	FileLogging bool `split_words:"true"`
	// Directory to log to to when file logging is enabled
	Directory string `split_words:"true" default:"/var/log"`
	// Filename is the name of the logfile which will be placed inside the directory
	Filename string `split_words:"true" default:"service.log"`
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int `split_words:"true" default:"500"`
	// MaxBackups the max number of rolled files to keep
	MaxBackups int `split_words:"true" default:"3"`
	// MaxAge the max age in days to keep a logfile
	MaxAge int `split_words:"true" default:"30"`
}
