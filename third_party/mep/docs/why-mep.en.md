# MEP: why AI assistants need a software workbench

**MEP (Motion Edit Protocol)** is an open AI–software collaboration protocol initiated by ADu Labs. ADuDir is an early participant and the reference implementation for the MEP 1 video-editing specification. Video editing is the first use case; [MEP Core](../CORE.md) defines the shared rules for authorization, state, actions, and receipts.

In one line: **MCP gives the assistant a cable. MEP gives it a workbench.**

This guide uses video editing as an example. The app opens a collaboration folder; the assistant reads state and submits changes; the app validates and executes them, then returns a receipt. The example also shows how MEP differs from MCP and Computer Use.

Chinese version: [why-mep.md](why-mep.md).

---

## 1. The problem: how does an AI assistant operate an editor?

Today there are two main ways to let an AI assistant operate software. Neither fits editing well.

**Computer Use (look at the screen, click the mouse)**
At every step the assistant takes a screenshot, works out the interface, then moves the mouse and clicks. An editor's interface is a timeline with tracks and hundreds or thousands of clips. Recognizing them from screenshots and dragging them with a mouse is slow and expensive, easy to get wrong, and hard to undo cleanly.

**MCP (Model Context Protocol)**
MCP is the open standard Anthropic released in 2024, and it is now the common way to connect AI assistants to tools. It is good, but it has preconditions: the assistant must support MCP, the user has to configure it once in every assistant and usually restart the assistant afterwards, and tool definitions usually sit in the assistant's context. For ordinary creators, that configuration step is the barrier.

Editing has its own nature: a project is versioned structured data, changes must be checkable and undoable, and the picture must be visible. Editing needs a path designed for it.

---

## 2. What MEP is

The whole of MEP is a folder.

```
project collaboration folder/
├── mep.json        protocol version, supported operations, limits, error codes (for programs)
├── README.md       short entry point (for the assistant)
├── docs/           chapters: timeline, rough cut, templates, media, code scenes (read when needed)
├── project.json    current timeline and revision
├── captions.json   every caption with per-character timing
├── status.json     the editor's heartbeat and what it is working on
├── requests/       the assistant writes changes here
├── results/        the editor writes a result for each request here
├── looks/          frames of the final video the editor renders for the assistant
└── inbox/          media the assistant brings
```

Four steps:

1. **Copy the command.** The user clicks "Copy command" in the editor and gets one sentence containing the folder path and the order key.
2. **Give it to the assistant.** The user pastes the sentence into their assistant and says what they want.
3. **The assistant submits.** It reads `project.json` and writes the change as a request file, carrying the revision it read.
4. **The editor executes.** It checks the key and revision, makes the change and writes a result. Changes appear on the timeline live and the whole order can be undone.

A request looks like this:

```json
{
  "mep": 1,
  "key": "order key",
  "who": "Codex",
  "base": "the project revision that was read",
  "ops": [{ "op": "cut", "clip": "clip-id", "at": 12.4 }]
}
```

Besides timeline edits, the same format covers rough-cutting talking-head footage from its transcript, making videos with templates, writing code scenes for motion graphics, voice-over, finding media, captions and translations, and asking the editor to render a few seconds of the final video for the assistant to see.

---

## 3. Three approaches compared

| | MEP | MCP | Computer Use |
|---|---|---|---|
| What the assistant needs | Read and write local files | MCP support | Screen control support |
| User setup | No install, no settings | Configure once per assistant, usually restart | Grant screen control |
| How the assistant learns the project | Reads structured text | Calls tools for data | Looks at screenshots |
| How the assistant edits | Writes one request file | Calls tools | Moves the mouse, types |
| Context used by the integration | None; docs read on demand | Tool definitions usually in context | One screenshot per step |
| Deterministic execution | Yes, the editor executes the request | Yes | No, clicks can miss |
| Editing alongside the user | Revision check; never overwrites newer edits | Depends on each implementation | No protection |
| Undo | Whole-order undo | Depends on each implementation | Hard |
| Assistant can see the final video | Yes, frames by second or a contact sheet | Depends on each implementation | Only the current screen |
| Runs in the background | Yes, the user can keep working | Yes | No, it occupies the screen |
| Permission scope | This one project's folder | Whatever the tools expose | The whole screen |

---

## 4. Advantages of MEP

### 1. Zero setup

No plugin, no configuration, no assistant restart. Copy one sentence and start. Claude Code, Codex, Cursor, WorkBuddy, Doubao desktop: any assistant that can read and write local files works. Assistants that do not exist yet will work tomorrow too, because all they need is to read and write files.

### 2. Tokens spent only where they matter

This is MEP's most concrete advantage. Tokens usually go to three places, and MEP saves in each:

- **The integration takes no context.** MCP tool definitions usually ride along in the session's context whether or not they are used. MEP has no standing tool definitions and costs nothing while idle.
- **No screenshots.** Computer Use sends a screenshot to the model at every step, and one edit can take dozens or hundreds of steps. MEP reads plain-text timeline data and fetches frames only when it needs to see the picture; one contact sheet covers ten seconds of pacing.
- **Execution calls no model.** Once the assistant hands in a request, validation, cutting and writing the timeline are all done by the editor with deterministic rules, with no further model calls.

The assistant spends tokens on only two things: reading the few pages of documentation it needs, and reading project data. The documentation is split into chapters read on demand, and the entry point is one page.

### 3. Deterministic, checkable, undoable

Every request carries the revision it read. If the user changed the project in the meantime, the request is refused and the assistant is told to reread, so it never overwrites the user's newer edits. Every request has an id and a result stating what changed, the ids of new clips and how many credits were spent. If the user is not happy, the whole order is undone.

### 4. The assistant can see what it made

The assistant can ask the editor to render the final video at given seconds, or one overview image of frames sampled across a span. Whether captions cover a face, whether the motion's pacing is right: the assistant checks its own work before handing it over.

### 5. Clear boundaries

The assistant gets only this one project's collaboration folder. It cannot touch other projects or other local folders, and it never receives account keys. Every order has a key and an expiry, and the user can "Stop this order" at any time.

### 6. Clear errors

Failed results carry a fixed error code and a suggested next step: reread on a version conflict, fix the request on a format error, relay to the user when confirmation is needed. If the editor is not running, the assistant knows from the heartbeat file within seconds instead of waiting forever.

### 7. Independent of models and vendors

MEP is not tied to any model, assistant or company. Its medium is files and JSON, the greatest common denominator of every system, language and assistant.

---

## 5. How MEP relates to MCP

MEP does not replace MCP. MCP answers "how does an assistant connect to tools" and is a general connection standard. MEP answers "how does an editing project get handed to an assistant" and is a working agreement for the editing domain.

ADuDir supports both:

- **MEP (Copy command):** the main route. No setup; any assistant that can read and write files can use it.
- **MCP direct connection:** for assistants you use often. Connect once, then just say "have Adu…" in the assistant to start editing.

---

## 6. Why it will become the industry standard

**First, it sits at the lowest possible barrier.** Whether a protocol becomes a standard depends on how many participants can implement it easily. On the assistant side, MEP only requires reading and writing files, which every mainstream coding and desktop assistant can already do. On the editor side, MEP only requires exporting the project as structured data, accepting request files and writing results, which any editor with a timeline can implement.

**Second, it captures the essence of editing.** An editing project is versioned timeline data. MEP's three core agreements, revision checks, whole-order undo and seeing frames by the second, are exactly the three things AI-assisted editing cannot do without. A general-purpose protocol will not define them for editors; the editing domain has to define them itself.

**Third, it follows where AI assistants are heading.** Assistants are moving from chatting to doing work on the local machine, and reading and writing files is their most basic and reliable ability. Turning a project into a folder turns the editor into something assistants already know how to use.

**Fourth, it is open.** The protocol has a version (`mep/1`) and a machine-readable manifest, and the specification is public. Any editor and any assistant may implement it without permission.

The industry needs an agreement everyone can use. MEP is designed for exactly that.

---

## 7. The MEP 1.0 specification

See [SPEC.en.md](../SPEC.en.md) for the full specification and [schema/](../schema/) for the JSON Schemas.

---

## 8. Scope

- MEP is for assistants that can read and write local files. Assistants that only chat in a web page and cannot access local files are out of scope.
- Exporting the final video is confirmed by the user in the editor; the protocol does not make the final delivery decision for the user.
- Operations that spend credits, such as voice-over and automatic template composition, state each amount in the result.

---

## Three lines for presenting MEP

1. **MCP is a cable; MEP is a workbench.** One needs configuration; the other works from one copied sentence.
2. **Computer Use teaches AI to use an interface designed for people; MEP gives AI an interface it already knows how to use.**
3. **Turn the project into a folder, and any assistant can edit video.** This is the protocol for the next generation of editing software.
