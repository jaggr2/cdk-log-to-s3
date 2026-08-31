import * as fs from "fs";
import * as path from "path";
import { Architecture } from "aws-cdk-lib/aws-lambda";

/**
 * Resolves the prebuilt layer zip for an architecture.
 *
 * The zips are committed and shipped inside the npm tarball, so a consumer
 * needs neither a Go toolchain nor Docker to synth.
 *
 * Internal: not exported from the package entry point.
 */
export function resolveLayerAsset(architecture: Architecture): string {
  const suffix = LAYER_BY_ARCHITECTURE[architecture.name];
  if (!suffix) {
    throw new Error(
      `Unsupported architecture "${architecture.name}" for the log-to-s3 extension. ` +
        `Supported architectures are: ${Object.keys(LAYER_BY_ARCHITECTURE).join(", ")}.`,
    );
  }

  // __dirname is <package>/lib/private once compiled, and <repo>/src/private
  // when the tests run from source; both are two levels below the root.
  const asset = path.join(
    __dirname,
    "..",
    "..",
    "assets",
    `layer-${suffix}.zip`,
  );
  if (!fs.existsSync(asset)) {
    throw new Error(
      `The log-to-s3 layer asset is missing at ${asset}. ` +
        "If you are working in a clone of the repository, run: npx projen build:layers",
    );
  }
  return asset;
}

const LAYER_BY_ARCHITECTURE: { [name: string]: string } = {
  [Architecture.ARM_64.name]: "arm64",
  [Architecture.X86_64.name]: "x86_64",
};
