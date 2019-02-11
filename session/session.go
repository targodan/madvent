package session

import (
	"errors"
	"path"
	"sync"
	"time"

	"github.com/targodan/madvent/adventure"
)

const (
	saveFileExtension = ".advent"
)

type Session struct {
	ID       string
	saveFile string

	manager *Manager

	timer   *time.Timer
	timeout time.Duration
	abort   chan struct{}
	mutex   sync.Locker

	adventure *adventure.Adventure

	valid bool
}

func getSaveName(savePath, id string) string {
	return path.Join(savePath, id+saveFileExtension)
}

func newSession(id string, manager *Manager, config Config) (*Session, error) {
	saveName := getSaveName(config.SavePath, id)
	adv, err := adventure.NewOrResume(config.ExecutablePath, saveName)
	if err != nil {
		return nil, err
	}
	sess := &Session{
		ID:        id,
		saveFile:  saveName,
		manager:   manager,
		timer:     time.NewTimer(config.SessionTimeout),
		timeout:   config.SessionTimeout,
		abort:     make(chan struct{}),
		mutex:     &sync.Mutex{},
		adventure: adv,
		valid:     true,
	}

	_, _, err = sess.adventure.Start()
	if err != nil {
		return nil, err
	}

	go func() {
		select {
		case <-sess.abort:
			return
		case <-sess.timer.C:
			sess.Close()
		}
	}()

	return sess, nil
}

func (sess *Session) invalidate() {
	sess.valid = false
	close(sess.abort)
	sess.adventure = nil
}

func (sess *Session) Valid() bool {
	return sess.valid
}

func (sess *Session) resetTimer() error {
	if !sess.Valid() {
		return errors.New("already expired, create a new session instead")
	}

	sess.timer.Reset(sess.timeout)

	return nil
}

func (sess *Session) ResetTimer() error {
	sess.mutex.Lock()
	defer sess.mutex.Unlock()

	return sess.resetTimer()
}

func (sess *Session) Output() <-chan string {
	sess.mutex.Lock()
	defer sess.mutex.Unlock()

	return sess.adventure.Output()
}

func (sess *Session) Writeln(text string) error {
	sess.mutex.Lock()
	defer sess.mutex.Unlock()

	return sess.adventure.Writeln(text)
}

func (sess *Session) Close() {
	sess.mutex.Lock()
	defer sess.mutex.Unlock()

	if !sess.Valid() {
		return
	}

	sess.adventure.SaveAndClose(sess.saveFile)

	sess.invalidate()
	sess.manager.removeSession(sess.ID)
}
