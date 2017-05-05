package deluge

import (
	"github.com/ofux/deluge-dsl/parser"
	log "github.com/sirupsen/logrus"
)

func PrintParserErrors(errors []parser.ParseError) {
	log.Errorln("Syntax error:")
	for _, err := range errors {
		log.Errorf("\t%s (line %d, col %d)\n", err.Message, err.Line, err.Column)
	}
}
