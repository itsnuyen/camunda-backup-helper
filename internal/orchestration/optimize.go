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

// CreateOptimizeBackup initiates an Optimize backup operation with the given backup ID
// Corresponds to: curl --fail -XPOST "$OPTIMIZE_MANAGEMENT_API/actuator/backups" -H "Content-Type: application/json" -d '{"backupId": $BACKUP_ID}'
func CreateOptimizeBackup(backupId int64) error {
	url := fmt.Sprintf("%s/actuator/backups", config.OptimizeUrl)
	log.Printf("Creating optimize backup with ID %d... for url: %s", backupId, url)

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
		return fmt.Errorf("failed to create optimize backup: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetOptimizeBackupStatus retrieves the current status of an Optimize backup by ID
// Corresponds to: curl --fail -s "$OPTIMIZE_MANAGEMENT_API/actuator/backups/$BACKUP_ID"
func GetOptimizeBackupStatus(backupId int64) (*domain.BackupStatusResponse, error) {
	url := fmt.Sprintf("%s/actuator/backups/%d", config.OptimizeUrl, backupId)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimize backup status: %w", err)
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

// WaitForOptimizeBackupCompletion polls the Optimize backup status until it reaches COMPLETED state
// Corresponds to: while [[ "$(curl --fail -s ... | jq -r .state)" != "COMPLETED" ]]; do echo "Waiting for Optimize..."; sleep 5; done
func WaitForOptimizeBackupCompletion(backupId int64) error {
	for {
		status, err := GetOptimizeBackupStatus(backupId)
		if err != nil {
			return fmt.Errorf("failed to check optimize backup status: %w", err)
		}

		if status.State == "COMPLETED" {
			fmt.Printf("\nFinished Optimize backup with ID %d\n", backupId)
			return nil
		}

		if status.State == "FAILED" {
			return fmt.Errorf("optimize backup failed with reason: %v", status)
		}

		fmt.Println("\nWaiting for Optimize...")
		time.Sleep(5 * time.Second)
	}
}
