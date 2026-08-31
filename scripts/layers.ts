import * as crypto from 'node:crypto';
import * as fs from 'node:fs';
import * as path from 'node:path';

/** Name of the executable inside the layer, i.e. /opt/extensions/<BINARY>. */
export const BINARY = 'log-to-s3';

export const ROOT = path.resolve(__dirname, '..');
export const GO_DIR = path.join(ROOT, 'extension-go');
export const ASSETS_DIR = path.join(ROOT, 'assets');

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

export function zipPath(target: Target): string {
  return path.join(ASSETS_DIR, `layer-${target.layer}.zip`);
}

export function sha256Path(target: Target): string {
  return `${zipPath(target)}.sha256`;
}

export function sha256(file: string): string {
  return crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
}
