module.exports = {
  theme: 'cosmos',
  title: 'Ethermint Documentation',
  locales: {
    '/': {lang: 'en-US'},
  },
  base: process.env.VUEPRESS_BASE || '/',
  themeConfig: {
    repo: 'ChainSafe/ethermint',
    docsRepo: 'ChainSafe/ethermint',
    docsBranch: 'development',
    docsDir: 'docs',
    editLinks: true,
    custom: true,
    logo: {
      src: '/logo.svg',
    },
    algolia: {id: 'BH4D9OD16A', key: 'c5da4dd3636828292e3c908a0db39688', index: 'ethermint'},
    topbar: {banner: false},
    sidebar: {
      auto: false,
      nav: [
        {
          title: 'Reference',
          children: [
            {title: 'Introduction', directory: true, path: '/intro'},
            {title: 'Quick Start', directory: true, path: '/quickstart'},
            {title: 'Basics', directory: true, path: '/basics'},
            {title: 'Core Concepts', directory: true, path: '/core'},
            {title: 'Guides', directory: true, path: '/guides'}
          ]
        },
        {title: 'Specifications', children: [{title: 'Modules', directory: true, path: '/modules'}]}, {
          title: 'Resources',
          children: [
            {title: 'Ethermint API Reference', path: 'https://godoc.org/github.com/cosmos/ethermint'},
            {title: 'Cosmos REST API Spec', path: 'https://cosmos.network/rpc/'},
            {title: 'Ethereum JSON RPC API Reference', path: 'https://eth.wiki/json-rpc/API'}
          ]
        }
      ]
    },
    gutter: {
      title: 'Help & Support',
      chat: {
        title: 'Developer Chat',
        text: 'Chat with Ethermint developers on Discord.',
        url: 'https://discordapp.com/channels/669268347736686612',
        bg: 'linear-gradient(103.75deg, #1B1E36 0%, #22253F 100%)'
      },
      forum: {
        title: 'Ethermint Developer Forum',
        text: 'Join the Ethermint Developer Forum to learn more.',
        url: 'https://forum.cosmos.network/',
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
      textLink: {text: 'ethermint.zone', url: 'https://ethermint.zone'},
      services: [
        {service: 'github', url: 'https://github.com/ChainSafe/ethermint'},
        {service: 'twitter', url: 'https://twitter.com/chainsafeth'},
        {service: 'linkedin', url: 'https://www.linkedin.com/company/chainsafe-systems'},
        {service: 'medium', url: 'https://medium.com/chainsafe-systems'},
      ],
      smallprint:
          'This website is maintained by [ChainSafe Systems](https://chainsafe.io). The contents and opinions of this website are those of Chainsafe Systems.',
      links: [
        {
          title: 'Documentation',
          children: [
            {title: 'Cosmos SDK Docs', url: 'https://docs.cosmos.network'},
            {title: 'Ethermint Docs', url: 'https://ethereum.org/developers'},
            {title: 'Tendermint Core Docs', url: 'https://docs.tendermint.com'}
          ]
        },
        {
          title: 'Community',
          children: [
            {title: 'Cosmos Community', url: 'https://discord.gg/W8trcGV'},
            {title: 'Ethermint Forum', url: 'https://forum.cosmos.network/c/ethermint'},
            {title: 'Chainsafe Blog', url: 'https://medium.com/chainsafe-systems'}
          ]
        },
        {
          title: 'Contributing',
          children: [
            {title: 'Contributing to the docs', url: 'https://github.com/ChainSafe/ethermint/tree/development/docs'},
            {title: 'Careers at Chainsafe', url: 'https://chainsafe.io/#careers'}, {
              title: 'Source code on GitHub',
              url: 'https://github.com/Chainsafe/ethermint/blob/development/docs/DOCS_README.md'
            }
          ]
        }
      ]
    }
  },
};
