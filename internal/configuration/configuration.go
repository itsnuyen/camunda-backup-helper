package configuration

import "os"

var Config *BackupHelperConfig

func init() {
	_, err := LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
}

type BackupHelperConfig struct {
	// Placeholder for configuration fields
	SecondarystorageUrl  string
	OptimizeUrl          string
	OrchestrationUrl     string
	SnapshotOptimizeRepo string
	SnapshotCamundaRepo  string
	SnapshotZeebeRepo    string
	SnapshotRepository   string
	Serverport           string
	KeepBackUpItems      string
}

func LoadConfig() (*BackupHelperConfig, error) {
	config := &BackupHelperConfig{
		SecondarystorageUrl:  getEnvOrDefault("SECONDARYSTORAGE_URL", "http://localhost:9200"),
		OptimizeUrl:          getEnvOrDefault("OPTIMIZE_URL", "http://localhost:8092"),
		OrchestrationUrl:     getEnvOrDefault("ORCHESTRATION_URL", "http://localhost:9600/core"),
		SnapshotOptimizeRepo: getEnvOrDefault("SNAPSHOT_OPTIMIZE_REPO", "optimize-backup"),
		SnapshotCamundaRepo:  getEnvOrDefault("SNAPSHOT_CAMUNDA_REPO", "camunda-backup"),
		SnapshotZeebeRepo:    getEnvOrDefault("SNAPSHOT_ZEEBE_REPO", "zeebe-backup"),
		SnapshotRepository:   getEnvOrDefault("SNAPSHOT_REPOSITORY", "azure"),
		Serverport:           getEnvOrDefault("SERVER_PORT", "8080"),
		KeepBackUpItems:      getEnvOrDefault("KEEP_BACKUP_ITEMS", "1"),
	}
	Config = config
	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
