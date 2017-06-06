package simplenet

import (
	"errors"
	"fmt"
	"io"
	"time"
)

const MaxDevices int = 15

type SimpleNetCore struct {
	Devices [MaxDevices]Device

	readBuf       []byte
	byteAvailable bool
	waitingOnByte bool
	running       bool
	byteRecieved  chan bool
}

func NewSimpleNetCore() *SimpleNetCore {
	nsc := new(SimpleNetCore)
	nsc.readBuf = make([]byte, 1, 1)
	nsc.byteAvailable = false
	nsc.waitingOnByte = false
	nsc.byteRecieved = make(chan bool)
	return nsc
}

//  Read a byte or time out - Assume blocking i/o?????
func (snc *SimpleNetCore) ReadByteWithDeadline(port io.ReadWriteCloser) (byte, error) {
	/*if snc.byteAvailable {
		snc.byteAvailable = false
		snc.waitingOnByte = false
		return snc.readBuf[0], nil
	}*/

	if !snc.waitingOnByte {
		go func() {
			snc.waitingOnByte = true
			port.Read(snc.readBuf)
			//		snc.byteAvailable = true
			snc.byteRecieved <- true
		}()
	}

	select {
	case <-snc.byteRecieved:
		//	snc.byteAvailable = false
		snc.waitingOnByte = false
		return snc.readBuf[0], nil
	case <-time.After(10 * time.Millisecond):
		return 0, errors.New("Timeout")
	}
}

func (snc *SimpleNetCore) SimpleNetInnerLoop(port io.ReadWriteCloser) {
	var p byte
	for i := 0; i < MaxDevices; i++ {
		// Only process this device if there is enough buffer space
		// to read packet and the there is not any data available
		if snc.Devices[i].ReadLength >= 768 {
			fmt.Printf("Buffer full for device %d\n", i)
			//  This is dubious as we potentaially loose data
			snc.Devices[i].DataAvailable = true
		} else if !snc.Devices[i].DataAvailable {

			port.Write([]byte{byte(0xFF), byte(i)})
			data, _ := snc.Devices[i].GetBytesForWire()
			port.Write([]byte{byte(len(data))})
			//for p := 0; p < len(data); p++ {
			port.Write(data)
			//}
			crc := 0
			port.Write([]byte{byte(crc)})

			var b byte
			var err error
			b, err = snc.ReadByteWithDeadline(port)
			//	port.Read(snc.readBuf)
			//	b = snc.readBuf[0]
			if err == nil {
				if b != 0xff {
					snc.Devices[i].Online = false
				} else {
					//  Assume at this point io is alway
					snc.Devices[i].Online = true
					//	port.Read(snc.readBuf)
					//		l := snc.readBuf[0]
					l, _ := snc.ReadByteWithDeadline(port)
					for p = 0; p < l; p++ {
						//	port.Read(snc.readBuf)
						//		b = snc.readBuf[0]
						b, err = snc.ReadByteWithDeadline(port)
						snc.Devices[i].ByteRecieved(b)
					}
					// CRC
					snc.ReadByteWithDeadline(port)
					if snc.Devices[i].ReadLength != 0 && l < 255 {
						snc.Devices[i].DataAvailable = true
					}
					//port.Read(snc.readBuf)
				}
			} else {
				snc.Devices[i].Online = false
			}
		}
	}

}
func (snc *SimpleNetCore) SimpleNetCoreLoop(port io.ReadWriteCloser) {
	snc.running = true
	fmt.Println("Started")
	for snc.running {
		snc.SimpleNetInnerLoop(port)
	}
	fmt.Println("Finished")
}
