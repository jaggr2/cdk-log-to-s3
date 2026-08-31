/**
 * Cross-compiles the Go binaries for every supported Lambda architecture and
 * packages each as a zip under assets/.
 *
 * Two artifacts come out of this: the telemetry extension, packaged as a
 * layer, and the compaction function, packaged as a provided.al2023 handler.
 *
 * This is a Node script rather than a Makefile because neither `make` nor
 * `zip` can be assumed present on a Windows development machine, and because
 * the executables have to carry mode 0755 inside the archive - Lambda fails
 * INIT otherwise, with an error that never mentions permissions.
 */
import { execFileSync } from 'node:child_process';
import * as crypto from 'node:crypto';
import * as fs from 'node:fs';
import * as path from 'node:path';
import {
  ARTIFACTS,
  ASSETS_DIR,
  Artifact,
  GO_DIR,
  TARGETS,
  Target,
  sha256,
  sha256Path,
  zipPath,
} from './layers';
import { buildZip } from './zip';

/**
 * The toolchain version, e.g. "go1.26.4".
 *
 * Deliberately `go env GOVERSION` and not `go version`: the latter appends the
 * host OS and architecture, so a Windows and a Linux build of byte-identical
 * cross-compiled binaries would still disagree here and trip the CI staleness
 * check.
 */
function goVersion(): string {
  return execFileSync('go', ['env', 'GOVERSION'], { encoding: 'utf8' }).trim();
}

function compile(artifact: Artifact, target: Target): Buffer {
  const outDir = path.join(GO_DIR, 'dist', artifact.name, target.layer);
  fs.rmSync(outDir, { recursive: true, force: true });
  fs.mkdirSync(outDir, { recursive: true });

  const binPath = path.join(outDir, path.basename(artifact.entry));
  console.log(`[build] ${artifact.name} GOARCH=${target.goarch}`);

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
      artifact.pkg,
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

function pack(artifact: Artifact, target: Target, binary: Buffer): void {
  const out = zipPath(artifact, target);
  fs.mkdirSync(ASSETS_DIR, { recursive: true });

  // 0o755 is not optional. Lambda executes the file directly.
  fs.writeFileSync(out, buildZip([{ name: artifact.entry, data: binary, mode: 0o755 }]));
  fs.writeFileSync(sha256Path(artifact, target), `${sha256(out)}\n`, 'utf8');

  const kb = Math.round(fs.statSync(out).size / 1024);
  console.log(`[pack ] ${path.basename(out)} (${kb} KB)`);
}

/**
 * Fingerprint of the Go source. Recorded so a deployed artifact can be traced
 * back to the code it was built from without a git checkout.
 */
function extensionSourceSha(): string {
  const hash = crypto.createHash('sha256');

  const walk = (dir: string): void => {
    for (const name of fs.readdirSync(dir).sort()) {
      // dist/ is build output, and hashing it would make the fingerprint
      // depend on itself.
      if (name === 'dist') continue;

      const full = path.join(dir, name);
      if (fs.statSync(full).isDirectory()) {
        walk(full);
        continue;
      }
      if (!/\.go$/.test(name) && name !== 'go.mod' && name !== 'go.sum') continue;
      if (name.endsWith('_test.go')) continue;

      hash.update(path.relative(GO_DIR, full).split(path.sep).join('/'));
      hash.update(fs.readFileSync(full));
    }
  };
  walk(GO_DIR);

  return hash.digest('hex');
}

function main(): void {
  console.log(`[go   ] ${goVersion()}`);

  for (const artifact of ARTIFACTS) {
    for (const target of TARGETS) {
      pack(artifact, target, compile(artifact, target));
    }
  }

  fs.writeFileSync(
    path.join(ASSETS_DIR, 'BUILDINFO.json'),
    `${JSON.stringify(
      {
        goVersion: goVersion(),
        artifacts: ARTIFACTS.map((a) => a.name),
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
