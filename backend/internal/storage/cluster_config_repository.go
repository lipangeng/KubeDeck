package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ClusterConfig represents a Kubernetes cluster configuration
type ClusterConfig struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	Name        string     `gorm:"uniqueIndex;size:255" json:"name"`
	Server      string     `gorm:"size:500" json:"server"`
	Token       string     `gorm:"size:2000" json:"-"`    // Don't return in JSON
	Kubeconfig  string     `gorm:"type:text" json:"-"`    // Don't return in JSON
	CACert      string     `gorm:"type:text" json:"-"`    // Don't return in JSON
	Status      string     `gorm:"size:20" json:"status"` // connected, disconnected, error
	Version     string     `gorm:"size:50" json:"version,omitempty"`
	Nodes       int        `gorm:"default:0" json:"nodes,omitempty"`
	LastConnect *time.Time `json:"last_connect,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ClusterConfigRepository defines the interface for cluster config data access
type ClusterConfigRepository interface {
	Create(*ClusterConfig) error
	GetByID(id uuid.UUID) (*ClusterConfig, error)
	GetAll() ([]ClusterConfig, error)
	Update(*ClusterConfig) error
	Delete(id uuid.UUID) error
	GetByName(name string) (*ClusterConfig, error)
}

type clusterConfigRepository struct {
	db *gorm.DB
}

// NewClusterConfigRepository creates a new cluster config repository
func NewClusterConfigRepository(db *gorm.DB) ClusterConfigRepository {
	return &clusterConfigRepository{db: db}
}

func (r *clusterConfigRepository) Create(cluster *ClusterConfig) error {
	return r.db.Create(cluster).Error
}

func (r *clusterConfigRepository) GetByID(id uuid.UUID) (*ClusterConfig, error) {
	var cluster ClusterConfig
	err := r.db.First(&cluster, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *clusterConfigRepository) GetAll() ([]ClusterConfig, error) {
	var clusters []ClusterConfig
	err := r.db.Find(&clusters).Error
	return clusters, err
}

func (r *clusterConfigRepository) Update(cluster *ClusterConfig) error {
	return r.db.Save(cluster).Error
}

func (r *clusterConfigRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&ClusterConfig{ID: id}).Error
}

func (r *clusterConfigRepository) GetByName(name string) (*ClusterConfig, error) {
	var cluster ClusterConfig
	err := r.db.Where("name = ?", name).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}
