#!/usr/bin/env node

/**
 * validate-diagrams.mjs
 * 
 * Validates all Mermaid diagrams across all Markdown files in the repository
 * using the official Mermaid.js parser.
 */

import { readFileSync, readdirSync, statSync } from 'node:fs';
import { resolve, relative, join } from 'node:path';
import { JSDOM } from 'jsdom';

// Setup browser DOM environment for Mermaid & DOMPurify in Node.js
const dom = new JSDOM('<!DOCTYPE html><html><body></body></html>');
global.window = dom.window;
global.document = dom.window.document;
global.Element = dom.window.Element;
global.HTMLElement = dom.window.HTMLElement;
global.SVGElement = dom.window.SVGElement;

// Import mermaid after globals are set
const { default: mermaid } = await import('mermaid');

const REPO_ROOT = resolve(process.cwd());

// Initialize mermaid
mermaid.initialize({
  startOnLoad: false,
  suppressErrorRendering: true,
  securityLevel: 'loose',
});

function getMarkdownFiles(dir) {
  let files = [];
  const entries = readdirSync(dir);

  for (const entry of entries) {
    if (entry === 'node_modules' || entry === '.git' || entry === '.next' || entry === 'dist') {
      continue;
    }

    const fullPath = join(dir, entry);
    const stat = statSync(fullPath);

    if (stat.isDirectory()) {
      files = files.concat(getMarkdownFiles(fullPath));
    } else if (entry.endsWith('.md')) {
      files.push(fullPath);
    }
  }

  return files;
}

function extractMermaidBlocks(filePath) {
  const content = readFileSync(filePath, 'utf8');
  const lines = content.split('\n');
  const blocks = [];

  let inBlock = false;
  let currentBlock = [];
  let startLine = 0;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];

    if (!inBlock && line.trim().startsWith('```mermaid')) {
      inBlock = true;
      startLine = i + 1;
      currentBlock = [];
    } else if (inBlock && line.trim().startsWith('```')) {
      inBlock = false;
      blocks.push({
        code: currentBlock.join('\n').trim(),
        line: startLine,
      });
    } else if (inBlock) {
      currentBlock.push(line);
    }
  }

  return blocks;
}

async function validate() {
  const mdFiles = getMarkdownFiles(REPO_ROOT);
  let totalDiagrams = 0;
  let failedDiagrams = 0;
  const errors = [];

  console.log(`Scanning ${mdFiles.length} markdown files for Mermaid diagrams...\n`);

  for (const file of mdFiles) {
    const relPath = relative(REPO_ROOT, file);
    const blocks = extractMermaidBlocks(file);

    for (const block of blocks) {
      totalDiagrams++;
      try {
        await mermaid.parse(block.code);
        console.log(`  ✓ ${relPath}:${block.line}`);
      } catch (err) {
        failedDiagrams++;
        const errorMsg = err.message || String(err);
        console.error(`  ✗ ${relPath}:${block.line} - FAILED`);
        console.error(`    Error: ${errorMsg.split('\n')[0]}`);
        errors.push({
          file: relPath,
          line: block.line,
          code: block.code,
          error: errorMsg,
        });
      }
    }
  }

  console.log(`\n========================================`);
  console.log(`Diagram Validation Summary:`);
  console.log(`Total diagrams: ${totalDiagrams}`);
  console.log(`Passed:         ${totalDiagrams - failedDiagrams}`);
  console.log(`Failed:         ${failedDiagrams}`);
  console.log(`========================================\n`);

  if (failedDiagrams > 0) {
    console.error(`Found ${failedDiagrams} broken diagrams:`);
    for (const e of errors) {
      console.error(`\n--- ${e.file}:${e.line} ---`);
      console.error(e.error);
    }
    process.exit(1);
  } else {
    console.log(`All ${totalDiagrams} Mermaid diagrams parsed successfully without errors!`);
    process.exit(0);
  }
}

validate().catch(err => {
  console.error('Fatal error running diagram validator:', err);
  process.exit(1);
});
