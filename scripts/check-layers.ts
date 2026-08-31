/**
 * Verifies the committed layer zips without needing a Go toolchain, so that
 * `npx projen test` is runnable by anyone.
 *
 * The mode check is the important one. A layer whose extension file lacks the
 * execute bit fails at Lambda INIT with an error that never mentions
 * permissions, and no CDK assertion or synth test can catch it - the zip is
 * opaque to them.
 */
import * as fs from 'node:fs';
import * as path from 'node:path';
import { BINARY, TARGETS, sha256, sha256Path, zipPath } from './layers';

const MIN_SIZE_BYTES = 1024 * 1024;

interface ZipEntry {
  readonly name: string;
  /** UNIX permission bits, or undefined if the archive stored none. */
  readonly mode?: number;
}

/**
 * Reads the central directory of a zip file. Written out by hand rather than
 * pulled from a dependency because the one field that matters here - the UNIX
 * mode in the external attributes - is the field most zip libraries drop.
 */
function readZipEntries(file: string): ZipEntry[] {
  const buf = fs.readFileSync(file);

  const EOCD_SIG = 0x06054b50;
  const CDH_SIG = 0x02014b50;

  let eocd = -1;
  for (let i = buf.length - 22; i >= 0; i--) {
    if (buf.readUInt32LE(i) === EOCD_SIG) {
      eocd = i;
      break;
    }
  }
  if (eocd < 0) throw new Error(`${file}: no end-of-central-directory record; the file is not a zip`);

  const count = buf.readUInt16LE(eocd + 10);
  let offset = buf.readUInt32LE(eocd + 16);

  const entries: ZipEntry[] = [];
  for (let i = 0; i < count; i++) {
    if (buf.readUInt32LE(offset) !== CDH_SIG) {
      throw new Error(`${file}: corrupt central directory at entry ${i}`);
    }
    const nameLen = buf.readUInt16LE(offset + 28);
    const extraLen = buf.readUInt16LE(offset + 30);
    const commentLen = buf.readUInt16LE(offset + 32);
    const externalAttr = buf.readUInt32LE(offset + 38);
    const name = buf.toString('utf8', offset + 46, offset + 46 + nameLen);

    const mode = (externalAttr >>> 16) & 0xffff;
    entries.push({ name, mode: mode === 0 ? undefined : mode });

    offset += 46 + nameLen + extraLen + commentLen;
  }
  return entries;
}

const problems: string[] = [];

for (const target of TARGETS) {
  const zip = zipPath(target);
  const rel = path.basename(zip);

  if (!fs.existsSync(zip)) {
    problems.push(`${rel} is missing - run: npx projen build:layers`);
    continue;
  }

  const stat = fs.statSync(zip);
  if (stat.size < MIN_SIZE_BYTES) {
    problems.push(`${rel} is only ${stat.size} bytes; a compiled extension is larger than 1 MB`);
  }

  const head = Buffer.alloc(2);
  const fd = fs.openSync(zip, 'r');
  fs.readSync(fd, head, 0, 2, 0);
  fs.closeSync(fd);
  if (head.toString('latin1') !== 'PK') {
    problems.push(`${rel} does not start with the zip magic bytes`);
    continue;
  }

  const sidecar = sha256Path(target);
  if (!fs.existsSync(sidecar)) {
    problems.push(`${rel}.sha256 is missing - run: npx projen build:layers`);
  } else {
    const expected = fs.readFileSync(sidecar, 'utf8').trim();
    const actual = sha256(zip);
    if (expected !== actual) {
      problems.push(`${rel} does not match its sha256 sidecar (expected ${expected}, got ${actual})`);
    }
  }

  const entries = readZipEntries(zip);
  const wantName = `extensions/${BINARY}`;
  const entry = entries.find((e) => e.name === wantName);

  if (!entry) {
    problems.push(
      `${rel} does not contain ${wantName} (found: ${entries.map((e) => e.name).join(', ') || 'nothing'})`,
    );
  } else if (entry.mode === undefined) {
    problems.push(`${rel}: ${wantName} carries no UNIX mode; Lambda will not be able to execute it`);
  } else if ((entry.mode & 0o111) === 0) {
    problems.push(
      `${rel}: ${wantName} has mode 0${entry.mode.toString(8)} and is not executable; Lambda INIT will fail`,
    );
  }
}

if (problems.length > 0) {
  console.error('Layer asset check failed:');
  for (const p of problems) console.error(`  - ${p}`);
  process.exit(1);
}

console.log(`Layer assets OK (${TARGETS.map((t) => t.layer).join(', ')}), extensions/${BINARY} is executable.`);
