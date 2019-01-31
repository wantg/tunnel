cd test
if [[ "$1" == "-bench" ]]; then
   go test -v -bench=. -benchmem -run=Benchma*
else
   go test -v -bench=. -benchmem -timeout=0
fi
