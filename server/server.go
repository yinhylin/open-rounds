package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"rounds/pb"
	"rounds/world"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

// TODO: non-global subscribers. rooms or something?
type subscriber struct {
	Messages chan []byte
	PlayerID string
	c        *websocket.Conn
}

type event struct {
	*pb.ClientEvent
	*subscriber
}

type Server struct {
	subscribers map[*subscriber]struct{}
	mu          sync.Mutex
	serveMux    http.ServeMux
	events      chan *event
	state       *world.StateBuffer
}

func NewServer() *Server {
	s := &Server{
		subscribers: make(map[*subscriber]struct{}),
		events:      make(chan *event, 1024),
		state:       world.NewStateBuffer(60),
	}

	s.state.Add(&world.State{
		Entities: make(map[string]world.Entity),
		Tick:     0,
	})

	go func() {
		sync := time.Tick(250 * time.Second)
		tick := time.Tick(17 * time.Millisecond)
		ticks := 0
		for {
			select {
			case <-sync:
				// If we're behind, catch up.
				for i := 0; i < 15-ticks; i++ {
					s.onTick()
				}
				s.publish(&pb.ServerEvent{
					Tick:  s.state.CurrentTick(),
					Event: &pb.ServerEvent_ServerTick{},
				})
				ticks = 0
			case <-tick:
				s.onTick()
				ticks++
			}
		}
	}()

	s.serveMux.HandleFunc("/", s.onConnection)
	return s
}

func (s *Server) onEvent(e *event) (*pb.ServerEvent, error) {
	switch e.Event.(type) {
	case *pb.ClientEvent_Intents:
		s.state.ApplyIntents(&world.IntentsUpdate{
			ID:      e.Id,
			Tick:    e.Tick,
			Intents: world.IntentsFromProto(e.GetIntents()),
		})

		return &pb.ServerEvent{
			Tick: e.Tick,
			Event: &pb.ServerEvent_EntityEvents{
				EntityEvents: &pb.EntityEvents{
					Id:      e.Id,
					Intents: e.GetIntents(),
				},
			},
		}, nil

	case *pb.ClientEvent_Connect:
		e.subscriber.PlayerID = e.Id
		log.Println("adding entity")
		s.state.AddEntity(&world.AddEntity{
			ID:   e.Id,
			Tick: s.state.CurrentTick(),
		})

		return &pb.ServerEvent{
			Tick: s.state.CurrentTick(),
			Event: &pb.ServerEvent_AddEntity{
				AddEntity: &pb.AddEntity{
					Entity: &pb.Entity{
						Id: e.Id,
					},
				},
			},
		}, nil
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
	states := &pb.States{}
	s.state.ForEachEntity(func(ID string, entity *world.Entity) {
		states.States = append(states.States, entity.ToProto())
	})
	sub.Messages <- toBytesOrDie(&pb.ServerEvent{
		Tick: s.state.CurrentTick(),
		Event: &pb.ServerEvent_States{
			States: states,
		},
	})

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
		Messages: make(chan []byte, 1024),
		c:        c,
	}
	s.addSubscriber(sub)

	defer s.removeSubscriber(sub)

	// this is the magic event loop
	go func() {
		defer func() {
			if sub.PlayerID == "" {
				return
			}
			s.removeSubscriber(sub)
			s.publish(&pb.ServerEvent{
				Event: &pb.ServerEvent_RemoveEntity{
					RemoveEntity: &pb.RemoveEntity{
						Id: sub.PlayerID,
					},
				},
			})
			s.state.RemoveEntity(&world.RemoveEntity{
				ID: sub.PlayerID,
			})
		}()

		for {
			messageType, reader, err := c.Reader(ctx)
			if err != nil {
				return
			}

			if messageType != websocket.MessageBinary {
				log.Println("unexpected message type", messageType)
				return
			}

			b, err := ioutil.ReadAll(reader)
			if err != nil {
				log.Println(err)
				return
			}

			var clientEvent pb.ClientEvent
			err = proto.Unmarshal(b, &clientEvent)
			if err != nil {
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
			c.Write(ctx, websocket.MessageBinary, msg)
		case <-ctx.Done():
			fmt.Println(":(")
			return ctx.Err()
		}
	}
}

func toBytesOrDie(event *pb.ServerEvent) []byte {
	bytes, err := proto.Marshal(event)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}

func (s *Server) publish(event *pb.ServerEvent) {
	bytes := toBytesOrDie(event)
	s.mu.Lock()
	defer s.mu.Unlock()
	for sub := range s.subscribers {
		select {
		case sub.Messages <- bytes:
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
