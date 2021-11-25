package world

import "rounds/pb"

type UpdateBuffer struct {
	Intents map[string]map[pb.Intents_Intent]struct{}
	Angles  map[string]float64
	Add     map[string]struct{}
	Remove  map[string]struct{}
	Shots   map[string][]string
}

func UpdateBufferFromProto(p *pb.UpdateBuffer) UpdateBuffer {
	u := NewUpdateBuffer()
	for _, intent := range p.Intents {
		u.Intents[intent.Id] = IntentsFromProto(intent.Intents)
	}
	for _, ID := range p.Add {
		u.Add[ID] = struct{}{}
	}
	for _, ID := range p.Remove {
		u.Remove[ID] = struct{}{}
	}
	for _, angle := range p.Angles {
		u.Angles[angle.Id] = angle.Angle
	}
	for _, shot := range p.Shots {
		u.Shots[shot.SourceId] = append(u.Shots[shot.SourceId], shot.Id)
	}
	return u
}

func (u *UpdateBuffer) ToProto(tick int64) *pb.UpdateBuffer {
	p := &pb.UpdateBuffer{
		Tick: tick,
	}
	for ID, intents := range u.Intents {
		p.Intents = append(p.Intents, &pb.EntityIntents{
			Id:      ID,
			Intents: IntentsToProto(intents),
		})
	}
	for ID := range u.Add {
		p.Add = append(p.Add, ID)
	}
	for ID := range u.Remove {
		p.Remove = append(p.Remove, ID)
	}
	for ID, angle := range u.Angles {
		p.Angles = append(p.Angles, &pb.EntityAngle{
			Id:    ID,
			Angle: angle,
		})
	}
	for source, shots := range u.Shots {
		for _, ID := range shots {
			p.Shots = append(p.Shots, &pb.EntityShoot{
				Id:       ID,
				SourceId: source,
			})
		}
	}
	return p
}

func NewUpdateBuffer() UpdateBuffer {
	return UpdateBuffer{
		Intents: make(map[string]map[pb.Intents_Intent]struct{}),
		Angles:  make(map[string]float64),
		Add:     make(map[string]struct{}),
		Remove:  make(map[string]struct{}),
		Shots:   make(map[string][]string),
	}
}
