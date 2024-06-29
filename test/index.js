const {expect} = require('chai');
const {resolve} = require('path');
const {spawn} = require("child_process");
const {
	createConnection,
	StreamMessageReader,
	StreamMessageWriter,
} = require('vscode-languageserver/node');


describe('lsp', function () {
	const ROOT = resolve(__dirname, 'root');

	function connectServer() {
		const dir = resolve(__dirname, '..')

		const p = spawn('/usr/local/go/bin/go', ['run', 'main.go'], {cwd: dir});

		if (!p || !p.pid) {
			throw new Error(`can't spawn process`);
		}

		return createConnection(
			new StreamMessageReader(p.stdout),
			new StreamMessageWriter(p.stdin),
		);
	}

	/** @type {import('vscode-languageserver/node').Connection} */
	let server;

	before(function () {
		server = connectServer();

		server.listen();
	});

	after(function () {
		if (server) {
			server.dispose();
			server = null;
		}
	});

	it('initialize', async function () {
		const init = await server.sendRequest('initialize', {
			rootUri: 'file://' + resolve(__dirname, 'root'),
			processId: 1,
			capabilities: {},
			workspaceFolders: null,
			workDoneToken: 'token'
		});

		expect(init).to.have.property('capabilities');
	});

	it('semanticTokens', async function () {
		const textDocument = {
			uri: `file://` + resolve(ROOT, 'semanticTokens.txt')
		};

		await server.sendRequest('initialize', {
			rootUri: 'file://' + resolve(__dirname, 'root'),
			processId: 1,
			capabilities: {},
			workspaceFolders: null,
			workDoneToken: 'token'
		});

		let tokens = await server.sendRequest('textDocument/semanticTokens/full', {
			textDocument
		});

		expect(tokens)
		.to.have.property('data')
		.and.to.be.an('array')
		.with.lengthOf(18 * 5)

		tokens = await server.sendRequest('textDocument/semanticTokens/range', {
			textDocument,
			range: {
				start: {
					line: 0,
					character: 0,
				},
				end: {
					line: 2,
					character: 100,
				},
			}
		});

		expect(tokens)
		.to.have.property('data')
		.and.to.be.an('array')
		.with.lengthOf(5 * 5)

		const err = await server.sendRequest('textDocument/semanticTokens/full', {
			textDocument: {
				uri: `file://` + resolve(ROOT, 'not-exist.txt')
			},
		}).catch(err => err);

		expect(err).to.be.an('error');
	});
});