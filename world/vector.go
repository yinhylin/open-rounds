package world

import "github.com/sailormoon/open-rounds/pb"

type Vector struct {
	X, Y float64
}

func (v *Vector) ToTileCoordinates() (int64, int64) {
	return int64(v.X / tileSize), int64(v.Y / tileSize)
}

func (v *Vector) ToProto() *pb.Vector {
	return &pb.Vector{
		X: v.X,
		Y: v.Y,
	}
}

func VectorFromProto(v *pb.Vector) Vector {
	return Vector{
		X: v.X,
		Y: v.Y,
	}
}
