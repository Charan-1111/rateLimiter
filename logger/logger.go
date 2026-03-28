package logger

import (
	"os"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type AsyncWriter struct {
	ch     chan []byte
	writer *os.File
	quit   chan struct{}
	wg     sync.WaitGroup
}

func NewAsyncWriter(w *os.File, bufferSize int) *AsyncWriter {
	aw := &AsyncWriter{
		ch:     make(chan []byte, bufferSize),
		writer: w,
		quit:   make(chan struct{}),
	}
	aw.wg.Add(1)
	go aw.run()
	return aw
}

func (aw *AsyncWriter) Write(p []byte) (n int, err error) {
	select {
	case aw.ch <- append([]byte(nil), p...):
		return len(p), nil
	default:
		// Buffer full: drop log (non-blocking)
		return 0, nil
	}
}

func (aw *AsyncWriter) Close() error {
	close(aw.quit)
	aw.wg.Wait() // wait until background writer finishes
	return nil
}

func (aw *AsyncWriter) run() {
	defer aw.wg.Done()
	for {
		select {
		case p := <-aw.ch:
			aw.writer.Write(p)
		case <-aw.quit:
			// Drain remaining logs
			for {
				select {
				case p := <-aw.ch:
					aw.writer.Write(p)
				default:
					return
				}
			}
		}
	}
}



func InitLogger() (zerolog.Logger, func()) {
	asyncWriter := NewAsyncWriter(os.Stdout, 1024)
	log := zerolog.New(asyncWriter).With().Timestamp().Logger()

	closer := func() {
		asyncWriter.Close()
	}

	return log, closer
}

func GetRequestLogger(c *fiber.Ctx, baseLogger zerolog.Logger) zerolog.Logger {
	requestID := c.GetRespHeader("X-Request-ID")
	return baseLogger.With().Str("request_id", requestID).Logger()
}
