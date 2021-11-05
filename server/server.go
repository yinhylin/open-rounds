package main

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
	msgs chan []byte
	c    *websocket.Conn
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
	messageType, reader, err := c.Reader(ctx)
	if err != nil {
		return err
	}
	if messageType != websocket.MessageBinary {
		return fmt.Errorf("unexpected message type: %v", messageType)
	}
	sub := &subscriber{
		msgs: make(chan []byte, 1024),
		c:    c,
	}
	s.addSubscriber(sub)

	defer s.removeSubscriber(sub)

	// this is the magic event loop
	go func() {
		for {
			b, err := ioutil.ReadAll(reader)
			if err != nil {
				log.Println(err)
				// TODO: close connection
				return
			}
			if len(b) <= 0 {
				continue
			}
			var clientEvent pb.ClientEvent
			err = proto.Unmarshal(b, &clientEvent)
			if err != nil {
				// TODO: close connection
				log.Println(err)
				return
			}

			// TODO: we should do movement vectors and validation
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
			}

			if serverEvent != nil {
				bytes, err := proto.Marshal(serverEvent)
				if err != nil {
					log.Println(err)
					return
				}
				s.publish(bytes)
			}

			// TODO: better handling of server events in a separate area.
			log.Println(clientEvent.String())

			_, reader, err = c.Reader(ctx)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}()
	for {
		select {
		case msg := <-sub.msgs:
			c.Write(ctx, websocket.MessageBinary, msg)
		case <-ctx.Done():
			fmt.Println(":(")
			return ctx.Err()
		}
	}
}

func (s *Server) publish(msg []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for sub := range s.subscribers {
		select {
		case sub.msgs <- msg:
		default:
			sub.c.Close(websocket.StatusPolicyViolation, "write would block")
		}
	}
}

func main() {
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
