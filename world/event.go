package world

import "rounds/pb"

func ClientEventToServerEvent(serverTick int64, e *pb.ClientEvent) *pb.ServerEvent {
	switch e.Event.(type) {
	case *pb.ClientEvent_Intents:
		return &pb.ServerEvent{
			Tick:       e.Tick,
			ServerTick: serverTick,
			Event: &pb.ServerEvent_PlayerIntents{
				PlayerIntents: &pb.PlayerIntents{
					Id:      e.Id,
					Intents: e.GetIntents(),
				},
			},
		}

	case *pb.ClientEvent_Angle:
		return &pb.ServerEvent{
			Tick:       e.Tick,
			ServerTick: serverTick,
			Event: &pb.ServerEvent_PlayerAngle{
				PlayerAngle: &pb.PlayerAngle{
					Id:    e.Id,
					Angle: e.GetAngle().Angle,
				},
			},
		}

	case *pb.ClientEvent_Shoot:
		return &pb.ServerEvent{
			Tick:       e.Tick,
			ServerTick: serverTick,
			Event: &pb.ServerEvent_PlayerShoot{
				PlayerShoot: &pb.PlayerShoot{
					Id:       e.GetShoot().Id,
					SourceId: e.Id,
				},
			},
		}
	}
	return nil
}
