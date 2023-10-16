package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"
)

var GameDict []string = []string{"paper", "rock", "scissor", "random"}

func checkMessage(message string) bool {
	for _, val := range GameDict {
		if message == val {
			return true
		}
	}
	return false
}

func randomMove() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := rand.Intn(len(GameDict))
	return GameDict[idx]
}

func main() {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		var message string
		fmt.Print("Client: ")
		fmt.Scan(&message)
		for !checkMessage(message) {
			fmt.Printf("You entered wrong word (%s), program accepts only: %v\n", message, GameDict)
			fmt.Print("Client: ")
			fmt.Scan(&message)
		}
		if message == "random" {
			message = randomMove()
		}

		_, err := conn.Write([]byte(message))
		if err != nil {
			opErr, ok := err.(*net.OpError)
			if ok && opErr.Op == "write" && opErr.Err.Error() == "broken pipe" {
				fmt.Println("Server has closed the connection.")
				return
			}
		}

		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Server has closed the connection.")
				return
			} else {
				fmt.Println(err)
			}
		}

		fmt.Printf("Server: %s\n", string(buf[:n]))
	}
}
