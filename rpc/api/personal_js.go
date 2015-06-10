package api

const Personal_JS = `
web3._extend({
	property: 'personal',
	methods:
	[
		new web3._extend.Method({
			name: 'listAccounts',
			call: 'personal_listAccounts',
			params: 0,
			inputFormatter: [],
			outputFormatter: function(obj) { return obj; }
		}),
		new web3._extend.Method({
			name: 'newAccount',
			call: 'personal_newAccount',
			params: 1,
			inputFormatter: [web3._extend.formatters.formatInputString],
			outputFormatter: web3._extend.formatters.formatOutputString
		}),
		new web3._extend.Method({
			name: 'unlockAccount',
			call: 'personal_unlockAccount',
			params: 3,
			inputFormatter: [web3._extend.formatters.formatInputString,web3._extend.formatters.formatInputString,web3._extend.formatters.formatInputInt],
			outputFormatter: web3._extend.formatters.formatOutputBool
		})
	],
	properties:
	[
	]
});
`
