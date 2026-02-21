package orchestration

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/itsnuyen/camunda-backup-helper/internal/domain"
)

func GetMergedBackupHistory() ([]domain.BackupStatusResponse, error) {
	webappHistory, err := GetBackupHistory(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get webapp backup history: %w", err)
	}

	zeebeHistory, err := GetBackupHistory(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get zeebe backup history: %w", err)
	}

	mergedHistory := mergeAndSortBackups(webappHistory, zeebeHistory)
	return mergedHistory, nil
}

func GetBackupHistory(isRuntimeData bool) ([]domain.BackupStatusResponse, error) {

	url := fmt.Sprintf("%s/actuator/backupHistory", config.OrchestrationUrl)

	if isRuntimeData {
		url = fmt.Sprintf("%s/actuator/backupRuntime", config.OrchestrationUrl)
	}

	log.Printf("Get backup history from url %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var history []domain.BackupStatusResponse
	if err := json.Unmarshal(body, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	sort.Slice(history, func(i, j int) bool {
		return history[i].BackupId > history[j].BackupId
	})
	return history, nil
}
