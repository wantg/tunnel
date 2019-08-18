APP_ID=tunnel
BIN_PATH=./bin

mkdir -p $BIN_PATH

ln -sf ../config.json $BIN_PATH

cd ./src
go build -v -o ../$BIN_PATH/$APP_ID

cd ..
$BIN_PATH/$APP_ID $1 $2
