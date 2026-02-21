package orchestration

import (
	"log"

	"github.com/itsnuyen/camunda-backup-helper/internal/configuration"
	"github.com/itsnuyen/camunda-backup-helper/internal/domain"
)

var config *configuration.BackupHelperConfig

func init() {
	config = configuration.Config

	log.Printf("Orchestration module initialized with configuration: %s\n", config.OrchestrationUrl)
}

type Orchestration interface {
	PerformBackup() error
	GetCurrentBackupState() ([]domain.CreateBackupResponse, error)
	GetHistoryById(id string) (domain.BackupStatusResponse, error)
	CreateBackup(backupId string) error
}
