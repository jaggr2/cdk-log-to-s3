/**
 * Cross-compiles the Go extension for every supported Lambda architecture and
 * packages each binary as a layer zip under assets/.
 *
 * This is a Node script rather than a Makefile because neither `make` nor
 * `zip` can be assumed present on a Windows development machine, and because
 * the extension file has to carry mode 0755 inside the archive - Lambda fails
 * INIT otherwise, with an error that never mentions permissions.
 */
import { execFileSync } from 'node:child_process';
import * as crypto from 'node:crypto';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { ASSETS_DIR, BINARY, GO_DIR, TARGETS, Target, sha256, sha256Path, zipPath } from './layers';
import { buildZip } from './zip';

function goVersion(): string {
  return execFileSync('go', ['version'], { encoding: 'utf8' }).trim();
}

function compile(target: Target): Buffer {
  const outDir = path.join(GO_DIR, 'dist', target.layer);
  fs.rmSync(outDir, { recursive: true, force: true });
  fs.mkdirSync(outDir, { recursive: true });

  const binPath = path.join(outDir, BINARY);
  console.log(`[build] GOARCH=${target.goarch}`);

  execFileSync(
    'go',
    [
      'build',
      // -trimpath and an empty buildid strip absolute paths and the build
      // fingerprint, which is what makes the output byte-reproducible and lets
      // CI verify that assets/ is not stale.
      '-trimpath',
      '-buildvcs=false',
      '-ldflags',
      '-s -w -buildid=',
      '-o',
      binPath,
      '.',
    ],
    {
      cwd: GO_DIR,
      stdio: 'inherit',
      env: {
        ...process.env,
        GOOS: 'linux',
        GOARCH: target.goarch,
        CGO_ENABLED: '0',
        GOFLAGS: '-mod=readonly',
      },
    },
  );

  return fs.readFileSync(binPath);
}

function pack(target: Target, binary: Buffer): void {
  const out = zipPath(target);
  fs.mkdirSync(ASSETS_DIR, { recursive: true });

  // 0o755 is not optional. Lambda executes /opt/extensions/<name> directly.
  fs.writeFileSync(out, buildZip([{ name: `extensions/${BINARY}`, data: binary, mode: 0o755 }]));
  fs.writeFileSync(sha256Path(target), `${sha256(out)}\n`, 'utf8');

  const kb = Math.round(fs.statSync(out).size / 1024);
  console.log(`[pack ] ${path.basename(out)} (${kb} KB)`);
}

/**
 * Fingerprint of the extension source. Recorded so a deployed layer can be
 * traced back to the code it was built from without a git checkout.
 */
function extensionSourceSha(): string {
  const files = fs
    .readdirSync(GO_DIR)
    .filter((f) => (f.endsWith('.go') && !f.endsWith('_test.go')) || f === 'go.mod' || f === 'go.sum')
    .sort();

  const hash = crypto.createHash('sha256');
  for (const f of files) {
    hash.update(f);
    hash.update(fs.readFileSync(path.join(GO_DIR, f)));
  }
  return hash.digest('hex');
}

function main(): void {
  console.log(`[go   ] ${goVersion()}`);

  for (const target of TARGETS) {
    pack(target, compile(target));
  }

  fs.writeFileSync(
    path.join(ASSETS_DIR, 'BUILDINFO.json'),
    `${JSON.stringify(
      {
        goVersion: goVersion(),
        targets: TARGETS.map((t) => t.layer),
        extensionSourceSha: extensionSourceSha(),
      },
      null,
      2,
    )}\n`,
    'utf8',
  );

  console.log('[done ] assets are up to date');
}

main();
