import { awscdk, javascript, JsonFile } from "projen";
import { JobPermission } from "projen/lib/github/workflows-model";

const project = new awscdk.AwsCdkConstructLibrary({
  name: "@jaggr2/cdk-log-to-s3",
  description:
    "CDK constructs for a Lambda Telemetry API extension that writes structured logs to S3 as Parquet, queryable with Athena via partition projection.",
  author: "Roger Jaggi",
  authorAddress: "roger.jaggi@gmail.com",
  repositoryUrl: "https://github.com/jaggr2/cdk-log-to-s3.git",
  license: "MIT",
  stability: "experimental",
  keywords: [
    "aws",
    "cdk",
    "lambda",
    "lambda-extension",
    "telemetry",
    "logs",
    "s3",
    "parquet",
    "athena",
    "glue",
  ],

  defaultReleaseBranch: "main",
  majorVersion: 0,
  projenrcTs: true,

  cdkVersion: "2.150.0",
  constructsVersion: "10.3.0",
  jsiiVersion: "~6.0.0",
  minNodeVersion: "20.0.0",
  workflowNodeVersion: "20.x",
  packageManager: javascript.NodePackageManager.NPM,

  releaseToNpm: true,
  npmAccess: javascript.NpmAccess.PUBLIC,
  // Tokenless publishing via GitHub OIDC. Must be configured on npmjs.com
  // first, and npm only allows that once the package already exists - hence
  // the one manual publish that bootstraps each package.
  npmTrustedPublishing: true,
  docgen: true,
  sampleCode: false,
  prettier: true,

  devDeps: [
    "aws-cdk@^2",
    "@aws-sdk/client-s3",
    "@aws-sdk/client-athena",
    "@aws-sdk/client-lambda",
  ],

  // The logger package and the integ example are compiled by their own
  // toolchains, never by jsii.
  excludeTypescript: ["packages/**/*", "examples/**/*"],

  gitignore: [
    "/extension-go/dist/",
    "cdk.out/",
    "/packages/*/lib/",
    "/packages/*/node_modules/",
  ],
});

// -- Ship the prebuilt layer zips inside the npm tarball ---------------------
// Enforced by the `verify layer assets are packed` post-build step below;
// a tarball without these is a package that cannot synth.
project.npmignore?.exclude(
  "/packages",
  "/extension-go",
  "/examples",
  "/scripts",
  "/docs",
  "/.go-version",
);
project.npmignore?.include("/assets/", "/assets/**");
project.gitattributes.addAttributes(
  "/assets/*.zip",
  "binary",
  "-diff",
  "-merge",
);

// -- Build helpers ----------------------------------------------------------
// The generated root tsconfig only covers src/, so the build scripts and the
// integ example each need one of their own.
for (const dir of ["scripts", "examples"]) {
  new JsonFile(project, `${dir}/tsconfig.json`, {
    obj: {
      extends: "../tsconfig.json",
      compilerOptions: {
        noEmit: true,
        rootDir: "..",
        esModuleInterop: true,
        resolveJsonModule: true,
      },
      include: ["**/*.ts"],
      exclude: ["node_modules"],
    },
  });
}

// -- Tasks ------------------------------------------------------------------
project.addTask("build:layers", {
  description:
    "Cross-compile the Go extension for arm64 + x86_64 and package the layer zips (requires Go)",
  exec: "ts-node --project scripts/tsconfig.json scripts/build-layers.ts",
});

const checkLayers = project.addTask("check:layers", {
  description:
    "Verify the committed layer zips exist and match their sha256 sidecars (no Go required)",
  exec: "ts-node --project scripts/tsconfig.json scripts/check-layers.ts",
});
project.testTask.prependSpawn(checkLayers);

project.addTask("test:go", {
  description: "Run the Go extension test suite",
  cwd: "extension-go",
  // -race is left to CI: it requires cgo, and a Windows dev box normally has
  // no C toolchain installed.
  exec: "go vet ./... && go test ./...",
});

// Integration helpers - need real AWS credentials, never run in the default CI.
project.addTask("integ:deploy", {
  exec: 'cdk --app "ts-node --project examples/tsconfig.json examples/integ/app.ts" deploy --require-approval never',
});
project.addTask("integ:verify", {
  exec: "ts-node --project scripts/tsconfig.json scripts/verify-integ.ts",
});
project.addTask("integ:destroy", {
  exec: 'cdk --app "ts-node --project examples/tsconfig.json examples/integ/app.ts" destroy --force',
});

// -- Logger companion package (plain npm, not a projen subproject) ----------
project.addTask("logger:build", {
  cwd: "packages/logger",
  exec: "npm run build",
});
project.addTask("logger:test", { cwd: "packages/logger", exec: "npm test" });

// -- CI: rebuild the layers whenever the Go source changes ------------------
const gh = project.github!;
const layersWf = gh.addWorkflow("layers");
layersWf.on({
  pullRequest: {
    paths: [
      "extension-go/**",
      "scripts/build-layers.ts",
      "assets/**",
      ".go-version",
    ],
  },
  push: {
    branches: ["main"],
    paths: ["extension-go/**", "scripts/build-layers.ts", ".go-version"],
  },
  workflowDispatch: {},
});
layersWf.addJob("build-layers", {
  runsOn: ["ubuntu-latest"],
  permissions: { contents: JobPermission.READ },
  steps: [
    { uses: "actions/checkout@v4" },
    {
      uses: "actions/setup-go@v5",
      with: {
        // .go-version, not go.mod: setup-go reads go.mod's `go` directive
        // (the language floor) and ignores the `toolchain` line, so it would
        // install a different compiler than the one the build actually uses.
        "go-version-file": ".go-version",
        cache: true,
        "cache-dependency-path": "extension-go/go.sum",
      },
    },
    { uses: "actions/setup-node@v4", with: { "node-version": "20.x" } },
    {
      name: "go vet + test",
      run: "cd extension-go && go vet ./... && go test -race ./...",
    },
    { name: "install", run: "npm ci" },
    { name: "rebuild layers", run: "npx projen build:layers" },
    {
      name: "fail if assets are stale",
      run: 'git diff --exit-code --stat -- assets/ || (echo "::error::assets/ is stale - run: npx projen build:layers" && exit 1)',
    },
  ],
});

// -- CI: publish the logger package on a logger-v* tag ---------------------
const loggerWf = gh.addWorkflow("release-logger");
loggerWf.on({ push: { tags: ["logger-v*"] }, workflowDispatch: {} });
loggerWf.addJob("publish", {
  runsOn: ["ubuntu-latest"],
  permissions: {
    contents: JobPermission.READ,
    // Required for both npm trusted publishing and provenance attestation.
    idToken: JobPermission.WRITE,
  },
  steps: [
    { uses: "actions/checkout@v4" },
    {
      uses: "actions/setup-node@v4",
      with: {
        "node-version": "20.x",
        "registry-url": "https://registry.npmjs.org",
      },
    },
    {
      // Node 20 ships npm 10; trusted publishing needs npm >= 11.5.1.
      name: "upgrade npm for trusted publishing",
      run: "npm install -g npm@latest",
    },
    {
      // No NODE_AUTH_TOKEN. npm exchanges the workflow's OIDC token for
      // publish rights against the trusted publisher configured on npmjs.com.
      name: "build, test and publish",
      run: "cd packages/logger && npm install && npm test && npm run build && npm publish --access public --provenance",
    },
  ],
});

// -- Guard: the published tarball must actually contain the layer zips ------
project.buildWorkflow?.addPostBuildSteps({
  name: "verify layer assets are packed",
  run: "node scripts/verify-pack.js",
});

project.synth();
