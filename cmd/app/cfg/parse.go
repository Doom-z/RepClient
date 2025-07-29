package cfg

import (
	"log"
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

	if path == "" {
		f, createErr := os.Create(defaultConfPath)
		if createErr != nil {
			log.Fatalf("failed to create config file: %v", createErr)
		}
		defer f.Close()

		encErr := toml.NewEncoder(f).Encode(defaultConf)
		if encErr != nil {
			log.Fatalf("failed to write default config to file: %v", encErr)
		}
		log.Printf("using default config. written to: %s", defaultConfPath)

		path = defaultConfPath
	}

	_, err := toml.DecodeFile(path, &defaultConf)
	if err != nil {
		log.Fatalf("failed to load config from %s: %v", path, err)
	}

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
