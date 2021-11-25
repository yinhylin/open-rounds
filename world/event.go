package world

import "rounds/pb"

func ClientEventToServerEvent(serverTick int64, e *pb.ClientEvent) *pb.ServerEvent {
	switch e.Event.(type) {
	case *pb.ClientEvent_Intents:
		return &pb.ServerEvent{
			Tick: serverTick,
			Player: &pb.PlayerDetails{
				Tick: e.Tick,
				Id:   e.Id,
			},
			Event: &pb.ServerEvent_Intents{
				Intents: e.GetIntents(),
			},
		}

	case *pb.ClientEvent_Angle:
		return &pb.ServerEvent{
			Tick: serverTick,
			Player: &pb.PlayerDetails{
				Tick: e.Tick,
				Id:   e.Id,
			},
			Event: &pb.ServerEvent_Angle{
				Angle: e.GetAngle(),
			},
		}

	case *pb.ClientEvent_Shoot:
		return &pb.ServerEvent{
			Tick: serverTick,
			Player: &pb.PlayerDetails{
				Tick: e.Tick,
				Id:   e.Id,
			},
			Event: &pb.ServerEvent_Shoot{
				Shoot: e.GetShoot(),
			},
		}
	}
	return nil
}
