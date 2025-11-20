package lsp

import (
	"bytes"
	"strings"
	"sync"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"

	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
)

// TODO(skewb1k): make local.
var documents = make(map[protocol.DocumentUri]string)
var documentsMu sync.RWMutex

func NewServer() *server.Server {
	handler := protocol.Handler{
		Initialize:             initialize,
		Initialized:            initialized,
		Shutdown:               shutdown,
		TextDocumentDidOpen:    didOpen,
		TextDocumentDidChange:  didChange,
		TextDocumentDidClose:   didClose,
		TextDocumentFormatting: formatting,
	}

	return server.NewServer(&handler, "", false)
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	return protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: ptrTo(true),
				Change:    ptrTo(protocol.TextDocumentSyncKindFull),
			},
			DocumentFormattingProvider: true,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	return nil
}

func didOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	content := params.TextDocument.Text

	documentsMu.Lock()
	documents[uri] = content
	documentsMu.Unlock()
	return nil
}

func didChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI
	text := params.ContentChanges[0].(protocol.TextDocumentContentChangeEventWhole).Text

	documentsMu.Lock()
	documents[uri] = text
	documentsMu.Unlock()

	return nil
}

func didClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	uri := params.TextDocument.URI
	documentsMu.Lock()
	delete(documents, uri)
	documentsMu.Unlock()
	return nil
}

func formatting(context *glsp.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	uri := params.TextDocument.URI

	documentsMu.Lock()
	defer documentsMu.Unlock()

	text := documents[uri]

	parsed, err := parser.ParseBytes([]byte(text), uri)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := format.FormatFile(&buf, parsed); err != nil {
		return nil, err
	}

	formatted := buf.String()

	documents[uri] = formatted

	lines := strings.Split(formatted, "\n")
	lastLine := uint32(len(lines) - 1)
	lastChar := uint32(len(lines[lastLine]))

	return []protocol.TextEdit{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: lastLine, Character: lastChar},
			},
			NewText: formatted,
		},
	}, nil
}

func ptrTo[T any](v T) *T {
	return &v
}
