package generate

import (
	"encoding/json"
	"github.com/xuperchain/xuperchain/service/pb"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type file struct {
	total       int
	concurrency int
	path        string
}

func NewFile(config *Config) (Generator, error) {
	t := &file{
		total: config.Total,
		concurrency: config.Concurrency,
		path: config.Args["path"],
	}

	log.Printf("generate: type=file, total=%d, concurrency=%d path=%s", t.total, t.concurrency, t.path)
	return t, nil
}

func (t *file) Init() error {
	return nil
}

func (t *file) Generate() []chan *pb.Transaction {
	files, err := ioutil.ReadDir(t.path)
	if err != nil {
		log.Fatalf("generate read dir error: %v, path=%s", err, t.path)
	}

	if len(files) < t.concurrency {
		log.Fatalf("file number less than concurrency")
	}

	queues := make([]chan *pb.Transaction, t.concurrency)
	for i := 0; i < t.concurrency; i++ {
		queues[i] = make(chan *pb.Transaction, t.concurrency)
	}

	for i := 0; i < t.concurrency; i++ {
		filename := files[i].Name()
		file, err := os.Open(filepath.Join(t.path, filename))
		if err != nil {
			log.Fatalf("generate read file error: %v, filename=%s", err, filename)
		}

		go func(in io.Reader, queue chan *pb.Transaction) {
			if err := ReadFile(file, queue); err != nil {
				log.Fatalf("read tx error: %v, filename=%s", err, filename)
			}
			close(queue)
		}(file, queues[i])
	}

	return queues
}

func WriteFile(queue chan *pb.Transaction, out io.Writer) error {
	e := json.NewEncoder(out)
	for tx := range queue {
		err := e.Encode(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadFile(in io.Reader, queue chan *pb.Transaction) error {
	dec := json.NewDecoder(in)
	for {
		var m pb.Transaction
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		queue <- &m
	}

	return nil
}