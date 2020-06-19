// Copyright 2014 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package modbus

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// TLSClientHandler implements Packager and Transporter interface.
type TLSClientHandler struct {
	tcpPackager
	tlsTransporter
}

// NewTLSClientHandler allocates a new TLSClientHandler.
func NewTLSClientHandler(address, key, cert string, insecure bool) *TLSClientHandler {
	h := &TLSClientHandler{}

	h.Address = address
	h.Timeout = tcpTimeout
	h.IdleTimeout = tcpIdleTimeout

	h.key = key
	h.crt = cert
	h.insecure = insecure

	return h
}

// TLSClient creates TLS client with default handler and given connect string.
func TLSClient(address, key, cert string, insecure bool) Client {
	handler := NewTLSClientHandler(address, key, cert, insecure)
	return NewClient(handler)
}

// tlsTransporter implements Transporter interface.
type tlsTransporter struct {
	// Connect string
	Address string
	// Connect & Read timeout
	Timeout time.Duration
	// Idle timeout to close the connection
	IdleTimeout time.Duration
	// Transmission logger
	Logger *log.Logger

	// TCP connection
	mu           sync.Mutex
	conn         net.Conn
	closeTimer   *time.Timer
	lastActivity time.Time

	key, crt string
	insecure bool
}

// Send sends data to server and ensures response length is greater than header length.
func (mb *tlsTransporter) Send(aduRequest []byte) (aduResponse []byte, err error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	// Establish a new connection if not connected
	if err = mb.connect(); err != nil {
		return
	}
	// Set timer to close when idle
	mb.lastActivity = time.Now()
	mb.startCloseTimer()
	// Set write and read timeout
	var timeout time.Time
	if mb.Timeout > 0 {
		timeout = mb.lastActivity.Add(mb.Timeout)
	}
	if err = mb.conn.SetDeadline(timeout); err != nil {
		return
	}
	// Send data
	mb.logf("modbus: sending % x", aduRequest)
	if _, err = mb.conn.Write(aduRequest); err != nil {
		return
	}
	// Read header first
	var data [tcpMaxLength]byte
	if _, err = io.ReadFull(mb.conn, data[:tcpHeaderSize]); err != nil {
		return
	}
	// Read length, ignore transaction & protocol id (4 bytes)
	length := int(binary.BigEndian.Uint16(data[4:]))
	if length <= 0 {
		mb.flush(data[:])
		err = fmt.Errorf("modbus: length in response header '%v' must not be zero", length)
		return
	}
	if length > (tcpMaxLength - (tcpHeaderSize - 1)) {
		mb.flush(data[:])
		err = fmt.Errorf("modbus: length in response header '%v' must not greater than '%v'", length, tcpMaxLength-tcpHeaderSize+1)
		return
	}
	// Skip unit id
	length += tcpHeaderSize - 1
	if _, err = io.ReadFull(mb.conn, data[tcpHeaderSize:length]); err != nil {
		return
	}
	aduResponse = data[:length]
	mb.logf("modbus: received % x\n", aduResponse)
	return
}

// Connect establishes a new TLS connection to the address in Address.
// Connect and Close are exported so that multiple requests can be done with one session
func (mb *tlsTransporter) Connect() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	return mb.connect()
}

func (mb *tlsTransporter) connect() error {
	if mb.conn == nil {
		crt, err := tls.LoadX509KeyPair(mb.crt, mb.key)
		if err != nil {
			return err
		}

		config := &tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: mb.insecure,
		}

		dialer := &net.Dialer{Timeout: mb.Timeout}

		conn, err := tls.DialWithDialer(dialer, "tcp", mb.Address, config)
		if err != nil {
			return err
		}

		mb.conn = conn
	}

	return nil
}

func (mb *tlsTransporter) startCloseTimer() {
	if mb.IdleTimeout <= 0 {
		return
	}
	if mb.closeTimer == nil {
		mb.closeTimer = time.AfterFunc(mb.IdleTimeout, mb.closeIdle)
	} else {
		mb.closeTimer.Reset(mb.IdleTimeout)
	}
}

// Close closes current connection.
func (mb *tlsTransporter) Close() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	return mb.close()
}

// flush flushes pending data in the connection,
// returns io.EOF if connection is closed.
func (mb *tlsTransporter) flush(b []byte) (err error) {
	if err = mb.conn.SetReadDeadline(time.Now()); err != nil {
		return
	}
	// Timeout setting will be reset when reading
	if _, err = mb.conn.Read(b); err != nil {
		// Ignore timeout error
		if netError, ok := err.(net.Error); ok && netError.Timeout() {
			err = nil
		}
	}
	return
}

func (mb *tlsTransporter) logf(format string, v ...interface{}) {
	if mb.Logger != nil {
		mb.Logger.Printf(format, v...)
	}
}

// closeLocked closes current connection. Caller must hold the mutex before calling this method.
func (mb *tlsTransporter) close() (err error) {
	if mb.conn != nil {
		err = mb.conn.Close()
		mb.conn = nil
	}
	return
}

// closeIdle closes the connection if last activity is passed behind IdleTimeout.
func (mb *tlsTransporter) closeIdle() {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.IdleTimeout <= 0 {
		return
	}
	idle := time.Now().Sub(mb.lastActivity)
	if idle >= mb.IdleTimeout {
		mb.logf("modbus: closing connection due to idle timeout: %v", idle)
		mb.close()
	}
}
