local config = import 'default.jsonnet';

config {
  'ethermint_9000-1'+: {
    config+: {
      tx_index+: {
        indexer: 'null',
      },
    },
    'app-config'+: {
      'json-rpc'+: {
        'enable-indexer': true,
      },
    },
  },
}
