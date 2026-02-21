package packager

import (
	"fmt"
	"io"
	"os"
	fp "path/filepath"

	crawler "github.com/Yyote/gobsidizer/internal/entity_crawler"
	"github.com/Yyote/gobsidizer/internal/logger"
	// "github.com/Yyote/gobsidizer/internal/markdown_parser"
	// mdparser "github.com/Yyote/gobsidizer/internal/markdown_parser"
)

type Packager interface {
	Package(entityCrawler *crawler.Crawler) error
}

type FilePackager struct {
	Destination		string

	log 		 	logger.Logger
	maxBufferLen 	int
}

func NewFilePackager(Destination string, maxBufferLen int, log logger.Logger) *FilePackager {
	return &FilePackager{
		Destination: Destination,
		maxBufferLen: maxBufferLen,
		log: log,
	}
} 

func (this *FilePackager) Package(entityCrawler crawler.Crawler) error {
	fi, destStatErr := os.Stat(this.Destination)

	if destStatErr != nil {
		this.log.Error(destStatErr.Error())
		return fmt.Errorf("Destination stat error: %w", destStatErr)
	}

	if !fi.IsDir() {
		msg := "Destination is not a directory"
		this.log.Error(msg)
		return fmt.Errorf("%v", msg)
	}
	
	exportedEntitiesSlice := entityCrawler.ExportedEntities().Flatten()

	for _, entity := range exportedEntitiesSlice {
		rdFilePath := entity.FullPath()
		rdFile, err := os.OpenFile(rdFilePath, os.O_RDONLY, 0644)
		if err != nil {
			this.log.Error(fmt.Sprintf("Could not open file for reading %v", rdFilePath))
			return fmt.Errorf("Could not open file %v", rdFilePath)
		}
		defer rdFile.Close()

		wrFilePath := fp.Join(this.Destination, entity.FullFileName())
		wrFile, err := os.OpenFile(wrFilePath, os.O_CREATE | os.O_WRONLY, 0644)
		if err != nil {
			this.log.Error(fmt.Sprintf("Could not open file for writing %v", wrFilePath))
			return fmt.Errorf("Could not open file for writing %v", wrFilePath)
		}
		defer wrFile.Close()


		readBuffer := make([]byte, this.maxBufferLen)
		
		readBytesNum, readErr := rdFile.Read(readBuffer)
		for ; readErr == nil; {
			writeBuffer := readBuffer
			if readBytesNum < this.maxBufferLen {
				writeBuffer = readBuffer[:readBytesNum]
			}

			writtenBytesNum, writeError := wrFile.Write(writeBuffer)
			if writeError != nil {
				this.log.Error(writeError.Error())
				return writeError
			}
			if writtenBytesNum != readBytesNum {
				msg := fmt.Sprintf("During copying of file %v to destination file %v, there was a underwrite of bytes in quantity of %v", rdFilePath, wrFilePath, readBytesNum - writtenBytesNum)
				this.log.Error(msg)
				return fmt.Errorf("%v", msg)
			}
			readBytesNum, readErr = rdFile.Read(readBuffer)
		}

		if readErr != io.EOF {
			msg := fmt.Sprintf("During copying of file %v to destination file %v, there was an unexpected error: ", rdFilePath, wrFilePath, )
			this.log.Error(msg)
			return fmt.Errorf("%v", msg)
		}
		
		this.log.Debug(fmt.Sprintf("Successfully copied file %v to %v", rdFilePath, wrFilePath))
	}

	return nil
}
