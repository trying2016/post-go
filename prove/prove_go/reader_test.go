package post_go

import "testing"

func TestNewBatchingReader(t *testing.T) {
	err := ReadData("/Volumes/172.16.7.43/post_91ca4e37742c565193b0bbdb5a4f36b99a3e9ac8bc3c7d8e56e1b8bcb0c50b0e", 1024*1024, 32*1024*1024*1024, func(batch *Batch) bool {

		return true
	})
	if err != nil {
		t.Fatal(err)
	}
}
