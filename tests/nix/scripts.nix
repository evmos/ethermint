{ pkgs
, config
}:
rec {
  start-ethermint = pkgs.writeShellScriptBin "start-ethermint"
    ''
      set -e
      if [ -z "$1" ]; then
        echo "No data directory supplied"
        exit 1
      fi
      export PATH=${pkgs.pystarport}/bin:${pkgs.ethermint}/bin:$PATH
      pystarport serve --config ${config.ethermint-config} --data $1
    '';
  start-geth = pkgs.writeShellScriptBin "start-geth"
    ''
      set -e
      if [ -z "$1" ]; then
        echo "No data directory supplied"
        exit 1
      fi
      export PATH=${pkgs.go-ethereum}/bin:$PATH
      geth --datadir $1 init ${config.geth-genesis}
      pwdfile=$(mktemp /tmp/password.XXXXXX)
      tmpfile=$(mktemp /tmp/validator-key.XXXXXX)

      cat > $pwdfile << EOF
      123456
      EOF

      # import validator key
      cat > $tmpfile << EOF
      826E479F5385C8C32CD96B0C0ACCDB8CC4FA5CACCC1BE54C1E3AA4D676A6EFF5
      EOF
      geth --datadir $1 account import $tmpfile --password $pwdfile

      # import community key
      cat > $tmpfile << EOF
      5D665FBD2FB40CB8E9849263B04457BA46D5F948972D0FE4C1F19B6B0F243574
      EOF
      geth --datadir $1 account import $tmpfile --password $pwdfile

      rm $tmpfile

      # start up
      geth --datadir $1 --http --http.addr localhost --http.api 'personal,eth,net,web3,txpool,miner' -unlock '0x57f96e6b86cdefdb3d412547816a82e3e0ebf9d2' --password $pwdfile --mine --allow-insecure-unlock

      rm $pwdfile
    '';
  scripts-env = pkgs.symlinkJoin {
    name = "scripts-env";
    paths = [ start-ethermint start-geth ];
  };
}
