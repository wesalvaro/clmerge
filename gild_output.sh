if ! go build; then
  echo "Did not build"
  exit 1
fi

./clmerge \
  --output=tests/1/o.py \
  --local=tests/1/a.py \
  --other=tests/1/b.py \
  --base=tests/1/x.py

./clmerge \
  --output=tests/2/o.go \
  --local=tests/2/a.go \
  --other=tests/2/b.go \
  --base=tests/2/x.go

./clmerge \
  --appetite=1 \
  --output=tests/3/SameLine.O.java \
   --local=tests/3/SameLine.A.java \
   --other=tests/3/SameLine.B.java \
    --base=tests/3/SameLine.X.java <<< "
m
u
"
