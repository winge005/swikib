package helpers

import "os"

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true // file exists
	}
	if os.IsNotExist(err) {
		return false // file does not exist
	}
	// some other error (e.g. permission denied)
	return false
}
