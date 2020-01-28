package gorm

import (
	"github.com/jinzhu/gorm"
	"github.com/slim-crown/issue-1-website/internal/services/session"
	"time"
)

// SessionGormRepo implements session.Repository interface
type sessionRepo struct {
	db *gorm.DB
}

// NewSessionRepo  returns a new SessionGormRepo object
func NewSessionRepo(db *gorm.DB) session.Repository {
	return &sessionRepo{db: db}
}

// GetSession returns a given stored session
func (repo *sessionRepo) GetSession(sessionID string) (*session.Session, []error) {
	s := session.Session{Data: make([]session.MapPair, 0)}
	errs := repo.db.First(&s, "uuid=?", sessionID).GetErrors()
	if len(errs) > 0 {
		return nil, errs
	}
	err := repo.db.Model(&s).Association("Data").Find(&s.Data).Error
	if err != nil {
		return nil, []error{err}
	}
	return &s, errs
}

// AddSession stores a given session
func (repo *sessionRepo) AddSession(s *session.Session) (*session.Session, []error) {
	errs := repo.db.Save(s).GetErrors()
	if len(errs) > 0 {
		return nil, errs
	}
	return s, errs
}

// DeleteSession deletes a given session
func (repo *sessionRepo) DeleteSession(sessionID string) (*session.Session, []error) {
	s, errs := repo.GetSession(sessionID)
	if len(errs) > 0 {
		return nil, errs
	}
	errs = repo.db.Delete(s, "uuid=?", s.UUID).GetErrors()
	if len(errs) > 0 {
		return nil, errs
	}
	return s, errs
}

// StartSessionGC launches an infinite recursive routine that cleans
// expired sessions every interval of the specified duration.
func (repo *sessionRepo) StartSessionGC(duration time.Duration) {
	time.AfterFunc(duration, func() {
		_ = repo.db.Delete(session.Session{}, "expires<", time.Now())
		repo.StartSessionGC(duration)
	})
}

// UpdateSession stores a given session
func (repo *sessionRepo) UpdateSession(s *session.Session) (*session.Session, []error) {
	errs := repo.db.Save(s).GetErrors()
	if len(errs) > 0 {
		return nil, errs
	}
	return s, errs
}
