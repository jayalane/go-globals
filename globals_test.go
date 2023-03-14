// -*- tab-width: 2 -*-

package globals

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
	"testing"
	"time"
)

var g Global

func writeStringToFile(fn string, content string) error {
	binaryFilename, err := os.Executable()
	if err != nil {
		panic(err)
	}
	filePath := path.Dir(binaryFilename) + "/"

	fmt.Println("Writing to", filePath, fn, ": content", content)
	err = os.MkdirAll(filePath, 0770)
	if err != nil {
		fmt.Println("Mkdir failed", filePath, err)
		return err
	}
	f, err := os.CreateTemp(filePath, fn+"tmp")
	if err != nil {
		fmt.Println("create temp failed", filePath, fn, err)
		return err
	}
	fmt.Println("got temp file", f.Name())
	// no defer because closing 1/2 thru

	total := len(content + "\n")
	nn := 0
	for {
		n, err := f.WriteString(content + "\n")
		if err != nil {
			g.Ml.La("error on write", fn, err)
			_ = f.Close()
			return err
		}
		nn = nn + n
		if nn >= total {
			break
		}
	}
	err = f.Close()
	if err != nil {
		fmt.Println("close file got err", fn, err)
	}
	// now move it into place
	err = os.Rename(f.Name(), filePath+fn)
	if err != nil {
		fmt.Println("rename file got err", fn, err)
		return err
	}
	cmd := exec.Command("/bin/cat", filePath+fn)
	o, err := cmd.Output()
	if err != nil {
		fmt.Println("cat got error:", err)
		return err
	}
	fmt.Println("cat got", string(o))
	cmd = exec.Command("/bin/ls", filePath+fn)
	o, err = cmd.Output()
	if err != nil {
		fmt.Println("ls got error:", err)
		return err
	}
	fmt.Println("ls got", string(o))
	return nil
}

func TestSignal(t *testing.T) {

	err := writeStringToFile("config.txt", `debugLevel = always
`)
	if err != nil {
		fmt.Println("Could not write string to file", err.Error())
		t.Fail()
		return
	}
	g = NewGlobal(`debugLevel = never`, false)

	fmt.Println("Before signal level is", g.Ml.GetLevel())

	err = writeStringToFile("config.txt", `debugLevel = network
`)

	if err != nil {
		fmt.Println("Could not write string to file", err.Error())
		t.Fail()
		return
	}

	pid := os.Getpid()

	proc, _ := os.FindProcess(pid)

	err = proc.Signal(syscall.SIGHUP)

	if err != nil {
		fmt.Println("Signal failed to deliver:", err)
	}
	time.Sleep(2 * time.Second)
	fmt.Println("After signal level is", g.Ml.GetLevel())
}
