package simplenet

import (
	"errors"
	//"fmt"
	"testing"
)

type MockPort struct {
	ReadBuffer  [10000]byte
	WriteBuffer [10000]byte

	ReadErrorPosition int
	ReadPosition      int
	ReadLength        int
	WritePosition     int
	WriteLength       int
}

func (mp *MockPort) Read(p []byte) (n int, err error) {
	if mp.ReadPosition == mp.ReadErrorPosition {
		return 0, errors.New("No Data")
	}
	if mp.ReadPosition > mp.ReadLength {
		return 0, nil
	}
	b := mp.ReadBuffer[mp.ReadPosition]
	mp.ReadPosition++
	p[0] = b
	return 1, nil
}

func (mp *MockPort) Write(p []byte) (n int, err error) {
	//fmt.Printf("OUT -> %d\n", c)
	for i := 0; i < len(p); i++ {
		mp.WriteBuffer[mp.WritePosition] = p[i]
		mp.WritePosition++
	}
	return len(p), nil
}

func (mp *MockPort) Close() error {
	return nil
}

func NewMockPort() *MockPort {
	mockport := new(MockPort)
	mockport.ReadErrorPosition = -1
	return mockport
}

func FillWireBuffer(mockport *MockPort, noOfDLoops int, did int, noBytes int) int {
	data := make([]byte, 0, 10000)
	for i := 0; i < (noOfDLoops * 15); i++ {
		data = append(data, 0xff)
		if i%15 == did {
			bb := 255
			if noBytes < 255 {
				bb = noBytes
			}
			data = append(data, byte(bb))
			for d := 0; d < bb; d++ {
				data = append(data, byte(d%53))
			}
			noBytes = noBytes - bb
		} else {
			data = append(data, 0)
		}
		data = append(data, 0)
	}

	mockport.ReadLength = len(data)
	for i := 0; i < len(data); i++ {
		mockport.ReadBuffer[i] = data[i]
	}
	return len(data)
}

// Test read from one device
func TestReadOneDevice(t *testing.T) {
	mockport := NewMockPort()
	data := []byte{0xff, 0x00, 0x00, 0xff, 11, 'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', 0x00}
	mockport.ReadLength = len(data)
	for i := 0; i < len(data); i++ {
		mockport.ReadBuffer[i] = data[i]
	}
	simplenetcore := NewSimpleNetCore()
	device := &simplenetcore.Devices[1]
	//go
	simplenetcore.SimpleNetInnerLoop(mockport)
	recdata := make([]byte, 1024, 1024)
	n, _ := device.Read(recdata)

	if n != 11 {
		t.Errorf("Recieved %d bytes expected 11", n)
	}
}

// Test read from one device
func TestWriteOneDevice(t *testing.T) {
	mockport := NewMockPort()
	data := []byte{0xff, 0x00, 0x00, 0xff, 0x00, 0x00}
	mockport.ReadLength = len(data)
	for i := 0; i < len(data); i++ {
		mockport.ReadBuffer[i] = data[i]
	}
	simplenetcore := NewSimpleNetCore()
	device := &simplenetcore.Devices[1]
	device.Write([]byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', 0x00})

	expectedData := []byte{0xff, 0, 0, 0, 0xff, 1, 12, 'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', 0x00, 0x00}
	//go
	simplenetcore.SimpleNetInnerLoop(mockport)
	expectedlen := 4*15 + 12
	if mockport.WritePosition != expectedlen {
		t.Errorf("Recieved %d bytes expected %d", mockport.WritePosition, expectedlen)
	}
	for i := 0; i < len(expectedData); i++ {
		if mockport.WriteBuffer[i] != expectedData[i] {
			t.Errorf("Bad write data")
			break
		}
	}
}

func TestReadMultiDevice(t *testing.T) {
	mockport := NewMockPort()
	data := []byte{0xff, 0x00, 0x00, 0xff, 11, 'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', 0x00, 0xff, 10, 'S', 'e', 'c', 'o', 'n', 'd', ' ', 'M', 's', 'g', 0x00}
	mockport.ReadLength = len(data)
	for i := 0; i < len(data); i++ {
		mockport.ReadBuffer[i] = data[i]
	}
	simplenetcore := NewSimpleNetCore()
	deviceOne := &simplenetcore.Devices[1]
	deviceTwo := &simplenetcore.Devices[2]
	//go
	simplenetcore.SimpleNetInnerLoop(mockport)
	recdata := make([]byte, 1024, 1024)
	n, _ := deviceOne.Read(recdata)
	if n != 11 {
		t.Errorf("Device One Recieved %d bytes expected 11", n)
	}
	n, _ = deviceTwo.Read(recdata)
	if n != 10 {
		t.Errorf("Device Two Recieved %d bytes expected 10", n)
	}
}

func TestWriteMultiDevice(t *testing.T) {
	mockport := NewMockPort()
	data := []byte{0xff, 0x00, 0x00, 0xff, 0x00, 0x00, 0xff, 0x00, 0x00, 0xff, 0x00, 0x00}
	mockport.ReadLength = len(data)
	for i := 0; i < len(data); i++ {
		mockport.ReadBuffer[i] = data[i]
	}
	simplenetcore := NewSimpleNetCore()
	deviceOne := &simplenetcore.Devices[1]
	deviceTwo := &simplenetcore.Devices[2]
	deviceOne.Write([]byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', 0x00})
	deviceTwo.Write([]byte{'S', 'e', 'c', 'o', 'n', 'd', ' ', 'M', 's', 'g', 0x00})

	expectedData := []byte{0xff, 0, 0, 0, 0xff, 1, 12, 'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', 0x00, 0x00, 0xff, 2, 11, 'S', 'e', 'c', 'o', 'n', 'd', ' ', 'M', 's', 'g', 0x00, 0x00}
	//go
	simplenetcore.SimpleNetInnerLoop(mockport)
	expectedlen := 4*15 + 12 + 11
	if mockport.WritePosition != expectedlen {
		t.Errorf("Recieved %d bytes expected %d", mockport.WritePosition, expectedlen)
	}
	for i := 0; i < len(expectedData); i++ {
		if mockport.WriteBuffer[i] != expectedData[i] {
			t.Errorf("Bad write data")
			break
		}

	}
}

func TestMultiPacketRead(t *testing.T) {
	mockport := NewMockPort()
	mockport.ReadLength = FillWireBuffer(mockport, 2, 1, 306)
	simplenetcore := NewSimpleNetCore()
	device := &simplenetcore.Devices[1]
	//go
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	recdata := make([]byte, 1024, 1024)
	n, _ := device.Read(recdata)

	if n != 306 {
		t.Errorf("Recieved %d bytes expected 306", n)
	}
}

func TestWriteMultiPacketWrite(t *testing.T) {
	mockport := NewMockPort()
	pos := 0
	for i := 0; i < 32; i++ {
		mockport.ReadBuffer[pos] = 0xff
		pos++
		mockport.ReadBuffer[pos] = 0x00
		pos++
		mockport.ReadBuffer[pos] = 0x00
		pos++
	}

	simplenetcore := NewSimpleNetCore()
	device := &simplenetcore.Devices[1]

	for i := 0; i < 320; i++ {
		device.Write([]byte{byte(i % 51)})
	}

	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	expectedlen := 4*30 + 320
	if mockport.WritePosition != expectedlen {
		t.Errorf("Recieved %d bytes expected %d", mockport.WritePosition, expectedlen)
	}

}

func TestNoDevice(t *testing.T) {
	mockport := NewMockPort()
	data := make([]byte, 15*3, 15*3)
	pos := 0
	for i := 0; i < 14; i++ {
		data[pos] = 0xff
		pos++
		data[pos] = 0x00
		pos++
		data[pos] = 0x00
		pos++
	}
	mockport.ReadErrorPosition = 6
	mockport.ReadLength = len(data)
	for i := 0; i < len(data); i++ {
		mockport.ReadBuffer[i] = data[i]
	}

	simplenetcore := NewSimpleNetCore()
	//device := &simplenetcore.devices[1]
	//go
	simplenetcore.SimpleNetInnerLoop(mockport)

	if !simplenetcore.Devices[1].Online {
		t.Errorf("Device 1 offline but should be online")
	}
	if simplenetcore.Devices[2].Online {
		t.Errorf("Device 1 online but should be offline")
	}
}

func TestOverBufferReadAndStopRec(t *testing.T) {
	mockport := NewMockPort()
	mockport.ReadLength = FillWireBuffer(mockport, 5, 1, 1400)
	simplenetcore := NewSimpleNetCore()
	device := &simplenetcore.Devices[1]
	//go
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)

	recdata := make([]byte, 2024, 2024)
	n, _ := device.Read(recdata)
	if n != 1020 {
		t.Errorf("Recieved %d bytes expected 1020", n)
	}
}

/*
func TestOverBufferReadWithFlush(t *testing.T) {
	mockport := NewMockPort()
	mockport.ReadLength = FillWireBuffer(mockport, 6, 2, 1350)
	simplenetcore := NewSimpleNetCore()
	device := &simplenetcore.Devices[2]
	recdata := make([]byte, 2024, 2024)
	//go
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	_, _ = device.Read(recdata)

	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	n, _ := device.Read(recdata)

	if n != 330 {
		t.Errorf("Recieved %d bytes expected 1020", n)
	}
}
*/

func TestOverBufferWrite(t *testing.T) {
	mockport := NewMockPort()
	mockport.ReadLength = FillWireBuffer(mockport, 6, 2, 0)
	simplenetcore := NewSimpleNetCore()
	device := &simplenetcore.Devices[2]
	//	recdata := make([]byte, 2024, 2024)

	data := make([]byte, 1400, 1400)
	for i := 0; i < 1400; i++ {
		data[i] = byte(i % 53)
	}

	n, _ := device.Write(data)
	if n != 1024 {
		t.Errorf("Written %d bytes but expected 1024", n)
	}

	//device.Write([]byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', 0x00})

	//go
	simplenetcore.SimpleNetInnerLoop(mockport)

	n, _ = device.Write(data)
	if n != 255 {
		t.Errorf("Written %d bytes but expected 255", n)
	}

	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)
	simplenetcore.SimpleNetInnerLoop(mockport)

}
