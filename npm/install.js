#!/usr/bin/env node

const { execSync } = require("child_process");
const fs = require("fs");
const https = require("https");
const path = require("path");

const pkg = require("./package.json");
const version = pkg.version;
const repo = "AlexGladkov/harnest";

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

function getBinaryName() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    console.error(
      `Unsupported platform: ${process.platform}-${process.arch}`
    );
    console.error("Install manually: https://github.com/" + repo + "/releases");
    process.exit(1);
  }

  return `harnest-${platform}-${arch}`;
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const follow = (url) => {
      https
        .get(url, { headers: { "User-Agent": "harnest-npm" } }, (res) => {
          if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
            follow(res.headers.location);
            return;
          }
          if (res.statusCode !== 200) {
            reject(new Error(`HTTP ${res.statusCode} for ${url}`));
            return;
          }
          const file = fs.createWriteStream(dest);
          res.pipe(file);
          file.on("finish", () => {
            file.close();
            resolve();
          });
        })
        .on("error", reject);
    };
    follow(url);
  });
}

async function main() {
  const binDir = path.join(__dirname, "bin");
  const binPath = path.join(binDir, "harnest");

  // Skip if binary already exists (e.g. CI cache)
  if (fs.existsSync(binPath)) {
    return;
  }

  fs.mkdirSync(binDir, { recursive: true });

  const binaryName = getBinaryName();
  const url = `https://github.com/${repo}/releases/download/v${version}/${binaryName}`;

  console.log(`Downloading harnest v${version} (${process.platform}-${process.arch})...`);

  try {
    await download(url, binPath);
    fs.chmodSync(binPath, 0o755);
    console.log("harnest installed successfully.");
  } catch (err) {
    console.error(`Failed to download: ${err.message}`);
    console.error(`URL: ${url}`);
    console.error("Install manually: https://github.com/" + repo + "/releases");
    // Don't fail npm install — user can get binary manually
    process.exit(0);
  }
}

main();
