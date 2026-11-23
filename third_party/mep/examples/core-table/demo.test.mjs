import assert from 'node:assert/strict';
import test from 'node:test';
import { createDemoHost, fixtures } from './demo.mjs';

const host = () => createDemoHost({ now: () => 1790000000000 });
const request = fixtures.request;
test('two rows update atomically, matching the wire fixture', () => {
  const h = host();
  assert.deepEqual(h.submit(request), fixtures.result);
  assert.deepEqual(h.snapshot().data.rows.map(row => row.status), ['done', 'done']);
});
test('a stale write and an unknown row do not change state', () => {
  const h = host();
  h.submit(request);
  const state = h.snapshot();
  assert.deepEqual(h.submit({ ...request, id: 'r-02' }), fixtures.conflict);
  const rejection = h.submit({ ...request, id: 'r-04', baseRevision: 'rev-4', params: { rowIds: ['row-17', 'missing'], status: 'todo' } });
  assert.equal(rejection.error.code, 'row_not_found');
  assert.deepEqual(h.snapshot(), state);
});
test('equivalent replay returns its receipt; same-ID collision cannot replace it', () => {
  const h = host();
  const result = h.submit(request);
  const state = h.snapshot();
  const { id, ...rest } = request;
  assert.deepEqual(h.submit({ ...rest, id }), result);
  assert.equal(h.submit({ ...request, baseRevision: 'rev-4' }).error.code, 'id_conflict');
  assert.deepEqual(h.submit(request), result);
  assert.deepEqual(h.snapshot(), state);
});
test('missing base, unknown action, wrong credential, expired or revoked sessions are rejected', () => {
  const h = host();
  assert.equal(h.submit({ ...request, id: 'r-version', baseRevision: undefined }).error.code, 'bad_request');
  assert.equal(h.submit({ ...request, id: 'r-unknown', action: 'table.remove' }).error.code, 'unknown_action');
  assert.equal(h.submit({ ...request, credential: 'wrong' }).error.code, 'not_authorized');
  assert.equal(createDemoHost({ now: () => fixtures.manifest.expiresAt }).submit(request).error.code, 'not_authorized');
  h.revoke();
  assert.equal(h.submit(request).error.code, 'not_authorized');
  assert.deepEqual(h.snapshot(), fixtures.state);
});
