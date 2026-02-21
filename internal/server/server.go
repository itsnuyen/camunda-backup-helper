package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/itsnuyen/camunda-backup-helper/internal/configuration"
)

var config configuration.BackupHelperConfig

func init() {
	config = *configuration.Config
}

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		responseMap := map[string]string{"key": "value"}
		responseJSON, err := json.Marshal(responseMap)
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)
	})

	mux.HandleFunc("/create-camundabackup", createCamundaBackupHandler)
	mux.HandleFunc("/backup-webapp-history", historyHandlerWebApps)
	mux.HandleFunc("/backup-zeebe-history", historyHandlerZeebe)
	mux.HandleFunc("/backup-cleanup", cleanUpHandler)
	mux.HandleFunc("/backup-merged-history", backupMergedHistoryHandler)

	if err := http.ListenAndServe(readServerPortEnv(), mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func readServerPortEnv() string {
	log.Printf("Server will start on port: %s", config.Serverport)
	return fmt.Sprintf(":%s", config.Serverport)
}
