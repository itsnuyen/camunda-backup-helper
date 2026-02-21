package orchestration

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/itsnuyen/camunda-backup-helper/internal/domain"
)

func Cleanup() ([]string, error) {
	// Implement cleanup logic here, e.g., delete old backups, clear temporary files, etc.
	// Return a list of cleaned items and any error that occurs during the cleanup process.

	maxKeepItems, err := strconv.Atoi(config.KeepBackUpItems)
	if err != nil {
		return nil, fmt.Errorf("invalid value for KeepBackUpItems: %w", err)
	}
	log.Printf("Starting cleanup process, keeping the latest %d backup items", maxKeepItems)

	zeebeBackups, err := GetBackupHistory(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get Zeebe backup history: %w", err)
	}
	log.Printf("Retrieved %d Zeebe backup items", len(zeebeBackups))
	camundaBackups, err := GetBackupHistory(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get Camunda backup history: %w", err)
	}
	log.Printf("Retrieved %d Camunda backup items", len(camundaBackups))

	mergedBackups := mergeAndSortBackups(zeebeBackups, camundaBackups)
	log.Printf("Merged and sorted total of %d backup items", len(mergedBackups))

	if len(mergedBackups) <= maxKeepItems {
		log.Printf("No cleanup needed, total backup items (%d) is within the limit (%d)", len(mergedBackups), maxKeepItems)
		return []string{}, nil
	}

	itemsDeleted := []string{}
	// ignore the first item and start with the second item, as the first item is the latest backup which we want to keep
	for i := 1; i < len(mergedBackups); i++ {
		backup := mergedBackups[i]
		DeleteBackup("backupRuntime", int64(backup.BackupId))
		DeleteBackup("backupHistory", int64(backup.BackupId))
		itemsDeleted = append(itemsDeleted, fmt.Sprintf("Deleted backup with ID %d", backup.BackupId))
		log.Printf("Deleted backup with ID %d", backup.BackupId)
	}

	log.Printf("Cleanup completed, deleted %d old backup items", len(itemsDeleted))
	return itemsDeleted, nil
}

func mergeAndSortBackups(zeebeBackups, camundaBackups []domain.BackupStatusResponse) []domain.BackupStatusResponse {
	merged := append(zeebeBackups, camundaBackups...)
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].BackupId > merged[j].BackupId
	})
	return removeDuplicates(merged)
}

// removeDuplicates removes duplicate backups based on BackupId
func removeDuplicates(backups []domain.BackupStatusResponse) []domain.BackupStatusResponse {
	seen := make(map[int64]bool)
	result := []domain.BackupStatusResponse{}

	for _, b := range backups {
		if !seen[int64(b.BackupId)] {
			seen[int64(b.BackupId)] = true
			result = append(result, b)
		}
	}

	return result
}

// DeleteBackup sends a DELETE request to remove a backup by ID
// Corresponds to: curl --request DELETE 'http://localhost:9600/actuator/backupRuntime/100'
// endpoint: the base URL endpoint (e.g., "/actuator/backupRuntime" or "/actuator/backupHistory")
// backupId: the ID of the backup to delete
func DeleteBackup(endpoint string, backupId int64) error {

	url := fmt.Sprintf("%s/actuator/%s/%d", config.OrchestrationUrl, endpoint, backupId)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}
	defer resp.Body.Close()
	// need to check if the response status code indicates success (e.g., 200 OK or 204 No Content)
	log.Printf("Delete request sent for backup ID %d to endpoint %s, received status code: %d", backupId, endpoint, resp.StatusCode)
	// if resp.StatusCode < 200 || resp.StatusCode >= 300 {
	// 	return fmt.Errorf("delete request failed with status: %d", resp.StatusCode)
	// }

	log.Printf("Successfully deleted backup %d from %s", backupId, endpoint)
	return nil
}
