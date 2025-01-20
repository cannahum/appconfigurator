package testutils

import "runtime"

func GetCurrentTestFilePath() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("Could not get current test file path")
	}
	return file
}
