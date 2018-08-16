if ! go build; then
  echo 'Could not build.'
  exit 1
fi

verify() {
  local name="$1"
  local expected="$2"
  if diff /tmp/clmerge.out "$expected"; then
    echo "[OK] \c"
  else
    echo "[!!] \c"
  fi
  echo "$name"
}

run_sameline() {
  local name="$1"
  local output="$2"
  ./clmerge \
    --local tests/conins/SameLine.A.java \
    --other tests/conins/SameLine.B.java \
    --base tests/conins/SameLine.X.java \
    --output /tmp/clmerge.out > /dev/null 2>&1
  verify "$name" "tests/conins/SameLine.O$output.java"
}

run_sameline "Appetite 1" "" <<< "
h1
m
u
"
run_sameline "Appetite 4" "" <<< "
h4
m
u
"
run_sameline "Appetite 5" ".h5" <<< "
h5
m
u
"
