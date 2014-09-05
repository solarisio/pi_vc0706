package vc0706

const (
	PORT              = "/dev/ttyAMA0"
	BAUD              = 38400
	SERIAL_NUM        = 0x00
	BUFFER_CHUNK_SIZE = uint32(256) // bytes

	CMD_END                  = 0x00
	CMD_GET_VERSION          = 0x11
	CMD_SET_SERIAL_NUMBER    = 0x21
	CMD_SET_PORT             = 0x24
	CMD_SYSTEM_RESET         = 0x26
	CMD_READ_DATA            = 0x30
	CMD_WRITE_DATA           = 0x31
	CMD_READ_BUF             = 0x32
	CMD_GET_BUF_LEN          = 0x34
	CMD_TAKE_PHOTO           = 0x36
	CMD_COMM_MOTION_CTRL     = 0x37
	CMD_COMM_MOTION_STATUS   = 0x38
	CMD_COMM_MOTION_DETECTED = 0x39
	CMD_COLOR_CTRL           = 0x3C
	CMD_COKOR_STATUS         = 0x3D
	CMD_MOTION_CTRL          = 0x42
	CMD_MOTION_STATUS        = 0x43
	CMD_TVOUT_CTRL           = 0x44
	CMD_OSD_ADD_CHAR         = 0x45
	CMD_SET_ZOOM             = 0x52
	CMD_GET_ZOOM             = 0x53
	CMD_DOWNSIZE_CTRL        = 0x54
	CMD_DOWNSIZE_STATUS      = 0x55
	CMD_SEND                 = 0x56
	CMD_REPLY                = 0x76
)

const (
	STOP_CURRENT_FRAME byte = iota
	STOP_NEXT_FRAME
	RESUME_FRAME
	STEP_FRAME
)

const (
	// Status:
	// 0: successful; 1: doesn't receive the cmd; 2: data length error;
	// 3:data format error; 4: cmd cannot exec now; 5: cmd received but
	// exec wrong
	STATUS_SUCCESS byte = iota
	STATUS_NOT_RECEIVED
	STATUS_DATA_LEN_ERROR
	STATUS_DATA_FMT_ERROR
	STATUS_CMD_NOT_EXEC
	STATUS_CMD_EXEC_ERROR
)

const (
	// data transfer mode
	MCU_MODE = 0x0A
	DMA_MODE = 0x0F
)

const (
	IMAGE_SIZE_LARGE  = 0x00
	IMAGE_SIZE_MEDIUM = 0x11
	IMAGE_SIZE_SMALL  = 0x22
)

const (
	COLOR_CTRL_MODE_GPIO byte = iota
	COLOR_CTRL_MODE_UART
)

const (
	COLOR_SHOW_MODE_AUTO byte = iota
	COLOR_SHOW_MODE_COLOR
	COLOR_SHOW_MODE_BLACK
)

const (
	DEVICE_TYPE_CHIP_REGISTER byte = iota
	DEVICE_TYPE_SENSOR_REGISTER
	DEVICE_TYPE_CCIR656_REGISTER
	DEVICE_TYPE_I2C_EEPROM
	DEVICE_TYPE_SPI_EEPROM
	DEVICE_TYPE_SPI_FLASH
)
