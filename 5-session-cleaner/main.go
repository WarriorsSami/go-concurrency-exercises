//////////////////////////////////////////////////////////////////////
//
// Given is a SessionManager that stores session information in
// memory. The SessionManager itself is working, however, since we
// keep on adding new sessions to the manager our program will
// eventually run out of memory.
//
// Your task is to implement a session cleaner routine that runs
// concurrently in the background and cleans every session that
// hasn't been updated for more than 5 seconds (of course usually
// session times are much longer).
//
// Note that we expect the session to be removed anytime between 5 and
// 7 seconds after the last update. Also, note that you have to be
// very careful in order to prevent race conditions.
//

package main

import (
	"errors"
	"log"
	"sync"
	"time"
)

// ICleaner is an interface for cleaning idle sessions
type ICleaner interface {
	Clean(sessionManager *SessionManager)
}

// SessionCleaner is a cleaner that cleans idle sessions
type SessionCleaner struct {
	clock *time.Ticker
	done  chan interface{}
}

// NewSessionCleaner creates a new sessionCleaner
func NewSessionCleaner() *SessionCleaner {
	return &SessionCleaner{
		clock: time.NewTicker(1 * time.Second),
	}
}

// Clean cleans idle sessions
func (c *SessionCleaner) Clean(sm *SessionManager) {
	for {
		select {
		case <-c.clock.C:
			sm.rwLock.Lock()

			for sessionID, session := range sm.sessions {
				if time.Since(session.updatedAt) > 5*time.Second {
					delete(sm.sessions, sessionID)
				}
			}

			sm.rwLock.Unlock()
		case <-c.done:
			return
		}
	}
}

// SessionManager keeps track of all sessions from creation, updating
// to destroying.
type SessionManager struct {
	sessions       map[string]Session
	rwLock         sync.RWMutex
	sessionCleaner ICleaner
}

// Session stores the session's data
type Session struct {
	Data      map[string]interface{}
	updatedAt time.Time
}

// NewSessionManager creates a new sessionManager
func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions:       make(map[string]Session),
		sessionCleaner: NewSessionCleaner(),
	}

	go m.sessionCleaner.Clean(m)
	return m
}

// CreateSession creates a new session and returns the sessionID
func (m *SessionManager) CreateSession() (string, error) {
	sessionID, err := MakeSessionID()
	if err != nil {
		return "", err
	}

	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	m.sessions[sessionID] = Session{
		Data:      make(map[string]interface{}),
		updatedAt: time.Now(),
	}

	return sessionID, nil
}

// ErrSessionNotFound returned when sessionID not listed in
// SessionManager
var ErrSessionNotFound = errors.New("SessionID does not exists")

// GetSessionData returns data related to session if sessionID is
// found, errors otherwise
func (m *SessionManager) GetSessionData(sessionID string) (map[string]interface{}, error) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session.Data, nil
}

// UpdateSessionData overwrites the old session data with the new one
func (m *SessionManager) UpdateSessionData(sessionID string, data map[string]interface{}) error {
	m.rwLock.RLock()

	_, ok := m.sessions[sessionID]
	if !ok {
		defer m.rwLock.RUnlock()
		return ErrSessionNotFound
	}
	m.rwLock.RUnlock()

	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	// Hint: you should renew expiry of the session here
	m.sessions[sessionID] = Session{
		Data:      data,
		updatedAt: time.Now(),
	}

	return nil
}

func main() {
	// Create new sessionManager and new session
	m := NewSessionManager()
	sID, err := m.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Created new session with ID", sID)

	// Update session data
	data := make(map[string]interface{})
	data["website"] = "longhoang.de"

	err = m.UpdateSessionData(sID, data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Update session data, set website to longhoang.de")

	// Retrieve data from manager again
	updatedData, err := m.GetSessionData(sID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Get session data:", updatedData)
}
