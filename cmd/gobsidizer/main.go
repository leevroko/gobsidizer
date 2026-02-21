package main 

import (
	"fmt" 
	crawler 	"github.com/Yyote/gobsidizer/internal/entity_crawler"
	mdparser 	"github.com/Yyote/gobsidizer/internal/markdown_parser"
	log 		"github.com/Yyote/gobsidizer/internal/logger"
				"github.com/Yyote/gobsidizer/internal/packager"
)

func testCrawler() {
	pkg := packager.NewFilePackager("/home/leev/testVault/", 256, log.Logger(log.NewPrintLogger("WorkParser", log.ERROR)))

	er := crawler.NewEntityRegistry()
	fr := crawler.NewFileRegistry()
	
	mdlinkParser := mdparser.LinkParser(mdparser.NewWikiLinkParser())
	linkParsers := []mdparser.LinkParser{mdlinkParser}

	workForbiddenParser := mdparser.NewStringForbiddingParser("#Работа", log.Logger(log.NewPrintLogger("WorkParser", log.ERROR)))
	markedForbiddenParser := mdparser.NewStringForbiddingParser("#excluded", log.Logger(log.NewPrintLogger("ExcludedParser", log.ERROR)))

	forbiddingParsers := []mdparser.BooleanParser{
		workForbiddenParser,
		markedForbiddenParser,
	}

	ec := crawler.NewEntityCrawler("/home/leev/work/Vaults/", "/home/leev/work/Vaults/Obsidian/NumPy.md", 3, er, fr, linkParsers, forbiddingParsers, log.Logger(log.NewPrintLogger("Crawler", log.INFO))) // ec := crawler.NewEntityCrawler("/home/leev/work/Vaults/", "/home/leev/work/Vaults/Зенхаб.md", 0, er, fr, linkParsers, forbiddingParsers, log.Logger(log.NewPrintLogger("Crawler", log.INFO)))
	// ec := crawler.NewEntityCrawler("/home/leev/work/Vaults/", "/home/leev/work/Vaults/Нейросети.md", 1, er, fr, linkParsers, forbiddingParsers, log.Logger(log.NewPrintLogger("Crawler", log.INFO))) // ec := crawler.NewEntityCrawler("/home/leev/work/Vaults/", "/home/leev/work/Vaults/Зенхаб.md", 0, er, fr, linkParsers, forbiddingParsers, log.Logger(log.NewPrintLogger("Crawler", log.INFO)))
	// ec := crawler.NewEntityCrawler("/home/leev/Yandex.Disk/Vaults/", "/home/leev/Yandex.Disk/Vaults/Материалы по работе.md", 2, er, fr, linkParsers, forbiddingParsers, log.Logger(log.NewPrintLogger("Crawler", log.INFO)))
	// ec := crawler.NewEntityCrawler("/home/leev/Yandex.Disk/Vaults/", "/home/leev/Yandex.Disk/Vaults/Obsidian/golang OOP.md", 3, er, fr, linkParsers, forbiddingParsers, log.Logger(log.NewPrintLogger("Crawler", log.INFO)))
	initErr := ec.Initialize()

	if initErr != nil {
		fmt.Printf("Error: %v\n", initErr.Error())
		return
	}

	err := ec.Crawl()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	}

	pkgErr := pkg.Package(ec)
	if pkgErr != nil {
		fmt.Printf("Error: %v\n", pkgErr.Error())
	}

	fmt.Println("Success")
	// newEr, ok := ec.ExportedEntities().(*crawler.EntityRegistry)
	// if !ok {
	// 	fmt.Printf("Error: could not convert the interface\n")
	// 	return 
	// } 
	//
	// for key, value := range newEr.GetStorage() {
	// 	fmt.Printf("Key = %v, Value = %v\n", key, value)
	// }
}

func main () {
	testCrawler()
}
