package markdown_parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	fp "path/filepath"
	"unicode/utf8"
)

const MaxRunesInFileExtension int = 5

type FileTypeHint int

const (
	FileTypeHintMarkdown FileTypeHint = iota
	FileTypeHintDirectory
	FileTypeHintUnknown
)

type Linker interface {
	Link() (*EntityDescription, error)
}

type LinkParser interface {
	// Parse(newRune rune) (*Linker, error)
	Parse(newRune rune) (bool, Linker)
}

type BooleanParser interface {
	Parse(newRune rune) bool
}

var (
	// ErrNothingParsed = errors.New("LinkParser: nothing parsed yet")
	ErrRuneIsMultiple = errors.New("Parse(): currentRune has multiple runes which should not be possible")
)

type EntityDescription struct {
	location 	string 		// path to the object 
	fileName	string 		// The name of the file
	fileType 	string 		// starts with a "." and ends in a char+num sequence
}

func NewEntityDescription(location string, fileName string, fileType string) *EntityDescription {
	return &EntityDescription{
		location: location,
		fileName: fileName,
		fileType: fileType,
	}
}

func (this *EntityDescription) SetLocation(location string) {
	this.location = location
}

func (this EntityDescription) Location() string {
	return this.location
}

func (this EntityDescription) FileName() string {
	return this.fileName
}

func (this EntityDescription) FileType() string {
	return this.fileType
}

func (this EntityDescription) FullFileName() string {
	return this.fileName + this.fileType
}

func (this EntityDescription) FullPath() string {
	return fp.Join(this.location, this.fileName + this.fileType)
}

type Entity struct {
	EntityDescription 				// info about the file
	visited				bool 		// a flag that shows if the object has been visited by the crawler
	isPrivate			bool		// a flag that shows if the object should be excluded from the packaging
	links				[]Linker 	// a list of links that this file has
}

func NewEntity(ed EntityDescription) *Entity {
	return &Entity{
		EntityDescription: ed, 
		visited: false, 
		isPrivate: false,
		links: []Linker{},
	}
}

func (this *Entity) Forbidden() bool {
	return this.isPrivate
}

func (this *Entity) Links() []Linker {
	// retLinks := make([]Linker, len(this.links))
	// for i, p := range this.links {
	// 	retLinks[i] = p
	// }
	return this.links
}

type WikiLinkParser struct {
	currentLink 			string
	previousRune			*rune
	isLinkStarted   		bool
	noSymbolsAfterLinkStart bool
}

func NewWikiLinkParser() *WikiLinkParser {
	return &WikiLinkParser{
		currentLink: 				"",
		previousRune: 				nil,
		isLinkStarted: 				false,
		noSymbolsAfterLinkStart: 	true,
	}
}

func (this *WikiLinkParser) Parse(newRune rune) (bool, Linker) {
	if this.previousRune != nil {

		if !this.isLinkStarted {
			if newRune == '[' && *this.previousRune == '[' {
				this.isLinkStarted = true
			}
		} else {
			if newRune == '|' {
				if *this.previousRune == '[' {
					this.currentLink = ""
					this.isLinkStarted = false
					this.previousRune = nil
					return false, nil
				}
				this.isLinkStarted = false 
				fileEx := fp.Ext(this.currentLink)
				extensionLength := utf8.RuneCountInString(fileEx)
				for ; extensionLength > MaxRunesInFileExtension; { // TODO: implement some better way of fightin names with .
					fileEx = fp.Ext(fileEx[1:])
					extensionLength = utf8.RuneCountInString(fileEx)
				}
				fth := FileTypeHintUnknown
				if fileEx == "" {
					fth = FileTypeHintMarkdown
				}
				ret := Linker(NewWikiLink(this.currentLink, fth))
				this.currentLink = ""
				this.previousRune = nil
				return true, ret
			} else if newRune == ']' {
				if *this.previousRune == '[' {
					this.currentLink = ""
					this.isLinkStarted = false
					this.previousRune = nil
					return false, nil
				}
				if *this.previousRune == ']' {
					this.isLinkStarted = false
					fileEx := fp.Ext(this.currentLink)
					extensionLength := utf8.RuneCountInString(fileEx)
					for ; extensionLength > MaxRunesInFileExtension; { // TODO: implement some better way of fightin names with .
						fileEx = fp.Ext(fileEx[1:])
						extensionLength = utf8.RuneCountInString(fileEx)
					}
					fth := FileTypeHintUnknown
					if fileEx == "" {
						fth = FileTypeHintMarkdown
					}
					ret := Linker(NewWikiLink(this.currentLink, fth))
					this.currentLink = ""
					this.previousRune = nil
					return true, ret
				}
			} else {
				this.currentLink += string(newRune)
			}
		}
	
	} 
	
	this.previousRune = &newRune
	return false, nil
}

// Examples:
// [[File name]]
// [[Directory/File name]]
// [[Directory/File name|Alias]]
type WikiLink struct {
	contents	string 	// the contents of the link
	typeHint 	FileTypeHint
}

func NewWikiLink(contents string, typeHint FileTypeHint) *WikiLink {
	return &WikiLink{
		contents: contents,
		typeHint: typeHint,
	}
}

func (this *WikiLink) Link() (*EntityDescription, error) {
	switch this.typeHint {
	case FileTypeHintDirectory:
		dir, _ := fp.Split(this.contents)
		return NewEntityDescription(dir, "", ""), nil
	case FileTypeHintMarkdown:
		return NewEntityDescription("", this.contents, ".md"), nil
	case FileTypeHintUnknown:
		return NewEntityDescription("", this.contents, ""), nil
	}
	return nil, fmt.Errorf("WikiLink.Link(): no suitable linkage found for %v", this.contents)
}

// Examples:
// ???
type StorageLink struct {
	contents 	string	// the contents of the link
}

func (this *StorageLink) Link() (*EntityDescription, error) {
	return nil, errors.New("Not implemented")
}

// type EntityParser interface {
func Parse(entity *Entity, parsers []LinkParser, forbiddingParsers []BooleanParser) error {
	if entity.visited {
		return nil
	}

	if entity.fileType != ".md" {
		return nil
	} 

	// Opening the file 
	file, err := os.OpenFile(entity.EntityDescription.FullPath(), os.O_RDONLY, 0400)
	if err != nil {
		err = fmt.Errorf("Parse(): could not open file: %w", err)
		return err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)	
	scanner.Split(bufio.ScanRunes)
	
	entity.isPrivate = false

	linkers := make([]Linker, 0)
	for scanner.Scan() {
			if !entity.isPrivate {
			currentRune := []rune(scanner.Text())
			
			if len(currentRune) > 1 {
				return ErrRuneIsMultiple
			}

			for _, forbiddingParser := range forbiddingParsers {
				forbidden := forbiddingParser.Parse(currentRune[0])
				if forbidden {
					entity.isPrivate = true
					fmt.Printf("Warning: file %v is forbidden\n", entity.EntityDescription.FullPath())
				}
			}
			
			for _, parser := range parsers {
				ok, linker := parser.Parse(currentRune[0])
				if ok {
					linkers = append(linkers, linker)
				}
			}
		}
	}
	
	entity.visited = true 
	entity.links = linkers

	return nil
}
