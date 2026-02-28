package main

import "github.com/libp2p/go-libp2p/core/network"

type Listen struct {
	listener bool
	s        network.Stream
}

type ProtocolMessage interface{}

/*
listener:
wait for key
assert key
send index
receive index

for each stale file open stream for transfer
*/

func (p *Protocol) ReadKCV() ([]byte, error) {
	kcv := make([]byte, 3)
	_, err := p.s.Read(kcv)
	return kcv, err
}

func (p *Protocol) SendKCV(kcv []byte) error {
	_, err := p.s.Write(kcv)
	return err
}
