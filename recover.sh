. ./setup.sh
echo $MYMNEMONICS
$CLI keys add ibc1 --recover --keyring-backend test --index=0 
$CLI keys add ibc2 --recover --keyring-backend test --index=1 
