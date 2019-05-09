#!/bin/bash

runtime=$(cat ./runtime.js)

echo "
${runtime}
"


cat > runtime.go <<EOF
package linker

const runtime = \`
${runtime}
\`
EOF
