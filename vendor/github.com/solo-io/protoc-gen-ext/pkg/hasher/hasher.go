package hasher

import (
	"hash"
)

type SafeHasher interface {
	Hash(hasher hash.Hash64) (uint64, error)
}
