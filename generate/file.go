package generate

import (
	"encoding/json"
	"fmt"
	"github.com/xuperchain/xuperchain/service/pb"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type files struct {
	total       int
	concurrency int
	path        string

	queues      []chan *pb.Transaction
}

func NewFiles(config *Config) (Generator, error) {
	t := &files{
		total: config.Total,
		concurrency: config.Concurrency,
		path: config.Args["path"],
	}

	log.Printf("generate: type=file, total=%d, concurrency=%d path=%s", t.total, t.concurrency, t.path)
	return t, nil
}

func (t *files) Init() error {
	files, err := ioutil.ReadDir(t.path)
	if err != nil {
		return fmt.Errorf("generate read dir error: %v", err)
	}

	if len(files) < t.concurrency {
		return fmt.Errorf("file number less than concurrency")
	}

	queues := make([]chan *pb.Transaction, t.concurrency)
	for i := 0; i < t.concurrency; i++ {
		queues[i] = make(chan *pb.Transaction, t.concurrency)
	}

	for i := 0; i < t.concurrency; i++ {
		filename := files[i].Name()
		file, err := os.Open(filepath.Join(t.path, filename))
		if err != nil {
			return fmt.Errorf("generate read file error: %v", err)
		}

		go func(in io.Reader, queue chan *pb.Transaction) {
			if err := ReadFile(file, queue, -1); err != nil {
				log.Fatalf("read tx error: %v, filename=%s", err, filename)
			}
			close(queue)
		}(file, queues[i])
	}

	return nil
}

func (t *files) Generate(id int) (*pb.Transaction, error) {
	if tx, ok := <-t.queues[id]; ok {
		return tx, nil
	}

	return nil, fmt.Errorf("read empty tx from file")
}

func WriteFile(queue chan *pb.Transaction, w io.Writer, total int) error {
	count := 0
	e := json.NewEncoder(w)
	for tx := range queue {
		err := e.Encode(tx)
		if err != nil {
			return err
		}

		count++
		if total > 0 && count >= total {
			return nil
		}
	}

	return nil
}

func ReadFile(r io.Reader, queue chan *pb.Transaction, total int) error {
	count := 0
	dec := json.NewDecoder(r)
	for {
		var m pb.Transaction
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		queue <- &m

		count++
		if total > 0 && count >= total {
			return nil
		}
	}

	return nil
}

func init() {
	RegisterGenerator(BenchmarkFile, NewFiles)
}
