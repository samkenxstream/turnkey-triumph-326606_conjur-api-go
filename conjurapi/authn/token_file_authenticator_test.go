package authn

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
	"github.com/spf13/afero"
)

func TestTokenFileAuthenticator_RefreshToken(t *testing.T) {
	AppFS = afero.NewMemMapFs()

	Convey("Given existent token filename", t, func() {
		token_file, _ := afero.TempFile(AppFS, "", "existent-token-file")
		token_file_name := token_file.Name()
		token_file_contents := "token-from-file-contents"
		token_file.Write([]byte(token_file_contents))
		token_file.Close()
		defer AppFS.Remove(token_file_name)

		Convey("Return the token from the file", func() {
			authenticator := TokenFileAuthenticator{
				TokenFile: token_file_name,
			}

			token, err := authenticator.RefreshToken()

			So(err, ShouldBeNil)
			So(string(token), ShouldEqual, "token-from-file-contents")
		})
	})

	Convey("Given an eventually existent token filename", t, func() {
		token_file, _ := afero.TempFile(AppFS, "", "existent-token-file")
		token_file_name := token_file.Name()

		token_file_contents := "token-from-file-contents"
		AppFS.Remove(token_file_name)
		go func() {
			afero.WriteFile(AppFS, token_file_name, []byte(token_file_contents), 0600)
		}()
		defer AppFS.Remove(token_file_name)

		Convey("Return the token from the file", func() {
			authenticator := TokenFileAuthenticator{
				TokenFile:   token_file_name,
				MaxWaitTime: 500 * time.Millisecond,
			}

			token, err := authenticator.RefreshToken()

			So(err, ShouldBeNil)
			So(string(token), ShouldEqual, "token-from-file-contents")
		})
	})

	Convey("Given a non-existent token filename", t, func() {
		token_file := "/path/to/non-existent-token-file"

		Convey("Return nil with error", func() {
			authenticator := TokenFileAuthenticator{
				TokenFile: token_file,
			}

			token, err := authenticator.RefreshToken()

			So(token, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Operation waitForTextFile timed out.")
		})
	})
}

func TestTokenFileAuthenticator_NeedsTokenRefresh(t *testing.T) {
	AppFS = afero.NewMemMapFs()

	Convey("Given existent token filename", t, func() {
		token_file, _ := afero.TempFile(AppFS, "", "existent-token-file")
		token_file_name := token_file.Name()
		token_file_contents := "token-from-file-contents"
		token_file.Write([]byte(token_file_contents))
		defer AppFS.Remove(token_file_name)

		Convey("Return true for recently modified file", func() {
			authenticator := TokenFileAuthenticator{
				TokenFile: token_file_name,
			}
			authenticator.RefreshToken()

			time.Sleep(1000 * time.Millisecond)
			token_file.Write([]byte("recent modification"))

			So(authenticator.NeedsTokenRefresh(), ShouldBeTrue)
		})

		Convey("Return false for unmodified file", func() {
			authenticator := TokenFileAuthenticator{
				TokenFile: token_file_name,
			}
			authenticator.RefreshToken()

			So(authenticator.NeedsTokenRefresh(), ShouldBeFalse)
		})
	})
}
