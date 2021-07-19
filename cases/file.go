package cases

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/xuperchain/xuperchain/service/pb"
)

type file struct {
	concurrency int
	path        string

	decoders    []*json.Decoder
}

func NewFile(config *Config) (Generator, error) {
	t := &file{
		concurrency: config.Concurrency,
		path: config.Args["path"],
		decoders: make([]*json.Decoder, config.Concurrency),
	}

	log.Printf("generate: type=file, concurrency=%d path=%s", t.concurrency, t.path)
	return t, nil
}

func (t *file) Init() error {
	files, err := ioutil.ReadDir(t.path)
	if err != nil {
		return fmt.Errorf("generate read dir error: %v", err)
	}

	if len(files) < t.concurrency {
		return fmt.Errorf("file number less than concurrency")
	}

	for i := 0; i < t.concurrency; i++ {
		filename := files[i].Name()
		file, err := os.Open(filepath.Join(t.path, filename))
		if err != nil {
			return fmt.Errorf("generate read file error: %v", err)
		}

		t.decoders[i] = json.NewDecoder(file)
	}

	return nil
}

func (t *file) Generate(id int) (*pb.Transaction, error) {
	var tx pb.Transaction
	err := t.decoders[id].Decode(&tx)
	if err != nil {
		return nil, fmt.Errorf("read tx file error: %v", err)
	}

	return &tx, nil
}

func init() {
	RegisterGenerator(CaseFile, NewFile)
}
