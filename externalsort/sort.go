//go:build !solution

package externalsort

import (
	"bytes"
	"container/heap"
	"errors"
	"io"
	"os"
	"sort"
	"strings"
)

type LineReaderImpl struct {
	inStream          io.Reader
	inStreamExhausted bool
	readBuf           bytes.Buffer
}

var _ LineReader = (*LineReaderImpl)(nil)

func (lr *LineReaderImpl) ReadLine() (string, error) {
	line, err := lr.readBuf.ReadString('\n')
	if err == nil {
		return line[:len(line)-1], nil
	}
	if !errors.Is(err, io.EOF) {
		return line, err
	}
	if lr.inStreamExhausted {
		return line, err
	}

	inBytes := make([]byte, 1024)
LoopUntilDelim:
	for {
		readN, err := lr.inStream.Read(inBytes)
		lr.readBuf.Write(inBytes[:readN])
		if errors.Is(err, io.EOF) {
			lr.inStreamExhausted = true
			break LoopUntilDelim
		}
		if err != nil {
			return "", err
		}
		for _, b := range inBytes[:readN] {
			if b == '\n' {
				break LoopUntilDelim
			}
		}
	}

	next, err := lr.ReadLine()
	return line + next, err
}

func NewReader(r io.Reader) LineReader {
	return &LineReaderImpl{inStream: r}
}

type LineWriterImpl struct {
	outStream io.Writer
}

var _ LineWriter = (*LineWriterImpl)(nil)

func (lw *LineWriterImpl) Write(l string) error {
	for _, line := range strings.Split(l, "\n") {
		_, err := lw.outStream.Write([]byte(line + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func NewWriter(w io.Writer) LineWriter {
	return &LineWriterImpl{outStream: w}
}

type Elem struct {
	Value     string
	ReaderIdx int
}

type ElemHeap []Elem

func (s *ElemHeap) Len() int {
	return len(*s)
}

func (s *ElemHeap) Less(i, j int) bool {
	return (*s)[i].Value < (*s)[j].Value
}

func (s *ElemHeap) Swap(i, j int) {
	(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
}

func (s *ElemHeap) Push(x any) {
	*s = append(*s, x.(Elem))
}

func (s *ElemHeap) Pop() any {
	size := len(*s)
	last := (*s)[size-1]
	*s = (*s)[:size-1]
	return last
}

var _ heap.Interface = (*ElemHeap)(nil)

func Merge(w LineWriter, readers ...LineReader) error {
	minHeap := &ElemHeap{}
	heap.Init(minHeap)
	exhausted := make(map[int]bool)

	for i, r := range readers {
		line, err := r.ReadLine()
		if err == nil {
			heap.Push(minHeap, Elem{Value: line, ReaderIdx: i})
			continue
		}
		if !errors.Is(err, io.EOF) {
			return err
		}
		if len(line) > 0 {
			heap.Push(minHeap, Elem{Value: line, ReaderIdx: i})
		}
		exhausted[i] = true
	}

	for minHeap.Len() > 0 {
		minimum := heap.Pop(minHeap).(Elem)
		err := w.Write(minimum.Value)
		if err != nil {
			return err
		}
		if exhausted[minimum.ReaderIdx] {
			continue
		}
		line, err := readers[minimum.ReaderIdx].ReadLine()
		if err == nil {
			heap.Push(minHeap, Elem{Value: line, ReaderIdx: minimum.ReaderIdx})
			continue
		}
		if !errors.Is(err, io.EOF) {
			return err
		}
		if len(line) > 0 {
			heap.Push(minHeap, Elem{Value: line, ReaderIdx: minimum.ReaderIdx})
		}
		exhausted[minimum.ReaderIdx] = true
	}
	return nil
}

func Sort(w io.Writer, in ...string) error {
	readers := make([]LineReader, 0, len(in))
	for _, path := range in {
		if err := sortFile(path); err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()

		readers = append(readers, NewReader(f))
	}
	return Merge(NewWriter(w), readers...)
}

func sortFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	lineReader := NewReader(f)
	lines := make([]string, 0)
	for {
		line, err := lineReader.ReadLine()
		if err == nil {
			lines = append(lines, line)
			continue
		}
		if !errors.Is(err, io.EOF) {
			return err
		}
		if len(line) > 0 {
			lines = append(lines, line)
		}
		break
	}

	if _, err = f.Seek(0, 0); err != nil {
		return err
	}

	sort.Strings(lines)
	for i, line := range lines {
		if i == 0 {
			_, err = f.WriteString(line)
		} else {
			_, err = f.WriteString("\n" + line)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
