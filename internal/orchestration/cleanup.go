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

	mergedBackups, corruptedOnes := mergeAndSortBackups(zeebeBackups, camundaBackups)
	log.Printf("Merged and sorted total of %d backup items", len(mergedBackups))

	if len(mergedBackups) <= maxKeepItems {
		log.Printf("No cleanup needed, total backup items (%d) is within the limit (%d)", len(mergedBackups), maxKeepItems)
		return []string{}, nil
	}

	// TODO need to specify where to delete
	for _, backup := range corruptedOnes {
		DeleteBackup("backupRuntime", int64(backup.BackupId))
		DeleteBackup("backupHistory", int64(backup.BackupId))
	}
	log.Printf("delete all faulty backup first %v", len(corruptedOnes))

	itemsToDelete := mergedBackups[maxKeepItems:]
	// ignore the first item and start with the second item, as the first item is the latest backup which we want to keep
	for i := range itemsToDelete {
		backup := itemsToDelete[i]
		DeleteBackup("backupRuntime", int64(backup.BackupId))
		DeleteBackup("backupHistory", int64(backup.BackupId))
		log.Printf("Deleted backup with ID %d", backup.BackupId)
	}

	log.Printf("Cleanup completed, deleted %d old backup items", len(itemsToDelete))
	deletedIds := []string{}
	for _, backup := range itemsToDelete {
		log.Printf("Deleted backup ID: %d, State: %s", backup.BackupId, backup.State)
		deletedIds = append(deletedIds, fmt.Sprintf("%d", backup.BackupId))
	}
	return deletedIds, nil
}

func mergeAndSortBackups(zeebeBackups, camundaBackups []domain.BackupStatusResponse) ([]domain.BackupStatusResponse, []domain.BackupStatusResponse) {
	merged := append(zeebeBackups, camundaBackups...)
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].BackupId > merged[j].BackupId
	})
	return removeDuplicates(merged)
}

// removeDuplicates removes duplicate backups based on BackupId
func removeDuplicates(backups []domain.BackupStatusResponse) ([]domain.BackupStatusResponse, []domain.BackupStatusResponse) {
	seen := make(map[int64]int) // tracks count of each BackupId
	result := []domain.BackupStatusResponse{}
	trulyUnique := []domain.BackupStatusResponse{}

	// First pass: count occurrences
	for _, b := range backups {
		seen[int64(b.BackupId)]++
	}

	// Second pass: build both lists
	added := make(map[int64]bool)
	for _, b := range backups {
		id := int64(b.BackupId)

		// Deduplicated list (first occurrence only)
		requiredSeen := 2
		if config.OptimizeBackup != "false" {
			requiredSeen = 3
		}

		if !added[id] {
			added[id] = true
			if seen[id] == requiredSeen {
				result = append(result, b)
			}
		}

		// Truly unique list (only appears once)
		if seen[id] == 1 {
			trulyUnique = append(trulyUnique, b)
		}
	}

	for _, tru := range trulyUnique {
		log.Printf("Truely unique ones with no duplication %v", tru.BackupId)
	}

	return result, trulyUnique
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
