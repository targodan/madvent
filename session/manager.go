package session

import "github.com/targodan/madvent/adventure"

type Manager struct {
	sessions map[string]*adventure.Adventure
}
