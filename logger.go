package main

import ()

// const (
// 	LOG_CLEAR_SYMBOL = "\033[0m"
// 	LOG_TYPE_ERROR = "ERROR"
// 	LOG_MODE_ERROR   = "\033[91m"
// 	LOG_TYPE_INFO = "INFO"
// 	LOG_MODE_INFO      = "\033[94m"
// )

const (
	LOG_MODE_INFO    = "INFO"
	LOG_MODE_ERROR   = "ERROR"
	LOG_MODE_WARNING = "WARN"
)

// func log(msg string, mode string) {
// 	if _, file, line, ok := runtime.Caller(0); ok {
// 		fmt.Printf("[%s %s:%d] %s\n", mode, filepath.Base(file), line, msg)
// 	} else {
// 		fmt.Printf("[%s] %s\n", mode, msg)
// 	}
// }
