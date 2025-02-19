package terminal

const (
	KeyArrowUp         = '\x10'
	KeyArrowDown       = '\x10'
	KeyEnter           = '\r'
	KeyArrowLeft       = '\x02'
	KeyArrowRight      = '\x06'
	KeySpace           = ' '
	KeyBackspace       = '\b'
	KeyDelete          = '\x7f'
	KeyInterrupt       = '\x03'
	KeyEndTransmission = '\x04'
	KeyEscape          = '\x1b'
	KeyDeleteWord      = '\x17' // Ctrl+W
	KeyDeleteLine      = '\x18' // Ctrl+X
	SpecialKeyHome     = '\x01'
	SpecialKeyEnd      = '\x11'
	SpecialKeyDelete   = '\x12'
	IgnoreKey          = '\000'
	KeyTab             = '\t'
)

const (
	normalKeypad      = '['
	applicationKeypad = 'O'
)

const ioctlReadTermios = 0x5401  // syscall.TCGETS
const ioctlWriteTermios = 0x5402 // syscall.TCSETS
