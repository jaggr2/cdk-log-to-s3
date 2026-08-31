import * as fs from "fs";
import * as path from "path";
import { Architecture } from "aws-cdk-lib/aws-lambda";

/**
 * Resolves a prebuilt Go artifact for an architecture.
 *
 * The zips are committed and shipped inside the npm tarball, so a consumer
 * needs neither a Go toolchain nor Docker to synth.
 *
 * Internal: not exported from the package entry point.
 */
export function resolveArtifactAsset(
  name: string,
  architecture: Architecture,
): string {
  const suffix = SUFFIX_BY_ARCHITECTURE[architecture.name];
  if (!suffix) {
    throw new Error(
      `Unsupported architecture "${architecture.name}" for the log-to-s3 ${name}. ` +
        `Supported architectures are: ${Object.keys(SUFFIX_BY_ARCHITECTURE).join(", ")}.`,
    );
  }

  // __dirname is <package>/lib/private once compiled, and <repo>/src/private
  // when the tests run from source; both are two levels below the root.
  const asset = path.join(
    __dirname,
    "..",
    "..",
    "assets",
    `${name}-${suffix}.zip`,
  );
  if (!fs.existsSync(asset)) {
    throw new Error(
      `The log-to-s3 ${name} asset is missing at ${asset}. ` +
        "If you are working in a clone of the repository, run: npx projen build:layers",
    );
  }
  return asset;
}

/** The Lambda extension layer. */
export function resolveLayerAsset(architecture: Architecture): string {
  return resolveArtifactAsset("layer", architecture);
}

/** The compaction function, packaged for the provided.al2023 runtime. */
export function resolveCompactorAsset(architecture: Architecture): string {
  return resolveArtifactAsset("compactor", architecture);
}

const SUFFIX_BY_ARCHITECTURE: { [name: string]: string } = {
  [Architecture.ARM_64.name]: "arm64",
  [Architecture.X86_64.name]: "x86_64",
};
