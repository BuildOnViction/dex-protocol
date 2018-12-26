DATA_DIR=$PWD/datadir
KEYSTORE_DIR=$DATA_DIR/keystore

account=$(
  tomo account list --datadir $DATA_DIR  --keystore $KEYSTORE_DIR \
  2> /dev/null \
  | head -n 1 \
  | cut -d"{" -f 2 | cut -d"}" -f 1
)

tomo \
  --verbosity 4 \
  --datadir $DATA_DIR \
  --keystore $KEYSTORE_DIR \
  --identity $NAME \
  --unlock 0x$account \
  --networkid 89 \
  --port 30303 \
  --rpc \
  --rpccorsdomain "*" \
  --rpcaddr 0.0.0.0 \
  --rpcport 8545 \
  --rpcvhosts "*" \
  --ws \
  --wsaddr 0.0.0.0 \
  --wsport 8546 \
  --wsorigins "*" \
  --mine \
  --gasprice "1" \
  --targetgaslimit "420000000"