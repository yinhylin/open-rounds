package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/sailormoon/open-rounds/pb"
	"github.com/sailormoon/open-rounds/world"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wspb"
)

//go:embed maps/*
var mapsFS embed.FS

func loadMaps() (map[string]*world.Map, error) {
	maps := make(map[string]*world.Map)
	files, err := mapsFS.ReadDir("maps")
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			return nil, errors.New("directories not supported")
		}

		if _, ok := maps[f.Name()]; ok {
			return nil, fmt.Errorf("already seen %s", f.Name())
		}

		file, err := mapsFS.Open(strings.Join([]string{"maps", f.Name()}, "/"))
		if err != nil {
			return nil, err
		}

		contents, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}

		m, err := world.LoadMap(string(contents))
		if err != nil {
			return nil, err
		}
		maps[f.Name()] = m
	}
	return maps, nil
}

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
	maps        map[string]*world.Map
}

func NewServer() *Server {
	maps, err := loadMaps()
	if err != nil {
		log.Fatal(err)
	}
	s := &Server{
		subscribers: make(map[*subscriber]struct{}),
		events:      make(chan *event, 1024),
		state:       world.NewStateBuffer(32, maps["basic"]),
		maps:        maps,
	}

	s.state.Add(&world.State{
		Players: make(map[string]world.Player),
		Tick:    0,
	})

	go func() {
		tick := time.NewTicker(time.Second / 60)
		for {
			select {
			case <-tick.C:
				if s.onTick() == 0 {
					now := time.Now()
					defer func() {
						if time.Since(now) > 10*time.Millisecond {
							log.Println("long update", time.Since(now))
						}
					}()
					// Send all the clients the current tick the server is on.
					// TODO: Sync less often.
					s.publish(&pb.ServerEvent{
						Tick: s.state.CurrentTick(),
					})
				}
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

func (s *Server) onEvent(e *event) *pb.ServerEvent {
	if event := world.ClientEventToServerEvent(s.state.CurrentTick(), e.ClientEvent); event != nil {
		return event
	}

	switch e.Event.(type) {
	case *pb.ClientEvent_Connect:
		e.subscriber.PlayerID = e.Id
		return &pb.ServerEvent{
			Tick: s.state.CurrentTick(),
			Player: &pb.PlayerDetails{
				Id:   e.Id,
				Tick: s.state.CurrentTick(),
			},
			Event: &pb.ServerEvent_AddPlayer{},
		}

	case *pb.ClientEvent_RequestState:
		e.subscriber.Messages <- &pb.ServerEvent{
			Tick: s.state.CurrentTick(),
			Event: &pb.ServerEvent_State{
				State: s.state.ToProto(),
			},
		}

	}
	return nil
}

func (s *Server) onTick() int {
	s.state.Next()
	var serverEvents []*pb.ServerEvent
	for len(s.events) > 0 {
		event := <-s.events
		if serverEvent := s.onEvent(event); serverEvent != nil {
			s.state.OnEvent(serverEvent)
			serverEvents = append(serverEvents, serverEvent)
		}
	}
	for _, event := range serverEvents {
		s.publish(event)
	}
	return len(serverEvents)
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
			event := &pb.ServerEvent{
				Tick: s.state.CurrentTick(),
				Player: &pb.PlayerDetails{
					Tick: s.state.CurrentTick(),
					Id:   sub.PlayerID,
				},
				Event: &pb.ServerEvent_RemovePlayer{
					RemovePlayer: &pb.RemovePlayer{},
				},
			}
			if err := s.state.OnEvent(event); err != nil {
				log.Println(err)
			}
			s.publish(event)
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
