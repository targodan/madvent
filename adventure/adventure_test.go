package adventure

import (
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	use, _ := os.LookupEnv("LOG_DURING_TESTS")
	if use != "" && (use[0] == 'y' || use[0] == 'Y') {
		log.SetLevel(log.DebugLevel)
	}
}

const adventExecutable = "advent"

func TestInitialMessage(t *testing.T) {
	Convey("Starting a new advent the initial message should be read.", t, func() {
		adv, err := New(adventExecutable)
		So(err, ShouldBeNil)
		defer adv.Close()

		out, _, err := adv.Start()
		So(err, ShouldBeNil)

		var text string
		select {
		case text = <-out:
		case <-time.After(5000 * time.Millisecond):
			t.Error("Read from channel timed out.")
		}
		So(text, ShouldEqual, `Welcome to Adventure!!  Would you like instructions?`)

		// Only one output expected
		text = ""
		select {
		case text = <-out:
			t.Error("Channel should not have any more messages!")
		case <-time.After(100 * time.Millisecond):
		}
		So(text, ShouldEqual, "")
	})
}

func TestStartsWithDelimiter(t *testing.T) {
	Convey("If it starts with the delimiter", t, func() {
		text := []byte("> test 123")
		delim := []byte("> ")
		Convey("it should return true.", func() {
			So(startsWithDelimiter(text, delim), ShouldBeTrue)
		})
	})
	Convey("If it does not start with the delimiter", t, func() {
		text := []byte(">")
		delim := []byte("> ")
		Convey("it should return false.", func() {
			So(startsWithDelimiter(text, delim), ShouldBeFalse)
		})
	})
}
