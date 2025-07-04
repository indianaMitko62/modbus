// Copyright 2014 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package modbus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// ClientHandler is the interface that groups the Packager and Transporter methods.
type ClientHandler interface {
	Packager
	Transporter
}

type client struct {
	packager    Packager
	transporter Transporter
}

// NewClient creates a new modbus client with given backend handler.
func NewClient(handler ClientHandler) Client {
	return &client{packager: handler, transporter: handler}
}

// NewClient2 creates a new modbus client with given backend packager and transporter.
func NewClient2(packager Packager, transporter Transporter) Client {
	return &client{packager: packager, transporter: transporter}
}

// Request:
//
//	Function code         : 1 byte (0x01)
//	Starting address      : 2 bytes
//	Quantity of coils     : 2 bytes
//
// Response:
//
//	Function code         : 1 byte (0x01)
//	Byte count            : 1 byte
//	Coil status           : N* bytes (=N or N+1)
func (mb *client) ReadCoils(address, quantity uint16) (results []byte, err error) {
	if quantity < 1 || quantity > 2000 {
		err = fmt.Errorf("modbus: quantity '%v' must be between '%v' and '%v',", quantity, 1, 2000)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadCoils,
		Data:         dataBlock(address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	length := len(response.Data) - 1
	if count != length {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", length, count)
		return
	}
	results = response.Data[1:]
	return
}

// Request:
//
//	Function code         : 1 byte (0x02)
//	Starting address      : 2 bytes
//	Quantity of inputs    : 2 bytes
//
// Response:
//
//	Function code         : 1 byte (0x02)
//	Byte count            : 1 byte
//	Input status          : N* bytes (=N or N+1)
func (mb *client) ReadDiscreteInputs(address, quantity uint16) (results []byte, err error) {
	if quantity < 1 || quantity > 2000 {
		err = fmt.Errorf("modbus: quantity '%v' must be between '%v' and '%v',", quantity, 1, 2000)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadDiscreteInputs,
		Data:         dataBlock(address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	length := len(response.Data) - 1
	if count != length {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", length, count)
		return
	}
	results = response.Data[1:]
	return
}

// Request:
//
//	Function code         : 1 byte (0x03)
//	Starting address      : 2 bytes
//	Quantity of registers : 2 bytes
//
// Response:
//
//	Function code         : 1 byte (0x03)
//	Byte count            : 1 byte
//	Register value        : Nx2 bytes
func (mb *client) ReadHoldingRegisters(address, quantity uint16) (results []byte, err error) {
	if quantity < 1 || quantity > 125 {
		err = fmt.Errorf("modbus: quantity '%v' must be between '%v' and '%v',", quantity, 1, 125)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadHoldingRegisters,
		Data:         dataBlock(address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	length := len(response.Data) - 1
	if count != length {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", length, count)
		return
	}
	results = response.Data[1:]
	return
}

// Request:
//
//	Function code         : 1 byte (0x04)
//	Starting address      : 2 bytes
//	Quantity of registers : 2 bytes
//
// Response:
//
//	Function code         : 1 byte (0x04)
//	Byte count            : 1 byte
//	Input registers       : N bytes
func (mb *client) ReadInputRegisters(address, quantity uint16) (results []byte, err error) {
	if quantity < 1 || quantity > 125 {
		err = fmt.Errorf("modbus: quantity '%v' must be between '%v' and '%v',", quantity, 1, 125)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadInputRegisters,
		Data:         dataBlock(address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	length := len(response.Data) - 1
	if count != length {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", length, count)
		return
	}
	results = response.Data[1:]
	return
}

// Request:
//
//	Function code         : 1 byte (0x05)
//	Output address        : 2 bytes
//	Output value          : 2 bytes
//
// Response:
//
//	Function code         : 1 byte (0x05)
//	Output address        : 2 bytes
//	Output value          : 2 bytes
func (mb *client) WriteSingleCoil(address, value uint16) (results []byte, err error) {
	// The requested ON/OFF state can only be 0xFF00 and 0x0000
	if value != 0xFF00 && value != 0x0000 {
		err = fmt.Errorf("modbus: state '%v' must be either 0xFF00 (ON) or 0x0000 (OFF)", value)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteSingleCoil,
		Data:         dataBlock(address, value),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = fmt.Errorf("modbus: response data size '%v' does not match expected '%v'", len(response.Data), 4)
		return
	}
	respValue := binary.BigEndian.Uint16(response.Data)
	if address != respValue {
		err = fmt.Errorf("modbus: response address '%v' does not match request '%v'", respValue, address)
		return
	}
	results = response.Data[2:]
	respValue = binary.BigEndian.Uint16(results)
	if value != respValue {
		err = fmt.Errorf("modbus: response value '%v' does not match request '%v'", respValue, value)
		return
	}
	return
}

// Request:
//
//	Function code         : 1 byte (0x06)
//	Register address      : 2 bytes
//	Register value        : 2 bytes
//
// Response:
//
//	Function code         : 1 byte (0x06)
//	Register address      : 2 bytes
//	Register value        : 2 bytes
func (mb *client) WriteSingleRegister(address, value uint16) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteSingleRegister,
		Data:         dataBlock(address, value),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = fmt.Errorf("modbus: response data size '%v' does not match expected '%v'", len(response.Data), 4)
		return
	}
	respValue := binary.BigEndian.Uint16(response.Data)
	if address != respValue {
		err = fmt.Errorf("modbus: response address '%v' does not match request '%v'", respValue, address)
		return
	}
	results = response.Data[2:]
	respValue = binary.BigEndian.Uint16(results)
	if value != respValue {
		err = fmt.Errorf("modbus: response value '%v' does not match request '%v'", respValue, value)
		return
	}
	return
}

// Request:
//
//	Function code         : 1 byte (0x0F)
//	Starting address      : 2 bytes
//	Quantity of outputs   : 2 bytes
//	Byte count            : 1 byte
//	Outputs value         : N* bytes
//
// Response:
//
//	Function code         : 1 byte (0x0F)
//	Starting address      : 2 bytes
//	Quantity of outputs   : 2 bytes
func (mb *client) WriteMultipleCoils(address, quantity uint16, value []byte) (results []byte, err error) {
	if quantity < 1 || quantity > 1968 {
		err = fmt.Errorf("modbus: quantity '%v' must be between '%v' and '%v',", quantity, 1, 1968)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteMultipleCoils,
		Data:         dataBlockSuffix(value, address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = fmt.Errorf("modbus: response data size '%v' does not match expected '%v'", len(response.Data), 4)
		return
	}
	respValue := binary.BigEndian.Uint16(response.Data)
	if address != respValue {
		err = fmt.Errorf("modbus: response address '%v' does not match request '%v'", respValue, address)
		return
	}
	results = response.Data[2:]
	respValue = binary.BigEndian.Uint16(results)
	if quantity != respValue {
		err = fmt.Errorf("modbus: response quantity '%v' does not match request '%v'", respValue, quantity)
		return
	}
	return
}

// Request:
//
//	Function code         : 1 byte (0x10)
//	Starting address      : 2 bytes
//	Quantity of outputs   : 2 bytes
//	Byte count            : 1 byte
//	Registers value       : N* bytes
//
// Response:
//
//	Function code         : 1 byte (0x10)
//	Starting address      : 2 bytes
//	Quantity of registers : 2 bytes
func (mb *client) WriteMultipleRegisters(address, quantity uint16, value []byte) (results []byte, err error) {
	if quantity < 1 || quantity > 123 {
		err = fmt.Errorf("modbus: quantity '%v' must be between '%v' and '%v',", quantity, 1, 123)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteMultipleRegisters,
		Data:         dataBlockSuffix(value, address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = fmt.Errorf("modbus: response data size '%v' does not match expected '%v'", len(response.Data), 4)
		return
	}
	respValue := binary.BigEndian.Uint16(response.Data)
	if address != respValue {
		err = fmt.Errorf("modbus: response address '%v' does not match request '%v'", respValue, address)
		return
	}
	results = response.Data[2:]
	respValue = binary.BigEndian.Uint16(results)
	if quantity != respValue {
		err = fmt.Errorf("modbus: response quantity '%v' does not match request '%v'", respValue, quantity)
		return
	}
	return
}

// Request:
//
//	Function code			: 1 byte (0x2B)
//	MEI Type				: 1 byte
//	Read Device ID code		: 1 byte
//	Object Id				: 1 byte
//
// Response:
// 	Function code 			: 1 byte (0x2B)
// 	MEI Type 				: 1 byte (0x0E)
// 	Read Device ID code 	: 1 byte
// 	Conformity level 		: 1 byte
// 	More Follows  			: 1 byte
// 	Next Object Id  		: 1 byte
// 	Number of objects  		: 1 byte
// 	List Of
// 			Object ID  		: 1 byte
// 			Object length  	: 1 byte
// 			Object Value 	: Object length

func (mb *client) readDeviceIdentification(object_id uint8, readDeviceIDCode uint8) (vendorName string, productCode string, majorMinorVersion string, err error) {
	var meiType uint8 = 0x0E
	data := []byte{meiType, readDeviceIDCode, object_id}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadDeviceIdentification,
		Data:         data,
	}
	response, err := mb.send(&request)
	if err != nil {
		return // add error message
	}

	r := bytes.NewReader(response.Data)

	respMeiType, err := r.ReadByte()
	if err != nil {
		return
	}
	if meiType != uint8(respMeiType) {
		err = fmt.Errorf("modbus: response mei type '%v' does not match request '%v'", respMeiType, meiType)
		return
	}

	respDeviceIDCode, err := r.ReadByte()
	if err != nil {
		return
	}
	if readDeviceIDCode != uint8(respDeviceIDCode) {
		err = fmt.Errorf("modbus: response device ID code '%v' does not match request '%v'", respDeviceIDCode, readDeviceIDCode)
		return
	}

	respConformityLevel, err := r.ReadByte()
	if err != nil {
		return
	}
	if respConformityLevel&0x01 > 3 {
		err = fmt.Errorf("modbus: invalid response conformity level '%v'", respConformityLevel)
		return
	}

	moreFollows, err := r.ReadByte()
	if err != nil {
		return
	}
	if moreFollows != 0 && moreFollows != 0xFF {
		err = fmt.Errorf("modbus: invalid response more follows flag '%v'", moreFollows)
		return
	}

	nextObjectID, err := r.ReadByte()
	if err != nil {
		return
	}
	numberOfObjects, err := r.ReadByte()
	if err != nil {
		return
	}
	if nextObjectID != 0 {
		err = fmt.Errorf("modbus: currently not supporting multi-transaction responses. Received first '%v' objects", numberOfObjects)
	}

	var objID uint8
	var objLen uint8

	for i := 0; i < int(numberOfObjects); i++ {
		objID, err = r.ReadByte()
		if err != nil {
			return
		}
		objLen, err = r.ReadByte()
		if err != nil {
			return
		}
		switch objID {
		case 0x00:
			vendorNameBytes := make([]byte, objLen)
			_, err = io.ReadFull(r, vendorNameBytes)
			if err != nil {
				return
			}
			vendorName = string(vendorNameBytes)
		case 0x01:
			vendorPorductCodeBytes := make([]byte, objLen)
			_, err = io.ReadFull(r, vendorPorductCodeBytes)
			if err != nil {
				return
			}
			productCode = string(vendorPorductCodeBytes)
		case 0x02:
			majorMinorVersionBytes := make([]byte, objLen)
			_, err = io.ReadFull(r, majorMinorVersionBytes)
			if err != nil {
				return
			}
			majorMinorVersion = string(majorMinorVersionBytes)
		}
	}

	return
}

func (mb *client) ReadDeviceIdentificationBasic() (vendorName string, productCode string, majorMinorVersion string, err error) {
	return mb.readDeviceIdentification(0, 0x01)
}

func (mb *client) ReadDeviceIdentificationSpecific(object_id uint8) (vendorName string, productCode string, majorMinorVersion string, err error) {
	return mb.readDeviceIdentification(object_id, 0x04)
}

// Request:
//
//	Function code         : 1 byte (0x15)
//	Request data length   : 1 byte
//Subrequest 1:
//  Reference type:       : 1 byte
//	File number           : 2 bytes (0x0001 to 0xFFFF)
//	File record           : 2 bytes (0x0000 to 0x270F)
// 	Record length         : 2 bytes (N count of 16-bit registers)
//  Record data:          : 2 * N bytes
//Subrequest 2:
//  ...
//
// Response:
//  The normal response is an echo of the request.

func (mb *client) WriteFileRecord(fileNumber uint16, recordNumber uint16, value []uint16, count uint16) (err error) {
	if fileNumber == 0x0000 {
		return fmt.Errorf("modbus: invalid file number: %v", fileNumber)
	}
	if recordNumber > 0x270F {
		return fmt.Errorf("modbus: invalid record number: %v", recordNumber)
	}
	if count > 122 {
		return fmt.Errorf("modbus: invalid record count: %v", count)
	}

	dataSize := uint8(count) * 2
	data := []byte{7 + dataSize, 6,
		byte((fileNumber >> 8) & 0xFF), byte(fileNumber & 0xFF),
		byte((recordNumber >> 8) & 0xFF), byte(recordNumber & 0xFF),
		0x00, uint8(count)}

	recordBytes := make([]byte, dataSize)
	putRegs(recordBytes, value, count)
	data = append(data, recordBytes...)

	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteFileRecord,
		Data:         data,
	}

	response, err := mb.send(&request)
	if err != nil {
		return
	}

	r := bytes.NewReader(response.Data)

	responseSize, err := r.ReadByte()
	if err != nil {
		return
	}
	if responseSize > 251 {
		err = fmt.Errorf("modbus: response size invalid: %v", responseSize)
		return
	}

	respReferenceType, err := r.ReadByte()
	if err != nil {
		return
	}
	if respReferenceType != 6 {
		err = fmt.Errorf("modbus: response reference type invalid: %v", respReferenceType)
		return
	}

	var buf [2]byte
	_, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}

	respFileNumber := binary.BigEndian.Uint16(buf[:])
	if respFileNumber != fileNumber {
		err = fmt.Errorf("modbus: response file number invalid: %v", respFileNumber)
		return
	}

	_, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}

	respRecordNumber := binary.BigEndian.Uint16(buf[:])
	if respRecordNumber != recordNumber {
		err = fmt.Errorf("modbus: response record number invalid: %v", respRecordNumber)
		return
	}

	_, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}

	respRecordLength := binary.BigEndian.Uint16(buf[:])
	if respRecordLength != count {
		err = fmt.Errorf("modbus: response record length invalid: %v", respRecordLength)
		return
	}

	responseDataSize := respRecordLength * 2
	responseRecordDataBytes := make([]byte, responseDataSize)

	_, err = io.ReadFull(r, responseRecordDataBytes)
	if err != nil {
		return
	}

	responseRecordDataUint16 := bytesToUint16s(responseRecordDataBytes, binary.LittleEndian)
	if !equalUint16Slices(value, responseRecordDataUint16) {
		err = fmt.Errorf("modbus: request and response file record does not match")
		return
	}

	return nil
}

// Request:
//
//	Function code         : 1 byte (0x16)
//	Reference address     : 2 bytes
//	AND-mask              : 2 bytes
//	OR-mask               : 2 bytes
//
// Response:
//
//	Function code         : 1 byte (0x16)
//	Reference address     : 2 bytes
//	AND-mask              : 2 bytes
//	OR-mask               : 2 bytes
func (mb *client) MaskWriteRegister(address, andMask, orMask uint16) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeMaskWriteRegister,
		Data:         dataBlock(address, andMask, orMask),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 6 {
		err = fmt.Errorf("modbus: response data size '%v' does not match expected '%v'", len(response.Data), 6)
		return
	}
	respValue := binary.BigEndian.Uint16(response.Data)
	if address != respValue {
		err = fmt.Errorf("modbus: response address '%v' does not match request '%v'", respValue, address)
		return
	}
	respValue = binary.BigEndian.Uint16(response.Data[2:])
	if andMask != respValue {
		err = fmt.Errorf("modbus: response AND-mask '%v' does not match request '%v'", respValue, andMask)
		return
	}
	respValue = binary.BigEndian.Uint16(response.Data[4:])
	if orMask != respValue {
		err = fmt.Errorf("modbus: response OR-mask '%v' does not match request '%v'", respValue, orMask)
		return
	}
	results = response.Data[2:]
	return
}

// Request:
//
//	Function code         : 1 byte (0x17)
//	Read starting address : 2 bytes
//	Quantity to read      : 2 bytes
//	Write starting address: 2 bytes
//	Quantity to write     : 2 bytes
//	Write byte count      : 1 byte
//	Write registers value : N* bytes
//
// Response:
//
//	Function code         : 1 byte (0x17)
//	Byte count            : 1 byte
//	Read registers value  : Nx2 bytes
func (mb *client) ReadWriteMultipleRegisters(readAddress, readQuantity, writeAddress, writeQuantity uint16, value []byte) (results []byte, err error) {
	if readQuantity < 1 || readQuantity > 125 {
		err = fmt.Errorf("modbus: quantity to read '%v' must be between '%v' and '%v',", readQuantity, 1, 125)
		return
	}
	if writeQuantity < 1 || writeQuantity > 121 {
		err = fmt.Errorf("modbus: quantity to write '%v' must be between '%v' and '%v',", writeQuantity, 1, 121)
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadWriteMultipleRegisters,
		Data:         dataBlockSuffix(value, readAddress, readQuantity, writeAddress, writeQuantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	if count != (len(response.Data) - 1) {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", len(response.Data)-1, count)
		return
	}
	results = response.Data[1:]
	return
}

// Request:
//
//	Function code         : 1 byte (0x18)
//	FIFO pointer address  : 2 bytes
//
// Response:
//
//	Function code         : 1 byte (0x18)
//	Byte count            : 2 bytes
//	FIFO count            : 2 bytes
//	FIFO count            : 2 bytes (<=31)
//	FIFO value register   : Nx2 bytes
func (mb *client) ReadFIFOQueue(address uint16) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadFIFOQueue,
		Data:         dataBlock(address),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	if len(response.Data) < 4 {
		err = fmt.Errorf("modbus: response data size '%v' is less than expected '%v'", len(response.Data), 4)
		return
	}
	count := int(binary.BigEndian.Uint16(response.Data))
	if count != (len(response.Data) - 1) {
		err = fmt.Errorf("modbus: response data size '%v' does not match count '%v'", len(response.Data)-1, count)
		return
	}
	count = int(binary.BigEndian.Uint16(response.Data[2:]))
	if count > 31 {
		err = fmt.Errorf("modbus: fifo count '%v' is greater than expected '%v'", count, 31)
		return
	}
	results = response.Data[4:]
	return
}

// Helpers

// send sends request and checks possible exception in the response.
func (mb *client) send(request *ProtocolDataUnit) (response *ProtocolDataUnit, err error) {
	aduRequest, err := mb.packager.Encode(request)
	if err != nil {
		return
	}
	aduResponse, err := mb.transporter.Send(aduRequest)
	if err != nil {
		return
	}
	if err = mb.packager.Verify(aduRequest, aduResponse); err != nil {
		return
	}
	response, err = mb.packager.Decode(aduResponse)
	if err != nil {
		return
	}
	// Check correct function code returned (exception)
	if response.FunctionCode != request.FunctionCode {
		err = responseError(response)
		return
	}
	if response.Data == nil || len(response.Data) == 0 {
		// Empty response
		err = fmt.Errorf("modbus: response data is empty")
		return
	}
	return
}

// dataBlock creates a sequence of uint16 data.
func dataBlock(value ...uint16) []byte {
	data := make([]byte, 2*len(value))
	for i, v := range value {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	return data
}

// dataBlockSuffix creates a sequence of uint16 data and append the suffix plus its length.
func dataBlockSuffix(suffix []byte, value ...uint16) []byte {
	length := 2 * len(value)
	data := make([]byte, length+1+len(suffix))
	for i, v := range value {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	data[length] = uint8(len(suffix))
	copy(data[length+1:], suffix)
	return data
}

func responseError(response *ProtocolDataUnit) error {
	mbError := &ModbusError{FunctionCode: response.FunctionCode}
	if response.Data != nil && len(response.Data) > 0 {
		mbError.ExceptionCode = response.Data[0]
	}
	return mbError
}

func putRegs(buf []byte, data []uint16, n uint16) {
	for i := 0; i < int(n); i++ {
		if 2*i+1 >= len(buf) {
			break
		}
		val := data[i]
		buf[2*i] = byte(val >> 8)
		buf[2*i+1] = byte(val & 0x00FF)
		// buf[2*i] = byte(val & 0x00FF) // low byte
		// buf[2*i+1] = byte(val >> 8)   // high byte
	}
}

func bytesToUint16s(b []byte, order binary.ByteOrder) []uint16 {
	n := len(b) / 2
	out := make([]uint16, n)
	for i := 0; i < n; i++ {
		out[i] = order.Uint16(b[2*i : 2*i+2])
	}
	return out
}

func equalUint16Slices(a, b []uint16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
