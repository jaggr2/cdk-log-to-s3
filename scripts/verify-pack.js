/**
 * Fails the build when the npm tarball would not contain the layer zips.
 *
 * .npmignore is an exclude list with negations, which is easy to get subtly
 * wrong and impossible to notice until a consumer runs `cdk synth` against a
 * published package and hits a missing asset.
 */
const { execSync } = require('node:child_process');

const required = [
  'assets/layer-arm64.zip',
  'assets/layer-x86_64.zip',
  'assets/compactor-arm64.zip',
  'assets/compactor-x86_64.zip',
];

const out = execSync('npm pack --dry-run --json', { encoding: 'utf8' });

const files = JSON.parse(out)[0].files.map((f) => f.path.split('\\').join('/'));

const missing = required.filter((r) => !files.includes(r));
if (missing.length > 0) {
  console.error('The npm tarball is missing required layer assets:');
  for (const m of missing) console.error(`  - ${m}`);
  console.error('\nCheck the .npmignore negations in .projenrc.ts.');
  process.exit(1);
}

console.log(`npm tarball contains all ${required.length} Go artifacts`);
