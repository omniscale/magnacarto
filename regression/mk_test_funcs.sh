DEST=regression_cases_test.go
cat <<EOF > $DEST
package regression

import (
    "testing"
)

EOF


ls -1 cases | perl -ne 's/\n$//g; $fname = $_; $fname =~ s/-/_/g; print "func Test_$fname(t *testing.T) { testIt(t, testCase{Name: \"$_\"}) }\n"' | gofmt >> $DEST