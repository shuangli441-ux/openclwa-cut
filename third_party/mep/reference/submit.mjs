// MEP 1 reference client — usage: node submit.mjs draft.json
// Lives at the root of a collaboration folder. Checks heartbeat and order, checks base,
// picks an id, writes the request atomically and waits for its result.
import * as files from 'node:fs/promises';
import * as paths from 'node:path';
import * as urls from 'node:url';
import * as timers from 'node:timers/promises';
import * as crypto from 'node:crypto';

const STALE_MS = 20_000;               // heartbeat.staleMs in mep.json
const POLL_MS = 500;
const ROUND_MS = 30 * 60 * 1000;       // give up waiting after 30 minutes
const ID_RULE = /^[\w-]{1,100}$/;

const folder = paths.dirname(urls.fileURLToPath(import.meta.url));

async function readJson(name) {
  try { return JSON.parse(await files.readFile(paths.join(folder, name), 'utf8')); }
  catch { return null; }
}

function quit(message, code = 1) {
  console.error(message);
  process.exit(code);
}

const draftPath = process.argv[2];
if (!draftPath) quit('Usage: node submit.mjs draft.json');

const draft = JSON.parse(await files.readFile(draftPath, 'utf8'));
if (draft.mep !== 1 || typeof draft.key !== 'string' || typeof draft.who !== 'string') {
  quit('The draft needs "mep": 1, "key" (the order key) and "who" (your name).');
}

const isAlive = status => Boolean(status) && Date.now() - status.alive < STALE_MS;

const status = await readJson('status.json');
if (!isAlive(status)) quit('The host is not running or this project is not open. Ask the user to open it.', 2);
if (status.state !== 'open') quit('This order is stopped or expired. Ask the user to copy the command again and use the new key.', 2);

const needsBase = draft.ops || draft.roughcut || draft.template || (draft.look && !draft.look.src);
if (needsBase) {
  const project = await readJson('project.json');
  if (!project || draft.base !== project.revision) {
    quit('The project has changed since you read it (base is not the current revision). Read project.json again; never force an overwrite.');
  }
}

const { id: givenId, ...request } = draft;
const id = givenId ?? `r-${crypto.randomBytes(6).toString('hex')}`;
if (!ID_RULE.test(id)) quit('An id may contain only letters, digits, "-" and "_".');
if (await readJson(paths.join('results', `${id}.json`))) quit('This id already has a result. Choose a new id.');

const requests = paths.join(folder, 'requests');
await files.mkdir(requests, { recursive: true });
const staging = paths.join(requests, `${id}.tmp`);
await files.writeFile(staging, JSON.stringify(request), { mode: 0o600 });
await files.rename(staging, paths.join(requests, `${id}.json`));

let lastNote = '';
const deadline = Date.now() + ROUND_MS;
while (true) {
  const result = await readJson(paths.join('results', `${id}.json`));
  if (result) {
    console.log(JSON.stringify(result, null, 2));
    process.exitCode = result.ok ? 0 : 1;
    break;
  }
  const now = await readJson('status.json');
  if (now?.refused?.id === id) quit(now.refused.message);
  if (now?.working?.id === id && now.working.text !== lastNote) {
    lastNote = now.working.text;
    console.error(`host: ${lastNote}`);
  }
  if (now && Date.now() - now.alive >= 3 * STALE_MS) {
    quit('The host stopped reporting (it quit or the project closed) before answering. Ask the user to reopen the project, and do not resubmit this request until its result appears.', 2);
  }
  if (Date.now() > deadline) {
    const last = lastNote ? ` Last host status: ${lastNote}.` : '';
    quit(`No result after 30 minutes.${last} The request is still with the host and its result will be written to results/${id}.json. Tell the user, and do not resubmit it.`, 2);
  }
  await timers.setTimeout(POLL_MS);
}
