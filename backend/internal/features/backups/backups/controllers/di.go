package backups_controllers

import (
	backups_services "databasus-backend/internal/features/backups/backups/services"
	"databasus-backend/internal/features/databases"
)

var backupController = &BackupController{
	backups_services.GetBackupService(),
}

func GetBackupController() *BackupController {
	return backupController
}

var postgresWalBackupController = &PostgreWalBackupController{
	databases.GetDatabaseService(),
	backups_services.GetWalService(),
}

func GetPostgresWalBackupController() *PostgreWalBackupController {
	return postgresWalBackupController
}
