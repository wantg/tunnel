APP_ID=tunnel
BIN_PATH=./bin

ln -sf ../config.json $BIN_PATH

cd ./src
go build -o ../$BIN_PATH/$APP_ID

cd ..
$BIN_PATH/$APP_ID