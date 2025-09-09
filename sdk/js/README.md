[![GitHub Workflow Status: CI [SDK/JS]](https://img.shields.io/github/actions/workflow/status/erodez/stremthru/ci-sdk-js.yml?branch=main&label=CI%20%5BSDK%2FJS%5D&style=for-the-badge)](https://github.com/erodez/stremthru/actions/workflows/ci-sdk-js.yml)
[![NPM Version](https://img.shields.io/npm/v/stremthru?style=for-the-badge)](https://npmjs.com/package/stremthru)
[![License](https://img.shields.io/github/license/erodez/stremthru?style=for-the-badge)](https://github.com/erodez/stremthru/blob/main/sdk/js/LICENSE)

# StremThru - JavaScript SDK

## Installation

```sh
# using pnpm:
pnpm add stremthru

# using npm:
npm install --save stremthru

# using yarn:
yarn add stremthru
```

## Usage

**Basic Usage:**

```ts
import { StremThru } from "stremthru";

const st = new StremThru({
  baseUrl: "http://127.0.0.1:8080",
  auth: "user:pass",
});
```

## License

Licensed under the MIT License. Check the [LICENSE](./LICENSE) file for details.
