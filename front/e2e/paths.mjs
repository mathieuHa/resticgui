import os from 'node:os';
import path from 'node:path';

/** The folder the server of a port runs in; the tests look at files there. */
export const e2eRoot = (/** @type {number | string} */ port) => path.join(os.tmpdir(), `resticgui-e2e-${port}`);
