package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLogRepository defines the interface for audit log data access
type AuditLogRepository interface {
	Create(*AuditLog) error
	GetByID(id uuid.UUID) (*AuditLog, error)
	GetByUserID(userID uuid.UUID, limit int) ([]AuditLog, error)
	GetByResource(resource, namespace, clusterID string, limit int) ([]AuditLog, error)
	GetAll(limit int) ([]AuditLog, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

func (r *auditLogRepository) Create(log *AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepository) GetByID(id uuid.UUID) (*AuditLog, error) {
	var log AuditLog
	err := r.db.First(&log, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *auditLogRepository) GetByUserID(userID uuid.UUID, limit int) ([]AuditLog, error) {
	if limit == 0 {
		limit = 100
	}
	var logs []AuditLog
	err := r.db.Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *auditLogRepository) GetByResource(resource, namespace, clusterID string, limit int) ([]AuditLog, error) {
	if limit == 0 {
		limit = 100
	}
	var logs []AuditLog
	query := r.db.Where("resource = ?", resource)
	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}
	if clusterID != "" {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("timestamp DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *auditLogRepository) GetAll(limit int) ([]AuditLog, error) {
	if limit == 0 {
		limit = 100
	}
	var logs []AuditLog
	err := r.db.Order("timestamp DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// LogAction creates an audit log entry for a user action
func (r *auditLogRepository) LogAction(userID uuid.UUID, action, resource, resourceName, namespace, clusterID, status string, details map[string]interface{}) error {
	log := &AuditLog{
		ID:           uuid.New(),
		UserID:       userID,
		Action:       action,
		Resource:     resource,
		ResourceName: resourceName,
		Namespace:    namespace,
		ClusterID:    clusterID,
		Status:       status,
		Timestamp:    time.Now(),
	}

	// Serialize details as JSON string
	if details != nil {
		// GORM will handle JSON serialization for the Details field
		// This is a simplified version - in production you'd want proper JSON handling
		log.Details = "{}"
	}

	return r.Create(log)
}
