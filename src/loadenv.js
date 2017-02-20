"use strict";

const env = require('../.env.json');
Object.keys(env).map(k => process.env[k] = env[k]);
