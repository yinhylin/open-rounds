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

type Server struct {
	subscribers map[*subscriber]struct{}
	mu          sync.Mutex
	serveMux    http.ServeMux
}

func NewServer() *Server {
	s := &Server{
		subscribers: make(map[*subscriber]struct{}),
	}
	s.serveMux.HandleFunc("/", s.onConnection)
	return s
}

func (s *Server) addSubscriber(sub *subscriber) {
	s.mu.Lock()
	s.subscribers[sub] = struct{}{}
	for other := range s.subscribers {
		if other.PlayerID == "" {
			continue
		}
		sub.Messages <- toBytesOrDie(&pb.ServerEvent{
			PlayerId: other.PlayerID,
			Event:    &pb.ServerEvent_AddPlayer{},
		})
	}
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
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "")

	log.Println("connection")
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

			s.publish(&pb.ServerEvent{
				PlayerId: sub.PlayerID,
				Event:    &pb.ServerEvent_RemovePlayer{},
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

			// TODO: we should do movement vectors and validation
			// TODO: better handling of server events in a separate area.
			var serverEvent *pb.ServerEvent
			switch clientEvent.Event.(type) {
			case *pb.ClientEvent_Move:
				serverEvent = &pb.ServerEvent{
					// TODO: Send player numbers to clients, not UUIDs.
					PlayerId: clientEvent.PlayerUuid,
					Event: &pb.ServerEvent_Move{
						Move: clientEvent.GetMove(),
					},
				}

			case *pb.ClientEvent_Connect:
				sub.PlayerID = clientEvent.PlayerUuid
				serverEvent = &pb.ServerEvent{
					PlayerId: clientEvent.PlayerUuid,
					Event:    &pb.ServerEvent_AddPlayer{},
				}
			}

			if serverEvent != nil {
				s.publish(serverEvent)
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

func Run() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	address := "localhost:4242"
	if len(os.Args) > 1 {
		address = os.Args[1]
	}
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
}
