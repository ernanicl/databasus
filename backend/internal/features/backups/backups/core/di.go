package backups_core

var backupRepository = &BackupRepository{}

func GetBackupRepository() *BackupRepository {
	return backupRepository
}
