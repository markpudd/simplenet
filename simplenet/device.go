package simplenet

import (
	//"fmt"
	"sync"
	"time"
)

type Device struct {
	ReadBuffer  [1024]byte
	WriteBuffer [1024]byte

	ReadPosition  int
	ReadLength    int
	WritePosition int
	WriteLength   int

	Online        bool
	DataAvailable bool

	readLock  sync.Mutex
	writeLock sync.Mutex
}

func NewDevice() *Device {
	device := new(Device)
	device.Online = false
	device.DataAvailable = false
	return device
}

// GetBytesForWire - this gets a byte array to write to wire
func (d *Device) GetBytesForWire() (b []byte, err error) {
	d.writeLock.Lock()
	bytesToRead := d.WriteLength
	if bytesToRead > 255 {
		bytesToRead = 255
	}
	data := make([]byte, bytesToRead, bytesToRead)
	pos := 0
	sPos := d.WritePosition
	if sPos < 0 {
		sPos = sPos + 1024
	}

	for pos < bytesToRead {
		data[pos] = d.WriteBuffer[sPos]
		pos++
		sPos++
		if sPos >= 1024 {
			sPos = 0
		}
	}
	d.WritePosition = sPos
	d.WriteLength = d.WriteLength - bytesToRead
	d.writeLock.Unlock()
	return data, nil
}

// ByteRecieved - When a byte is recieved write it to the devices recieve buffer
func (d *Device) ByteRecieved(b byte) {
	d.readLock.Lock()
	// TODO block on overwrite - actually probably should drop because blocking
	// here will block wire......
	pos := d.ReadPosition + d.ReadLength
	if pos >= 1024 {
		pos = pos - 1024
	}
	d.ReadBuffer[pos] = b
	d.ReadLength++
	d.readLock.Unlock()
}

// Read - Read any buffered data, this confirms to Reader IF
func (d *Device) Read(b []byte) (n int, err error) {
	// TODO Check threading iomplication here
	for !d.DataAvailable {
		time.Sleep(3 * time.Millisecond)
	}
	d.readLock.Lock()

	ll := d.ReadLength
	//	fmt.Printf("RL - %d\n", d.ReadLength)
	for i := 0; i < d.ReadLength; i++ {
		b[i] = d.ReadBuffer[d.ReadPosition]
		d.ReadPosition++
		if d.ReadPosition >= 1024 {
			d.ReadPosition = 0
		}
	}
	d.ReadLength = 0
	d.DataAvailable = false
	d.readLock.Unlock()
	return ll, nil

}

// Write - Write bytes to device output buffer -  Writer IF
func (d *Device) Write(p []byte) (n int, err error) {
	d.writeLock.Lock()
	availableBuffer := 1024 - d.WriteLength
	if len(p) < availableBuffer {
		availableBuffer = len(p)
	}
	pos := d.WritePosition + d.WriteLength
	for i := 0; i < availableBuffer; i++ {
		if pos >= 1024 {
			pos = 0
		}
		d.WriteBuffer[pos] = p[i]
		pos++
	}
	d.WriteLength = d.WriteLength + availableBuffer
	d.writeLock.Unlock()
	return availableBuffer, nil
}
