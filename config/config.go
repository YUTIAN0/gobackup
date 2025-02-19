package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/gobackup/gobackup/logger"
)

var (
	// Exist Is config file exist
	Exist bool
	// Models configs
	Models []ModelConfig
	// gobackup base dir
	GoBackupDir string = getGoBackupDir()
)

type ScheduleConfig struct {
	Enabled bool
	// Cron expression
	Cron string
	// Every
	Every string
	// At time
	At string
}

func (sc ScheduleConfig) String() string {
	if sc.Enabled {
		if len(sc.Cron) > 0 {
			return fmt.Sprintf("cron %s", sc.Cron)
		} else {
			if len(sc.At) > 0 {
				return fmt.Sprintf("every %s at %s", sc.Every, sc.At)
			} else {
				return fmt.Sprintf("every %s", sc.Every)
			}
		}
	}

	return "disabled"
}

// ModelConfig for special case
type ModelConfig struct {
	Name         string
	TempPath     string
	DumpPath     string
	Schedule     ScheduleConfig
	CompressWith SubConfig
	EncryptWith  SubConfig
	Archive      *viper.Viper
	Environment  map[string]SubConfig
	Splitter     *viper.Viper
	Databases    map[string]SubConfig
	Storages     map[string]SubConfig
	Notifiers    map[string]SubConfig
	Viper        *viper.Viper
}

func getGoBackupDir() string {
	dir := os.Getenv("GOBACKUP_DIR")
	if len(dir) == 0 {
		dir = filepath.Join(os.Getenv("HOME"), ".gobackup")
	}
	return dir
}

// SubConfig sub config info
type SubConfig struct {
	Name  string
	Type  string
	Viper *viper.Viper
}

// loadConfig from:
// - ./gobackup.yml
// - ~/.gobackup/gobackup.yml
// - /etc/gobackup/gobackup.yml
func Init(configFile string) {
	logger := logger.Tag("Config")

	viper.SetConfigType("yaml")

	// set config file directly
	if len(configFile) > 0 {
		logger.Info("Load config:", configFile)
		viper.SetConfigFile(configFile)
	} else {
		logger.Info("Load config from default path.")
		viper.SetConfigName("gobackup")

		// ./gobackup.yml
		viper.AddConfigPath(".")
		// ~/.gobackup/gobackup.yml
		viper.AddConfigPath("$HOME/.gobackup") // call multiple times to add many search paths
		// /etc/gobackup/gobackup.yml
		viper.AddConfigPath("/etc/gobackup/") // path to look for the config file in
	}

	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("Load gobackup config failed: ", err)
		return
	}

	viperConfigFile := viper.ConfigFileUsed()
	if info, _ := os.Stat(viperConfigFile); info.Mode()&(1<<2) != 0 {
		// max permission: 0770
		logger.Warnf("Other users are able to access %s with mode %v", viperConfigFile, info.Mode())
	}

	viper.Set("useTempWorkDir", false)
	if workdir := viper.GetString("workdir"); len(workdir) == 0 {
		// use temp dir as workdir
		dir, err := os.MkdirTemp("", "gobackup")
		if err != nil {
			logger.Fatal(err)
		}
		viper.Set("workdir", dir)
		viper.Set("useTempWorkDir", true)
	}

	Exist = true
	Models = []ModelConfig{}
	for key := range viper.GetStringMap("models") {
		Models = append(Models, loadModel(key))
	}

	if len(Models) == 0 {
		logger.Fatalf("No model found in %s", viperConfigFile)
	}
}

func loadModel(key string) (model ModelConfig) {
	model.Name = key
	model.TempPath = filepath.Join(viper.GetString("workdir"), fmt.Sprintf("%d", time.Now().UnixNano()))
	model.DumpPath = filepath.Join(model.TempPath, key)
	model.Viper = viper.Sub("models." + key)

	model.Schedule = ScheduleConfig{Enabled: false}

	model.CompressWith = SubConfig{
		Type:  model.Viper.GetString("compress_with.type"),
		Viper: model.Viper.Sub("compress_with"),
	}

	model.EncryptWith = SubConfig{
		Type:  model.Viper.GetString("encrypt_with.type"),
		Viper: model.Viper.Sub("encrypt_with"),
	}

	model.Archive = model.Viper.Sub("archive")
	model.Splitter = model.Viper.Sub("split_with")

	loadScheduleConfig(&model)
	loadDatabasesConfig(&model)
	loadStoragesConfig(&model)
	//logger.Info("loadEnvironmentConfig_pre")

	loadEnvironmentConfig(&model)

	if len(model.Storages) == 0 {
		logger.Fatalf("No storage found in model %s", model.Name)
	}

	loadNotifiersConfig(&model)

	return
}

func loadScheduleConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("schedule")
	model.Schedule = ScheduleConfig{Enabled: false}
	if subViper == nil {
		return
	}

	model.Schedule = ScheduleConfig{
		Enabled: true,
		Cron:    subViper.GetString("cron"),
		Every:   subViper.GetString("every"),
		At:      subViper.GetString("at"),
	}
}

func loadEnvironmentConfig(model *ModelConfig) {
	//	logger.Info("loadEnvironmentConfig")

	subViper := model.Viper.Sub("environment")
	//	logger.Info("loadEnvironmentConfi11:")
	model.Environment = map[string]SubConfig{}

	for key := range model.Viper.GetStringMap("environment") {
		//logger.Info("key:" + key)
		//envViper := subViper.Sub(key)
		//var env string = subViper.GetString(key).
		//logger.Info(subViper.Get(key))
		os.Setenv(strings.ToTitle(key), subViper.GetString(key)+":"+os.Getenv(strings.ToTitle(key)))
		//logger.Info(os.Getenv(strings.ToTitle(key)))

		//	logger.Info(os.Environ())
		//modddel.Environment[key] = SubConfig{
		//	Name: key,
		//Type:  envViper.GetString("type"),
		//	Viper: envViper,
		//}
		//logger.Info(envViper)

	}

}

func loadDatabasesConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("databases")
	model.Databases = map[string]SubConfig{}
	for key := range model.Viper.GetStringMap("databases") {
		dbViper := subViper.Sub(key)
		model.Databases[key] = SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		}
	}
}

func loadStoragesConfig(model *ModelConfig) {
	storageConfigs := map[string]SubConfig{}
	// Backward compatible with `store_with` config
	storeWith := model.Viper.Sub("store_with")
	if storeWith != nil {
		logger.Warn(`[Deprecated] "store_with" is deprecated now, please use "storages" which supports multiple storages. Cycler config which usually located in "~/.gobackup/cycler" and named "MODEL.json" should be renamed to "MODEL_STORAGENAME.json", or cycler will start from scratch.`)
		storageConfigs["store_with"] = SubConfig{
			Name:  "",
			Type:  model.Viper.GetString("store_with.type"),
			Viper: model.Viper.Sub("store_with"),
		}
	}

	subViper := model.Viper.Sub("storages")
	for key := range model.Viper.GetStringMap("storages") {
		storageViper := subViper.Sub(key)
		//logger.Info(key)
		//logger.Info(storageViper)
		storageConfigs[key] = SubConfig{
			Name:  key,
			Type:  storageViper.GetString("type"),
			Viper: storageViper,
		}
	}
	model.Storages = storageConfigs
}

func loadNotifiersConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("notifiers")
	model.Notifiers = map[string]SubConfig{}
	for key := range model.Viper.GetStringMap("notifiers") {
		dbViper := subViper.Sub(key)
		model.Notifiers[key] = SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		}
	}
}

// GetModelConfigByName get model config by name
func GetModelConfigByName(name string) (model *ModelConfig) {
	for _, m := range Models {
		if m.Name == name {
			model = &m
			return
		}
	}
	return
}

// GetDatabaseByName get database config by name
func (model *ModelConfig) GetDatabaseByName(name string) (subConfig *SubConfig) {
	for _, m := range model.Databases {
		if m.Name == name {
			subConfig = &m
			return
		}
	}
	return
}
