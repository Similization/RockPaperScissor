package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

type Message struct {
	from    net.Addr
	payload []byte
}

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan Message
	peerMap    map[net.Addr]struct{}
	connch     chan net.Conn
}

func NewServer(listenAddr string, playerCount int) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan Message, 10),
		peerMap:    make(map[net.Addr]struct{}),
		connch:     make(chan net.Conn, playerCount),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	fmt.Println("Server is running...")
	s.ln = ln

	go s.acceptLoop()

	<-s.quitch
	close(s.msgch)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		select {
		case s.connch <- conn:
		default:
			conn.Write([]byte("Room is full. Refusing connection."))
			conn.Close()
			continue
		}

		s.peerMap[conn.RemoteAddr()] = struct{}{}
		fmt.Printf("New connection: %s\n", conn.RemoteAddr().String())

		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer func() {
		delete(s.peerMap, conn.RemoteAddr())
	}()
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client (%v) closed the connection.", conn.RemoteAddr())
			} else {
				fmt.Println("Some errors occured: ", err)
			}
			return
		}

		s.msgch <- Message{
			from:    conn.RemoteAddr(),
			payload: buf[:n],
		}

		conn.Write([]byte("Thank you for message!"))
	}
}

func main() {
	server := NewServer(":3000", 2)

	go func() {
		for msg := range server.msgch {
			fmt.Printf("Received from (%v): %s\n", msg.from, string(msg.payload))
		}
	}()

	log.Fatal(server.Start())
}
