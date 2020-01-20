package session

// Service specifies logged in user session related service
type Service interface {
	GetSession(sessionID string) (*Session, []error)
	AddSession(session *Session) (*Session, []error)
	DeleteSession(sessionID string) (*Session, []error)
}

// Repository specifies logged in user session related database operations
type Repository interface {
	GetSession(sessionID string) (*Session, []error)
	AddSession(session *Session) (*Session, []error)
	DeleteSession(sessionID string) (*Session, []error)
}

type service struct {
	repo *Repository
}

// NewService  returns a new SessionService object
func NewService(r *Repository) Service {
	return &service{repo: r}
}

// GetSession returns a given stored session
func (s *service) GetSession(sessionID string) (*Session, []error) {
	sess, errs := (*s.repo).GetSession(sessionID)
	if len(errs) > 0 {
		return nil, errs
	}
	return sess, errs
}

// AddSession stores a given session
func (s *service) AddSession(session *Session) (*Session, []error) {
	sess, errs := (*s.repo).AddSession(session)
	if len(errs) > 0 {
		return nil, errs
	}
	return sess, errs
}

// DeleteSession deletes a given session
func (s *service) DeleteSession(sessionID string) (*Session, []error) {
	sess, errs := (*s.repo).DeleteSession(sessionID)
	if len(errs) > 0 {
		return nil, errs
	}
	return sess, errs
}
