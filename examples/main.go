package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"actshad.dev/modbus"
)

const (
	rtuDevice = "/tmp/virtualClient"
)

func main() {
	handler := modbus.NewRTUClientHandler(rtuDevice)
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.Timeout = 1 * time.Second
	handler.SlaveId = 2
	handler.Logger = log.New(os.Stdout, "rtu: ", log.LstdFlags)
	err := handler.Connect()
	if err != nil {
		fmt.Println("error connecting to serial")
		return
	}
	defer handler.Close()

	client := modbus.NewClient(handler)

	value := []uint16{0x55, 0xAA}
	err = client.WriteFileRecord(1, 1, value, 2)
	if err != nil {
		fmt.Printf("error writing file record: %v", err.Error())
		return
	}
}
