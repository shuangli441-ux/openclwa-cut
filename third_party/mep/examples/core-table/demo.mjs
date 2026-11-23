import { readFileSync } from 'node:fs';
import { isDeepStrictEqual } from 'node:util';
import { pathToFileURL } from 'node:url';

export const fixtures = JSON.parse(readFileSync(new URL('./fixtures.json', import.meta.url), 'utf8'));
const id = /^[A-Za-z0-9_-]{1,100}$/;
const isObject = value => value !== null && typeof value === 'object' && !Array.isArray(value);

// An in-memory example of host semantics, not a production transport or authorization system.
export function createDemoHost({ now = Date.now } = {}) {
  const { manifest, request } = fixtures;
  const { protocol, session, workspace } = manifest;
  const results = new Map();
  let state = structuredClone(fixtures.state);
  let nextRevision = 4;
  let revoked = false;
  const snapshot = () => structuredClone(state);
  const reject = (requestId, code, next, message) => ({
    protocol, session, workspace, id: requestId, status: 'rejected',
    error: { code, next, message },
  });

  return {
    snapshot,
    revoke() { revoked = true; },
    submit(input) {
      if (!isObject(input) || typeof input.id !== 'string' || !id.test(input.id)) {
        throw new TypeError('A valid request ID is required for a receipt.');
      }
      if (input.protocol !== protocol || input.session !== session || input.workspace !== workspace ||
          input.credential !== request.credential || revoked || now() >= manifest.expiresAt) {
        return reject(input.id, 'not_authorized', 'ask_user', 'Start an authorized session.');
      }
      const previous = results.get(input.id);
      if (previous) {
        // This return value is a caller-local response; it never replaces the original receipt.
        return isDeepStrictEqual(previous.input, input) ? structuredClone(previous.result)
          : reject(input.id, 'id_conflict', 'fix', 'This ID belongs to a different request.');
      }
      const record = result => {
        results.set(input.id, { input: structuredClone(input), result: structuredClone(result) });
        return result;
      };
      const allowed = ['protocol', 'session', 'workspace', 'id', 'credential', 'action', 'version', 'baseRevision', 'params'];
      if (Object.keys(input).some(k => !allowed.includes(k))) {
        return record(reject(input.id, 'bad_request', 'fix', 'Remove undeclared request fields.'));
      }
      if (input.action !== 'table.set_status' || input.version !== 1) {
        return record(reject(input.id, 'unknown_action', 'fix', 'Choose an action and version from the manifest.'));
      }
      if (typeof input.baseRevision !== 'string' || !input.baseRevision) {
        return record(reject(input.id, 'bad_request', 'fix', 'A write needs a base revision.'));
      }
      if (input.baseRevision !== state.revision) {
        return record({ ...reject(input.id, 'conflict', 'reread', 'Read the new snapshot before editing.'), revision: state.revision });
      }
      const { params } = input;
      if (!isObject(params) || Object.keys(params).length !== 2 ||
          !Array.isArray(params.rowIds) || params.rowIds.length < 1 || params.rowIds.length > 100 ||
          params.rowIds.some(row => typeof row !== 'string' || !id.test(row)) ||
          new Set(params.rowIds).size !== params.rowIds.length || !['todo', 'done'].includes(params.status)) {
        return record(reject(input.id, 'invalid_params', 'fix', 'Supply unique row IDs and a valid status.'));
      }
      const known = new Set(state.data.rows.map(row => row.id));
      if (params.rowIds.some(row => !known.has(row))) {
        return record(reject(input.id, 'row_not_found', 'fix', 'At least one row does not exist; nothing changed.'));
      }
      const selected = new Set(params.rowIds);
      const changed = state.data.rows.filter(row => selected.has(row.id) && row.status !== params.status).map(row => row.id);
      const revision = changed.length ? `rev-${nextRevision++}` : state.revision;
      state = { ...state, revision, data: { rows: state.data.rows.map(row => selected.has(row.id) ? { ...row, status: params.status } : row) } };
      return record({ protocol, session, workspace, id: input.id, status: 'succeeded', revision, output: { changed } });
    },
  };
}

if (process.argv[1] && pathToFileURL(process.argv[1]).href === import.meta.url) {
  const host = createDemoHost({ now: () => 1790000000000 });
  console.log(JSON.stringify({ receipt: host.submit(fixtures.request), state: host.snapshot() }, null, 2));
}
