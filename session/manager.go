package session

import (
	"sync"
	"time"
)

type Config struct {
	SessionTimeout time.Duration
	SavePath       string
	ExecutablePath string
}

type Manager struct {
	config        *Config
	sessions      map[string]*Session
	sessionsMutex sync.Locker
}

func NewManager(config *Config) *Manager {
	return &Manager{
		config:        config,
		sessions:      make(map[string]*Session),
		sessionsMutex: &sync.Mutex{},
	}
}

func (man *Manager) HasSession(id string) bool {
	man.sessionsMutex.Lock()
	defer man.sessionsMutex.Unlock()

	sess, ok := man.sessions[id]
	return ok && sess.Valid()
}

func (man *Manager) GetOrCreateSession(id string) (*Session, error) {
	man.sessionsMutex.Lock()
	defer man.sessionsMutex.Unlock()

	var err error
	sess, ok := man.sessions[id]
	if !ok || !sess.Valid() {
		sess, err = newSession(id, man, *man.config)
		if err != nil {
			return nil, err
		}
		man.sessions[id] = sess
	}

	return sess, err
}

func (man *Manager) removeSession(id string) {
	man.sessionsMutex.Lock()
	defer man.sessionsMutex.Unlock()

	delete(man.sessions, id)
}
