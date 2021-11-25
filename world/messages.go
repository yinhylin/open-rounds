package world

import "rounds/pb"

type IntentsUpdate struct {
	ID      string
	Intents map[pb.Intents_Intent]struct{}
	Tick    int64
}

type EntityUpdate struct {
	ID     string
	Entity Entity
	Tick   int64
}

type AngleUpdate struct {
	ID    string
	Angle float64
	Tick  int64
}

type AddEntity struct {
	ID   string
	Tick int64
}

type AddBullet struct {
	Source string
	ID     string
	Tick   int64
}

type RemoveEntity struct {
	ID   string
	Tick int64
}
