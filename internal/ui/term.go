package ui

import (
	"os"
	"syscall"
	"unsafe"
)

// Key constants for special keys.
const (
	KeyEnter     = '\r'
	KeyNewline   = '\n'
	KeyEscape    = 27
	KeyBackspace = 127
	KeyCtrlC     = 3

	// Virtual key codes for arrow keys (outside ASCII range).
	KeyUp    = 1000
	KeyDown  = 1001
	KeyRight = 1002
	KeyLeft  = 1003
)

// termios mirrors the C termios struct for darwin/arm64.
type termios struct {
	Iflag  uint64
	Oflag  uint64
	Cflag  uint64
	Lflag  uint64
	Cc     [20]byte
	Ispeed uint64
	Ospeed uint64
}

// Terminal flag constants for darwin.
const (
	tICRNL  = 0x00000100
	tIXON   = 0x00000200
	tOPOST  = 0x00000001
	tECHO   = 0x00000008
	tICANON = 0x00000100
	tIEXTEN = 0x00000400
	tISIG   = 0x00000080
	tVMIN   = 16
	tVTIME  = 17
)

var (
	origTermios *termios
	stdinFd     = int(os.Stdin.Fd())
	rawMode     bool
)

// EnableRawMode puts the terminal into raw mode (no echo, no canonical, char-at-a-time).
// ISIG is kept enabled so Ctrl+C still generates SIGINT.
func EnableRawMode() error {
	if rawMode {
		return nil
	}

	var old termios
	if err := getTermios(stdinFd, &old); err != nil {
		return err
	}
	origTermios = &old

	raw := old
	// Disable echo, canonical mode, extended processing
	raw.Lflag &^= (tECHO | tICANON | tIEXTEN)
	// Keep ISIG so Ctrl+C works
	// Keep OPOST so \n still produces \r\n (avoids broken output formatting)

	// Disable CR-to-NL translation, XON/XOFF
	raw.Iflag &^= (tICRNL | tIXON)

	// Read returns after 1 byte, no timeout
	raw.Cc[tVMIN] = 1
	raw.Cc[tVTIME] = 0

	if err := setTermios(stdinFd, &raw); err != nil {
		return err
	}
	rawMode = true
	return nil
}

// DisableRawMode restores the terminal to its original state.
func DisableRawMode() {
	if origTermios == nil {
		return
	}
	setTermios(stdinFd, origTermios)
	rawMode = false
}

// IsRawMode returns whether the terminal is currently in raw mode.
func IsRawMode() bool {
	return rawMode
}

// ReadKey reads a single keypress from stdin. Handles escape sequences for arrow keys.
func ReadKey() (int, error) {
	var buf [1]byte
	_, err := os.Stdin.Read(buf[:])
	if err != nil {
		return 0, err
	}

	b := buf[0]

	// Escape sequence
	if b == KeyEscape {
		var seq [2]byte
		// Try to read 2 more bytes (arrow key sequence)
		n, _ := os.Stdin.Read(seq[:])
		if n < 2 {
			return KeyEscape, nil
		}
		if seq[0] == '[' {
			switch seq[1] {
			case 'A':
				return KeyUp, nil
			case 'B':
				return KeyDown, nil
			case 'C':
				return KeyRight, nil
			case 'D':
				return KeyLeft, nil
			}
		}
		return KeyEscape, nil
	}

	return int(b), nil
}

func getTermios(fd int, t *termios) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TIOCGETA),
		uintptr(unsafe.Pointer(t)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func setTermios(fd int, t *termios) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TIOCSETA),
		uintptr(unsafe.Pointer(t)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}
