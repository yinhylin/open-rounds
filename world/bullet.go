package world

import "rounds/pb"

type Bullet struct {
	Coords   Vector
	Velocity Vector
}

func (b *Bullet) ToProto() *pb.Bullet {
	return &pb.Bullet{
		Position: b.Coords.ToProto(),
		Velocity: b.Velocity.ToProto(),
	}
}

func BulletFromProto(b *pb.Bullet) *Bullet {
	if b == nil {
		return nil
	}
	return &Bullet{
		Coords:   VectorFromProto(b.Position),
		Velocity: VectorFromProto(b.Velocity),
	}
}

func BulletsFromProto(b []*pb.Bullet) []Bullet {
	bullets := make([]Bullet, 0, len(b))
	for _, bullet := range b {
		bullets = append(bullets, *BulletFromProto(bullet))
	}
	return bullets
}

func BulletsToProto(b []Bullet) []*pb.Bullet {
	bullets := make([]*pb.Bullet, 0, len(b))
	for _, bullet := range b {
		bullets = append(bullets, bullet.ToProto())
	}
	return bullets
}
