package errors
import("errors";"testing")
func TestCodeOfReadsWrappedPublicCode(t *testing.T){cause:=errors.New("missing");if CodeOf(Wrap(NotFound,"load",cause))!=NotFound{t.Fatal("public code lost")}}
