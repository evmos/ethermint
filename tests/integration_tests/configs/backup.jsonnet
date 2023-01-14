local config = import 'main.jsonnet';

config {
  'ethermint_9000-1'+: {
    'app-config'+: {
      'json-rpc'+: {
        'backup-grpc-address-block-range': {
          '0.0.0.0:26754': [0, 10],
        },
      },
    },
  },
}
