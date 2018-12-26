DATA_DIR=$PWD/datadir
GENESIS_PATH=$PWD/genesis.json

tomo removedb --datadir $DATA_DIR
tomo init $GENESIS_PATH --datadir $DATA_DIR