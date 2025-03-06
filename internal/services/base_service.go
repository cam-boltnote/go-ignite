package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/cam-boltnote/go-ignite/internal/utils"

	"gorm.io/gorm"
)

// BaseService provides common CRUD operations for services
type BaseService[T any] struct {
	db     *gorm.DB
	logger *utils.Logger
}

// NewBaseService creates a new base service instance
func NewBaseService[T any](db *gorm.DB) *BaseService[T] {
	return &BaseService[T]{
		db:     db,
		logger: utils.GetLogger().WithService("base_service"),
	}
}

// Create creates a new record
func (s *BaseService[T]) Create(ctx context.Context, model *T) error {
	s.logger.Info("Creating new record", map[string]interface{}{
		"model_type": fmt.Sprintf("%T", *model),
	})

	err := s.db.WithContext(ctx).Create(model).Error
	if err != nil {
		s.logger.Error("Failed to create record", err, map[string]interface{}{
			"model_type": fmt.Sprintf("%T", *model),
		})
	}
	return err
}

// GetByID retrieves a record by ID
func (s *BaseService[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	s.logger.Debug("Fetching record by ID", map[string]interface{}{
		"id": id,
	})

	var model T
	err := s.db.WithContext(ctx).First(&model, id).Error
	if err != nil {
		s.logger.Error("Failed to fetch record", err, map[string]interface{}{
			"id": id,
		})
		return nil, err
	}
	return &model, nil
}

// List retrieves a list of records with pagination
func (s *BaseService[T]) List(ctx context.Context, page, pageSize int) ([]T, int64, error) {
	s.logger.Debug("Listing records", map[string]interface{}{
		"page":      page,
		"page_size": pageSize,
	})

	var models []T
	var total int64

	offset := (page - 1) * pageSize
	err := s.db.WithContext(ctx).Model(new(T)).Count(&total).Error
	if err != nil {
		s.logger.Error("Failed to count records", err, nil)
		return nil, 0, err
	}

	err = s.db.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&models).Error
	if err != nil {
		s.logger.Error("Failed to fetch records", err, map[string]interface{}{
			"page":      page,
			"page_size": pageSize,
		})
		return nil, 0, err
	}

	return models, total, nil
}

// Update updates a record
func (s *BaseService[T]) Update(ctx context.Context, model *T) error {
	s.logger.Info("Updating record", map[string]interface{}{
		"model_type": fmt.Sprintf("%T", *model),
	})

	err := s.db.WithContext(ctx).Save(model).Error
	if err != nil {
		s.logger.Error("Failed to update record", err, map[string]interface{}{
			"model_type": fmt.Sprintf("%T", *model),
		})
	}
	return err
}

// Delete deletes a record
func (s *BaseService[T]) Delete(ctx context.Context, id uint) error {
	s.logger.Info("Deleting record", map[string]interface{}{
		"id": id,
	})

	result := s.db.WithContext(ctx).Delete(new(T), id)
	if result.Error != nil {
		s.logger.Error("Failed to delete record", result.Error, map[string]interface{}{
			"id": id,
		})
		return result.Error
	}
	if result.RowsAffected == 0 {
		err := errors.New("record not found")
		s.logger.Warn("No record found to delete", map[string]interface{}{
			"id": id,
		})
		return err
	}
	return nil
}

// ServiceError represents a service-level error
type ServiceError struct {
	Code    int
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Common service error codes
const (
	ErrNotFound           = 404
	ErrInvalidInput       = 400
	ErrUnauthorized       = 401
	ErrForbidden          = 403
	ErrInternalServer     = 500
	ErrServiceUnavailable = 503
)
