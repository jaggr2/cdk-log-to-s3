import * as crypto from 'node:crypto';
import * as fs from 'node:fs';
import * as path from 'node:path';

export const ROOT = path.resolve(__dirname, '..');
export const GO_DIR = path.join(ROOT, 'extension-go');
export const ASSETS_DIR = path.join(ROOT, 'assets');

export interface Artifact {
  /** Base name of the zip, without the architecture suffix. */
  readonly name: string;
  /** Go package to build, relative to extension-go. */
  readonly pkg: string;
  /**
   * Path of the executable inside the zip.
   *
   * A Lambda extension must sit at extensions/<name>; a provided.al2023
   * function must be a file called `bootstrap` at the archive root. Neither is
   * negotiable, and both fail at INIT with errors that say nothing useful.
   */
  readonly entry: string;
}

export const ARTIFACTS: Artifact[] = [
  { name: 'layer', pkg: '.', entry: 'extensions/log-to-s3' },
  { name: 'compactor', pkg: './compactor', entry: 'bootstrap' },
];

export interface Target {
  /** Suffix used in the asset file name; matches lambda.Architecture.name. */
  readonly layer: string;
  /** GOARCH value passed to the Go compiler. */
  readonly goarch: string;
}

export const TARGETS: Target[] = [
  { layer: 'arm64', goarch: 'arm64' },
  { layer: 'x86_64', goarch: 'amd64' },
];

export function zipPath(artifact: Artifact, target: Target): string {
  return path.join(ASSETS_DIR, `${artifact.name}-${target.layer}.zip`);
}

export function sha256Path(artifact: Artifact, target: Target): string {
  return `${zipPath(artifact, target)}.sha256`;
}

export function sha256(file: string): string {
  return crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
}
