if ! go build; then
  echo 'Could not build.'
  exit 1
fi

./clmerge \
  --local tests/conins/SameLine.A.java \
  --other tests/conins/SameLine.B.java \
  --base tests/conins/SameLine.X.java \
  --output /tmp/out.java <<< "
h1
m
u
" > /dev/null 2>&1
if diff /tmp/out.java tests/conins/SameLine.O.java; then
  echo "[OK] \c"
else
  echo "[!!] \c"
fi
echo "Appetite 1"

./clmerge \
  --local tests/conins/SameLine.A.java \
  --other tests/conins/SameLine.B.java \
  --base tests/conins/SameLine.X.java \
  --output /tmp/out.java <<< "
h4
m
u
" > /dev/null 2>&1
if diff /tmp/out.java tests/conins/SameLine.O.java; then
  echo "[OK] \c"
else
  echo "[!!] \c"
fi
echo "Appetite 4"

./clmerge \
  --local tests/conins/SameLine.A.java \
  --other tests/conins/SameLine.B.java \
  --base tests/conins/SameLine.X.java \
  --output /tmp/out.java <<< "
m
u
" > /dev/null 2>&1
if diff /tmp/out.java tests/conins/SameLine.O.h5.java; then
  echo "[OK] \c"
else
  echo "[!!] \c"
fi
echo "Appetite 5"
