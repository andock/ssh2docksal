package ssh2docksal

import (
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	. "github.com/smartystreets/goconvey/convey"
)

func TestServer_CheckConfig(t *testing.T) {
	log.SetHandler(text.New(os.Stderr))

	Convey("Testing Server.CheckConfig", t, FailureContinues, func() {
		// FIXME: check with a script
		server, err := NewServer()
		So(err, ShouldBeNil)
		server.AllowedImages = []string{"alpine", "ubuntu:trusty", "abcde123"}

	})
}
