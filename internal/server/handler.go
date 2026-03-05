package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/itsnuyen/camunda-backup-helper/internal/orchestration"
)

func historyHandlerWebApps(w http.ResponseWriter, r *http.Request) {
	historyData, err := orchestration.GetBackupHistory(false)
	if err != nil {
		log.Printf("Error performing backup: %v", err)
		http.Error(w, fmt.Sprintf("Failed to perform backup: %v", err), http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(historyData)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func historyHandlerZeebe(w http.ResponseWriter, r *http.Request) {
	historyData, err := orchestration.GetBackupHistory(true)
	if err != nil {
		log.Printf("Error performing backup: %v", err)
		http.Error(w, fmt.Sprintf("Failed to perform backup: %v", err), http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(historyData)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func createCamundaBackupHandler(w http.ResponseWriter, r *http.Request) {
	orchestration.PauseZeebeExport()
	backupId, err := orchestration.PerformBackup()
	if err != nil {
		log.Printf("Error performing backup: %v", err)
		orchestration.ResumeZeebeExport()
		http.Error(w, fmt.Sprintf("Failed to perform backup: %v", err), http.StatusInternalServerError)
		return
	}
	orchestration.ResumeZeebeExport()

	responseMap := map[string]string{"backup": fmt.Sprint(backupId)}
	responseJSON, err := json.Marshal(responseMap)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func cleanUpHandler(w http.ResponseWriter, r *http.Request) {

	cleanupIds, err := orchestration.Cleanup()
	if err != nil {
		log.Printf("Error performing backup: %v", err)
		http.Error(w, fmt.Sprintf("Failed to perform backup: %v", err), http.StatusInternalServerError)
		return
	}

	responseMap := map[string]interface{}{"status": "success", "deleted": cleanupIds}
	responseJSON, err := json.Marshal(responseMap)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func backupMergedHistoryHandler(w http.ResponseWriter, r *http.Request) {
	historyData, err := orchestration.GetMergedBackupHistory()
	if err != nil {
		log.Printf("Error performing backup: %v", err)
		http.Error(w, fmt.Sprintf("Failed to perform backup: %v", err), http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(historyData)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}
