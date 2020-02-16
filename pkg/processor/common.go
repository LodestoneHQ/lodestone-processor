package processor

import "os"

type CommonProcessor struct{}

func (c *CommonProcessor) IsEmptyFile(filepath string) bool {
	fi, err := os.Stat(filepath)
	if err != nil {
		return false //error occurred while attempting to read file info, just assume this file is not empty.
	}
	// get the size
	return fi.Size() == 0
}
