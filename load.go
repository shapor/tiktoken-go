package tiktoken

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"errors"
)

var (
	//go:embed cl100k.gob
	cl100k []byte
	//go:embed p50k.gob
	p50k []byte
	//go:embed r50k.gob
	r50k          []byte
	embedded_maps = func() (s struct {
		Cl100k_base map[string]int
		P50k_base   map[string]int
		R50k_base   map[string]int
	}) {
		dec := gob.NewDecoder(bytes.NewReader(cl100k))
		if err := dec.Decode(&s.Cl100k_base); err != nil {
			panic(err)
		}
		dec = gob.NewDecoder(bytes.NewReader(p50k))
		if err := dec.Decode(&s.P50k_base); err != nil {
			panic(err)
		}
		dec = gob.NewDecoder(bytes.NewReader(r50k))
		if err := dec.Decode(&s.R50k_base); err != nil {
			panic(err)
		}
		return
	}()
)

type BpeLoader interface {
	LoadTiktokenBpe(tiktokenBpeFile string) (map[string]int, error)
}

type defaultBpeLoader struct{}

func (l *defaultBpeLoader) LoadTiktokenBpe(tiktokenBpeFile string) (map[string]int, error) {
	switch tiktokenBpeFile {
	case "https://openaipublic.blob.core.windows.net/encodings/cl100k_base.tiktoken":
		return embedded_maps.Cl100k_base, nil
	case "https://openaipublic.blob.core.windows.net/encodings/p50k_base.tiktoken":
		return embedded_maps.P50k_base, nil
	case "https://openaipublic.blob.core.windows.net/encodings/r50k_base.tiktoken":
		return embedded_maps.R50k_base, nil
	default:
		return nil, errors.New("Invalid vocabulary")
	}
}

func NewDefaultBpeLoader() BpeLoader {
	return &defaultBpeLoader{}
}
