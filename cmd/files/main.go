package main

import (
	"fmt"
	"os"
	fp "path/filepath"
	logger "github.com/Yyote/gobsidizer/internal/logger"
)

func init() {
	fmt.Println("First init")
}

func init() {
	fmt.Println("Second init")
}

func init() {
	fmt.Println("Third init")
}

func returnsInterface(isNil bool) any {
	if isNil {
		return nil
	}

	return !isNil
}

func someKindOfTest() {
	file, err := os.OpenFile("./noway/artifact.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		return
	}

	defer func (file *os.File) error {
		fmt.Println("File is closed.")
		return file.Close()
	}(file)

	_, writeErr := file.WriteString("I love Paris\nnew line\n")

	if writeErr != nil {
		fmt.Printf("Write error: %v\n", writeErr.Error())
	} else {
		fmt.Println("Write successful")
	}
}

func splitOnFilename() {
	filename := "file.txt"
	dir, file := fp.Split(filename)
	fmt.Printf("Dir = %v, File = %v\n", dir, file)

	filename = "/home/leev/Yandex.Disk/Vaults/"
	dir, file = fp.Split(filename)
	fmt.Printf("Dir = %v, File = %v\n", dir, file)
}

func testLogger() {
	for i := int(logger.DEBUG); i <= int(logger.ERROR); i++ {
		fmt.Println("---")
		logger := logger.Logger(logger.NewPrintLogger("prefix", logger.LogLevel(i)))
		logger.Debug("This is debug msg")
		logger.Info("This is info msg")
		logger.Warn("This is warn msg")
		logger.Error("This is error msg")
	}
}

func main () {
	// someKindOfTest()
	// splitOnFilename()
	// testLogger()
	// ext := fp.Ext("mat.ndim - numpy - количество измерений массива")
	// fmt.Println(ext)
	intrfc := returnsInterface(true)
	if intrfc != nil {
		fmt.Printf("interface is not nil = %v", intrfc)
	} else {
		fmt.Printf("interace is nil")
	}
}
