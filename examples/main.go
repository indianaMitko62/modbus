package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"actshad.dev/modbus"
)

const (
	rtuDevice = "/dev/ttyUSB0"
)

func main() {
	handler := modbus.NewRTUClientHandler(rtuDevice)
	handler.BaudRate = 115200
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 2
	handler.Logger = log.New(os.Stdout, "rtu: ", log.LstdFlags)
	err := handler.Connect()
	if err != nil {
		fmt.Println("error connecting to serial")
		return
	}
	defer handler.Close()

	client := modbus.NewClient(handler)

	vendorName, productCode, majorMinorVersion, err := client.ReadDeviceIdentificationBasic()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("vendor name: %s\n", vendorName)
	fmt.Printf("product code: %s\n", productCode)
	fmt.Printf("version: %s\n", majorMinorVersion)

	for i := 0; i < 5; i++ {
		results, err := client.ReadHoldingRegisters(8, 2)
		if err != nil || results == nil {
			fmt.Println(err.Error())
			return
		}
		for i := 0; i < len(results); i++ {
			fmt.Println(results[i])
		}
		time.Sleep(1 * time.Second)
	}
}
