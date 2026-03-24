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
	OptimizeBackup       string
	OrchestrationUrl     string
	SnapshotOptimizeRepo string
	SnapshotWebappRepo   string
	SnapshotZeebeRepo    string
	SnapshotRepository   string
	Serverport           string
	KeepBackUpItems      string
	ElasticUsername      string
	ElasticPassword      string
}

func LoadConfig() (*BackupHelperConfig, error) {
	config := &BackupHelperConfig{
		SecondarystorageUrl:  getEnvOrDefault("SECONDARYSTORAGE_URL", "http://localhost:9200"),
		OptimizeUrl:          getEnvOrDefault("OPTIMIZE_URL", "http://localhost:8092"),
		OrchestrationUrl:     getEnvOrDefault("ORCHESTRATION_URL", "http://localhost:9600/core"),
		SnapshotOptimizeRepo: getEnvOrDefault("SNAPSHOT_OPTIMIZE_REPO", "optimize-backup"),
		OptimizeBackup:       getEnvOrDefault("OPTIMIZE_BACKUP_ENABLED", "false"),
		SnapshotWebappRepo:   getEnvOrDefault("SNAPSHOT_CAMUNDA_REPO", "webapp-backup"),
		SnapshotZeebeRepo:    getEnvOrDefault("SNAPSHOT_ZEEBE_REPO", "camunda-backup"),
		SnapshotRepository:   getEnvOrDefault("SNAPSHOT_REPOSITORY", "azure"),
		Serverport:           getEnvOrDefault("SERVER_PORT", "8080"),
		KeepBackUpItems:      getEnvOrDefault("KEEP_BACKUP_ITEMS", "3"),
		ElasticUsername:      getEnvOrDefault("ELASTIC_USERNAME", "elastic"),
		ElasticPassword:      getEnvOrDefault("ELASTIC_PASSWORD", "changeme"),
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
