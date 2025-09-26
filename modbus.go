// Copyright 2014 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

/*
Package modbus provides a client for MODBUS TCP and RTU/ASCII.
*/
package modbus

import (
	"fmt"
)

const (
	// Bit access
	FuncCodeReadDiscreteInputs = 2  // 0x02
	FuncCodeReadCoils          = 1  // 0x01
	FuncCodeWriteSingleCoil    = 5  // 0x05
	FuncCodeWriteMultipleCoils = 15 // 0x0F

	// 16-bit access
	FuncCodeReadInputRegisters         = 4  // 0x04
	FuncCodeReadHoldingRegisters       = 3  // 0x03
	FuncCodeWriteSingleRegister        = 6  // 0x06
	FuncCodeWriteMultipleRegisters     = 16 // 0x10
	FuncCodeWriteFileRecord            = 21 // 0x15
	FuncCodeMaskWriteRegister          = 22 // 0x16
	FuncCodeReadWriteMultipleRegisters = 23 // 0x17
	FuncCodeReadFIFOQueue              = 24 // 0x18
	FuncCodeReadDeviceIdentification   = 43 // 0x2B
)

const (
	ExceptionCodeIllegalFunction                    = 1
	ExceptionCodeIllegalDataAddress                 = 2
	ExceptionCodeIllegalDataValue                   = 3
	ExceptionCodeServerDeviceFailure                = 4
	ExceptionCodeAcknowledge                        = 5
	ExceptionCodeServerDeviceBusy                   = 6
	ExceptionCodeMemoryParityError                  = 8
	ExceptionCodeGatewayPathUnavailable             = 10
	ExceptionCodeGatewayTargetDeviceFailedToRespond = 11
)

// ModbusError implements error interface.
type ModbusError struct {
	FunctionCode  byte
	ExceptionCode byte
}

// Error converts known modbus exception code to error message.
func (e *ModbusError) Error() string {
	var name string
	switch e.ExceptionCode {
	case ExceptionCodeIllegalFunction:
		name = "illegal function"
	case ExceptionCodeIllegalDataAddress:
		name = "illegal data address"
	case ExceptionCodeIllegalDataValue:
		name = "illegal data value"
	case ExceptionCodeServerDeviceFailure:
		name = "server device failure"
	case ExceptionCodeAcknowledge:
		name = "acknowledge"
	case ExceptionCodeServerDeviceBusy:
		name = "server device busy"
	case ExceptionCodeMemoryParityError:
		name = "memory parity error"
	case ExceptionCodeGatewayPathUnavailable:
		name = "gateway path unavailable"
	case ExceptionCodeGatewayTargetDeviceFailedToRespond:
		name = "gateway target device failed to respond"
	default:
		name = "unknown"
	}
	return fmt.Sprintf("modbus: exception '%v' (%s), function '%v'", e.ExceptionCode, name, e.FunctionCode)
}

// ProtocolDataUnit (PDU) is independent of underlying communication layers.
type ProtocolDataUnit struct {
	FunctionCode byte
	Data         []byte
}

// Structures that wrap the Read Device Identification Message reponses
// BasicDeviceID contains the Basic objects (0x00–0x02)
type BasicDeviceID struct {
	VendorName        []byte // 0x00
	ProductCode       []byte // 0x01
	MajorMinorVersion []byte // 0x02
}

// RegularDeviceID contains Basic + Regular objects (0x03–0x06)
type RegularDeviceID struct {
	Basic BasicDeviceID

	VendorURL           []byte // 0x03
	ProductName         []byte // 0x04
	ModelName           []byte // 0x05
	UserApplicationName []byte // 0x06
}

// ExtendedDeviceID contains Regular + Extended objects (all supported IDs)
type ExtendedDeviceID struct {
	Regular RegularDeviceID

	// Extended objects (0x07+) as map for flexibility
	ExtendedObjects map[uint8][]byte
}

// Packager specifies the communication layer.
type Packager interface {
	Encode(pdu *ProtocolDataUnit) (adu []byte, err error)
	Decode(adu []byte) (pdu *ProtocolDataUnit, err error)
	Verify(aduRequest []byte, aduResponse []byte) (err error)
}

// Transporter specifies the transport layer.
type Transporter interface {
	Send(aduRequest []byte) (aduResponse []byte, err error)
}
