package vc0706

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/golang/glog"
	"github.com/tarm/goserial"
)

type SerialIO io.ReadWriteCloser

var (
	EMPTY_DATA = []byte{}

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
	cmd = []byte{CMD_SEND, SERIAL_NUM, c, dl}
	cmd = append(cmd, d...)
	return
}

func MakeSimpleSendCmd(c byte) (cmd []byte) {
	cmd = MakeSendCmd(c, CMD_END, EMPTY_DATA)
	return
}

// Communication protocol for return:
// protocol sign(1B) + serial number (1B) + cmd (1B) + status(1B) +
// data length(1B) + Data(0~16B)
func MakeReplyCmd(c, s, dl byte, d []byte) (cmd []byte) {
	cmd = []byte{CMD_REPLY, SERIAL_NUM, c, s, dl}
	cmd = append(cmd, d...)
	return
}

func MakeSimpleReplyCmd(c byte) (cmd []byte) {
	cmd = MakeReplyCmd(c, STATUS_SUCCESS, CMD_END, EMPTY_DATA)
	return
}

// Check the reply to make sure the command executes successfully
// Only check the first byte (reply), 3rd byte (command), the status
func CheckReply(c byte, r []byte) (err error) {
	switch {
	case r[0] != CMD_REPLY:
		err = errors.New(hex.EncodeToString(r[0:1]) + " is not a reply cmd")
	case r[1] != SERIAL_NUM:
		err = errors.New("Unexpected serial number " + hex.EncodeToString(r[1:2]))
	case r[2] != c:
		err = errors.New("Expect cmd " + hex.EncodeToString([]byte{c}) + ", but got " +
			hex.EncodeToString(r[2:3]))
	case r[3] != STATUS_SUCCESS:
		err = errors.New("Error code " + hex.EncodeToString(r[3:4]))
	}
	return
}

func GetVersion(s SerialIO) (v string, err error) {
	cmd := MakeSimpleSendCmd(CMD_GET_VERSION)
	buf, err := RunCmd(s, cmd, 16, 10)
	v = string(buf[5:])
	return
}

// Do not trust the doc
func Reset(s SerialIO) (err error) {
	cmd := MakeSimpleSendCmd(CMD_SYSTEM_RESET)
	_, err = RunCmd(s, cmd, 80, 1000)
	return
}

// Actually run a command
// length: expected buffer length
func RunCmd(s SerialIO, cmd []byte, length uint32, ms int) (buf []byte, err error) {
	_, err = s.Write(cmd)
	if err != nil {
		glog.Warning(err)
		return
	}
	// wait a bit
	time.Sleep(time.Duration(ms) * time.Millisecond)
	buf = make([]byte, length)
	n, err := s.Read(buf)
	if err != nil {
		glog.Warning(err)
		return
	}
	buf = buf[:n]
	err = CheckReply(cmd[2], buf)
	if err != nil {
		glog.Warning(err)
	}
	return
}

// Default: medium (320x240)
// Command format: 0x56 + serial number + 0x31 + 0x05 + 0x04 + 0x01 + 0x00 +
// 0x19 + size
// Return format: 0x76 + serial number + 0x31 + 0x00 + 0x00
func SetPhotoSize(s SerialIO, sz string) (err error) {
	size, ok := IMAGE_SIZES[sz]
	if !ok {
		size = IMAGE_SIZE_MEDIUM
	}
	data := []byte{0x04, 0x01, 0x00, 0x19, size}
	cmd := MakeSendCmd(CMD_WRITE_DATA, 0x05, data)
	_, err = RunCmd(s, cmd, 5, 100)
	return
}

// register address: 0x12 0x04
func SetCompression(s SerialIO, rate byte) (err error) {
	data := []byte{DEVICE_TYPE_CHIP_REGISTER, 0x01, 0x12, 0x04, rate}
	cmd := MakeSendCmd(CMD_WRITE_DATA, 0x05, data)
	_, err = RunCmd(s, cmd, 5, 10)
	return
}

func SetColorMode(s SerialIO, ctrlMode, showMode byte) (err error) {
	cmd := MakeSendCmd(CMD_COLOR_CTRL, 0x02, []byte{ctrlMode, showMode})
	_, err = RunCmd(s, cmd, 5, 10)
	return
}

func InitCamera() (s SerialIO, err error) {
	c := &serial.Config{Name: PORT, Baud: BAUD}
	s, err = serial.OpenPort(c)
	if err != nil {
		glog.Fatal(err)
		return
	}
	glog.Infoln("Reset camera")
	if err = Reset(s); err != nil {
		glog.Fatal(err)
		return
	}
	// version, err := GetVersion(s)
	// if err != nil {
	// 	glog.Fatal(err)
	// 	return
	// }
	// glog.Infoln("Camera version " + version)
	// glog.Infoln("Set photo size")
	// if err = SetPhotoSize(s, "m"); err != nil {
	// 	glog.Warning(err)
	// }
	// if err = SetColorMode(s, COLOR_CTRL_MODE_GPIO, COLOR_SHOW_MODE_COLOR); err != nil {
	// 	glog.Warning(err)
	// }
	return
}

// Command format: 0x56 + serial number + 0x34 + 0x01 + FBUF type(1 byte)
// 0 for current frame, 1 for next frame
// Return format: 0x76 + serial number + 0x34 + 0x00 + 0x04 +
// FBUF data-lengths (4 bytes)
func GetBufferLen(s SerialIO) (length uint32, err error) {
	cmd := MakeSendCmd(CMD_GET_BUF_LEN, 0x01, []byte{0x00})
	buf, err := RunCmd(s, cmd, 9, 500)
	if err != nil {
		return
	}
	length = binary.BigEndian.Uint32(buf[5:])
	return
}

func VerifyFrame(buf []byte) (err error) {
	bufLen := len(buf)
	header := MakeSimpleReplyCmd(CMD_READ_BUF)
	if bytes.Compare(header, buf[:5]) != 0 || bytes.Compare(header, buf[bufLen-5:]) != 0 {
		err = errors.New("Frame headers are incorrect")
	}
	return
}

// Command format: 0x56 + serial number + 0x32 + 0x0C + FBUF type (1 byte)
// + control mode(1 byte) + starting address(4 bytes) + data-length(4 bytes)
// + delay(2 bytes)
// Return format:  0x76 + serial number + 0x32 + 0x00 + 0x00 + image data
// + 0x76 + serial number + 0x32 + 0x00 + 0x00
func ReadBuffer(s SerialIO) (buf []byte, err error) {
	length, err := GetBufferLen(s)
	if err != nil {
		glog.Warning(err)
		return
	}
	delay := []byte{0x01, 0x00} // 0.01 millisecond
	mode := []byte{STOP_CURRENT_FRAME, MCU_MODE}
	frameLen := BUFFER_CHUNK_SIZE
	frameLenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(frameLenBuf, frameLen)
	data := []byte{}
	frame := []byte{}
	retry := 0
	for start := uint32(0); start < length; start += BUFFER_CHUNK_SIZE {
		if start > length {
			frameLen = start - length
			binary.BigEndian.PutUint32(frameLenBuf, frameLen)
		}
		offset := make([]byte, 4)
		binary.BigEndian.PutUint32(offset, start)
		data = append(append(append(mode, offset...), frameLenBuf...), delay...)
		cmd := MakeSendCmd(CMD_READ_BUF, 0x0C, data)
		frame, err = RunCmd(s, cmd, 5+frameLen+5, 500)
		if err != nil {
			glog.Warning(err)
			return
		}
		if err = VerifyFrame(frame); err != nil {
			glog.Warningln(err.Error(), "retrying...")
			if retry > 5 {
				return
			}
			start -= BUFFER_CHUNK_SIZE
			retry += 1
			continue
		}
		frame = frame[5 : frameLen+5]
		buf = append(buf, frame...)
	}
	return
}

// Command format: 0x56 + serial number + 0x36 + 0x01 + control flag(1 byte)
// Return format:
// - OK: 0x76 + serial number + 0x36 + 0x00 + 0x00
// - Error: 0x76 + serial number + 0x36 + 0x03 + 0x00
func TakePhoto(s SerialIO) (buf []byte, err error) {
	cmd := MakeSendCmd(CMD_TAKE_PHOTO, 0x01, []byte{STOP_CURRENT_FRAME})
	if _, err = RunCmd(s, cmd, 5, 5000); err != nil {
		glog.Warning(err)
		return
	}
	buf, err = ReadBuffer(s)
	return
}

func SaveBuffer(filename string, data []byte) (err error) {
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		glog.Warning(err)
		return
	}
	return
}
