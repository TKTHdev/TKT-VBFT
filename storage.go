package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type Storage struct {
	id         int
	stateFile  *os.File
	logFile    *os.File
	logWriter  *bufio.Writer
	logOffsets []int64
	async      bool
	inMemory   bool
}

func NewStorage(id int, async bool, inMemory bool) (*Storage, error) {
	prefix := ""
	if inMemory {
		prefix = "/dev/shm/"
	}
	stateFilename := fmt.Sprintf("%spbft_state_%d.bin", prefix, id)
	logFilename := fmt.Sprintf("%spbft_log_%d.bin", prefix, id)

	sFile, err := os.OpenFile(stateFilename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	lFile, err := os.OpenFile(logFilename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		sFile.Close()
		return nil, err
	}

	return &Storage{
		id:         id,
		stateFile:  sFile,
		logFile:    lFile,
		logWriter:  bufio.NewWriter(lFile),
		logOffsets: []int64{},
		async:      async,
		inMemory:   inMemory,
	}, nil
}

func (s *Storage) SaveState(view int) error {
	if _, err := s.stateFile.Seek(0, 0); err != nil {
		return err
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf[0:8], uint64(view))

	if _, err := s.stateFile.Write(buf); err != nil {
		return err
	}

	if !s.async {
		return s.stateFile.Sync()
	}
	return nil
}

func (s *Storage) LoadState() (int, error) {
	info, err := s.stateFile.Stat()
	if err != nil {
		return 0, err
	}
	if info.Size() == 0 {
		return 0, nil
	}

	if _, err := s.stateFile.Seek(0, 0); err != nil {
		return 0, err
	}

	buf := make([]byte, 8)
	if _, err := io.ReadFull(s.stateFile, buf); err != nil {
		return 0, err
	}

	view := int(binary.LittleEndian.Uint64(buf[0:8]))

	return view, nil
}

func (s *Storage) AppendEntry(entry LogEntry) error {
	offset, err := s.logFile.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	s.logOffsets = append(s.logOffsets, offset)

	if err := binary.Write(s.logWriter, binary.LittleEndian, int64(entry.View)); err != nil {
		return err
	}
	cmdLen := int64(len(entry.Command))
	if err := binary.Write(s.logWriter, binary.LittleEndian, cmdLen); err != nil {
		return err
	}
	// Write Command
	if _, err := s.logWriter.Write(entry.Command); err != nil {
		return err
	}

	if err := s.logWriter.Flush(); err != nil {
		return err
	}
	if !s.async {
		return s.logFile.Sync()
	}
	return nil
}

func (s *Storage) LoadLog() ([]LogEntry, error) {
	if _, err := s.logFile.Seek(0, 0); err != nil {
		return nil, err
	}

	var logs []LogEntry
	s.logOffsets = []int64{}

	reader := bufio.NewReader(s.logFile)
	offset := int64(0)

	for {
		startOffset := offset

		var view int64
		err := binary.Read(reader, binary.LittleEndian, &view)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		offset += 8

		var cmdLen int64
		err = binary.Read(reader, binary.LittleEndian, &cmdLen)
		if err != nil {
			return nil, err
		}
		offset += 8

		cmd := make([]byte, cmdLen)
		if _, err := io.ReadFull(reader, cmd); err != nil {
			return nil, err
		}
		offset += cmdLen

		s.logOffsets = append(s.logOffsets, startOffset)
		logs = append(logs, LogEntry{
			View:    int(view),
			Command: cmd,
		})
	}

	return logs, nil
}

func (s *Storage) Close() error {
	s.logWriter.Flush()
	s.stateFile.Close()
	return s.logFile.Close()
}
