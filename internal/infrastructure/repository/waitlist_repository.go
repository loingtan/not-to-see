package repository

import (
	"context"

	domain "cobra-template/internal/domain/registration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WaitlistRepository implements WaitlistRepository using GORM
type WaitlistRepository struct {
	db *gorm.DB
}

// NewWaitlistRepository creates a new GORM waitlist repository
func NewWaitlistRepository(db *gorm.DB) domain.WaitlistRepository {
	return &WaitlistRepository{
		db: db,
	}
}

// Create creates a new waitlist entry
func (r *WaitlistRepository) Create(ctx context.Context, entry *domain.WaitlistEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

// GetByStudentAndSection retrieves a waitlist entry by student and section
func (r *WaitlistRepository) GetByStudentAndSection(ctx context.Context, studentID, sectionID uuid.UUID) (*domain.WaitlistEntry, error) {
	var entry domain.WaitlistEntry
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("Section").
		Where("student_id = ? AND section_id = ?", studentID, sectionID).
		First(&entry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

// GetNextInLine retrieves the next student in line for a section
func (r *WaitlistRepository) GetNextInLine(ctx context.Context, sectionID uuid.UUID) (*domain.WaitlistEntry, error) {
	var entry domain.WaitlistEntry
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("Section").
		Where("section_id = ?", sectionID).
		Order("position ASC").
		First(&entry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

// GetNextPosition retrieves the next available position for a section
func (r *WaitlistRepository) GetNextPosition(ctx context.Context, sectionID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.WaitlistEntry{}).
		Where("section_id = ?", sectionID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count) + 1, nil
}

// Delete deletes a waitlist entry by ID
func (r *WaitlistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.WaitlistEntry{}, "waitlist_id = ?", id).Error
}

// GetBySectionID retrieves all waitlist entries for a section
func (r *WaitlistRepository) GetBySectionID(ctx context.Context, sectionID uuid.UUID) ([]*domain.WaitlistEntry, error) {
	var entries []*domain.WaitlistEntry
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("Section").
		Where("section_id = ?", sectionID).
		Order("position ASC").
		Find(&entries).Error
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// GetByStudentID retrieves all waitlist entries for a student
func (r *WaitlistRepository) GetByStudentID(ctx context.Context, studentID uuid.UUID) ([]*domain.WaitlistEntry, error) {
	var entries []*domain.WaitlistEntry
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("Section").
		Where("student_id = ?", studentID).
		Order("timestamp ASC").
		Find(&entries).Error
	if err != nil {
		return nil, err
	}
	return entries, nil
}
