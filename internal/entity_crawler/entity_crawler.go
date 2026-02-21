package entity_crawler

import (
	"errors"
	"os"
	"fmt"
	fp 			"path/filepath"

	mdparser 	"github.com/Yyote/gobsidizer/internal/markdown_parser"
	log 		"github.com/Yyote/gobsidizer/internal/logger"

	r "regexp"
)

var (
	ErrEntityRegTryingToAddExistingEntity	= 	errors.New("EntityRegistry: trying to add an already existing entity")
	ErrEntityRegNotFound					= 	errors.New("EntityRegistry: entity not found")
)

type EntityRegister interface {
	AddEntity(newEntity *mdparser.Entity) error
	RemoveEntity(filename string, pathToFile string) error
	GetEntity(filename string, pathToFile string) (*mdparser.Entity, error)
	HasEntity(filename string, pathToFile string) bool
	Flatten() []*mdparser.Entity
}

type EntityRegistry struct {
	storage 	map[string]*mdparser.Entity
}

func NewEntityRegistry() *EntityRegistry {
	return &EntityRegistry{
		storage: make(map[string]*mdparser.Entity),
	}
}

func (this *EntityRegistry) GetStorage() map[string]*mdparser.Entity {
	return this.storage
}

func (this *EntityRegistry) AddEntity(newEntity *mdparser.Entity) error {
	name := newEntity.EntityDescription.FullPath()
	_, found := this.storage[name]
	if found {
		return ErrEntityRegTryingToAddExistingEntity
	}

	this.storage[name] = newEntity
	return nil
}

func (this *EntityRegistry) RemoveEntity(filename string, pathToFile string) error {
	name := pathToFile + filename
	_, found := this.storage[name]
	if !found {
		return ErrEntityRegNotFound
	}

	delete(this.storage, name)
	return nil
}

func (this *EntityRegistry) GetEntity(filename string, pathToFile string) (*mdparser.Entity, error) {
	name := pathToFile + filename
	entity, found := this.storage[name]
	if !found { 
		return nil, ErrEntityRegNotFound
	}
	return entity, nil
}

func (this *EntityRegistry) HasEntity(filename string, pathToFile string) bool {
	_, found := this.storage[pathToFile + filename]
	return found
}

func (this *EntityRegistry) Flatten() []*mdparser.Entity {
	retEntities := make([]*mdparser.Entity, len(this.storage))
	i := 0
	for _, entity := range this.storage {
		retEntities[i] = entity
		i++
	}
	return retEntities
}

var (
	ErrCrawlerInitFailed 		= 	errors.New("Crawler: initialization failed")
	ErrCrawlerVaultIsNotADir 	= 	errors.New("Crawler: the given Vault path is not a directory")
	ErrCrawlerRootNotMarkdown 	= 	errors.New("Crawler: initialization failed")
	ErrCrawlerInvalidDepth 		= 	errors.New("Crawler: invalid depth")
	ErrCrawlerFileSearchInvalid = 	errors.New("Crawler: file search yielded an invalid result")
	ErrCrawlerNoSuchFile 		=	errors.New("Crawler: no such file is present in the vault")
)

type Crawler interface {
	Crawl() 			error
	Initialize() 		error
	ExportedEntities() 	EntityRegister
}

type EntityCrawler struct {
	VaultPath 				string 
	RootFilePath 			string
	MaxDepth				int
	
	linkParsers   			[]mdparser.LinkParser
	forbiddingParsers			[]mdparser.BooleanParser
	exportedEntities   		EntityRegister
	allEntitiesPaths		FileRegister
	logger 					log.Logger
}

func NewEntityCrawler(
	vaultPath string, 
	rootFilePath string, 
	maxDepth int, 
	entityRegister EntityRegister, 
	fileRegister FileRegister, 
	linkParsers []mdparser.LinkParser, 
	forbiddingParsers []mdparser.BooleanParser, 
	logger log.Logger) *EntityCrawler {
	return &EntityCrawler{
		VaultPath: vaultPath, 
		RootFilePath: rootFilePath,
		MaxDepth: maxDepth,
		exportedEntities: entityRegister,
		allEntitiesPaths: fileRegister,
		linkParsers: linkParsers,
		forbiddingParsers: forbiddingParsers,
		logger: logger,
	}
}

func (this *EntityCrawler) ExportedEntities() EntityRegister {
	return this.exportedEntities
}

func (this *EntityCrawler) Initialize() error {
	if this.MaxDepth < 0 {
		return ErrCrawlerInvalidDepth
	}

	fileInfo, err := os.Stat(this.VaultPath)
	if err != nil {
		err = fmt.Errorf("EntityCrawler.Initialize() error: %w", err)
		return err
	}

	if !fileInfo.IsDir() {
		return ErrCrawlerVaultIsNotADir
	}
	
	rootFullpath := this.RootFilePath
	_, rootStatErr := os.Stat(rootFullpath)
	if rootStatErr != nil {
		rootStatErr = fmt.Errorf("EntityCrawler.Initialize() rootStat error: %w", rootStatErr)
		return rootStatErr
	}

	rootExtension := fp.Ext(rootFullpath)
	if rootExtension != ".md" {
		return ErrCrawlerRootNotMarkdown
	}

	re := r.MustCompile(`^mat\.ndim`)
		
	fileProcessorFunc := func (path string, info os.FileInfo, err error) error {
		if err != nil {
			this.logger.Error(fmt.Sprintf("Got a problem on file %v : %v", fp.Join(fp.Dir(path), fp.Base(path)), err.Error()))
			return nil
		}

		fileName := fp.Join(fp.Dir(path), fp.Base(path))
		matched := re.MatchString(fileName)
		if matched {
			this.logger.Info(fmt.Sprintf("Found the file %v", fileName))
		} 

		if info.Mode().IsRegular() {
			addedFile := fileName
			this.allEntitiesPaths.AddFile(fp.Dir(path), fp.Base(path))
			this.logger.Info(fmt.Sprintf("Adding file: %v", addedFile))
		}
		// } else {
		// 	// missedFile := fp.Join(fp.Dir(path), fp.Base(path))
		// 	// this.logger.Warn(fmt.Sprintf("Not adding file: %v", missedFile))
		// }

		return nil
	}
	
	walkErr := fp.Walk(this.VaultPath, fileProcessorFunc)
	if walkErr != nil {
		walkErr = fmt.Errorf("EntityCrawler.Initialize() walk error: %w", walkErr)
		return walkErr 
	}
	return nil
}

func (this *EntityCrawler) findFile(link mdparser.Linker) (string, error) {
	ed, error := link.Link()
	if error != nil {
		error = fmt.Errorf("EntityCrawler.findFile() link error: %w", error)
		return "", error
	}

	paths, pathsError := this.allEntitiesPaths.GetPaths(ed.FullFileName())
	if pathsError != nil {
		pathsError = fmt.Errorf("EntityCrawler.findFile() get paths error for file %v: %w", ed.FullFileName(), pathsError)
		return "", pathsError
	}
	
	smallestI := -1
	smallestLen := -1
	for i, path := range paths {
		els := fp.SplitList(path)
		lenEls := len(els)
		if smallestI == -1 {
			smallestI = i
			smallestLen = lenEls
		} else {
			if lenEls < smallestLen {
				smallestI = i 
				smallestLen = lenEls
			}
		}
	}

	if smallestI == -1 {
		return "", ErrCrawlerFileSearchInvalid
	}
	return paths[smallestI], nil
}

func (this *EntityCrawler) digestEntity(currentEntity *mdparser.Entity) ([]*mdparser.Entity, error) {
	this.logger.Debug(fmt.Sprintf("digestEntity on entity %v\n", currentEntity.EntityDescription.FullPath()))
	newEntities := make([]*mdparser.Entity, 0)
	for _, link := range currentEntity.Links() {
		filePath, searchErr := this.findFile(link)
		if searchErr == nil {
			ed, _ := link.Link()
			ed.SetLocation(filePath)
			entityAlreadyExists := this.exportedEntities.HasEntity(ed.FullFileName(), ed.Location())
			if !entityAlreadyExists {
				newEntities = append(newEntities, mdparser.NewEntity(*ed))
			}
		} else { 
			ed, _ := link.Link()

			this.logger.Warn(fmt.Sprintf("one of the searched files is missing. File is %v\n. Error is %v", ed.FullPath(), searchErr.Error()))
		}
	}
	
	for _, entity := range newEntities {
		parsingError := mdparser.Parse(entity, this.linkParsers, this.forbiddingParsers)
		if parsingError != nil {
			parsingError = fmt.Errorf("EntityCrawler.digestEntity() parsingError: %w", parsingError)
			return nil, parsingError
		}
		addFileErr := this.exportedEntities.AddEntity(entity)
		if addFileErr != nil && addFileErr != ErrEntityRegTryingToAddExistingEntity {
			addFileErr = fmt.Errorf("EntityCrawler.digestEntity() error: %w", addFileErr)
			return nil, addFileErr
		}
	}
	return newEntities, nil
}

func (this *EntityCrawler) parseEntityTree(currentEntity *mdparser.Entity, maxDepth int, currentDepth int) error {
	if currentEntity.Forbidden() {
		return nil
	}

	newEntities, digestErr := this.digestEntity(currentEntity)
	if digestErr != nil {
		return digestErr
	}
	
	if currentDepth < maxDepth {
		for _, entity := range newEntities {
			parsingError := this.parseEntityTree(entity, maxDepth, currentDepth + 1)
			if parsingError != nil {
				parsingError = fmt.Errorf("EntityCrawler.parseEntityTree() error on file %v, currentDepth %v, maxDepth %v: %w", currentEntity.FullPath(), currentDepth, maxDepth, parsingError)
				return parsingError
			}
		}
	}

	// fmt.Printf("Crawler.parseEntityTree(): exited. currentDepth %v, maxDepth %v\n", currentDepth, maxDepth)

	return nil	
}

func (this *EntityCrawler) Crawl() error {
	currentDepth := 0
	rootFilePath, fullRootFileName := fp.Split(this.RootFilePath)
	rootFileExtension := fp.Ext(fullRootFileName)
	rootFileName := fullRootFileName[:len(fullRootFileName)-len(rootFileExtension)]

	this.logger.Info(fmt.Sprintf("rootFilePath = %v\nrootFileName = %v\nrootFileExtension = %v\n", rootFilePath, rootFileName, rootFileExtension))

	rootDescr := mdparser.NewEntityDescription(rootFilePath, rootFileName, rootFileExtension)
	rootEntity := mdparser.NewEntity(*rootDescr)
	parsingError := mdparser.Parse(rootEntity, this.linkParsers, this.forbiddingParsers)
	if parsingError != nil {
		parsingError = fmt.Errorf("EntityCrawler.Crawl() error: %w", parsingError)
		return parsingError
	}
	if !rootEntity.Forbidden() {
		if currentDepth < this.MaxDepth {
			treeParsingError := this.parseEntityTree(rootEntity, this.MaxDepth, currentDepth + 1)
			if treeParsingError != nil {
				treeParsingError = fmt.Errorf("EntityCrawler.Crawl() error: %w", treeParsingError)
				return treeParsingError
			}
			return nil
		}
		return fmt.Errorf("Max depth violated\n") 
	} else {
		return fmt.Errorf("Root is a forbidden file\n") 
	}
}
