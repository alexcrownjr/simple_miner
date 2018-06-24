package simple_miner

import (
	"encoding/binary"
	"fmt"
	"crypto/sha256"
)

// Hash concatentates a message and a nonce and generate a hash
// miners would need to call this method
func Hash(msg string, nonce uint64) uint64 {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s %d", msg, nonce)))
	return binary.BigEndian.Uint64(hasher.Sum(nil))
}