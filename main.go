package main

import (
	"encoding/json"
	"log"
	"sort"

	"github.com/itsnuyen/camunda-backup-helper/internal/domain"
	"github.com/itsnuyen/camunda-backup-helper/internal/server"
)

func main() {
	// simpleTest()
	server.StartServer()
}

func simpleTest() {
	webappjson := `[
  {
    "backupId": 1771709316,
    "state": "COMPLETED"
  },
  {
    "backupId": 1771709272,
    "state": "COMPLETED"
  },
  {
    "backupId": 1771709173,
    "state": "COMPLETED"
  },
  {
    "backupId": 1771708887,
    "state": "COMPLETED"
  },
  {
    "backupId": 1771704701,
    "state": "COMPLETED"
  }
]`

	zeebejson := `{
    "backupId": 1771709316,
    "state": "COMPLETED"
  },
  {
    "backupId": 1771709272,
    "state": "COMPLETED"
  },
  {
    "backupId": 1771704701,
    "state": "COMPLETED"
  }
]`

	var camundaBackups []domain.BackupStatusResponse
	var zeebeBackups []domain.BackupStatusResponse

	// Unmarshal the JSON data into the respective slices
	// Handle errors appropriately in production code
	_ = json.Unmarshal([]byte(webappjson), &camundaBackups)
	_ = json.Unmarshal([]byte(zeebejson), &zeebeBackups)
	mergedData := mergeAndSortBackups(camundaBackups, zeebeBackups)

	log.Printf("Amount of data: %i\n", len(mergedData))

	log.Printf("Merged and sorted backup data: %+v\n", mergedData)

	mynewData := mergedData[1:]

	log.Printf("mynewData: %+v\n", mynewData)

}

func mergeAndSortBackups(zeebeBackups, camundaBackups []domain.BackupStatusResponse) []domain.BackupStatusResponse {
	merged := append(zeebeBackups, camundaBackups...)
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].BackupId > merged[j].BackupId
	})

	return removeDuplicates(merged)
}

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
