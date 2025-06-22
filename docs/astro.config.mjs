// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://shellme.github.io',
	base: '/esa-cli',
	integrations: [
		starlight({
			title: 'esa-cli',
			description: 'esa-cliのドキュメント',
			social: [
				{
					icon: 'github',
					label: 'GitHub',
					href: 'https://github.com/shellme/esa-cli'
				}
			],
			sidebar: [
				{
					label: 'はじめに',
					items: [
						{ label: 'はじめに', link: '/getting-started/' },
						{ label: 'インストール', link: '/getting-started/installation' },
						{ label: '初期設定と基本操作', link: '/getting-started/first-steps' },
					]
				},
				{
					label: 'ユーザーガイド',
					items: [
						{ label: 'ガイド', link: '/guides/' },
						{ label: '基本的な使い方', link: '/guides/basic-usage' },
						{ label: '高度な使い方', link: '/guides/advanced-usage' },
					]
				},
				{
					label: 'コマンドリファレンス',
					items: [
						{ label: 'コマンド一覧', link: '/commands/' },
						{ label: '初期設定', link: '/commands/setup' },
						{ label: '記事一覧', link: '/commands/list' },
						{ label: '記事取得', link: '/commands/fetch' },
						{ label: '記事更新', link: '/commands/update' },
						{ label: '記事一括移動', link: '/commands/move' },
					]
				},
				{
					label: 'サポート',
					items: [
						{ label: 'よくある質問', link: '/faq' },
						{ label: 'トラブルシューティング', link: '/troubleshooting' },
					]
				},
				{
					label: '開発者ガイド',
					items: [
						{ label: '開発者ガイド', link: '/developer/' },
						{ label: 'リリース手順', link: '/developer/release' },
						{ label: 'テスト', link: '/developer/test' },
					]
				}
			],
			editLink: {
				baseUrl: 'https://github.com/shellme/esa-cli/edit/main/docs/src/content/docs/'
			},
			head: [
				{
					tag: 'link',
					attrs: {
						rel: 'icon',
						href: '/esa-cli/favicon.ico'
					}
				}
			]
		}),
	],
});
