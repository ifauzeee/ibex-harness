import fs from "node:fs";
import path from "node:path";

type ReleaseType = "major" | "minor" | "patch";

export type ChangelogSection = Readonly<{
  title: string;
  items: string[];
}>;

export type ChangelogRelease = Readonly<{
  version: string;
  date: string | null;
  type: ReleaseType;
  summary: string | null;
  sections: ChangelogSection[];
}>;

const RELEASE_HEADER = /^## \[(\d+\.\d+\.\d+)\](?:\s+[-—]\s+(.+))?$/;
const SECTION_HEADER = /^###\s+(.+)$/;
const UNRELEASED_HEADER = "## [Unreleased]";

type MutableSection = { title: string; items: string[] };
type MutableRelease = {
  version: string;
  date: string | null;
  type: ReleaseType;
  summary: string | null;
  sections: MutableSection[];
};

function parseReleaseType(version: string): ReleaseType {
  const [major, minor] = version.split(".").map((part) => Number(part) || 0);
  if (major > 0) return "major";
  if (minor > 0) return "minor";
  return "patch";
}

function normalizeDate(raw: string | null): string | null {
  if (!raw) return null;
  const trimmed = raw.trim();
  if (!trimmed || trimmed === "YYYY-MM-DD") return null;
  return trimmed;
}

function isPlaceholderItem(item: string): boolean {
  const normalized = item.toLowerCase();
  return normalized === "_tbd_" || normalized === "(example)" || normalized.startsWith("(example) ");
}

function finalizeCurrentRelease(
  releases: ChangelogRelease[],
  currentRelease: MutableRelease | null,
): void {
  if (currentRelease) releases.push(currentRelease);
}

function createRelease(match: RegExpExecArray): MutableRelease {
  return {
    version: match[1],
    date: normalizeDate(match[2] ?? null),
    type: parseReleaseType(match[1]),
    summary: null,
    sections: [],
  };
}

function parseSection(line: string): MutableSection | null {
  const sectionMatch = SECTION_HEADER.exec(line);
  if (!sectionMatch) return null;
  return { title: sectionMatch[1], items: [] };
}

function appendListItem(line: string, currentSection: MutableSection | null): boolean {
  if (!line.startsWith("- ") || !currentSection) return false;
  const item = line.slice(2).trim();
  if (item && !isPlaceholderItem(item)) {
    currentSection.items.push(item);
  }
  return true;
}

function shouldSkipTextLine(line: string, hasSummary: boolean): boolean {
  if (!line) return true;
  if (line.startsWith("---")) return true;
  if (hasSummary) return true;
  return false;
}

function processReleaseLine(
  line: string,
  currentRelease: MutableRelease,
  currentSection: MutableSection | null,
): MutableSection | null {
  const parsedSection = parseSection(line);
  if (parsedSection) {
    currentRelease.sections.push(parsedSection);
    return parsedSection;
  }

  if (appendListItem(line, currentSection)) return currentSection;
  if (shouldSkipTextLine(line, Boolean(currentRelease.summary))) return currentSection;
  if (!currentSection && !currentRelease.summary) {
    currentRelease.summary = line;
  }
  return currentSection;
}

export function readChangelogReleases(): ChangelogRelease[] {
  const changelogPath = path.resolve(process.cwd(), "../CHANGELOG.md");
  const content = fs.readFileSync(changelogPath, "utf8");
  const lines = content.split(/\r?\n/);

  const releases: ChangelogRelease[] = [];
  let currentRelease: MutableRelease | null = null;
  let currentSection: MutableSection | null = null;

  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (line.startsWith(UNRELEASED_HEADER)) {
      finalizeCurrentRelease(releases, currentRelease);
      currentRelease = null;
      currentSection = null;
      continue;
    }

    const releaseMatch = RELEASE_HEADER.exec(line);
    if (releaseMatch) {
      finalizeCurrentRelease(releases, currentRelease);
      currentRelease = createRelease(releaseMatch);
      currentSection = null;
      continue;
    }

    if (!currentRelease) continue;

    currentSection = processReleaseLine(line, currentRelease, currentSection);
  }

  finalizeCurrentRelease(releases, currentRelease);

  return releases;
}
