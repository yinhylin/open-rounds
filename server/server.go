package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"rounds/pb"
	"rounds/world"
	"sync"
	"time"

	"net/http/pprof"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wspb"
)

// TODO: non-global subscribers. rooms or something?
type subscriber struct {
	Messages chan *pb.ServerEvent
	PlayerID string
	c        *websocket.Conn
}

type event struct {
	*pb.ClientEvent
	*subscriber
}

type Server struct {
	subscribers map[*subscriber]struct{}
	mu          sync.RWMutex
	serveMux    http.ServeMux
	events      chan *event
	state       *world.StateBuffer
}

func NewServer() *Server {
	s := &Server{
		subscribers: make(map[*subscriber]struct{}),
		events:      make(chan *event, 1024),
		state:       world.NewStateBuffer(32),
	}

	s.state.Add(&world.State{
		Entities: make(map[string]world.Entity),
		Tick:     0,
	})

	go func() {
		sync := time.NewTicker(25 * time.Millisecond)
		tick := time.NewTicker(17 * time.Millisecond)
		for {
			select {
			case <-sync.C:
				// Send all the clients the current tick the server is on.
				s.publish(&pb.ServerEvent{
					ServerTick: s.state.CurrentTick(),
				})

			case <-tick.C:
				s.onTick()
			}
		}
	}()

	s.serveMux.HandleFunc("/", s.onConnection)
	s.serveMux.HandleFunc("/debug/pprof/", pprof.Index)
	s.serveMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.serveMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.serveMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.serveMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return s
}

func (s *Server) onEvent(e *event) (*pb.ServerEvent, error) {
	switch e.Event.(type) {
	case *pb.ClientEvent_Intents:
		if err := s.state.ApplyIntents(&world.IntentsUpdate{
			ID:      e.Id,
			Tick:    e.Tick,
			Intents: world.IntentsFromProto(e.GetIntents()),
		}); err != nil {
			// Resync the client if there are any issues with their intents.
			e.subscriber.Messages <- &pb.ServerEvent{
				Tick:       s.state.CurrentTick(),
				ServerTick: s.state.CurrentTick(),
				Event: &pb.ServerEvent_State{
					State: s.state.ToProto(),
				},
			}
			return nil, nil
		}

		return &pb.ServerEvent{
			Tick:       e.Tick,
			ServerTick: s.state.CurrentTick(),
			Event: &pb.ServerEvent_EntityEvents{
				EntityEvents: &pb.EntityEvents{
					Id:      e.Id,
					Intents: e.GetIntents(),
				},
			},
		}, nil

	case *pb.ClientEvent_Connect:
		e.subscriber.PlayerID = e.Id
		s.state.AddEntity(&world.AddEntity{
			Tick: s.state.CurrentTick(),
			ID:   e.Id,
		})

		return &pb.ServerEvent{
			Tick:       s.state.CurrentTick(),
			ServerTick: s.state.CurrentTick(),
			Event: &pb.ServerEvent_AddEntity{
				AddEntity: &pb.AddEntity{
					Entity: &pb.Entity{
						Id: e.Id,
					},
				},
			},
		}, nil

	case *pb.ClientEvent_Angle:
		s.state.ApplyAngle(&world.AngleUpdate{
			ID:    e.Id,
			Tick:  e.Tick,
			Angle: e.GetAngle().Angle,
		})
		return &pb.ServerEvent{
			Tick:       e.Tick,
			ServerTick: s.state.CurrentTick(),
			Event: &pb.ServerEvent_EntityAngle{
				EntityAngle: &pb.EntityAngle{
					Id:    e.Id,
					Angle: e.GetAngle().Angle,
				},
			},
		}, nil

	case *pb.ClientEvent_Shoot:
		s.state.AddBullet(&world.AddBullet{
			ID:     e.GetShoot().Id,
			Tick:   e.Tick,
			Source: e.Id,
		})
		return &pb.ServerEvent{
			Tick:       e.Tick,
			ServerTick: s.state.CurrentTick(),
			Event: &pb.ServerEvent_EntityShoot{
				EntityShoot: &pb.EntityShoot{
					Id:       e.GetShoot().Id,
					SourceId: e.Id,
				},
			},
		}, nil

	case *pb.ClientEvent_RequestState:
		e.subscriber.Messages <- &pb.ServerEvent{
			Tick:       s.state.CurrentTick(),
			ServerTick: s.state.CurrentTick(),
			Event: &pb.ServerEvent_State{
				State: s.state.ToProto(),
			},
		}

	}
	return nil, nil
}

func (s *Server) onTick() {
	s.state.Next()
	var serverEvents []*pb.ServerEvent

	for len(s.events) > 0 {
		event := <-s.events
		serverEvent, err := s.onEvent(event)
		if err != nil {
			continue
		}
		if serverEvent != nil {
			serverEvents = append(serverEvents, serverEvent)
		}
	}

	for _, event := range serverEvents {
		s.publish(event)
	}
}

func (s *Server) addSubscriber(sub *subscriber) {
	s.mu.Lock()
	s.subscribers[sub] = struct{}{}
	s.mu.Unlock()
}

func (s *Server) removeSubscriber(sub *subscriber) {
	s.mu.Lock()
	delete(s.subscribers, sub)
	s.mu.Unlock()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.serveMux.ServeHTTP(w, r)
}

func (s *Server) onConnection(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{
			"tayrawr.com",
			"tayrawr.com:6969",
			"44.241.110.166",
			"44.241.110.166:6969",
			"localhost:8080",
		},
	})
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "")

	if err := s.handleConnection(r.Context(), c); err != nil {
		log.Println(err)
		return
	}
	log.Println("disconnect? idk")
}

func (s *Server) handleConnection(ctx context.Context, c *websocket.Conn) error {
	sub := &subscriber{
		Messages: make(chan *pb.ServerEvent, 1024),
		c:        c,
	}
	s.addSubscriber(sub)

	defer s.removeSubscriber(sub)

	go func() {
		defer func() {
			if sub.PlayerID == "" {
				return
			}
			s.removeSubscriber(sub)
			s.publish(&pb.ServerEvent{
				Tick:       s.state.CurrentTick(),
				ServerTick: s.state.CurrentTick(),
				Event: &pb.ServerEvent_RemoveEntity{
					RemoveEntity: &pb.RemoveEntity{
						Id: sub.PlayerID,
					},
				},
			})
			s.state.RemoveEntity(&world.RemoveEntity{
				Tick: s.state.CurrentTick(),
				ID:   sub.PlayerID,
			})
		}()

		for {
			var clientEvent pb.ClientEvent
			if err := wspb.Read(ctx, c, &clientEvent); err != nil {
				log.Println(err)
				return
			}
			s.events <- &event{
				&clientEvent,
				sub,
			}
		}
	}()
	for {
		select {
		case msg := <-sub.Messages:
			if err := wspb.Write(ctx, c, msg); err != nil {
				log.Println(err)
			}
		case <-ctx.Done():
			fmt.Println(":(")
			return ctx.Err()
		}
	}
}

func (s *Server) publish(event *pb.ServerEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for sub := range s.subscribers {
		select {
		case sub.Messages <- event:
		default:
			sub.c.Close(websocket.StatusPolicyViolation, "write would block???")
		}
	}
}

func Run(args []string) error {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	address := "localhost:4242"
	if len(args) > 1 {
		address = args[1]
	}
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	log.Printf("Listening on http://%v", l.Addr())
	server := NewServer()
	s := &http.Server{
		Handler:      server,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	errc := make(chan error, 1)
	go func() {
		errc <- s.Serve(l)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-errc:
		log.Println(err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}

	ctx := context.Background()
	if err := s.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
