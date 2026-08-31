import * as fs from "fs";
import * as path from "path";

/**
 * Guards the failure mode nothing else can see: a package published with a
 * missing or corrupt layer zip. Every other test synthesizes fine against a
 * broken asset, and the problem only surfaces at Lambda INIT.
 */
describe("layer assets", () => {
  const assetsDir = path.join(__dirname, "..", "assets");

  test.each(["arm64", "x86_64"])(
    "layer-%s.zip is present and plausible",
    (arch) => {
      const file = path.join(assetsDir, `layer-${arch}.zip`);

      expect(fs.existsSync(file)).toBe(true);

      const data = fs.readFileSync(file);
      expect(data.subarray(0, 2).toString("latin1")).toBe("PK");
      // A compiled Go binary compresses to several megabytes; anything much
      // smaller means the build produced a stub.
      expect(data.length).toBeGreaterThan(1024 * 1024);

      const sidecar = `${file}.sha256`;
      expect(fs.existsSync(sidecar)).toBe(true);
      expect(fs.readFileSync(sidecar, "utf8").trim()).toMatch(/^[0-9a-f]{64}$/);
    },
  );

  test("BUILDINFO records the toolchain and source fingerprint", () => {
    const info = JSON.parse(
      fs.readFileSync(path.join(assetsDir, "BUILDINFO.json"), "utf8"),
    );

    // Bare toolchain version, with no host OS/arch suffix - otherwise an
    // identical set of cross-compiled binaries would look stale to CI when
    // rebuilt on another platform.
    expect(info.goVersion).toMatch(/^go\d+\.\d+/);
    expect(info.goVersion).not.toMatch(/windows|linux|darwin/);
    expect(info.targets).toEqual(["arm64", "x86_64"]);
    expect(info.extensionSourceSha).toMatch(/^[0-9a-f]{64}$/);
    // No build timestamp: it would change on every run and make the CI
    // staleness check impossible to satisfy.
    expect(info.builtAt).toBeUndefined();
  });
});
