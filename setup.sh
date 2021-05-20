. ./setup0.sh
export S1=$($CLI keys show mykey -a --keyring-backend $KEYRING) 
export S2=$($CLI keys show mykey2 -a --keyring-backend $KEYRING) 

echo 'S1='$S1
echo 'S2='$S2
