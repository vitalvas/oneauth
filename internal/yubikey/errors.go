package yubikey

import "errors"

var (
	ErrCardReaderUnavailable = errors.New("the specified reader is not currently available for use") // error 0x80100017

	ErrYubikeyNotOpen = errors.New("yubikey not opened")
)
