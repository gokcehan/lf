#!/bin/sh
# This script is should be removed after we
# require Go 1.19+ and the `man.sh` script
# is updated to work with the new `go doc`
# output format.
if (go version | grep -E 'go1\.1[4-8]' > /dev/null);then
    exit 0
fi

cat <<'EOF'
`go generate` scripts in this repo require the
`go` binary in the PATH to be between Go 1.14
and Go 1.18. Currenlty, `go version` returns:
EOF
printf "  %s\n\n"  "$(go version)"

cat <<'EOF'
If Go 1.18 binary is in /lib/go-1.18/bin (for
example), you can use the following command:
  env PATH="/lib/go-1.18/bin:$PATH" go generate
EOF

exit 1
