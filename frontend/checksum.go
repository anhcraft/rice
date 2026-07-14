package frontend

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
)

// Checksum computes a SHA-256 hash over the token stream that is invariant to
// whitespace and comments. It includes token type, identifier names, and
// decoded literal values; it excludes source positions and the eof sentinel.
func Checksum(tokens []Token) string {
	h := sha256.New()
	for _, tok := range tokens {
		if tok.tokenType == eof {
			continue
		}
		hashToken(h, &tok)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func hashToken(h hash.Hash, tok *Token) {
	h.Write([]byte{byte(tok.tokenType)})

	switch tok.tokenType {
	case identifier, stringLiteral:
		s := tok.literal.(string)
		writeBytes(h, []byte(s))
	case integerLiteral:
		v := tok.literal.(int64)
		_ = binary.Write(h, binary.LittleEndian, v)
	case floatLiteral:
		v := tok.literal.(float64)
		_ = binary.Write(h, binary.LittleEndian, v)
	case booleanLiteral:
		if tok.literal.(bool) {
			h.Write([]byte{1})
		} else {
			h.Write([]byte{0})
		}
	}
}

func writeBytes(h hash.Hash, b []byte) {
	var length [8]byte
	binary.LittleEndian.PutUint64(length[:], uint64(len(b)))
	h.Write(length[:])
	h.Write(b)
}
