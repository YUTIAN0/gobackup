package model

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/huacnlee/gobackup/archive"
	"github.com/huacnlee/gobackup/compressor"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/database"
	"github.com/huacnlee/gobackup/encryptor"
	"github.com/huacnlee/gobackup/logger"
	"github.com/huacnlee/gobackup/splitter"
	"github.com/huacnlee/gobackup/storage"
)

// Model class
type Model struct {
	Config config.ModelConfig
}

// Perform model
func (m Model) Perform() {
	logger := logger.Tag(fmt.Sprintf("Model: %s", m.Config.Name))

	logger.Info("WorkDir:", m.Config.DumpPath)

	defer func() {
		if r := recover(); r != nil {
			m.cleanup()
		}

		m.cleanup()
	}()

	err := database.Run(m.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	if m.Config.Archive != nil {
		err = archive.Run(m.Config)
		if err != nil {
			logger.Error(err)
			return
		}
	}

	archivePath, err := compressor.Run(m.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	archivePath, err = encryptor.Run(archivePath, m.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	archivePath, err = splitter.Run(archivePath, m.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	err = storage.Run(m.Config, archivePath)
	if err != nil {
		logger.Error(err)
		return
	}

}

// Cleanup model temp files
func (m Model) cleanup() {
	logger := logger.Tag("Model")

	tempDir := m.Config.TempPath
	if viper.GetBool("useTempWorkDir") {
		tempDir = viper.GetString("workdir")
	}
	logger.Infof("Cleanup temp: %s/", tempDir)
	if err := os.RemoveAll(tempDir); err != nil {
		logger.Errorf("Cleanup temp dir %s error: %v", tempDir, err)
	}
}

// GetModelByName get model by name
func GetModelByName(name string) *Model {
	modelConfig := config.GetModelConfigByName(name)
	if modelConfig == nil {
		return nil
	}
	return &Model{
		Config: *modelConfig,
	}
}

// GetModels get models
func GetModels() (models []*Model) {
	for _, modelConfig := range config.Models {
		m := Model{
			Config: modelConfig,
		}
		models = append(models, &m)
	}
	return
}
