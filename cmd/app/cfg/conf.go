package cfg

import "github.com/sirupsen/logrus"

const Name = "repclient"

type Conf struct {
	App    App    `toml:"app"`
	Api    Api    `toml:"api"`
	Output Output `toml:"output"`
	Log    Log    `toml:"log"`
}

type Output struct {
	Format string `toml:"format"`
	Dir    string `toml:"dir"`
}

type App struct {
	Name string `toml:"name"`
}

type File struct {
	Format  LogFormat `toml:"format"`
	Path    string    `toml:"path"`
	MaxSize int       `toml:"max_size"`
	MaxAge  int       `toml:"max_age"`
}

type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

type LogOutput string

const (
	LogOutputStdout LogOutput = "stdout"
	LogOutputStderr LogOutput = "stderr"
	LogOutputFile   LogOutput = "file"
)

type Stdout struct {
	Format LogFormat `toml:"format"`
	Output LogOutput `toml:"output"`
}

type Api struct {
	Host   string `toml:"host"`
	Apikey string `toml:"api_key"`
}

type Log struct {
	Level  string   `toml:"level"`
	File   []File   `toml:"file"`
	Stdout []Stdout `toml:"stdout"`
}

func GetDefaultConf() Conf {
	return Conf{
		Api: Api{
			Host:   "https://repproject.world",
			Apikey: "@repproject",
		},
		Output: Output{
			Format: "ndjson",
			Dir:    "output",
		},
		Log: Log{
			Level:  "info",
			Stdout: []Stdout{{Format: LogFormatText, Output: LogOutputStdout}},
		},
	}
}

// levelToLogrusLevel converts a string to a logrus.Level
func LevelToLogrusLevel(level string) logrus.Level {
	switch level {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}
