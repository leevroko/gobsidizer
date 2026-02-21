package entity_crawler

import (
	"errors"

	utils "github.com/Yyote/gobsidizer/internal/utilities"
)

type FileRegister interface {
	AddFile(pathToFile string, filename string) error
	RemoveFile(filename string, pathToFile string) error
	GetPaths(filename string) ([]string, error)
}

var (
	ErrFileRegistryFileIsPresent		= errors.New("FileRegistry error: file is already present")
	ErrFileRegistryFileIsMissing		= errors.New("FileRegistry error: file is missing")
	ErrFileRegistryFileMapExistsEmpty 	= errors.New("FileRegistry error: file map entry exists but empty")
)

type FileRegistry struct {
	files map[string][]string
}

func NewFileRegistry() *FileRegistry {
	return &FileRegistry{
		files: make(map[string][]string),
	}
}

func (this *FileRegistry) AddFile(pathToFile string, filename string) error {
	paths, filePresent := this.files[filename]
	if !filePresent {
		this.files[filename] = make([]string, 1)
		this.files[filename][0] = pathToFile
	} else {
		_, found := utils.FindInSlice(paths, pathToFile)
		if !found {
			this.files[filename] = append(this.files[filename], pathToFile)
		} else {
			return ErrFileRegistryFileIsPresent
		}
	}
	return nil
}

func (this *FileRegistry) RemoveFile(filename string, pathToFile string) error {
	paths, filePresent := this.files[filename]
	
	if !filePresent {
		return ErrFileRegistryFileIsMissing
	} 

	if len(paths) > 1 {
		pos, found := utils.FindInSlice(paths, pathToFile)
		if !found {
			return ErrFileRegistryFileIsMissing
		}
		newPaths, delErr := utils.DeleteFromSlice(paths, pos)
		if delErr != nil {
			return delErr
		}
		this.files[filename] = *newPaths
	} else {
		if len(paths) == 1 {
			delete(this.files, filename)
			return nil
		}
		if len(paths) == 0 {
			return ErrFileRegistryFileMapExistsEmpty
		}
	}
	return nil
}

func (this *FileRegistry) GetPaths (filename string) ([]string, error) {
	v, ok := this.files[filename]	
	if ok {
		return v, nil
	} else {
		return []string{}, ErrFileRegistryFileIsMissing
	}
}
