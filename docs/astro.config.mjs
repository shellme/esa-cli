// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
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
						{
							label: 'はじめに',
							link: '/getting-started/'
						}
					]
				},
				{
					label: 'ユーザーガイド',
					autogenerate: { directory: 'guides' }
				},
				{
					label: '開発者ガイド',
					autogenerate: { directory: 'developer' }
				},
				{
					label: 'APIリファレンス',
					autogenerate: { directory: 'reference' }
				},
				{
					label: 'サンプル',
					autogenerate: { directory: 'examples' }
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
						href: '/favicon.ico'
					}
				}
			]
		}),
	],
});
