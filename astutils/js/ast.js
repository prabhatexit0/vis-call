const fs = require('fs');
const parser = require('@babel/parser');
const { parse } = parser;

const filePath = process.argv[2];
const code = fs.readFileSync(filePath, 'utf-8');
const ast = JSON.stringify(parse(code), null, 2);

console.log(ast);

