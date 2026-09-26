#!/usr/bin/env node

/**
 * generate-changelog.mjs
 * 
 * Standalone Conventional Commits Changelog Generator for Discord Bot Subscription Platform.
 * Requires 0 external npm dependencies (pure Node.js + Git CLI).
 *
 * Usage:
 *   node scripts/generate-changelog.mjs
 *   node scripts/generate-changelog.mjs --version v1.0.0
 *   node scripts/generate-changelog.mjs --check
 */

import { execSync } from 'node:child_process';
import { readFileSync, writeFileSync, existsSync } from 'node:fs';
import { resolve } from 'node:path';

const REPO_ROOT = resolve(process.cwd());
const CHANGELOG_PATH = resolve(REPO_ROOT, 'CHANGELOG.md');

// Determine GitHub repository URL from git remote
function getRepoUrl() {
  try {
    const remote = execSync('git remote get-url origin', { encoding: 'utf8' }).trim();
    if (remote.startsWith('git@github.com:')) {
      const path = remote.replace('git@github.com:', '').replace(/\.git$/, '');
      return `https://github.com/${path}`;
    }
    if (remote.startsWith('https://github.com/')) {
      return remote.replace(/\.git$/, '');
    }
  } catch {
    // fallback if no remote is configured
  }
  return 'https://github.com/Abdo30004/discord-subscriptions';
}

const REPO_URL = getRepoUrl();

// Parse CLI flags
const args = process.argv.slice(2);
const checkOnly = args.includes('--check');
const versionArgIndex = args.indexOf('--version');
const explicitVersion = versionArgIndex !== -1 && args[versionArgIndex + 1] ? args[versionArgIndex + 1] : null;

// Get Git log
function getCommits() {
  try {
    const output = execSync('git log --pretty=format:"%H|%h|%an|%ad|%s" --date=short', { encoding: 'utf8' });
    if (!output.trim()) return [];

    return output.split('\n').filter(Boolean).map(line => {
      const [fullHash, shortHash, author, date, subject] = line.split('|');
      return { fullHash, shortHash, author, date, subject };
    });
  } catch (err) {
    console.error('Failed to read git log:', err.message);
    process.exit(1);
  }
}

// Category mapping for Conventional Commits
const CATEGORIES = {
  feat: { title: '🚀 Features & Capabilities', items: [] },
  fix: { title: '🐛 Bug Fixes & Resilience', items: [] },
  gateway: { title: '🌐 Gateway, Traefik & Networking', items: [] },
  k8s: { title: '☁️ Kubernetes & Cloud Orchestration', items: [] },
  infra: { title: '🏗️ Infrastructure & Persistence', items: [] },
  perf: { title: '⚡ Performance Optimizations', items: [] },
  refactor: { title: '♻️ Refactoring & Clean Architecture', items: [] },
  docs: { title: '📚 Documentation & Architecture Guides', items: [] },
  test: { title: '🧪 Testing & Validation', items: [] },
  build: { title: '📦 Build System & Dependencies', items: [] },
  ci: { title: '🤖 CI/CD Automation', items: [] },
  chore: { title: '🔧 Maintenance & Scaffolding', items: [] },
  other: { title: '📋 Other Changes', items: [] },
};

function parseCommit(commit) {
  const match = commit.subject.match(/^([a-z]+)(?:\(([^)]+)\))?!?: (.+)$/i);
  if (!match) {
    return {
      type: 'other',
      scope: null,
      description: commit.subject,
      commit,
    };
  }

  const [, rawType, scope, description] = match;
  let type = rawType.toLowerCase();

  // Route specific scopes to top-level categories if appropriate
  if (scope === 'gateway' || scope === 'traefik') {
    type = 'gateway';
  } else if (scope === 'k8s' || scope === 'kubernetes') {
    type = 'k8s';
  } else if (scope === 'infra' || scope === 'database') {
    type = 'infra';
  }

  return {
    type: CATEGORIES[type] ? type : 'other',
    scope: scope || null,
    description: description.trim(),
    commit,
  };
}

function generateMarkdown(commits, releaseVersion, releaseDate) {
  // Reset items
  for (const key of Object.keys(CATEGORIES)) {
    CATEGORIES[key].items = [];
  }

  for (const c of commits) {
    const parsed = parseCommit(c);
    CATEGORIES[parsed.type].items.push(parsed);
  }

  let md = `# Changelog\n\n`;
  md += `All notable changes to the **Discord Bot Subscription & Turnkey Fleet Management Platform** will be documented in this file.\n\n`;
  md += `The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),\n`;
  md += `and this project strictly adheres to [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).\n\n`;
  md += `---\n\n`;

  md += `## [${releaseVersion}] - ${releaseDate}\n\n`;

  let hasSections = false;
  for (const [, section] of Object.entries(CATEGORIES)) {
    if (section.items.length === 0) continue;
    hasSections = true;

    md += `### ${section.title}\n\n`;
    for (const item of section.items) {
      const scopePrefix = item.scope ? `**${item.scope}**: ` : '';
      const commitLink = `([${item.commit.shortHash}](${REPO_URL}/commit/${item.commit.fullHash}))`;
      md += `- ${scopePrefix}${item.description} ${commitLink}\n`;
    }
    md += `\n`;
  }

  if (!hasSections) {
    md += `*No notable changes recorded.*\n\n`;
  }

  md += `---\n\n`;
  md += `*Generated automatically with \`scripts/generate-changelog.mjs\` based on Conventional Commits.*\n`;

  return md;
}

function run() {
  const commits = getCommits();
  if (commits.length === 0) {
    console.log('No commits found in repository.');
    return;
  }

  const today = new Date().toISOString().split('T')[0];
  const version = explicitVersion || 'v1.0.0';

  const newChangelog = generateMarkdown(commits, version, today);

  if (checkOnly) {
    if (!existsSync(CHANGELOG_PATH)) {
      console.error('CHANGELOG.md does not exist. Run "npm run changelog" or "make changelog" to generate it.');
      process.exit(1);
    }
    const current = readFileSync(CHANGELOG_PATH, 'utf8');
    if (current.trim() !== newChangelog.trim()) {
      console.error('CHANGELOG.md is out of date. Run "npm run changelog" or "make changelog".');
      process.exit(1);
    }
    console.log('✓ CHANGELOG.md is up to date.');
    return;
  }

  writeFileSync(CHANGELOG_PATH, newChangelog, 'utf8');
  console.log(`✓ CHANGELOG.md generated successfully for ${version} (${commits.length} commits parsed).`);
}

run();
