package sonzai

import (
	"encoding/json"
	"fmt"
)

const DefaultMaxChunkSize = 256 * 1024

type sseChunkEnvelope struct {
	Chunk *sseChunkMeta `json:"__chunk"`
	Data  string        `json:"data"`
}

type sseChunkMeta struct {
	Index int `json:"index"`
	Total int `json:"total"`
}

// ChunkPayload splits a JSON payload into multiple SSE-formatted lines when it
// exceeds maxChunkSize bytes. Each returned []byte is a complete "data: ...\n\n"
// SSE frame. If the payload fits in a single chunk, a single un-enveloped frame
// is returned (backward compatible with non-chunked consumers).
// Pass 0 for maxChunkSize to use DefaultMaxChunkSize.
func ChunkPayload(payload json.RawMessage, maxChunkSize int) [][]byte {
	if maxChunkSize <= 0 {
		maxChunkSize = DefaultMaxChunkSize
	}

	raw := string(payload)
	if len(raw) <= maxChunkSize {
		return [][]byte{[]byte(fmt.Sprintf("data: %s\n\n", raw))}
	}

	total := (len(raw) + maxChunkSize - 1) / maxChunkSize
	frames := make([][]byte, 0, total)

	for i := 0; i < total; i++ {
		start := i * maxChunkSize
		end := start + maxChunkSize
		if end > len(raw) {
			end = len(raw)
		}

		env := sseChunkEnvelope{
			Chunk: &sseChunkMeta{Index: i, Total: total},
			Data:  raw[start:end],
		}
		b, _ := json.Marshal(env)
		frames = append(frames, []byte(fmt.Sprintf("data: %s\n\n", b)))
	}

	return frames
}
