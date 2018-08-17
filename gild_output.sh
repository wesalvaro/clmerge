if ! go build; then
  echo "Did not build"
  exit 1
fi

./clmerge \
  --output=tests/o.py \
  --local=tests/a.py \
  --other=tests/b.py \
  --base=tests/x.py

./clmerge \
  --appetite=1 \
  --output=tests/conins/SameLine.O.java \
   --local=tests/conins/SameLine.A.java \
   --other=tests/conins/SameLine.B.java \
    --base=tests/conins/SameLine.X.java <<< "
m
u
"
