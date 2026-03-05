package orchestration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/itsnuyen/camunda-backup-helper/internal/domain"
)

func PerformBackup() (int64, error) {
	// Placeholder for backup creation logic

	// need to perform the backup here

	backupId := generateBackupID()
	
	if err := CreateOptimizeBackup(backupId); err != nil {
		return 0, fmt.Errorf("failed to create optimize backup: %w", err)
	}
	time.Sleep(time.Duration(2 * time.Second))
	if err := WaitForOptimizeBackupCompletion(backupId); err != nil {
		return 0, fmt.Errorf("optimize backup did not complete successfully: %w", err)
	}
	if err := CreateBackup(backupId, false); err != nil {
		return 0, fmt.Errorf("failed to create backup webapps: %w", err)
	}
	time.Sleep(time.Duration(2 * time.Second))
	if err := WaitForBackupCamundaWebAppsCompletion(backupId); err != nil {
		return 0, fmt.Errorf("backup webapps did not complete successfully: %w", err)
	}
	if err := CreateZeebeRecordsSnapshot(backupId, config.SnapshotZeebeRepo); err != nil {
		return 0, fmt.Errorf("failed to create Zeebe records snapshot: %w", err)
	}
	time.Sleep(time.Duration(2 * time.Second))
	if err := CreateBackup(backupId, true); err != nil {
		return 0, fmt.Errorf("failed to create backup zeebe runtime: %w", err)
	}
	time.Sleep(time.Duration(2 * time.Second))
	if err := WaitForBackupRuntimeCompletion(backupId); err != nil {
		return 0, fmt.Errorf("backup zeebe runtime did not complete successfully: %w", err)
	}

	return backupId, nil
}

// GetBackupHistoryStatus retrieves the current status of a backup history by ID
// Corresponds to: curl -s "$ORCHESTRATION_CLUSTER_MANAGEMENT_API/actuator/backupHistory/$BACKUP_ID"
func GetBackupHistoryStatus(backupId int64) (*domain.BackupStatusResponse, error) {
	url := fmt.Sprintf("%s/actuator/backupHistory/%d", config.OrchestrationUrl, backupId)
	log.Printf("Get Status from url %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup history status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var status domain.BackupStatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &status, nil
}

// WaitForBackupCamundaWebAppsCompletion polls the backup history status until it reaches COMPLETED state
// Corresponds to: while [[ "$(curl -s ... | jq -r .state)" != "COMPLETED" ]]; do echo "Waiting..."; sleep 5; done
func WaitForBackupCamundaWebAppsCompletion(backupId int64) error {
	fmt.Printf("Waiting for backup webapps %d to complete...\n", backupId)

	for {
		status, err := GetBackupHistoryStatus(backupId)
		if err != nil {
			return fmt.Errorf("failed to check backup status: %w", err)
		}

		if status.State == "COMPLETED" {
			fmt.Printf("Finished backup history with ID %d\n", backupId)
			return nil
		}

		if status.State == "FAILED" {
			return fmt.Errorf("backup failed with reason: %v", status)
		}

		fmt.Println("Waiting...")
		time.Sleep(5 * time.Second)
	}
}

// CreateZeebeRecordsSnapshot creates an Elasticsearch snapshot for Zeebe records
// Corresponds to: curl -XPUT "$ELASTIC_ENDPOINT/_snapshot/$ELASTIC_SNAPSHOT_REPOSITORY/camunda_zeebe_records_backup_$BACKUP_ID?wait_for_completion=true"
func CreateZeebeRecordsSnapshot(backupId int64, snapshotRepository string) error {
	snapshotName := fmt.Sprintf("camunda_zeebe_records_backup_%d", backupId)
	url := fmt.Sprintf("%s/_snapshot/%s/%s?wait_for_completion=true",
		config.SecondarystorageUrl, snapshotRepository, snapshotName)

	// Create request body with indices pattern and feature_states
	requestBody := map[string]interface{}{
		"indices":        "zeebe-record*",
		"feature_states": []string{"none"},
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create Zeebe records snapshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	fmt.Printf("Successfully created Zeebe records snapshot: %s\n", snapshotName)
	return nil
}

// CreateBackup initiates a runtime backup operation with the given backup ID
// Corresponds to: curl -XPOST "$ORCHESTRATION_CLUSTER_MANAGEMENT_API/actuator/backupRuntime" -d '{"backupId": $BACKUP_ID}'
func CreateBackup(backupId int64, runtimeBackup bool) error {
	log.Printf("Orchestration Url: %s\n", config.OrchestrationUrl)
	url := ""
	if runtimeBackup {
		url = fmt.Sprintf("%s/actuator/backupRuntime", config.OrchestrationUrl)
		log.Printf("Creating backup runtime with ID %d...\n for url: %s", backupId, url)
	} else {
		url = fmt.Sprintf("%s/actuator/backupHistory", config.OrchestrationUrl)
		log.Printf("Creating backup history with ID %d...\n for url: %s", backupId, url)
	}

	// Create request body with backupId
	requestBody := map[string]int64{"backupId": backupId}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create backup runtime: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetBackupRuntimeStatus retrieves the current status of a backup runtime by ID
// Corresponds to: curl -s "$ORCHESTRATION_CLUSTER_MANAGEMENT_API/actuator/backupRuntime/$BACKUP_ID"
func GetBackupRuntimeStatus(backupId int64) (*domain.BackupStatusResponse, error) {
	url := fmt.Sprintf("%s/actuator/backupRuntime/%d", config.OrchestrationUrl, backupId)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup runtime status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var status domain.BackupStatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &status, nil
}

// WaitForBackupRuntimeCompletion polls the backup runtime status until it reaches COMPLETED state
// Corresponds to: while [[ "$(curl -s ... | jq -r .state)" != "COMPLETED" ]]; do echo "Waiting..."; sleep 5; done
func WaitForBackupRuntimeCompletion(backupId int64) error {
	fmt.Printf("Waiting for backup runtime %d to complete...\n", backupId)

	for {
		status, err := GetBackupRuntimeStatus(backupId)
		if err != nil {
			return fmt.Errorf("failed to check backup runtime status: %w", err)
		}

		if status.State == "COMPLETED" {
			fmt.Printf("Finished backup runtime with ID %d\n", backupId)
			return nil
		}

		if status.State == "FAILED" {
			return fmt.Errorf("backup runtime failed with reason: %v", status)
		}

		fmt.Println("Waiting...")
		time.Sleep(5 * time.Second)
	}
}

// generateBackupID creates a new backup ID using Unix timestamp
// Corresponds to: export BACKUP_ID=$(date +%s)
func generateBackupID() int64 {
	return time.Now().Unix()
}

func PauseZeebeExport() error {
	return handleZeebeExportSoftPause(true)
}

func ResumeZeebeExport() error {
	return handleZeebeExportSoftPause(false)
}

func handleZeebeExportSoftPause(zeebeExport bool) error {
	url := fmt.Sprintf("%s/actuator/exporting/pause?soft=%t", config.OrchestrationUrl, zeebeExport)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to pause Zeebe export: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	return nil
}
