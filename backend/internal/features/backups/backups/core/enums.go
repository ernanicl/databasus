package backups_core

type BackupStatus string

const (
	BackupStatusInProgress BackupStatus = "IN_PROGRESS"
	BackupStatusCompleted  BackupStatus = "COMPLETED"
	BackupStatusFailed     BackupStatus = "FAILED"
	BackupStatusCanceled   BackupStatus = "CANCELED"
)

type PgWalUploadType string

const (
	PgWalUploadTypeBasebackup PgWalUploadType = "basebackup"
	PgWalUploadTypeWal        PgWalUploadType = "wal"
)
