package main

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/pkg/errors"
)

func (p *PBFT) dialRPCToPeer(peerID int) error {
	if peerID == p.id {
		return nil
	}
	client, err := rpc.Dial("tcp", p.peerIPPort[peerID])
	if err != nil {
		logMsg := fmt.Sprintf("Failed to connect to peer %d at %s: %v", peerID, p.peerIPPort[peerID], err)
		p.logPut(logMsg, PURPLE)
		return errors.WithStack(err)
	}
	p.mu.Lock()
	p.rpcConns[peerID] = client
	p.mu.Unlock()
	msg := fmt.Sprintf("Connected to peer %d at %s", peerID, p.peerIPPort[peerID])
	p.logPut(msg, GREEN)
	return nil
}

func (p *PBFT) dialRPCToAllPeers() error {
	for peerID := range p.peerIPPort {
		if peerID != p.id {
			logMsg := fmt.Sprintf("Dialing RPC to peer %d at %s", peerID, p.peerIPPort[peerID])
			p.logPut(logMsg, CYAN)
			go p.dialRPCToPeer(peerID)
		}
	}
	return nil
}

func (p *PBFT) listenRPC() error {
	_ = rpc.Register(p)
	l, err := net.Listen("tcp", p.peerIPPort[p.id])
	if err != nil {
		return errors.WithStack(err)
	}
	msg := fmt.Sprintf("Listening for RPC connections on %s", p.peerIPPort[p.id])
	p.logPut(msg, PURPLE)
	for {
		conn, err := l.Accept()
		if err != nil {
			logMsg := fmt.Sprintf("Failed to accept RPC connection: %v", err)
			p.logPut(logMsg, PURPLE)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
