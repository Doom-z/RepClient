package cfg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/Doom-z/RepClient/pkg/logger"
)

func getAppDir() string {
	dir, err := os.Executable()
	if err != nil {
		logger.Fatal(err)
	}
	return filepath.Dir(dir)
}

// loadConfValid loads the configuration from the given path.
// If the path is empty, the defaultConfPath is used.
// If the path is relative, the app executable dir is prepended.
func LoadConfValid(configFileName string, defaultConf Conf, defaultConfPath string) Conf {
	if configFileName == "" {
		configFileName = defaultConfPath
	}
	path := findSuitablePath(configFileName)

	_, err := toml.DecodeFile(path, &defaultConf)
	if err != nil {
		logger.Info("failed to load config file: ", err, " using default config")
	}
	fmt.Println(path)
	logger.WithField("conf", &defaultConf).Debug("configuration loaded")
	return defaultConf
}

func findSuitablePath(configFileName string) string {
	// config.toml in current dir > config.toml in app dir > $HOME/.config/appName/config.toml
	confPath := []string{configFileName, filepath.Join(getAppDir(), configFileName), filepath.Join(os.Getenv("HOME"), ".config", Name, configFileName)}
	path := ""
	for _, p := range confPath {
		if _, err := os.Stat(p); err == nil {
			path = p
			break
		}
	}
	return path
}
