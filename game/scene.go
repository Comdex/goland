package game

import (
	"fmt"
	"sync"
)

type Scene struct {
	byProperty map[PropType][]*Actor
	Actors     map[string]*Actor

	// lock for protecting data
	m sync.Mutex

	// systems running in this scene
	systems map[string]System

	// wg for synchronized systems shutdown
	wg sync.WaitGroup
}

func NewScene() *Scene {
	return &Scene{
		byProperty: make(map[PropType][]*Actor),
		Actors:     make(map[string]*Actor),
		systems:    make(map[string]System),
	}
}

func (s *Scene) Add(id string) *Actor {
	s.m.Lock()
	defer s.m.Unlock()

	a := NewActor(id)
	e := s.addActor(a)
	a.scene = s

	if e != nil {
		panic(e)
	}

	return a
}

func (s *Scene) addActor(a *Actor) error {
	if _, ok := s.Actors[a.ID]; ok {
		return fmt.Errorf("scene: addactor: already has actor id %s", a.ID)
	}

	s.Actors[a.ID] = a
	for t := range a.properties {
		s.cache(a, t)
	}

	return nil
}

// Removes a given actor from the scene.
func (s *Scene) Remove(a *Actor) {
	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.Actors[a.ID]; !ok {
		return
	}

	delete(s.Actors, a.ID)

	for t := range a.properties {
		s.uncache(a, t)
	}
}

func (s *Scene) cache(a *Actor, t PropType) {
	s.m.Lock()
	defer s.m.Unlock()
	if _, ok := s.byProperty[t]; !ok {
		s.byProperty[t] = []*Actor{}
	}
	s.byProperty[t] = append(s.byProperty[t], a)
}

func (s *Scene) uncache(a *Actor, t PropType) {
	s.m.Lock()
	defer s.m.Unlock()
	if actors, ok := s.byProperty[t]; ok {
		// TODO: pre-allocate right size rather than constant resizing
		newlist := []*Actor{}
		for _, actor := range actors {
			// keep all but the uncached one
			if actor.ID != a.ID {
				newlist = append(newlist, actor)
			}
		}
		s.byProperty[t] = newlist
	}
}

// Allows for very specialized query of the sccene by property type.
// Given one or more property types, will return a list of all actors
// that contain every given type. For very large scenes this will
// probably have to be improved in many ways, possibly by using a
// binary search tree.
func (s *Scene) Find(p ...PropType) (result []*Actor) {
	s.m.Lock()
	defer s.m.Unlock()

	// opt: exclude actors without first property
	if actors, ok := s.byProperty[p[0]]; ok {
		if len(p) == 1 {
			// opt: quit now if only looking for one property
			result = actors
			return
		}

		for _, a := range actors {
			if len(p) > len(a.properties) {
				// opt: exclude actors with less properties than requested
				// this assumes that actors only have one of each property type
				continue
			}

			// opt: we already checked prop at 0
			rest := p[1:]
			// opt: requested less or equal props than actor has
			// loop on those rather than all the props of the actor
			// look until we find a prop that doesn't match
			hit := true
			for _, wanted := range rest {
				if _, ok := a.properties[wanted]; !ok {
					hit = false
					break
				}
			}

			if hit {
				result = append(result, a)
			}
		}
	}

	return result
}

// Track our systems
func (s *Scene) AddSystem(sys System) {
	s.m.Lock()
	defer s.m.Unlock()

	s.wg.Add(1)
	s.systems[sys.String()] = sys
}

func (s *Scene) RemoveSystem(sys System) {
	s.m.Lock()
	defer s.m.Unlock()

	s.wg.Done()
	delete(s.systems, sys.String())
}

// Request stop of all systems and wait for them to shut down
func (s *Scene) StopSystems() {
	for _, sys := range s.systems {
		sys.Stop()
	}

	s.wg.Wait()
}
