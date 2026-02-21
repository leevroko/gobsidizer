package markdown_parser

import (
	"fmt"
	"unicode/utf8"

	log "github.com/Yyote/gobsidizer/internal/logger"
)

type StringForbiddingParser struct {
	forbiddenString			[]rune
	forbiddenStringLength 	int
	
	currentString			[]rune
	lastWrittenPosition		int
	logger 					log.Logger
}

func NewStringForbiddingParser(forbiddenString string, logger log.Logger) *StringForbiddingParser {
	return &StringForbiddingParser{
		forbiddenString: []rune(forbiddenString),
		forbiddenStringLength: utf8.RuneCountInString(forbiddenString),
		currentString: make([]rune, utf8.RuneCountInString(forbiddenString)),
		lastWrittenPosition: -1,
		logger: logger,
	}
}

func (this *StringForbiddingParser) Parse(newRune rune) bool {
	if this.lastWrittenPosition == this.forbiddenStringLength - 1 {
		for i := 0; i < this.forbiddenStringLength - 1; i++ {
			this.currentString[i] = this.currentString[i+1]
		}
		this.currentString[this.lastWrittenPosition] = newRune
	} else {
		this.lastWrittenPosition += 1
		this.currentString[this.lastWrittenPosition] = newRune
	}

	currentString := string(this.currentString)
	forbiddenString := string(this.forbiddenString)

	this.logger.Debug(fmt.Sprintf("currentString = %v, forbiddenString = %v\n", currentString, forbiddenString))

	return currentString == forbiddenString
}
