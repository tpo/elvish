package daemon

import (
	"syscall"
	"testing"
	"time"

	"src.elv.sh/pkg/daemon/client"
	"src.elv.sh/pkg/daemon/internal/api"
	"src.elv.sh/pkg/prog"
	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/store/storetest"
	"src.elv.sh/pkg/testutil"
)

func TestDaemon(t *testing.T) {
	// Set up filesystem.
	testutil.InTempDir(t)

	// Set up server.
	serverDone := make(chan struct{})
	go func() {
		Serve("sock", "db")
		close(serverDone)
	}()
	defer func() { <-serverDone }()

	// Set up client.
	client := client.NewClient("sock")
	defer client.Close()
	for i := 0; i < 100; i++ {
		client.ResetConn()
		_, err := client.Version()
		if err == nil {
			break
		} else if i == 99 {
			t.Fatal("Failed to connect after 1s")
		}
		time.Sleep(testutil.ScaledMs(10))
	}

	// Server state requests.
	gotVersion, err := client.Version()
	if gotVersion != api.Version || err != nil {
		t.Errorf(".Version() -> (%v, %v), want (%v, nil)", gotVersion, err, api.Version)
	}

	gotPid, err := client.Pid()
	wantPid := syscall.Getpid()
	if gotPid != wantPid || err != nil {
		t.Errorf(".Pid() -> (%v, %v), want (%v, nil)", gotPid, err, wantPid)
	}

	// Store requests.
	storetest.TestCmd(t, client)
	storetest.TestDir(t, client)
	storetest.TestSharedVar(t, client)
}

func TestProgram_SpuriousArgument(t *testing.T) {
	f := Setup(t)

	exit := prog.Run(f.Fds(), Elvish("-daemon", "x"), Program)

	TestError(t, f, exit, "arguments are not allowed with -daemon")
}
