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

run() {
  local name="$1"
  ./clmerge --local "$2" --other "$3" --base "$4" \
    --output /tmp/clmerge.out > /dev/null 2>&1
  verify "$name" "$5"
}

run_sameline() {
  local name="$1"
  local output="$2"
  run "$name" \
    tests/conins/SameLine.A.java \
    tests/conins/SameLine.B.java \
    tests/conins/SameLine.X.java \
    "tests/conins/SameLine.O$output.java"
}

run "no-conflict" tests/a.py tests/b.py tests/x.py tests/o.py
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
"
