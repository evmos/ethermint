. ./setup.sh
FROM=$S1
TO=$S2
AMOUNT=1aphoton
$CLI tx bank  send $FROM $TO $AMOUNT --chain-id $CHAINID --keyring-backend $KEYRING 
