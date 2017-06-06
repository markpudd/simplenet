package main

import (
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"github.com/markpudd/simplenet/simplenet"
	"log"
	"time"
)

func main() {
	simplenetcore := simplenet.NewSimpleNetCore()

	options := serial.OpenOptions{
		PortName:        "/dev/ttyUSB0",
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	go simplenetcore.SimpleNetCoreLoop(port)
	// Make sure to close it later.
	defer port.Close()
	fmt.Printf("Running\n")
	for true {
		/*	device := &simplenetcore.devices[1]
			recdata := make([]byte, 1024, 1024)
			n, _ := device.Read(recdata)
			if n > 0 {
				fmt.Printf("Recieved data %d\n",n)
			}*/
		time.Sleep(100 * time.Millisecond)
	}
}

/*
func main() {
    readBuf := make([]byte,3,256);
    // Set up options.
    options := serial.OpenOptions{
      PortName: "/dev/cu.wchusbserial1420",
      BaudRate: 9600,
      DataBits: 8,
      StopBits: 1,
      MinimumReadSize: 4,
    }

    // Open the port.
    port, err := serial.Open(options)
    if err != nil {
      log.Fatalf("serial.Open: %v", err)
    }

    // Make sure to close it later.
    defer port.Close()

    // Write 4 bytes to the port.
  //  var  x,y byte
    b := []byte{0xFF,0x02,0x03,'1','2','2',0x00}

    _, _ = port.Write(b)
            fmt.Printf("Reading \n")
    n,err :=port.Read(readBuf);
        fmt.Printf("Read %d\n",n)
    time.Sleep(1000000000)
    n,err =port.Read(readBuf);
    if err != io.EOF {
    fmt.Println("Error ",err)
  }


    n,_ =port.Read(readBuf);
    fmt.Printf("Read %d\n",n)



}
*/
