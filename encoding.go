package tiktoken

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"errors"
	"sync"
)

var (
	vocabCL100K, vocabP50K, vocabR50K map[string]int
	//go:embed cl100k.gob
	cl100k []byte
	//go:embed p50k.gob
	p50k []byte
	//go:embed r50k.gob
	r50k []byte
)

const ENDOFTEXT string = "<|endoftext|>"
const FIM_PREFIX string = "<|fim_prefix|>"
const FIM_MIDDLE string = "<|fim_middle|>"
const FIM_SUFFIX string = "<|fim_suffix|>"
const ENDOFPROMPT string = "<|endofprompt|>"

const (
	MODEL_CL100K_BASE string = "cl100k_base"
	MODEL_P50K_BASE   string = "p50k_base"
	MODEL_P50K_EDIT   string = "p50k_edit"
	MODEL_R50K_BASE   string = "r50k_base"
)

var MODEL_TO_ENCODING = map[string]string{
	// chat
	"gpt-4":         MODEL_CL100K_BASE,
	"gpt-3.5-turbo": MODEL_CL100K_BASE,
	// text
	"text-davinci-003": MODEL_P50K_BASE,
	"text-davinci-002": MODEL_P50K_BASE,
	"text-davinci-001": MODEL_R50K_BASE,
	"text-curie-001":   MODEL_R50K_BASE,
	"text-babbage-001": MODEL_R50K_BASE,
	"text-ada-001":     MODEL_R50K_BASE,
	"davinci":          MODEL_R50K_BASE,
	"curie":            MODEL_R50K_BASE,
	"babbage":          MODEL_R50K_BASE,
	"ada":              MODEL_R50K_BASE,
	// code
	"code-davinci-002": MODEL_P50K_BASE,
	"code-davinci-001": MODEL_P50K_BASE,
	"code-cushman-002": MODEL_P50K_BASE,
	"code-cushman-001": MODEL_P50K_BASE,
	"davinci-codex":    MODEL_P50K_BASE,
	"cushman-codex":    MODEL_P50K_BASE,
	// edit
	"text-davinci-edit-001": MODEL_P50K_EDIT,
	"code-davinci-edit-001": MODEL_P50K_EDIT,
	// embeddings
	"text-embedding-ada-002": MODEL_CL100K_BASE,
	// old embeddings
	"text-similarity-davinci-001":  MODEL_R50K_BASE,
	"text-similarity-curie-001":    MODEL_R50K_BASE,
	"text-similarity-babbage-001":  MODEL_R50K_BASE,
	"text-similarity-ada-001":      MODEL_R50K_BASE,
	"text-search-davinci-doc-001":  MODEL_R50K_BASE,
	"text-search-curie-doc-001":    MODEL_R50K_BASE,
	"text-search-babbage-doc-001":  MODEL_R50K_BASE,
	"text-search-ada-doc-001":      MODEL_R50K_BASE,
	"code-search-babbage-code-001": MODEL_R50K_BASE,
	"code-search-ada-code-001":     MODEL_R50K_BASE,
	// open source
	"gpt2": "gpt2",
}

var MODEL_PREFIX_TO_ENCODING = map[string]string{
	// chat
	"gpt-4-":         MODEL_CL100K_BASE, // e.g., gpt-4-0314, etc., plus gpt-4-32k
	"gpt-3.5-turbo-": MODEL_CL100K_BASE, // e.g, gpt-3.5-turbo-0301, -0401, etc.
}

var encodingMap map[string]*Encoding
var l *sync.Mutex

func init() {
	gob.NewDecoder(bytes.NewReader(cl100k)).Decode(&vocabCL100K)
	gob.NewDecoder(bytes.NewReader(p50k)).Decode(&vocabP50K)
	gob.NewDecoder(bytes.NewReader(r50k)).Decode(&vocabR50K)
	encodingMap = make(map[string]*Encoding)
	l = &sync.Mutex{}
}

type Encoding struct {
	Name           string
	PatStr         string
	MergeableRanks map[string]int
	SpecialTokens  map[string]int
	ExplicitNVocab int
}

func getEncoding(encodingName string) (*Encoding, error) {
	l.Lock()
	defer l.Unlock()
	if encoding, ok := encodingMap[encodingName]; ok {
		return encoding, nil
	}
	initEncoding, err := initEncoding(encodingName)
	if err != nil {
		return nil, err
	}
	encodingMap[encodingName] = initEncoding
	return encodingMap[encodingName], nil
}

func initEncoding(encodingName string) (*Encoding, error) {
	var encoding *Encoding
	var err error
	switch encodingName {
	case MODEL_CL100K_BASE:
		encoding, err = cl100k_base()
	case MODEL_P50K_BASE:
		encoding, err = p50k_base()
	case MODEL_R50K_BASE:
		encoding, err = r50k_base()
	case MODEL_P50K_EDIT:
		encoding, err = p50k_edit()
	default:
		return nil, errors.New("Unknown encoding: " + encodingName)
	}
	if err == nil && len(encoding.MergeableRanks) == 0 {
		return nil, errors.New("Invalid vocabulary: " + encodingName)
	}
	return encoding, err
}

func cl100k_base() (*Encoding, error) {
	special_tokens := map[string]int{
		ENDOFTEXT:   100257,
		FIM_PREFIX:  100258,
		FIM_MIDDLE:  100259,
		FIM_SUFFIX:  100260,
		ENDOFPROMPT: 100276,
	}
	return &Encoding{
		Name:           MODEL_CL100K_BASE,
		PatStr:         `(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+`,
		MergeableRanks: vocabCL100K,
		SpecialTokens:  special_tokens,
	}, nil
}

func p50k_edit() (*Encoding, error) {
	special_tokens := map[string]int{ENDOFTEXT: 50256, FIM_PREFIX: 50281, FIM_MIDDLE: 50282, FIM_SUFFIX: 50283}
	return &Encoding{
		Name:           MODEL_P50K_EDIT,
		PatStr:         `'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`,
		MergeableRanks: vocabP50K,
		SpecialTokens:  special_tokens,
	}, nil
}

func p50k_base() (*Encoding, error) {
	special_tokens := map[string]int{ENDOFTEXT: 50256}

	// ExplicitNVocab := 50281
	// max_tokens := int(math.Max(float64(len(special_tokens)), float64(len(ranks))))

	// if len(special_tokens)+len(ranks) != max_tokens {
	// 	return nil, errors.New("special_tokens and ranks must be disjoint")
	// }

	return &Encoding{
		Name:           MODEL_P50K_BASE,
		PatStr:         `'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`,
		MergeableRanks: vocabP50K,
		SpecialTokens:  special_tokens,
		ExplicitNVocab: 50281,
	}, nil
}

func r50k_base() (*Encoding, error) {
	special_tokens := map[string]int{ENDOFTEXT: 50256}
	return &Encoding{
		Name:           MODEL_R50K_BASE,
		MergeableRanks: vocabR50K,
		PatStr:         `'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`,
		SpecialTokens:  special_tokens,
		ExplicitNVocab: 50257,
	}, nil
}
