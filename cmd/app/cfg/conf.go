package cfg

import "github.com/sirupsen/logrus"

const Name = "repclient"

type Conf struct {
	App App `toml:"app"`
	Api Api `toml:"api"`
	Log Log `toml:"log"`
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
		Log: Log{
			Level:  "warn",
			Stdout: []Stdout{{Format: LogFormatJSON, Output: LogOutputStdout}},
		},
	}
}

// levelToLogrusLevel converts a string to a logrus.Level
func LevelToLogrusLevel(level string) logrus.Level {
	switch level {
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
