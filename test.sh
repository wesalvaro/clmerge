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
    tests/3/SameLine.A.java \
    tests/3/SameLine.B.java \
    tests/3/SameLine.X.java \
    "tests/3/SameLine.O$output.java"
}

f="tests/1"
run "no-conflict" "$f/a.py" "$f/b.py" "$f/x.py" "$f/o.py"

f="tests/2"
run "Too Small" "$f/a.go" "$f/b.go" "$f/x.go" "$f/o.go" <<< ""

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
run_sameline "Always mark" ".mbang" <<< "
h1
m!
"
