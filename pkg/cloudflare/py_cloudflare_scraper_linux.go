//go:build linux

package cloudflare

// #cgo pkg-config: python3
// #cgo CFLAGS : -I./ -I/usr/include/python3.10
// #cgo LDFLAGS: -L/usr/lib/python3.10/config-3.10-x86_64-linux-gnu -L/usr/lib -lpython3.10 -lpthread -ldl  -lutil -lm
// #include <Python.h>
import "C"
import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"
)

// /usr/include/python3.10/Python.h
func GetCloudFlareProtectedHTML(searchURL string) (string, error) {
	// context.c:56: warning: mpd_setminalloc: ignoring request to set MPD_MINALLOC a second time
	C.Py_Initialize()
	// check module
	moduleCString := C.CString("cloudscraper")
	defer C.free(unsafe.Pointer(moduleCString))
	// need pip3 install cloudscraper first
	cfModule := C.PyImport_ImportModule(moduleCString)
	if cfModule == nil {
		return "", errors.New("error when importing module, please install cloudscraper to your python first")
	}
	pycodeGo := fmt.Sprintf(`
import cloudscraper
scraper = cloudscraper.create_scraper()
print (scraper.get("%s").text)
		 `, searchURL)
	pycodeC := C.CString(pycodeGo)
	defer C.free(unsafe.Pointer(pycodeC))

	// Clone Stdout to origStdout.
	origStdout, err := syscall.Dup(syscall.Stdout)
	if err != nil {
		return "", fmt.Errorf("error when dup syscall.stdout: %w", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		return "", fmt.Errorf("error when open io pipe: %w", err)
	}

	// Clone the pipe's writer to the actual Stdout descriptor; from this point
	// on, writes to Stdout will go to w.
	if err = syscall.Dup2(int(w.Fd()), syscall.Stdout); err != nil {
		return "", fmt.Errorf("error when dup io pipe to stdout: %w", err)
	}

	// Background goroutine that drains the reading end of the pipe.
	out := make(chan []byte)
	go func() {
		var b bytes.Buffer
		io.Copy(&b, r)
		out <- b.Bytes()
	}()
	statusCode := C.PyRun_SimpleString(pycodeC)
	if statusCode != 0 {
		return "", fmt.Errorf("error running python with status code: %d", statusCode)
	}
	// close python first
	C.Py_Finalize()

	C.fflush(nil)
	w.Close()
	syscall.Close(syscall.Stdout)

	// Rendezvous with the reading goroutine.
	b := <-out

	// Restore original Stdout.
	syscall.Dup2(origStdout, syscall.Stdout)
	syscall.Close(origStdout)

	// fmt.Println("Captured:", string(b))

	return string(b), nil
}
