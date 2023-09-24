package post_go

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
)

// ReadBatch is a function that is called for each batch of data read from the file.
type ReadBatch func(batch *Batch) bool

type Batch struct {
	Data []byte
	Pos  uint64
}

type BatchingReader struct {
	reader      *os.File
	startingPos uint64
	pos         uint64
	batchSize   int
	totalSize   uint64
	tempData    []byte
}

func NewBatchingReader(reader *os.File, pos uint64, batchSize int, totalSize uint64) *BatchingReader {
	return &BatchingReader{
		reader:      reader,
		startingPos: pos,
		pos:         pos,
		batchSize:   batchSize,
		totalSize:   totalSize,
		tempData:    make([]byte, batchSize),
	}
}

func (r *BatchingReader) Next() (*Batch, error) {
	posInFile := r.pos - r.startingPos
	if posInFile >= r.totalSize {
		return nil, nil
	}
	remaining := r.totalSize - posInFile
	batchSize := r.batchSize
	if batchSize > int(remaining) {
		batchSize = int(remaining)
	}
	//data := make([]byte, batchSize)
	n, err := r.reader.Read(r.tempData[:batchSize])
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	batch := &Batch{
		Data: r.tempData[:n],
		Pos:  r.pos,
	}
	r.pos += uint64(n)
	return batch, nil
}

func PosFiles(datadir string) ([]os.FileInfo, error) {
	fileRe := regexp.MustCompile(`postdata_(\d+)\.bin`)
	entries, err := ioutil.ReadDir(datadir)
	if err != nil {
		return nil, err
	}
	var dirEntries []os.FileInfo
	for _, entry := range entries {
		if fileRe.MatchString(entry.Name()) {
			dirEntries = append(dirEntries, entry)
		}
	}
	sort.Slice(dirEntries, func(i, j int) bool {
		id1 := fileRe.FindStringSubmatch(dirEntries[i].Name())[1]
		id2 := fileRe.FindStringSubmatch(dirEntries[j].Name())[1]
		return id1 < id2
	})
	return dirEntries, nil
}

// ReadData reads all the data from the given directory and calls the given function for each batch of data read.
func ReadData(datadir string, batchSize int, fileSize uint64, fn ReadBatch) error {
	dirEntries, err := PosFiles(datadir)
	if err != nil {
		return err
	}
	var readers []*BatchingReader
	for id, entry := range dirEntries {
		pos := uint64(id) * fileSize
		file, err := os.Open(path.Join(datadir, entry.Name()))
		if err != nil {
			return err
		}
		fileInfo, err := file.Stat()
		if err != nil {
			return err
		}
		posFileSize := uint64(fileInfo.Size())
		if len(dirEntries) > id+1 && posFileSize != fileSize {
			log.Printf("invalid POS file, expected size: %d vs actual size: %d\n", fileSize, posFileSize)
		}
		reader := NewBatchingReader(file, pos, batchSize, posFileSize)
		readers = append(readers, reader)
	}
	defer func() {
		for _, reader := range readers {
			if reader.reader != nil {
				_ = reader.reader.Close()
			}
		}
	}()
	for _, reader := range readers {
		for {
			batch, err := reader.Next()
			if err != nil {
				return err
			}
			if batch == nil {
				break
			}
			if !fn(batch) {
				return nil
			}
		}
	}
	return nil
}

func ReadFrom(reader *os.File, batchSize int, maxSize uint64) ([]*Batch, error) {
	batchingReader := NewBatchingReader(reader, 0, batchSize, maxSize)
	var batches []*Batch
	for {
		batch, err := batchingReader.Next()
		if err != nil {
			return nil, err
		}
		if batch == nil {
			break
		}
		batches = append(batches, batch)
	}
	return batches, nil
}
