package parser

import "sync"

var parserPool = sync.Pool{
	New: func() any {
		return new(protoParserImpl)
	},
}

func parseWithPool(lexer *protoLex) {
	parser := parserPool.Get().(*protoParserImpl)
	defer func() {
		*parser = protoParserImpl{}
		parserPool.Put(parser)
	}()

	parser.Parse(lexer)
}
