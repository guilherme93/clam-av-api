package clamav

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"file-scan-api/internal/clamav/domain"
	"file-scan-api/internal/config"
)

const (
	retryInterval = 5 * time.Second
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=charges
type Service interface {
	ScanFile(file io.Reader) (*domain.ScanFileResp, error)
	Wait() error
}

type clamAVService struct {
	socketType         string
	socketAddr         string
	initialWaitTimeout time.Duration
}

func NewService(cfg config.ClamAVConfig) Service {
	return &clamAVService{
		socketType:         "tcp",
		socketAddr:         cfg.Address,
		initialWaitTimeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
	}
}

func (c *clamAVService) Wait() error {
	startTime := time.Now()

	const (
		pingCommand    = "zPING\x00"
		pongResponseOK = "PONG\x00"
	)

	for {
		if time.Since(startTime) > c.initialWaitTimeout {
			return errors.New("ClamAV health check failed: timeout after 30 seconds")
		}

		conn, err := net.DialTimeout(c.socketType, c.socketAddr, time.Second)
		if err != nil {
			time.Sleep(retryInterval)

			continue
		}
		defer conn.Close()

		if _, err = conn.Write([]byte(pingCommand)); err != nil {
			time.Sleep(retryInterval)

			continue
		}

		response := make([]byte, 64)

		n, err := conn.Read(response)
		if err != nil {
			time.Sleep(retryInterval)

			continue
		}

		if string(response[:n]) == pongResponseOK {
			return nil
		}

		time.Sleep(retryInterval)
	}
}

/*
ScanFile connects to ClamAV through TCP and scans the file
Documentation: https://linux.die.net/man/8/clamd
It's recommended to prefix clamd commands with the letter z to indicate that the command will be delimited by a NULL
character and that clamd should continue reading command data until a NULL character is read.
The null delimiter assures that the complete command and its entire argument will be processed as a single command.
Alternatively commands may be prefixed with the letter n to use a newline character as the delimiter.

zINSTREAM:

	Scans a stream of data.
	The format of the chunk is: '<length><data>' where <length> is the size of the following data in bytes expressed as
	a 4 byte unsigned integer in network byte order and <data> is the actual chunk. Streaming is terminated by sending
	a zero-length chunk.
	Note: do not exceed StreamMaxLength as defined in clamd.conf.
*/
func (c *clamAVService) ScanFile(file io.Reader) (*domain.ScanFileResp, error) {
	const (
		inStreamCommand = "zINSTREAM\x00"
		bufferSize      = 2048
	)

	conn, err := net.Dial(c.socketType, c.socketAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClamAV: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(inStreamCommand))
	if err != nil {
		return nil, fmt.Errorf("failed to send INSTREAM command: %w", err)
	}

	buffer := make([]byte, bufferSize)

	for {
		bytesRead, readErr := file.Read(buffer)

		err = writeBytes(bytesRead, conn, buffer)
		if err != nil {
			return nil, err
		}

		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return nil, fmt.Errorf("file read error: %w", readErr)
		}
	}

	err = writeZeroChuckEndStream(conn)
	if err != nil {
		return nil, err
	}

	return readResponse(conn)
}

func readResponse(conn net.Conn) (*domain.ScanFileResp, error) {
	const (
		bufferSize = 2048
		clamAVOK   = "OK"
	)

	readBuffer := make([]byte, bufferSize)

	var fullResponse []byte

	for {
		bytesRead, readErr := conn.Read(readBuffer)
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}

			return nil, fmt.Errorf("failed to read response: %w", readErr)
		}

		fullResponse = append(fullResponse, readBuffer[:bytesRead-1]...)

		// Check for null terminator
		if bytesRead > 0 && readBuffer[bytesRead-1] == 0 {
			break
		}
	}

	response := strings.TrimSpace(string(fullResponse))
	if response == clamAVOK {
		return &domain.ScanFileResp{HasVirus: false}, nil
	}

	return &domain.ScanFileResp{HasVirus: true, VirusText: response}, nil
}

func writeZeroChuckEndStream(conn net.Conn) error {
	zeroChunk := make([]byte, 4) // 4 bytes of zero for chunk size

	_, err := conn.Write(zeroChunk)
	if err != nil {
		return fmt.Errorf("failed to send end-of-stream marker: %w", err)
	}

	return err
}

func writeBytes(bytesRead int, conn net.Conn, buffer []byte) error {
	if bytesRead <= 0 {
		return nil
	}

	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(bytesRead))

	_, errWrite := conn.Write(sizeBytes)
	if errWrite != nil {
		return fmt.Errorf("failed to write size: %w", errWrite)
	}

	_, errWrite = conn.Write(buffer[:bytesRead])
	if errWrite != nil {
		return fmt.Errorf("failed to write size: %w", errWrite)
	}

	return nil
}
