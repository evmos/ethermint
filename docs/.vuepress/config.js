module.exports = {
  theme: 'cosmos',
  title: 'Ethermint Documentation',
  locales: {
    '/': {
      lang: 'en-US'
    },
  },
  markdown: {
    extendMarkdown: (md) => {
      md.use(require("markdown-it-katex"));
    },
  },
  head: [
      [
          "link",
          {
              rel: "stylesheet",
              href:
                  "https://cdnjs.cloudflare.com/ajax/libs/KaTeX/0.5.1/katex.min.css",
          },
      ],
      [
          "link",
          {
              rel: "stylesheet",
              href:
                  "https://cdn.jsdelivr.net/github-markdown-css/2.2.1/github-markdown.css",
          },
      ],
  ],
  base: process.env.VUEPRESS_BASE || '/',
  themeConfig: {
    repo: 'tharsis/ethermint',
    docsRepo: 'tharsis/ethermint',
    docsBranch: 'main',
    docsDir: 'docs',
    editLinks: true,
    custom: true,
    logo: {
      src: '/ethermint-logo-horizontal-alpha.svg',
    },
    algolia: {
      id: 'BH4D9OD16A',
      key: 'c5da4dd3636828292e3c908a0db39688',
      index: 'ethermint'
    },
    topbar: {
      banner: false
    },
    sidebar: {
      auto: false,
      nav: [
        {
          title: 'Reference',
          children: [
            {
              title: 'Introduction',
              directory: true,
              path: '/intro'
            },
            {
              title: 'Quick Start',
              directory: true,
              path: '/quickstart'
            },
            {
              title: 'Basics',
              directory: true,
              path: '/basics'
            },
            {
              title: 'Core Concepts',
              directory: true,
              path: '/core'
            },
          ]
        },
        {
          title: 'Guides',
          children: [
            {
              title: 'Localnet',
              directory: true,
              path: '/guides/localnet'
            },
            {
              title: 'Keys and Wallets',
              directory: true,
              path: '/guides/keys-wallets'
            },
            {
              title: 'Ethereum Tooling',
              directory: true,
              path: '/guides/tools'
            },
          ]
        },
        {
          title: 'APIs',
          children: [
            {
              title: 'JSON-RPC',
              directory: true,
              path: '/api/JSON-RPC'
            },
            {
              title: 'Protobuf Reference',
              directory: false,
              path: '/api/proto-docs'
            },
          ]
        },
        {
          title: 'Testnet',
          children: [
            {
              title: 'Guides',
              directory: true,
              path: '/testnet'
            },
          ]
        },
        {
          title: 'Specifications',
          children: [{
            title: 'Modules',
            directory: true,
            path: '/modules'
          }]
        }, {
          title: 'Resources',
          children: [
            {
              title: 'Ethermint API Reference',
              path: 'https://pkg.go.dev/github.com/tharsis/ethermint'
            },
            {
              title: 'Cosmos REST API Spec',
              path: 'https://cosmos.network/rpc/'
            },
            {
              title: 'JSON-RPC API Reference',
              path: '/api/JSON-RPC/endpoints'
            }
          ]
        }
      ]
    },
    gutter: {
      title: 'Help & Support',
      chat: {
        title: 'Developer Chat',
        text: 'Chat with Ethermint developers on Discord.',
        url: 'https://discord.gg/3ZbxEq4KDu',
        bg: 'linear-gradient(103.75deg, #1B1E36 0%, #22253F 100%)'
      },
      forum: {
        title: 'Ethermint Developer Forum',
        text: 'Join the Ethermint Developer Forum to learn more.',
        url: 'https://forum.cosmos.network/c/ethermint',
        bg: 'linear-gradient(221.79deg, #3D6B99 -1.08%, #336699 95.88%)',
        logo: 'ethereum-white'
      },
      github: {
        title: 'Found an Issue?',
        text: 'Help us improve this page by suggesting edits on GitHub.',
        bg: '#F8F9FC'
      }
    },
    footer: {
      logo: '/logo-bw.svg',
      textLink: {
        text: 'ethermint.dev',
        url: 'https://ethermint.dev'
      },
      services: [
        {
          service: 'github',
          url: 'https://github.com/tharsis/ethermint'
        },
        {
          service: "twitter",
          url: "https://twitter.com/ethermint",
        },
        {
          service: "linkedin",
          url: "https://www.linkedin.com/company/tharsis-finance/",
      },
      {
          service: "medium",
          url: "https://medium.com/@tharsis_labs",
      },
      ],
      smallprint: 'This website is maintained by Tharsis Labs Ltd.',
      links: [{
          title: 'Documentation',
          children: [{
              title: 'Cosmos SDK Docs',
              url: 'https://docs.cosmos.network/master/'
            },
            {
              title: 'Ethereum Docs',
              url: 'https://ethereum.org/developers'
            },
            {
              title: 'Tendermint Core Docs',
              url: 'https://docs.tendermint.com'
            }
          ]
        },
        {
          title: 'Community',
          children: [{
              title: 'Ethermint Community',
              url: 'https://discord.gg/3ZbxEq4KDu'
            },
            {
              title: 'Ethermint Forum',
              url: 'https://forum.cosmos.network/c/ethermint'
            }
          ]
        },
        {
          title: 'Tharsis',
          children: [
            {
              title: 'Jobs at Tharsis',
              url: 'https://tharsis.notion.site/Jobs-at-Tharsis-5a1642eb89b34747ae6f2db2d356fc0d'
            }
          ]
        }
      ]
    },
    versions: [
      {
        "label": "main",
        "key": "main"
      }
    ],
  }
};