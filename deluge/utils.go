package deluge

import (
	"fmt"
	"github.com/ofux/deluge-dsl/parser"
	log "github.com/sirupsen/logrus"
)

func PrintParserErrors(errors []parser.ParseError) {
	log.Errorln(SPrintParserErrors(errors))
}

func SPrintParserErrors(errors []parser.ParseError) string {
	msg := "Syntax error:\n"
	for _, err := range errors {
		msg += fmt.Sprintf("\t%s (line %d, col %d)\n", err.Message, err.Line, err.Column)
	}
	return msg
}
