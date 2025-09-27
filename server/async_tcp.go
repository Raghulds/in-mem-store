package server

import (
	"in-mem-store/config"
	"in-mem-store/core"
	"log"
	"net"
	"syscall"
	"time"
)

var con_clients = 0
var CRONFrequencyMs time.Duration = 60

func RunAsyncTcpServer() error {
	log.Println("starting an asynchronous TXP server on", config.Host, config.Port)

	max_clients := 20000
	lastCRONExecutionTime := time.Now()

	// IO ready connections
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_clients)

	// Create Socket for the server
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Println("serverFD", serverFD)
		return err
	}
	defer syscall.Close(serverFD)

	// Allow quick rebinding after restarts
	if err = syscall.SetsockoptInt(serverFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		log.Println("setsockopt SO_REUSEADDR", err)
		return err
	}

	// Set socket to operate in NON-BLOCKING way - TCP non blocking
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		log.Println("serverFD non blocking", err)
		return err
	}

	// Bind socket to Host & Port (ensure IPv4 bytes)
	ip := net.ParseIP(config.Host).To4()
	if ip == nil {
		// Fallback to 0.0.0.0 if parsing failed
		ip = net.IPv4zero
	}
	bindErr := syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Addr: [4]byte{ip[0], ip[1], ip[2], ip[3]},
		Port: config.Port,
	})
	if bindErr != nil {
		log.Println("bindErr", bindErr)
		return bindErr
	}

	// Server listens
	if err = syscall.Listen(serverFD, max_clients); err != nil {
		log.Println("listen", err)
		return err
	}

	// Create EPOLL
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(epollFD)

	var ServerEpollEvent syscall.EpollEvent = syscall.EpollEvent(syscall.EpollEvent{
		Fd:     int32(serverFD),
		Events: syscall.EPOLLIN,
	})

	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &ServerEpollEvent); err != nil {
		log.Println("epoll ctl", err)
		return err
	}

	log.Println("listening", epollFD)
	for {
		if time.Now().After(lastCRONExecutionTime.Add(CRONFrequencyMs)) {
			core.DeleteExpiredKeys()
			lastCRONExecutionTime = time.Now()
		}

		// Check if any FD is ready for IO
		nevents, err := syscall.EpollWait(epollFD, events, 1)
		if err != nil {
			continue
		}

		for i := 0; i < nevents; i++ {
			// Check if it's server FD IO
			if int(events[i].Fd) == serverFD {
				// Accept the connection & add the FD to the EPOLL
				fd, _, err := syscall.Accept(int(events[i].Fd))
				if err != nil {
					continue
				}

				con_clients++
				syscall.SetNonblock(fd, true)

				// Create EPOLL event for this client conn
				var clientEpollEvent syscall.EpollEvent = syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}

				if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &clientEpollEvent); err != nil {
					log.Fatal(err)
				}
			} else {
				// It's incoming IO from one of the FDs of existing IOs in the EPOLL
				comm := core.FDComm{Fd: int(events[i].Fd)}
				cmds, err := readCommands(comm)
				if err != nil {
					syscall.Close(int(events[i].Fd))
					con_clients -= 1
					continue
				}

				respond(cmds, comm)
			}
		}
	}
}
