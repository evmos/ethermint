local config = import 'default.jsonnet';

config {
  'ethermint_9000-1'+: {
    config+: {
      consensus+: {
        timeout_commit: '10s',
      },
    },
  },
}
