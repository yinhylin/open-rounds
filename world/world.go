package world

import "log"

type World struct {
	entities map[string]*Entity
}

func NewWorld() *World {
	return &World{
		entities: make(map[string]*Entity),
	}
}

func (w *World) Update() {
	for _, entity := range w.entities {
		entity.Update()
	}
}

func (w *World) AddEntity(ID string, e *Entity) {
	if _, ok := w.entities[ID]; ok {
		log.Fatalf("%s already exists", ID)
	}
	w.entities[ID] = e
}

func (w *World) Entity(ID string) *Entity {
	return w.entities[ID]
}

func (w *World) RemoveEntity(ID string) {
	delete(w.entities, ID)
}

func (w *World) ForEachEntity(callback func(string, *Entity)) {
	for ID, entity := range w.entities {
		callback(ID, entity)
	}
}
