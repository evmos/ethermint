local config = import 'main.jsonnet';

config {
  'ethermint_9000-1'+: {
    'app-config'+: {
        'backup-grpc-address': '0.0.0.0:26754',
    },
  },
}
