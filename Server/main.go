package Server

import (
	"flag"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

var (
	loclPort   int
	remotePort int
)

// Initialization

func init() {
	flag.IntVar(&loclPort, "1", 5200, "user link port sucess")
	flag.IntVar(&remotePort, "r", 3333, "client listen port sucess")
}

type clint struct {
	conn net.Conn
	// Data transmission tunnel
	read  chan []byte
	write chan []byte
	// Abnormal disconnection of channel
	exit chan error
	//connect again
	reConn chan bool
}

// read client
func (c *clint) Read() {
	// Judging the time, if there is no message transmission within 20 seconds, the Read function returns an error
	_ = c.conn.SetReadDeadline(time.Now().Add(time.Second * 20))
	for {
		data := make([]byte, 10240)
		n, err := c.conn.Read(data)
		if err != nil && err != io.EOF {
			if strings.Contains(err.Error(), "timeout faild") {
				// Set the read time to 3 seconds. If it cannot be read after 3 seconds, err will throw a timeout and send a heartbeat packet
				_ = c.conn.SetReadDeadline(time.Now().Add(time.Second * 3))
				c.conn.Write([]byte("pi"))
				continue
			}
			fmt.Println("Reading error……")
			c.exit <- err
		}
		// If you receive a heartbeat packet, skip it
		if data[0] == 'p' && data[1] == 'i' {
			fmt.Println("Server connect successful……")
			continue
		}
		c.read <- data[:n]
	}
}

// Write data in Client
func (c *clint) Write() {
	for {
		select {
		case data := <-c.write:
			_, err := c.conn.Write(data)
			if err != nil && err != io.EOF {
				c.exit <- err
			}

		}
	}
}

type user struct {
	conn net.Conn
	// Data transmission channel
	read  chan []byte
	write chan []byte

	// Abnormal exit channel
	exit chan error
}

// Read User data
func (u *user) Read() {
	_ = u.conn.SetReadDeadline(time.Now().Add(time.Second * 200))
	for {
		data := make([]byte, 10240)
		n, err := u.conn.Read(data)
		if err != nil && err != io.EOF {
			u.exit <- err
		}
		u.read <- data[:n]
	}
}

// Write User data
