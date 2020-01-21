package session

import (
	"fmt"
	"sync"
	"time"
)

//Session represents logged in user session
type Session struct {
	lock           sync.Mutex `gorm:"-"`
	sessionService Service    `gorm:"-"`

	UUID           string            `gorm:"type:text;not null;primary_key"`
	Expires        time.Time         `gorm:"not null"`
	LastAccessTime time.Time         `gorm:"not null"`
	Data           []MapPair         `gorm:"foreignkey:session_uuid;association_foreignkey:uuid"`
	dataMap        map[string]string `gorm:"-"`
	mapPopulated   bool              `gorm:"-"`
}

// MapPair is used to implement a map key-value pair.
type MapPair struct {
	SessionUUID string `gorm:"type:text;not null;primary_key"`
	Key         string `gorm:"type:text;not null;primary_key"`
	Value       string `gorm:"type:text"`
}

// TableName sets custom name for gorm tables.
func (MapPair) TableName() string {
	return "session_data"
}

// Set sets/replaces the given value for the given key and also updates the
// session persistence.
func (s *Session) Set(key, value string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.dataMap[key] = value

	s.syncFromMapToArray()
	_, errs := (s.sessionService).UpdateSession(s)
	if len(errs) > 0 {
		return fmt.Errorf("update session failed when setting cookie value because: %+v", errs)
	}
	return nil
}

// Get returns the value under the given key.
func (s *Session) Get(key string) string {
	if v, ok := s.dataMap[key]; ok {
		return v
	}
	return ""
}

// Delete removes any value from the given key.
func (s *Session) Delete(key string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.dataMap, key)

	s.syncFromMapToArray()
	_, errs := (s.sessionService).UpdateSession(s)

	if len(errs) > 0 {
		return fmt.Errorf("update session failed when deleting cookie value because: %+v", errs)
	}
	return nil
}

func (s *Session) syncFromMapToArray() {
	if !s.mapPopulated {
		// if not helper map (dataMap) not initially populated yet
		s.syncFromArrayToMap()
	}
	s.Data = make([]MapPair, 0)
	for k, v := range s.dataMap {
		s.Data = append(s.Data, MapPair{Key: k, Value: v})
	}
}

func (s *Session) syncFromArrayToMap() {
	s.dataMap = make(map[string]string)
	for _, mapPair := range s.Data {
		s.dataMap[mapPair.Key] = mapPair.Value
	}
	s.mapPopulated = true
}
