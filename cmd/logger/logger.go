package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	logFile    *os.File
	errLogFile *os.File
	vpnLogFile *os.File
)

func Init() error {
	var err error
	logFile, err = os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open app log file: %v", err)
	}

	errLogFile, err = os.OpenFile("logs/err.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open error log file: %v", err)
	}

	vpnLogFile, err = os.OpenFile("logs/vpn.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open VPN log file: %v", err)
	}

	return nil
}

func SetOutput(verbose bool) {
	if !verbose {
		log.SetOutput(logFile)
	} else {
		log.SetOutput(os.Stdout)
	}
}

func WriteVPNLog(message string) error {
	_, err := vpnLogFile.WriteString(message + "\n")
	return err
}

func WriteErrorLog(message string) error {
	_, err := errLogFile.WriteString(message + "\n")
	return err
}

func WriteAppLog(message string) error {
	_, err := logFile.WriteString(message + "\n")
	return err
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
	if errLogFile != nil {
		errLogFile.Close()
	}
	if vpnLogFile != nil {
		vpnLogFile.Close()
	}
}
