package gouring

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type genericTestingT interface {
	assert.TestingT
	require.TestingT
}

func testNewIoUring(t genericTestingT, entries uint32, flags uint32) *IoUring {
	h, err := New(entries, flags)
	require.NoError(t, err)
	require.NotNil(t, h)
	return h
}

func testNewIoUringWithParams(t genericTestingT, entries uint32, p *IoUringParams) *IoUring {
	h, err := NewWithParams(entries, p)
	require.NoError(t, err)
	require.NotNil(t, h)
	return h
}

func TestRingWrapper(t *testing.T) {
	h := testNewIoUring(t, 256, 0)
	defer h.Close()

	// O_RDWR|O_CREATE|O_EXCL
	ftmp, err := os.CreateTemp(os.TempDir(), "test_iouring_ring_wrapper_*")
	require.NoError(t, err)
	fd := ftmp.Fd()

	var whatToWrite = [][]byte{
		[]byte("hello\n"),
		[]byte("\tworld\n\n"),
		[]byte("io_uring\t\t\n"),
		[]byte("nice!\n!!!\n\x00"),
	}
	var off uint64 = 0
	for _, bs := range whatToWrite {
		sqe := h.GetSqe()
		PrepWrite(sqe, int(fd), &bs[0], len(bs), off)
		sqe.Flags = IOSQE_IO_LINK
		off = off + uint64(len(bs))
	}
	submitted, err := h.SubmitAndWait(uint32(len(whatToWrite)))
	require.NoError(t, err)
	require.Equal(t, len(whatToWrite), int(submitted))

	var readed = make([]byte, 1024)
	nb, err := ftmp.Read(readed)
	assert.NoError(t, err)
	readed = readed[:nb]

	joined := bytes.Join(whatToWrite, []byte{})
	assert.Equal(t, joined, readed)
}
