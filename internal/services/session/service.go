package session

import "time"

// Service specifies logged in user session related service
type Service interface {
	NewSession(sessionID string, maxHardLifetime time.Duration) (*Session, []error)
	GetSession(sessionID string) (*Session, []error)
	AddSession(session *Session) (*Session, []error)
	UpdateSession(session *Session) (*Session, []error)
	DeleteSession(sessionID string) (*Session, []error)
}

// Repository specifies logged in user session related database operations
type Repository interface {
	GetSession(sessionID string) (*Session, []error)
	AddSession(session *Session) (*Session, []error)
	UpdateSession(session *Session) (*Session, []error)
	DeleteSession(sessionID string) (*Session, []error)
}

type service struct {
	repo *Repository
}

// NewService  returns a new SessionService object
func NewService(r *Repository) Service {
	return &service{repo: r}
}

// NewSession returns a new session using the given sessionID.
func (s *service) NewSession(sessionID string, maxHardLifetime time.Duration) (*Session, []error) {
	sess := &Session{
		sessionService: s,
		UUID:           sessionID,
		Expires:        time.Now().Add(maxHardLifetime),
		LastAccessTime: time.Now(),
		Data:           make([]MapPair, 0),
		dataMap:        make(map[string]string, 0),
	}
	sess, err := s.AddSession(sess)
	return sess, err
}

// GetSession returns a given stored session
func (s *service) GetSession(sessionID string) (*Session, []error) {
	sess, errs := (*s.repo).GetSession(sessionID)
	if len(errs) > 0 {
		return nil, errs
	}
	// use UpdateSession to refresh last access time
	sess, errs = s.UpdateSession(sess)
	sess.sessionService = s
	sess.syncFromArrayToMap()
	return sess, errs
}

// AddSession stores a given session
func (s *service) AddSession(session *Session) (*Session, []error) {
	return (*s.repo).AddSession(session)
}

// UpdateSession stores a given session
func (s *service) UpdateSession(session *Session) (*Session, []error) {
	session.LastAccessTime = time.Now()
	return (*s.repo).UpdateSession(session)
}

// DeleteSession deletes a given session
func (s *service) DeleteSession(sessionID string) (*Session, []error) {
	return (*s.repo).DeleteSession(sessionID)
}
