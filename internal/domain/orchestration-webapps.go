package domain

type Snapshot struct {
	ID        string
	Timestamp int64
	Status    string
}

type BackupHistoryEntry struct {
	ID        string
	Timestamp int64
	Status    string
}

type CreateBackup struct {
	BackupId string `json:"backupId"`
}

type CreateBackupResponse struct {
	History []string `json:"scheduledSnapshots"`
}

type BackupStatusResponse struct {
	BackupId int    `json:"backupId"`
	State    string `json:"state"`
}
