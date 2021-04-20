SCRIPT_PATH=$(cd `dirname "${BASH_SOURCE[0]}"` && pwd)/`basename "${BASH_SOURCE[0]}"`
SD=$(dirname "$SCRIPT_PATH")
SRC_PATH=$SD/../src
DIST_PATH=$SD/../dist

APP_ID=tunnel

rm -rf $DIST_PATH
mkdir -p $DIST_PATH

cd $SRC_PATH
go mod tidy
go generate
                                        go build -v -o $DIST_PATH/$APP_ID-mac     main.go
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -v -o $DIST_PATH/$APP_ID-linux   main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -v -o $DIST_PATH/$APP_ID-win.exe main.go
# go build -v -o $DIST_PATH/$APP_ID-$APP_VERSION .
