package gorm

import (
	"github.com/jinzhu/gorm"
	"github.com/slim-crown/issue-1-website/internal/services/session"
)

// SessionGormRepo implements user.SessionRepository interface
type sessionRepo struct {
	conn *gorm.DB
}

// NewSessionRepo  returns a new SessionGormRepo object
func NewSessionRepo(db *gorm.DB) session.Repository {
	return &sessionRepo{conn: db}
}

// GetSession returns a given stored session
func (repo *sessionRepo) GetSession(sessionID string) (*session.Session, []error) {
	s := session.Session{}
	errs := repo.conn.Find(&s, "uuid=?", sessionID).GetErrors()
	if len(errs) > 0 {
		return nil, errs
	}
	return &s, errs
}

// AddSession stores a given session
func (repo *sessionRepo) AddSession(s *session.Session) (*session.Session, []error) {
	errs := repo.conn.Save(s).GetErrors()
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
	errs = repo.conn.Delete(s, s.ID).GetErrors()
	if len(errs) > 0 {
		return nil, errs
	}
	return s, errs
}
