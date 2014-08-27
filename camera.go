package vc0706

import (
	"encoding/hex"
	"errors"
	"io"
	"log"
	"time"

	"github.com/golang/glog"
	"github.com/tarm/goserial"
)

type SerialIO io.ReadWriteCloser

const (
	PORT = "/dev/ttyAMA0"
	BAUD = 38400

	DELAY             = 10 * time.Millisecond
	REPLY_BUFFER_SIZE = 21
	BUFFER_SIZE       = 128
	SERIAL_NUM        = byte(0x00)
	EMPTY_DATA        = [0]byte{}

	CMD_END                  = byte(0x00)
	CMD_VERSION              = byte(0x11)
	CMD_SET_SERIAL_NUMBER    = byte(0x21)
	CMD_SET_PORT             = byte(0x24)
	CMD_RESET                = byte(0x26)
	CMD_READ_DATA            = byte(0x30)
	CMD_WRITE_DATA           = byte(0x31)
	CMD_READ_BUF             = byte(0x32)
	CMD_GET_BUF_LEN          = byte(0x34)
	CMD_TAKE_PHOTO           = byte(0x36)
	CMD_COMM_MOTION_CTRL     = byte(0x37)
	CMD_COMM_MOTION_STATUS   = byte(0x38)
	CMD_COMM_MOTION_DETECTED = byte(0x39)
	CMD_MOTION_CTRL          = byte(0x42)
	CMD_MOTION_STATUS        = byte(0x43)
	CMD_TVOUT_CTRL           = byte(0x44)
	CMD_OSD_ADD_CHAR         = byte(0x45)
	CMD_SET_ZOOM             = byte(0x52)
	CMD_GET_ZOOM             = byte(0x53)
	CMD_DOWNSIZE_CTRL        = byte(0x54)
	CMD_DOWNSIZE_STATUS      = byte(0x55)
	CMD_SEND                 = byte(0x56)
	CMD_REPLY                = byte(0x76)

	STOP_CURRENT_FRAME = byte(0x00)
	STOP_NEXT_FRAME    = byte(0x01)
	RESUME_FRAME       = byte(0x02)
	STEP_FRAME         = byte(0x03)

	MOTION_CONTROL  = byte(0x00)
	UART_MOTION     = byte(0x01)
	ACTIVATE_MOTION = byte(0x01)

	// Status:
	// 0: successful; 1: doesn't receive the cmd; 2: data length error;
	// 3:data format error; 4: cmd cannot exec now; 5: cmd received but
	// exec wrong
	STATUS_SUCCESS        = byte(0x00)
	STATUS_NOT_RECEIVED   = byte(0x01)
	STATUS_DATA_LEN_ERROR = byte(0x02)
	STATUS_DATA_FMT_ERROR = byte(0x03)
	STATUS_CMD_NOT_EXEC   = byte(0x04)
	STATUS_CMD_EXEC_ERROR = byte(0x05)

	// data transfer mode
	MCU_MODE = byte(0x0A)
	DMA_MODE = byte(0x0F)

	IMAGE_SIZE_LARGE  = byte(0x00)
	IMAGE_SIZE_MEDIUM = byte(0x11)
	IMAGE_SIZE_SMALL  = byte(0x22)

	IMAGE_SIZES = map[string]byte{
		"l": IMAGE_SIZE_LARGE,
		"m": IMAGE_SIZE_MEDIUM,
		"s": IMAGE_SIZE_SMALL,
	}
)

// Communication protocol for receive:
// protocol sign(1B) + serial number(1B) + command(1B) + data length(1B) +
// data(0~16B)
func MakeSendCmd(c, dl byte, d []byte) (cmd []byte) {
	cmd = [4]byte{CMD_SEND, SERIAL_NUM, c, dl}
	cmd = append(cmd, d...)
}

func MakeSimpleSendCmd(c byte) (cmd []byte) {
	cmd = MakeSendCmd(c, CMD_END, EMPTY_DATA)
}

// Communication protocol for return:
// protocol sign(1B) + serial number (1B) + cmd (1B) + status(1B) +
// data length(1B) + Data(0~16B)
func MakeReplyCmd(c, s, dl byte, d []byte) (cmd []byte) {
	cmd = [5]byte{CMD_REPLY, SERIAL_NUM, c, s, dl}
	cmd = append(cmd, d...)
}

func MakeSimpleReplyCmd(c byte) (cmd []byte) {
	cmd = MakeReplyCmd(c, STATUS_SUCCESS, CMD_END, EMPTY_DATA)
}

// Check the reply to make sure the command executes successfully
func CheckReply(c byte, r []byte) (err error) {
	switch {
	case reply[2] != c:
		err = errors.New("Expect: " + hex.EncodeToString(c) + ", get: " +
			hex.EncodeToString(reply[2]))
	case reply[3] != STATUS_SUCCESS:
		err = errors.New("Error code: " + hex.EncodeToString(reply[3]))
	}
}

// Command format: 0x56 + serial number + 0x26 + 0x00
// Return format: 0x76 + serial number + 0x26 + 0x00 + 0x00
func Reset(s SerialIO) (err error) {
	cmd := MakeSimpleReplyCmd(CMD_RESET)
	if _, err := s.Write(cmd); err != nil {
		glog.Fatal(err)
	}
	buf := make([]byte, 5)
	if _, err := s.Read(buf); err != nil {
		glog.Fatal(err)
	}
	err = CheckReply(CMD_RESET, buf)
}

// Default: medium (320x240)
// Command format: 0x56 + serial number + 0x31 + 0x05 + 0x04 + 0x01 + 0x00 +
// 0x19 + size
// Return format: 0x76 + serial number + 0x31 + 0x00 + 0x00
func SetPhotoSize(s SerialIO, sz string) (err error) {
	size := IMAGE_SIZES[sz]
	if size == nil {
		size = IMAGE_SIZE_MEDIUM
	}
	data := [5]byte{0x04, 0x01, 0x00, 0x19, size}
	cmd := MakeSendCmd(CMD_WRITE_DATA, 0x05, data)
}

func InitCamera() (s SerialIO, err error) {
	c := &serial.Config{Name: PORT, Baud: BAUD}
	s, err := serial.OpenPort(c)
	if err != nil {
		glog.Fatal(err)
	}
}

// Command format: 0x56 + serial number + 0x36 + 0x01 + control flag(1 byte)
// Return format:
// - OK: 0x76 + serial number + 0x36 + 0x00 + 0x00
// - Error: 0x76 + serial number + 0x36 + 0x03 + 0x00
func TakePhoto(s SerialIO) (err error) {
	cmd := MakeSendCmd(CMD_TAKE_PHOTO, 0x01, STOP_CURRENT_FRAME)
	if _, err := s.Write(cmd); err != nil {
		glog.Warning(err)
	}
	buf := make([]byte, 5)
	if _, err := s.Read(buf); err != nil {
		log.Warning(err)
	}
	err = CheckReply(CMD_TAKE_PHOTO, buf)
}

// Command format: 0x56 + serial number + 0x34 + 0x01 + FBUF type(1 byte)
// 0 for current frame, 1 for next frame
// Return format: 0x76 + serial number + 0x34 + 0x00 + 0x04 +
// FBUF data-lengths (4 bytes)
func GetBufferLen(s SerialIO) (length []byte, err error) {
	cmd := MakeSendCmd(CMD_GET_BUF_LEN, 0x01, 0x00)
	if _, err := s.Write(cmd); err != nil {
		glog.Warning(err)
	}
	buf := make([]byte, 9)
	if _, err := s.Read(buf); err != nil {
		glog.Warning(err)
	}
	err = CheckReply(CMD_GET_BUF_LEN, buf)
	length = buf[5:]
	// if err == nil {
	// 	reader := bytes.NewReader(buf[5:])
	// 	err = binary.Read(reader, binary.BigEndian, &success)
	// 	if err != nil {
	// 		glog.Warning(err)
	// 	}
	// }
}

// Command format: 0x56 + serial number + 0x32 + 0x0C + FBUF type (1 byte)
// + control mode(1 byte) + starting address(4 bytes) + data-length(4 bytes)
// + delay(2 bytes)
// Return format:  0x76 + serial number + 0x32 + 0x00 + 0x00 + image data
// + 0x76 + serial number + 0x32 + 0x00 + 0x00
func ReadBuffer(s SerialIO, length []byte) (buf []byte, err error) {
	start := []byte{0x00, 0x00, 0x00, 0x00}
	delay := []byte{0x00, 0x0A} // 0.01 millisecond
	data := []byte{STOP_CURRENT_FRAME, DMA_MODE}
	data = append(append(append(data, start...), length...), delay...)
	cmd := MakeSendCmd(CMD_READ_BUF, 0x0C, data)

}
