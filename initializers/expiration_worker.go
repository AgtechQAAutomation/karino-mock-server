package initializers

import (
	"log"
	"time"

	"gorm.io/gorm"
)


func MarkExpiredRows(db *gorm.DB, expirationSeconds int) error {

	query := `
	UPDATE delivery_documents
	SET status = 'EXPIRED'
	WHERE status = 'NOT EXPIRED'
	  AND id_created_at IS NOT NULL
	  AND id_created_at + INTERVAL ? SECOND <= NOW()
	`

	return db.Exec(query, expirationSeconds).Error
}

func StartExpirationWorker(db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for range ticker.C {
			err := MarkExpiredRows(
				db,
				AppConfig.ExpirationTimeSeconds,
			)
			if err != nil {
				log.Println("expiration worker error:", err)
			}
		}
	}()
}

